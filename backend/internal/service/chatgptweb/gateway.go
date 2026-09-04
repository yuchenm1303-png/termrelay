package chatgptweb

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

const defaultAccountAcquireAttempts = 8

var (
	ErrConversationBindingMissing     = errors.New("chatgptweb: conversation sticky binding is missing")
	ErrConversationAccountUnavailable = errors.New("chatgptweb: conversation account is unavailable")
	ErrStickyBindingConflict          = errors.New("chatgptweb: sticky binding conflict")
)

type StickyKeyKind string

const (
	StickyKeyConversation  StickyKeyKind = "conversation"
	StickyKeyClientSession StickyKeyKind = "client_session"
)

type StickyKey struct {
	Kind StickyKeyKind
	ID   string
}

func (k StickyKey) String() string {
	return string(k.Kind) + ":" + strings.TrimSpace(k.ID)
}

func (k StickyKey) validate() error {
	if strings.TrimSpace(k.ID) == "" {
		return errors.New("chatgptweb: sticky key id is required")
	}
	switch k.Kind {
	case StickyKeyConversation, StickyKeyClientSession:
		return nil
	default:
		return fmt.Errorf("chatgptweb: unsupported sticky key kind %q", k.Kind)
	}
}

// StickyStore owns only persistence semantics. Account availability and retry
// policy deliberately live in the gateway rather than the Redis repository.
type StickyStore interface {
	GetBinding(ctx context.Context, key StickyKey) (accountID int64, found bool, err error)
	Bind(ctx context.Context, key StickyKey, accountID int64, ttl time.Duration) (boundAccountID int64, created bool, err error)
	Refresh(ctx context.Context, key StickyKey, accountID int64, ttl time.Duration) (refreshed bool, err error)
	Release(ctx context.Context, key StickyKey, accountID int64) (released bool, err error)
}

type AccountRef struct {
	ID int64
}

// AccountSelector separates scheduling from account credentials. Resolve must
// report unavailable accounts with available=false; Select chooses a new
// schedulable account while respecting excludeAccountIDs.
type AccountSelector interface {
	Resolve(ctx context.Context, accountID int64) (account AccountRef, available bool, err error)
	Select(ctx context.Context, excludeAccountIDs []int64) (AccountRef, error)
}

type TransportRequest struct {
	Conversation ConversationRequest
	ClientState  ClientState
}

// Transport is the account-scoped execution boundary. Credential/cookie/session
// material belongs behind this interface in the later credential layer.
type Transport interface {
	Send(ctx context.Context, account AccountRef, request TransportRequest) (Snapshot, error)
}

type SendRequest struct {
	ClientState  *ClientState
	Conversation ConversationRequest
}

type SendResult struct {
	AccountID int64
	Snapshot  Snapshot
	Session   SessionState
}

type ChatGPTWebGateway interface {
	Send(ctx context.Context, request SendRequest) (SendResult, error)
}

type gateway struct {
	stickyStore StickyStore
	selector    AccountSelector
	transport   Transport
	stickyTTL   time.Duration
	now         func() time.Time
}

func NewGateway(stickyStore StickyStore, selector AccountSelector, transport Transport, stickyTTL time.Duration) (ChatGPTWebGateway, error) {
	if stickyStore == nil {
		return nil, errors.New("chatgptweb: sticky store is required")
	}
	if selector == nil {
		return nil, errors.New("chatgptweb: account selector is required")
	}
	if transport == nil {
		return nil, errors.New("chatgptweb: transport is required")
	}
	if stickyTTL <= 0 {
		return nil, errors.New("chatgptweb: sticky TTL must be positive")
	}
	return &gateway{
		stickyStore: stickyStore,
		selector:    selector,
		transport:   transport,
		stickyTTL:   stickyTTL,
		now:         func() time.Time { return time.Now().UTC() },
	}, nil
}

