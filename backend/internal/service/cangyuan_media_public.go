package service

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var ErrCangyuanMediaNoAccount = errors.New("no eligible cangyuan media account in this API key group")
var ErrCangyuanMediaPriceChanged = errors.New("approved media price changed before balance reservation")

// Optional, caller-supplied maximum consent for a canvas payment. It is not
// trusted as a billable price; the server compares it to its frozen snapshot.
type CangyuanMediaPriceExpectation struct {
	TotalSaleUSD   string
	PricingVersion string
}

func cangyuanMediaPriceMatchesExpectation(snapshot *CangyuanMediaPriceSnapshot, expected CangyuanMediaPriceExpectation) bool {
	if snapshot == nil || !snapshot.VerifiedByAdmin ||
		snapshot.Currency != "USD" || snapshot.PricingVersion != expected.PricingVersion ||
		snapshot.Units <= 0 {
		return false
	}
	unit, err := parseCangyuanMediaMoney(snapshot.UnitSalePriceUSD)
	if err != nil {
		return false
	}
	consented, err := parseCangyuanMediaMoney(expected.TotalSaleUSD)
	if err != nil || !consented.IsPositive() {
		return false
	}
	return unit.Mul(decimal.NewFromInt(int64(snapshot.Units))).Equal(consented)
}

// CangyuanMediaPublicAccountLister selects exclusively from the API key's
// authenticated group, never from all accounts or an ungrouped fallback.
type CangyuanMediaPublicAccountLister interface {
	ListSchedulableByGroupIDAndPlatform(context.Context, int64, string) ([]Account, error)
}

// CangyuanMediaPublicService coordinates the verified, funded, one-POST-only
// media lifecycle. The HTTP handler must authenticate the API key and verify
// group / billing entitlements before calling this service.
type CangyuanMediaPublicService struct {
	Jobs     *CangyuanMediaJobService
	Prices   *CangyuanMediaPriceRegistry
	Accounts CangyuanMediaPublicAccountLister
	Billing  CangyuanMediaEscrowWriter

	ClientFactory func(*Account) *CangyuanMediaClient
}

func NewCangyuanMediaPublicService(
	jobs *CangyuanMediaJobService,
	prices *CangyuanMediaPriceRegistry,
	accounts AccountRepository,
	billing UsageBillingRepository,
) *CangyuanMediaPublicService {
	return &CangyuanMediaPublicService{
		Jobs: jobs, Prices: prices, Accounts: accounts, Billing: billing,
	}
}

// Candidates returns only actively schedulable, explicitly media-only
// accounts bound to the caller's group and whitelisting the exact media kind.
func (s *CangyuanMediaPublicService) Candidates(
	ctx context.Context, groupID int64, kind CangyuanMediaKind, model string,
) ([]Account, error) {
	if s == nil || s.Accounts == nil || groupID <= 0 ||
		strings.TrimSpace(model) != model || model == "" {
		return nil, ErrCangyuanMediaNoAccount
	}
	accounts, err := s.Accounts.ListSchedulableByGroupIDAndPlatform(ctx, groupID, PlatformOpenAI)
	if err != nil {
		return nil, err
	}
	eligible := make([]Account, 0)
	for _, account := range accounts {
		if account.IsSchedulable() && account.AllowsCangyuanMediaModel(kind, model) &&
			strings.TrimSpace(account.GetCredential("api_key")) != "" {
			eligible = append(eligible, account)
		}
	}
	sort.Slice(eligible, func(i, j int) bool { return eligible[i].ID < eligible[j].ID })
	return eligible, nil
}

// Submit creates or reuses an idempotent task and freezes the exact
// administrator-approved USD price before submitting it to the official API.
// A retry for an accepted or ambiguous task returns its existing local ID
// instead of performing another upstream POST.
func (s *CangyuanMediaPublicService) Submit(
	ctx context.Context, owner CangyuanMediaJobOwner, groupID int64,
	kind CangyuanMediaKind, model string, payload []byte, idempotencyKey string,
	expectations ...CangyuanMediaPriceExpectation,
) (*CangyuanMediaJob, error) {
	if s == nil || s.Jobs == nil || s.Prices == nil || s.Billing == nil {
		return nil, errors.New("cangyuan media public billing is unavailable")
	}
	if err := validateCangyuanMediaJobOwner(owner); err != nil {
		return nil, err
	}
	if len(expectations) > 1 {
		return nil, ErrCangyuanMediaPriceChanged
	}
	if idempotencyKey == "" || len(idempotencyKey) > 128 ||
		strings.TrimSpace(idempotencyKey) != idempotencyKey {
		return nil, errors.New("media generation requires an Idempotency-Key")
	}
	eligible, err := s.Candidates(ctx, groupID, kind, model)
	if err != nil {
		return nil, err
	}
	if len(eligible) == 0 {
		return nil, ErrCangyuanMediaNoAccount
	}
	// Deterministic account choice means retries favor the same account
	// without trusting account IDs or price snapshots from a client.
	seed := sha256.Sum256([]byte(fmt.Sprintf("%d:%d:%d:%s",
		owner.UserID, owner.APIKeyID, groupID, idempotencyKey)))
	account := &eligible[binary.BigEndian.Uint64(seed[:8])%uint64(len(eligible))]
	job, err := s.Jobs.ReserveWithApprovedPricing(ctx, owner, account, kind, model,
		payload, idempotencyKey, s.Prices)
	if err != nil {
		return nil, err
	}
	// The job's frozen snapshot is the actual amount SQL will hold and settle.
	// Check consent again HERE, after reserving the immutable snapshot and
	// before any balance hold or remote POST, closing the quote/submit race.
	if len(expectations) == 1 && !cangyuanMediaPriceMatchesExpectation(job.PriceSnapshot, expectations[0]) {
		return job, ErrCangyuanMediaPriceChanged
	}
	if job.Status != CangyuanMediaJobReserved {
		return job, nil
	}
	if job.BillingState == "pending" {
		job, _, err = s.Jobs.HoldApprovedCangyuanMediaBalance(ctx, owner, job.ID, s.Billing)
		if err != nil {
			return job, err
		}
	}
	if job.BillingState != "held" {
		return job, ErrCangyuanMediaNotBillable
	}
	var client *CangyuanMediaClient
	if s.ClientFactory != nil {
		client = s.ClientFactory(account)
	} else {
		client = &CangyuanMediaClient{
			APIKey: account.GetCredential("api_key"),
			HTTP: &http.Client{
				Timeout: 30 * time.Second,
				CheckRedirect: func(*http.Request, []*http.Request) error {
					return http.ErrUseLastResponse
				},
			},
		}
	}
	latest, _, submitErr := s.Jobs.SubmitReserved(ctx, owner, job.ID, client, payload)
	if submitErr != nil {
		// A definitive provider rejection means no generation was accepted.
		// Release the frozen price now; do not strand money waiting for a
		// poll that cannot succeed without a remote task ID.
		if latest != nil && latest.Status == CangyuanMediaJobFailed &&
			latest.BillingState == "held" {
			released, _, refundErr := s.Jobs.ReleaseFailedCangyuanMediaBalance(
				ctx, owner, job.ID, s.Billing)
			if refundErr == nil {
				return released, submitErr
			}
			return latest, errors.Join(submitErr, refundErr)
		}
		// A timeout or malformed accepted response is ambiguous. Return the
		// original task ID but never automatically retry the upstream POST.
		if stored, readErr := s.Jobs.Get(ctx, owner, job.ID); readErr == nil {
			return stored, submitErr
		}
		return job, submitErr
	}
	return latest, nil
}
