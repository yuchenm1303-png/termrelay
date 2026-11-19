package service

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CangyuanMediaRecoveryQueue is the atomic due-index and cross-instance lease
// used by the automatic media reconciliation worker. It contains only task IDs,
// not prompts, credentials or unredacted provider payloads.
type CangyuanMediaRecoveryQueue interface {
	Due(context.Context, int64) ([]string, error)
	Delay(context.Context, string, time.Duration) error
	Remove(context.Context, string) error
	Claim(context.Context, string, string, time.Duration) (bool, error)
	Release(context.Context, string, string) error
}

type CangyuanMediaRecoveryAccountReader interface {
	GetByID(context.Context, int64) (*Account, error)
}

type CangyuanMediaRecoveryWorker struct {
	Jobs     *CangyuanMediaJobService
	Accounts CangyuanMediaRecoveryAccountReader
	Billing  CangyuanMediaEscrowWriter
	Queue    CangyuanMediaRecoveryQueue

	// Optional injectable factory for deterministic, no-network tests.
	ClientFactory func(*Account) *CangyuanMediaClient

	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

func NewCangyuanMediaRecoveryWorker(
	jobs *CangyuanMediaJobService, accounts CangyuanMediaRecoveryAccountReader,
	billing CangyuanMediaEscrowWriter,
) *CangyuanMediaRecoveryWorker {
	w := &CangyuanMediaRecoveryWorker{Jobs: jobs, Accounts: accounts, Billing: billing}
	if jobs != nil {
		w.Queue, _ = jobs.Store.(CangyuanMediaRecoveryQueue)
	}
	return w
}

// Start is safe to call more than once. Only jobs with an approved, frozen
// SQL balance are indexed. It never creates or resubmits an upstream job.
func (w *CangyuanMediaRecoveryWorker) Start() {
	if w == nil || w.Queue == nil || w.Jobs == nil || w.Accounts == nil || w.Billing == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	w.done = make(chan struct{})
	go func() {
		defer close(w.done)
		ticker := time.NewTicker(6 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runCtx, stop := context.WithTimeout(ctx, 90*time.Second)
				if err := w.RunOnce(runCtx); err != nil && ctx.Err() == nil {
					log.Printf("[CangyuanMediaRecovery] tick failed: %v", err)
				}
				stop()
			}
		}
	}()
}

func (w *CangyuanMediaRecoveryWorker) Stop() {
	if w == nil {
		return
	}
	w.mu.Lock()
	cancel, done := w.cancel, w.done
	w.cancel = nil
	w.done = nil
	w.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
}

// RunOnce processes a small bounded batch. A Redis lease allows multiple
// TermRelay instances to operate simultaneously without duplicate polling.
// SQL hold/capture/release idempotency remains the money-safety backstop.
func (w *CangyuanMediaRecoveryWorker) RunOnce(ctx context.Context) error {
	if w == nil || w.Queue == nil || w.Jobs == nil || w.Accounts == nil || w.Billing == nil {
		return errors.New("cangyuan media recovery worker is not configured")
	}
	ids, err := w.Queue.Due(ctx, 8)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return err
		}
		token := uuid.NewString()
		claimed, err := w.Queue.Claim(ctx, id, token, 40*time.Second)
		if err != nil {
			return err
		}
		if !claimed {
			continue
		}
		err = w.reconcileOne(ctx, id)
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
		releaseErr := w.Queue.Release(releaseCtx, id, token)
		cancel()
		if err != nil {
			log.Printf("[CangyuanMediaRecovery] job %s: %v", id, err)
		}
		if releaseErr != nil {
			log.Printf("[CangyuanMediaRecovery] release lease %s: %v", id, releaseErr)
		}
	}
	return nil
}

func (w *CangyuanMediaRecoveryWorker) reconcileOne(ctx context.Context, id string) error {
	job, err := w.Jobs.Store.Get(ctx, id)
	if errors.Is(err, ErrCangyuanMediaJobNotFound) {
		return w.Queue.Remove(ctx, id)
	}
	if err != nil {
		return err
	}
	if job == nil || job.ID != id || job.BillingState != "held" {
		return w.Queue.Remove(ctx, id)
	}
	switch job.Status {
	case CangyuanMediaJobReserved:
		// Initial POST is controlled only by the original user request flow.
		return w.Queue.Delay(ctx, id, 60*time.Second)
	case CangyuanMediaJobSubmitting, CangyuanMediaJobSubmitUnknown:
		// Never re-POST an ambiguous submission. Keep it indexed for operator
		// inspection instead of treating missing remote ID as a failure.
		return w.Queue.Delay(ctx, id, 15*time.Minute)
	}

	owner := CangyuanMediaJobOwner{UserID: job.UserID, APIKeyID: job.APIKeyID}
	var client *CangyuanMediaClient
	if job.Status == CangyuanMediaJobQueued || job.Status == CangyuanMediaJobProcessing {
		account, err := w.Accounts.GetByID(ctx, job.AccountID)
		if err != nil || account == nil || !account.AllowsCangyuanMediaModel(job.Kind, job.Model) ||
			account.GetCredential("api_key") == "" {
			// An account can be paused/rotated after a billable task starts.
			// Do not refund an already accepted job without authoritative state.
			_ = w.Queue.Delay(ctx, id, time.Minute)
			return errors.New("original cangyuan media upstream account unavailable")
		}
		if w.ClientFactory != nil {
			client = w.ClientFactory(account)
		} else {
			client = &CangyuanMediaClient{
				APIKey: account.GetCredential("api_key"),
				HTTP: &http.Client{
					Timeout: 20 * time.Second,
					CheckRedirect: func(*http.Request, []*http.Request) error {
						return http.ErrUseLastResponse
					},
				},
			}
		}
	}
	workCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	updated, err := w.Jobs.ReconcileHeldCangyuanMediaJob(workCtx, owner, id, client, w.Billing)
	if err != nil {
		_ = w.Queue.Delay(ctx, id, 30*time.Second)
		return err
	}
	if updated == nil {
		_ = w.Queue.Delay(ctx, id, 30*time.Second)
		return errors.New("cangyuan media recovery returned no task")
	}
	if updated.BillingState == "settled" || updated.BillingState == "released" {
		return w.Queue.Remove(ctx, id)
	}
	return w.Queue.Delay(ctx, id, 6*time.Second)
}
