package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const defaultCangyuanMediaJobTTL = 48 * time.Hour

var (
	ErrCangyuanMediaJobNotFound = errors.New("cangyuan media job not found")
	ErrCangyuanMediaJobConflict = errors.New("cangyuan media job changed concurrently")
	ErrCangyuanMediaJobState    = errors.New("invalid cangyuan media job state transition")
)

// CangyuanMediaJobOwner is the tenant-scoped identity checked on every read
// or mutation. The upstream task ID alone never authorizes access to a job.
type CangyuanMediaJobOwner struct {
	UserID   int64
	APIKeyID int64
}

type CangyuanMediaJobState string

const (
	CangyuanMediaJobReserved      CangyuanMediaJobState = "reserved"
	CangyuanMediaJobSubmitting    CangyuanMediaJobState = "submitting"
	CangyuanMediaJobSubmitUnknown CangyuanMediaJobState = "submit_unknown"
	CangyuanMediaJobQueued        CangyuanMediaJobState = "queued"
	CangyuanMediaJobProcessing    CangyuanMediaJobState = "in_progress"
	CangyuanMediaJobCompleted     CangyuanMediaJobState = "completed"
	CangyuanMediaJobFailed        CangyuanMediaJobState = "failed"
)

// CangyuanMediaJob is the private persisted record. API credentials, prompts,
// source media and arbitrary HTTP responses are deliberately not stored here.
// BillingState remains "disabled" until a separate atomic ledger exists.
type CangyuanMediaJob struct {
	ID             string                `json:"id"`
	UserID         int64                 `json:"user_id"`
	APIKeyID       int64                 `json:"api_key_id"`
	AccountID      int64                 `json:"account_id"`
	Kind           CangyuanMediaKind     `json:"kind"`
	Model          string                `json:"model"`
	UpstreamTaskID string                `json:"upstream_task_id,omitempty"`
	Status         CangyuanMediaJobState `json:"status"`
	ResultURLs     []string              `json:"result_urls,omitempty"`
	BillingState   string                `json:"billing_state"`
	Version        int64                 `json:"version"`
	CreatedAt      int64                 `json:"created_at"`
	UpdatedAt      int64                 `json:"updated_at"`
	ExpiresAt      int64                 `json:"expires_at"`
}

// CangyuanMediaJobStore MUST atomically enforce uniqueness and version CAS,
// across goroutines and separate TermRelay instances. No plain Get+Set races.
type CangyuanMediaJobStore interface {
	Create(ctx context.Context, job *CangyuanMediaJob, ttl time.Duration) error
	Get(ctx context.Context, id string) (*CangyuanMediaJob, error)
	CompareAndSwap(ctx context.Context, job *CangyuanMediaJob, expectedVersion int64) error
}

type CangyuanMediaJobService struct {
	Store CangyuanMediaJobStore
	TTL   time.Duration
}

func (s *CangyuanMediaJobService) Reserve(ctx context.Context, owner CangyuanMediaJobOwner, accountID int64, kind CangyuanMediaKind, model string) (*CangyuanMediaJob, error) {
	if s == nil || s.Store == nil {
		return nil, errors.New("cangyuan media job storage unavailable")
	}
	if err := validateCangyuanMediaJobOwner(owner); err != nil {
		return nil, err
	}
	if accountID <= 0 || strings.TrimSpace(model) == "" || len(model) > 256 {
		return nil, errors.New("cangyuan media job requires an account and a model")
	}
	if _, err := kind.path(); err != nil {
		return nil, err
	}
	ttl := s.TTL
	if ttl <= 0 {
		ttl = defaultCangyuanMediaJobTTL
	}
	now := time.Now().UTC()
	job := &CangyuanMediaJob{
		ID:           "cyjob_" + strings.ReplaceAll(uuid.NewString(), "-", ""),
		UserID:       owner.UserID,
		APIKeyID:     owner.APIKeyID,
		AccountID:    accountID,
		Kind:         kind,
		Model:        strings.TrimSpace(model),
		Status:       CangyuanMediaJobReserved,
		BillingState: "disabled",
		Version:      1,
		CreatedAt:    now.Unix(),
		UpdatedAt:    now.Unix(),
		ExpiresAt:    now.Add(ttl).Unix(),
	}
	if err := s.Store.Create(ctx, job, ttl); err != nil {
		return nil, fmt.Errorf("reserve cangyuan media job: %w", err)
	}
	return job, nil
}

func (s *CangyuanMediaJobService) Get(ctx context.Context, owner CangyuanMediaJobOwner, id string) (*CangyuanMediaJob, error) {
	if s == nil || s.Store == nil {
		return nil, errors.New("cangyuan media job storage unavailable")
	}
	if err := validateCangyuanMediaJobOwner(owner); err != nil {
		return nil, err
	}
	if !strings.HasPrefix(id, "cyjob_") || len(id) != 38 {
		return nil, ErrCangyuanMediaJobNotFound
	}
	job, err := s.Store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if job == nil || job.UserID != owner.UserID || job.APIKeyID != owner.APIKeyID {
		return nil, ErrCangyuanMediaJobNotFound
	}
	return job, nil
}

