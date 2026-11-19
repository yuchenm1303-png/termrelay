package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	ID             string                      `json:"id"`
	UserID         int64                       `json:"user_id"`
	APIKeyID       int64                       `json:"api_key_id"`
	AccountID      int64                       `json:"account_id"`
	Kind           CangyuanMediaKind           `json:"kind"`
	Model          string                      `json:"model"`
	UpstreamTaskID string                      `json:"upstream_task_id,omitempty"`
	Status         CangyuanMediaJobState       `json:"status"`
	ResultURLs     []string                    `json:"result_urls,omitempty"`
	BillingState   string                      `json:"billing_state"`
	PriceSnapshot  *CangyuanMediaPriceSnapshot `json:"price_snapshot,omitempty"`
	// The original priced payload is pinned without persisting its prompt.
	// Every paid submission must match this digest before any upstream POST.
	PayloadSHA256 string `json:"payload_sha256,omitempty"`
	Version       int64  `json:"version"`
	CreatedAt     int64  `json:"created_at"`
	UpdatedAt     int64  `json:"updated_at"`
	ExpiresAt     int64  `json:"expires_at"`
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

func (s *CangyuanMediaJobService) Reserve(ctx context.Context, owner CangyuanMediaJobOwner, accountID int64, kind CangyuanMediaKind, model string, quotes ...CangyuanMediaPriceSnapshot) (*CangyuanMediaJob, error) {
	job, ttl, err := s.newReservation(owner, accountID, kind, model, quotes...)
	if err != nil {
		return nil, err
	}
	if err := s.Store.Create(ctx, job, ttl); err != nil {
		return nil, fmt.Errorf("reserve cangyuan media job: %w", err)
	}
	return job, nil
}