func (g *gateway) Send(ctx context.Context, request SendRequest) (SendResult, error) {
	if request.ClientState == nil {
		return SendResult{}, errors.New("chatgptweb: client state is required")
	}
	key, err := stickyKeyFor(request.ClientState, request.Conversation)
	if err != nil {
		return SendResult{}, err
	}

	excluded := make(map[int64]struct{})
	for attempt := 0; attempt < defaultAccountAcquireAttempts; attempt++ {
		account, err := g.acquireAccount(ctx, key, excluded)
		if err != nil {
			return SendResult{}, err
		}

		transportRequest := TransportRequest{
			Conversation: mergeConversationState(request.Conversation, request.ClientState),
			ClientState:  *request.ClientState,
		}
		snapshot, sendErr := g.transport.Send(ctx, account, transportRequest)
		if sendErr != nil {
			if shouldReleaseSticky(key, sendErr) {
				if _, releaseErr := g.stickyStore.Release(ctx, key, account.ID); releaseErr != nil {
					return SendResult{}, fmt.Errorf("chatgptweb: release failed account binding: %w", releaseErr)
				}
			}
			if key.Kind == StickyKeyClientSession && shouldRetryNextAccount(sendErr) {
				excluded[account.ID] = struct{}{}
				continue
			}
			return SendResult{}, sendErr
		}

		if err := g.keepBinding(ctx, key, account.ID); err != nil {
			return SendResult{}, err
		}
		if snapshot.ConversationID != "" {
			conversationKey := StickyKey{Kind: StickyKeyConversation, ID: snapshot.ConversationID}
			if err := g.bindConversation(ctx, conversationKey, account.ID); err != nil {
				return SendResult{}, err
			}
		}

		request.ClientState.NoteTurnResult(snapshot.ConversationID, snapshot.ParentMessageID)
		now := g.now()
		return SendResult{
			AccountID: account.ID,
			Snapshot:  snapshot,
			Session: SessionState{
				Route:           string(RoutingModeChatGPTWeb),
				AccountID:       account.ID,
				ConversationID:  request.ClientState.ConversationID,
				ParentMessageID: request.ClientState.ParentMessageID,
				ClientSessionID: request.ClientState.SessionID,
				UpdatedAt:       now,
			},
		}, nil
	}
	return SendResult{}, errors.New("chatgptweb: account acquisition attempts exhausted")
}

func (g *gateway) acquireAccount(ctx context.Context, key StickyKey, excluded map[int64]struct{}) (AccountRef, error) {
	for attempt := 0; attempt < defaultAccountAcquireAttempts; attempt++ {
		boundID, found, err := g.stickyStore.GetBinding(ctx, key)
		if err != nil {
			return AccountRef{}, fmt.Errorf("chatgptweb: get sticky binding: %w", err)
		}
		if found {
			account, available, err := g.selector.Resolve(ctx, boundID)
			if err != nil {
				return AccountRef{}, err
			}
			if available && account.ID == boundID {
				refreshed, err := g.stickyStore.Refresh(ctx, key, boundID, g.stickyTTL)
				if err != nil {
					return AccountRef{}, fmt.Errorf("chatgptweb: refresh sticky binding: %w", err)
				}
				if refreshed {
					return account, nil
				}
				continue
			}
			if _, err := g.stickyStore.Release(ctx, key, boundID); err != nil {
				return AccountRef{}, fmt.Errorf("chatgptweb: release unavailable account binding: %w", err)
			}
			if key.Kind == StickyKeyConversation {
				return AccountRef{}, fmt.Errorf("%w: account %d", ErrConversationAccountUnavailable, boundID)
			}
			excluded[boundID] = struct{}{}
			continue
		}

		// A known upstream conversation cannot be safely migrated to a random
		// account after its binding expires or is evicted. Active conversations
		// keep this mapping alive through Refresh and initial session promotion.
		if key.Kind == StickyKeyConversation {
			return AccountRef{}, ErrConversationBindingMissing
		}

		account, err := g.selector.Select(ctx, sortedAccountIDs(excluded))
		if err != nil {
			return AccountRef{}, err
		}
		if account.ID <= 0 {
			return AccountRef{}, errors.New("chatgptweb: selector returned invalid account id")
		}
		winnerID, _, err := g.stickyStore.Bind(ctx, key, account.ID, g.stickyTTL)
		if err != nil {
			return AccountRef{}, fmt.Errorf("chatgptweb: bind sticky account: %w", err)
		}
		if winnerID == account.ID {
			refreshed, err := g.stickyStore.Refresh(ctx, key, winnerID, g.stickyTTL)
			if err != nil {
				return AccountRef{}, fmt.Errorf("chatgptweb: refresh newly acquired binding: %w", err)
			}
			if refreshed {
				return account, nil
			}
			continue
		}

		// Another request won the atomic bind. Resolve and use that winner so all
		// concurrent requests for this client session converge on one account.
		winner, available, err := g.selector.Resolve(ctx, winnerID)
		if err != nil {
			return AccountRef{}, err
		}
		if available && winner.ID == winnerID {
			refreshed, err := g.stickyStore.Refresh(ctx, key, winnerID, g.stickyTTL)
			if err != nil {
				return AccountRef{}, fmt.Errorf("chatgptweb: refresh winning binding: %w", err)
			}
			if refreshed {
				return winner, nil
			}
			continue
		}
		if _, err := g.stickyStore.Release(ctx, key, winnerID); err != nil {
			return AccountRef{}, fmt.Errorf("chatgptweb: release unavailable winning binding: %w", err)
		}
		excluded[winnerID] = struct{}{}
	}
	return AccountRef{}, errors.New("chatgptweb: could not acquire available account")
}