// ClaimSubmission atomically reserves the right to send the one upstream POST.
// A concurrent worker cannot claim the same job, even across instances.
// A crash while submitting leaves it in "submitting" instead of retrying.
func (s *CangyuanMediaJobService) ClaimSubmission(ctx context.Context, owner CangyuanMediaJobOwner, id string) (*CangyuanMediaJob, error) {
	job, err := s.Get(ctx, owner, id)
	if err != nil {
		return nil, err
	}
	if job.Status != CangyuanMediaJobReserved {
		return nil, ErrCangyuanMediaJobState
	}
	return s.persistTransition(ctx, job, CangyuanMediaJobSubmitting, "", nil)
}

// MarkSubmissionUnknown marks a reservation after an ambiguous transport
// failure. Without an upstream task ID, the job MUST NEVER be auto-reposted.
func (s *CangyuanMediaJobService) MarkSubmissionUnknown(ctx context.Context, owner CangyuanMediaJobOwner, id string) (*CangyuanMediaJob, error) {
	job, err := s.Get(ctx, owner, id)
	if err != nil {
		return nil, err
	}
	if job.Status == CangyuanMediaJobSubmitUnknown {
		return job, nil
	}
	if job.Status != CangyuanMediaJobSubmitting {
		return nil, ErrCangyuanMediaJobState
	}
	return s.persistTransition(ctx, job, CangyuanMediaJobSubmitUnknown, "", nil)
}

// AttachSubmitted links the ID returned from the one-and-only upstream POST.
// A second submission or a different task ID must not overwrite that link.
func (s *CangyuanMediaJobService) AttachSubmitted(ctx context.Context, owner CangyuanMediaJobOwner, id string, task *CangyuanMediaTask) (*CangyuanMediaJob, error) {
	job, err := s.Get(ctx, owner, id)
	if err != nil {
		return nil, err
	}
	if job.Status != CangyuanMediaJobSubmitting || task == nil {
		return nil, ErrCangyuanMediaJobState
	}
	if !cangyuanTaskIDPattern.MatchString(task.IDValue()) || !validCangyuanMediaJobTask(job.Kind, task) {
		return nil, ErrCangyuanMediaJobState
	}
	return s.persistTransition(ctx, job, CangyuanMediaJobState(task.Status), task.IDValue(), task.ResultURLs(job.Kind))
}

// Observe persists a GET response. A completed/failed job is immutable and
// cannot regress to queued/processing after an out-of-order polling response.
func (s *CangyuanMediaJobService) Observe(ctx context.Context, owner CangyuanMediaJobOwner, id string, task *CangyuanMediaTask) (*CangyuanMediaJob, error) {
	job, err := s.Get(ctx, owner, id)
	if err != nil {
		return nil, err
	}
	if task == nil || job.UpstreamTaskID == "" || task.IDValue() != job.UpstreamTaskID || !validCangyuanMediaJobTask(job.Kind, task) {
		return nil, ErrCangyuanMediaJobState
	}
	switch job.Status {
	case CangyuanMediaJobCompleted, CangyuanMediaJobFailed:
		if string(job.Status) == task.Status {
			return job, nil
		}
		return nil, ErrCangyuanMediaJobState
	case CangyuanMediaJobQueued, CangyuanMediaJobProcessing:
		return s.persistTransition(ctx, job, CangyuanMediaJobState(task.Status), job.UpstreamTaskID, task.ResultURLs(job.Kind))
	default:
		return nil, ErrCangyuanMediaJobState
	}
}

func (s *CangyuanMediaJobService) persistTransition(ctx context.Context, original *CangyuanMediaJob, status CangyuanMediaJobState, taskID string, resultURLs []string) (*CangyuanMediaJob, error) {
	next := *original
	next.Status = status
	next.UpstreamTaskID = taskID
	next.ResultURLs = append([]string(nil), resultURLs...)
	next.Version++
	next.UpdatedAt = time.Now().UTC().Unix()
	if err := s.Store.CompareAndSwap(ctx, &next, original.Version); err != nil {
		return nil, err
	}
	return &next, nil
}

func validCangyuanMediaJobTask(kind CangyuanMediaKind, task *CangyuanMediaTask) bool {
	if task == nil {
		return false
	}
	switch task.Status {
	case "queued", "in_progress", "failed":
		return true
	case "completed":
		return len(task.ResultURLs(kind)) > 0
	default:
		return false
	}
}

func validateCangyuanMediaJobOwner(owner CangyuanMediaJobOwner) error {
	if owner.UserID <= 0 || owner.APIKeyID <= 0 {
		return errors.New("cangyuan media job requires authenticated owner")
	}
	return nil
}