// newReservation builds a complete private job without persisting it. This is
// reused by the atomic client-idempotency reservation path below, avoiding a
// two-phase create that could accidentally produce two local task records.
func (s *CangyuanMediaJobService) newReservation(owner CangyuanMediaJobOwner, accountID int64, kind CangyuanMediaKind, model string, quotes ...CangyuanMediaPriceSnapshot) (*CangyuanMediaJob, time.Duration, error) {
	if s == nil || s.Store == nil {
		return nil, 0, errors.New("cangyuan media job storage unavailable")
	}
	if err := validateCangyuanMediaJobOwner(owner); err != nil {
		return nil, 0, err
	}
	if accountID <= 0 || strings.TrimSpace(model) == "" || len(model) > 256 {
		return nil, 0, errors.New("cangyuan media job requires an account and a model")
	}
	if _, err := kind.path(); err != nil {
		return nil, 0, err
	}
	if len(quotes) > 1 {
		return nil, 0, errors.New("cangyuan media reservation accepts at most one price snapshot")
	}
	var snapshot *CangyuanMediaPriceSnapshot
	billingState := "disabled"
	if len(quotes) == 1 {
		quote := quotes[0]
		if err := validateCangyuanMediaPriceSnapshot(kind, &quote); err != nil {
			return nil, 0, err
		}
		snapshot = &quote
		billingState = "pending"
	}
	ttl := s.TTL
	if ttl <= 0 {
		ttl = defaultCangyuanMediaJobTTL
	}
	now := time.Now().UTC()
	job := &CangyuanMediaJob{
		ID:            "cyjob_" + strings.ReplaceAll(uuid.NewString(), "-", ""),
		UserID:        owner.UserID,
		APIKeyID:      owner.APIKeyID,
		AccountID:     accountID,
		Kind:          kind,
		Model:         strings.TrimSpace(model),
		Status:        CangyuanMediaJobReserved,
		BillingState:  billingState,
		PriceSnapshot: snapshot,
		Version:       1,
		CreatedAt:     now.Unix(),
		UpdatedAt:     now.Unix(),
		ExpiresAt:     now.Add(ttl).Unix(),
	}
	return job, ttl, nil
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
	case CangyuanMediaJobQueued:
		if task.Status == "queued" {
			return job, nil
		}
		return s.persistTransition(ctx, job, CangyuanMediaJobState(task.Status), job.UpstreamTaskID, task.ResultURLs(job.Kind))
	case CangyuanMediaJobProcessing:
		if task.Status == "queued" {
			return nil, ErrCangyuanMediaJobState
		}
		if task.Status == "in_progress" {
			return job, nil
		}
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

// SubmitReserved is the isolated job coordinator. It does not expose an API
// route, charge customers, or create a channel: those must be added only after
// pricing and user-entitlement checks. The local reservation is CAS-claimed
// BEFORE the upstream POST, so concurrent workers cannot double-submit.
//
// A non-nil upstream task is returned even when persisting its ID failed:
// operators can reconcile the ambiguous remote job without re-posting.
func (s *CangyuanMediaJobService) SubmitReserved(ctx context.Context, owner CangyuanMediaJobOwner, id string, client *CangyuanMediaClient, payload []byte) (*CangyuanMediaJob, *CangyuanMediaTask, error) {
	job, err := s.Get(ctx, owner, id)
	if err != nil {
		return nil, nil, err
	}
	if job.Status != CangyuanMediaJobReserved || client == nil {
		return nil, nil, ErrCangyuanMediaJobState
	}
	var input struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
	}
	if json.Unmarshal(payload, &input) != nil || input.Model != job.Model || strings.TrimSpace(input.Prompt) == "" {
		return nil, nil, errors.New("cangyuan media job payload differs from reserved model or lacks prompt")
	}
	var refs map[string]json.RawMessage
	if json.Unmarshal(payload, &refs) != nil || refs == nil {
		return nil, nil, errors.New("cangyuan media job requires a JSON object")
	}
	if err := validateCangyuanMediaReferences(refs); err != nil {
		return nil, nil, err
	}
	// Check again immediately before the single upstream POST. A request
	// cannot change quantity after the job's approved price was frozen.
	if job.PriceSnapshot != nil {
		if err := validateCangyuanMediaRequestQuantity(job.Kind, refs, job.PriceSnapshot); err != nil {
			return nil, nil, err
		}
		// The price rule was matched against the *whole* original request.
		// Checking only the model and quantity here would let a worker swap
		// expensive dimensions or reference material after price approval.
		if job.PayloadSHA256 == "" || job.PayloadSHA256 != cangyuanMediaPayloadDigest(job.ID, payload) {
			return nil, nil, errors.New("cangyuan media payload differs from the reserved priced request")
		}
	}
	// Paid jobs must freeze their full verified sale price in SQL before
	// their only upstream POST. A quote alone is not proof of a balance hold.
	if job.PriceSnapshot != nil && job.BillingState != "held" {
		return nil, nil, ErrCangyuanMediaNotBillable
	}
	if _, err := s.ClaimSubmission(ctx, owner, id); err != nil {
		return nil, nil, err
	}
	task, submitErr := client.Submit(ctx, job.Kind, payload)
	// This update must survive client cancellation. On ambiguous upstream
	// errors, a process must not leave the record looking ready to retry.
	storeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if submitErr != nil {
		_, storeErr := s.MarkSubmissionUnknown(storeCtx, owner, id)
		if storeErr != nil {
			return nil, nil, errors.Join(submitErr, fmt.Errorf("record ambiguous cangyuan media submission: %w", storeErr))
		}
		return nil, nil, submitErr
	}
	attached, err := s.AttachSubmitted(storeCtx, owner, id, task)
	if err != nil {
		return nil, task, fmt.Errorf("upstream accepted job but task ID could not be persisted: %w", err)
	}
	return attached, task, nil
}

// PollJob reconciles an existing upstream GET with the private Redis state.
// No request is sent for locally terminal or ambiguous-submission states.
func (s *CangyuanMediaJobService) PollJob(ctx context.Context, owner CangyuanMediaJobOwner, id string, client *CangyuanMediaClient) (*CangyuanMediaJob, error) {
	job, err := s.Get(ctx, owner, id)
	if err != nil {
		return nil, err
	}
	if job.Status == CangyuanMediaJobCompleted || job.Status == CangyuanMediaJobFailed {
		return job, nil
	}
	if (job.Status != CangyuanMediaJobQueued && job.Status != CangyuanMediaJobProcessing) || client == nil {
		return nil, ErrCangyuanMediaJobState
	}
	task, err := client.Poll(ctx, job.Kind, job.UpstreamTaskID)
	if err != nil {
		return nil, err
	}
	return s.Observe(ctx, owner, id, task)
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

// cangyuanMediaPayloadDigest ties the payload to the randomly generated local
// task ID, without storing the prompt or reference URLs in Redis.
func cangyuanMediaPayloadDigest(jobID string, payload []byte) string {
	sum := sha256.New()
	_, _ = sum.Write([]byte(jobID))
	_, _ = sum.Write([]byte{0})
	_, _ = sum.Write(payload)
	return hex.EncodeToString(sum.Sum(nil))
}

func validateCangyuanMediaJobOwner(owner CangyuanMediaJobOwner) error {
	if owner.UserID <= 0 || owner.APIKeyID <= 0 {
		return errors.New("cangyuan media job requires authenticated owner")
	}
	return nil
}