func (g *gateway) keepBinding(ctx context.Context, key StickyKey, accountID int64) error {
	refreshed, err := g.stickyStore.Refresh(ctx, key, accountID, g.stickyTTL)
	if err != nil {
		return fmt.Errorf("chatgptweb: refresh sticky binding after send: %w", err)
	}
	if refreshed {
		return nil
	}
	winnerID, _, err := g.stickyStore.Bind(ctx, key, accountID, g.stickyTTL)
	if err != nil {
		return fmt.Errorf("chatgptweb: restore expired sticky binding: %w", err)
	}
	if winnerID != accountID {
		return fmt.Errorf("%w: expected account %d, got %d", ErrStickyBindingConflict, accountID, winnerID)
	}
	return nil
}

func (g *gateway) bindConversation(ctx context.Context, key StickyKey, accountID int64) error {
	winnerID, _, err := g.stickyStore.Bind(ctx, key, accountID, g.stickyTTL)
	if err != nil {
		return fmt.Errorf("chatgptweb: promote conversation sticky binding: %w", err)
	}
	if winnerID != accountID {
		return fmt.Errorf("%w: conversation %s is bound to account %d, current account %d", ErrStickyBindingConflict, key.ID, winnerID, accountID)
	}
	refreshed, err := g.stickyStore.Refresh(ctx, key, accountID, g.stickyTTL)
	if err != nil {
		return fmt.Errorf("chatgptweb: refresh conversation sticky binding: %w", err)
	}
	if refreshed {
		return nil
	}
	winnerID, _, err = g.stickyStore.Bind(ctx, key, accountID, g.stickyTTL)
	if err != nil {
		return fmt.Errorf("chatgptweb: restore expired conversation sticky binding: %w", err)
	}
	if winnerID != accountID {
		return fmt.Errorf("%w: conversation %s was rebound to account %d, current account %d", ErrStickyBindingConflict, key.ID, winnerID, accountID)
	}
	return nil
}

func stickyKeyFor(state *ClientState, request ConversationRequest) (StickyKey, error) {
	conversationID := strings.TrimSpace(request.ConversationID)
	if conversationID == "" && state != nil {
		conversationID = strings.TrimSpace(state.ConversationID)
	}
	if conversationID != "" {
		key := StickyKey{Kind: StickyKeyConversation, ID: conversationID}
		return key, key.validate()
	}
	if state == nil || strings.TrimSpace(state.SessionID) == "" {
		return StickyKey{}, errors.New("chatgptweb: client session id is required before conversation creation")
	}
	key := StickyKey{Kind: StickyKeyClientSession, ID: state.SessionID}
	return key, key.validate()
}

func mergeConversationState(request ConversationRequest, state *ClientState) ConversationRequest {
	if state == nil {
		return request
	}
	if request.ConversationID == "" {
		request.ConversationID = state.ConversationID
	}
	if request.ParentMessageID == "" {
		request.ParentMessageID = state.ParentMessageID
	}
	return request
}

func sortedAccountIDs(values map[int64]struct{}) []int64 {
	ids := make([]int64, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func shouldRetryNextAccount(err error) bool {
	var upstream *UpstreamError
	return errors.As(err, &upstream) && upstream.RetryNextAccount
}

func shouldReleaseSticky(key StickyKey, err error) bool {
	var upstream *UpstreamError
	if !errors.As(err, &upstream) {
		return false
	}
	if upstream.RequiresReauth || upstream.ChallengeRequired {
		return true
	}
	return key.Kind == StickyKeyClientSession && upstream.RetryNextAccount
}
