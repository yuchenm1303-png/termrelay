package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/service/chatgptweb"
)

// ChatGPTWebAccountSelector is a thin adapter over the existing OpenAI
// scheduler. It preserves scheduler concurrency leases and filters out Codex-
// only accounts without introducing a second scheduling policy.
type ChatGPTWebAccountSelector struct {
	gateway     *OpenAIGatewayService
	accountRepo AccountRepository
	concurrency *ConcurrencyService
}

var _ chatgptweb.AccountSelector = (*ChatGPTWebAccountSelector)(nil)

func NewChatGPTWebAccountSelector(gateway *OpenAIGatewayService, accountRepo AccountRepository, concurrency *ConcurrencyService) chatgptweb.AccountSelector {
	return &ChatGPTWebAccountSelector{gateway: gateway, accountRepo: accountRepo, concurrency: concurrency}
}

func (s *ChatGPTWebAccountSelector) Resolve(ctx context.Context, request chatgptweb.AccountSelectionRequest, accountID int64) (chatgptweb.AccountRef, bool, error) {
	if s == nil || s.accountRepo == nil || s.concurrency == nil {
		return chatgptweb.AccountRef{}, false, errors.New("chatgptweb: account selector is not initialized")
	}
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return chatgptweb.AccountRef{}, false, fmt.Errorf("chatgptweb: resolve account %d: %w", accountID, err)
	}
	if !chatGPTWebAccountCompatible(account, request.GroupID) {
		return chatgptweb.AccountRef{}, false, nil
	}
	acquired, err := s.concurrency.AcquireAccountSlot(ctx, account.ID, account.Concurrency)
	if err != nil {
		return chatgptweb.AccountRef{}, false, fmt.Errorf("chatgptweb: acquire account %d: %w", account.ID, err)
	}
	if acquired == nil || !acquired.Acquired {
		return chatgptweb.AccountRef{}, false, nil
	}
	return chatgptweb.NewAccountRef(account.ID, acquired.ReleaseFunc), true, nil
}

func (s *ChatGPTWebAccountSelector) Select(ctx context.Context, request chatgptweb.AccountSelectionRequest, excludeAccountIDs []int64) (chatgptweb.AccountRef, error) {
	if s == nil || s.gateway == nil {
		return chatgptweb.AccountRef{}, errors.New("chatgptweb: OpenAI scheduler is not initialized")
	}
	excluded := make(map[int64]struct{}, len(excludeAccountIDs)+1)
	for _, id := range excludeAccountIDs {
		excluded[id] = struct{}{}
	}
	for attempt := 0; attempt < openAIAccountSelectionProbeLimit; attempt++ {
		selection, _, err := s.gateway.SelectAccountWithSchedulerForCapability(
			ctx,
			request.GroupID,
			"",
			request.SessionHash,
			request.RequestedModel,
			excluded,
			OpenAIUpstreamTransportHTTPSSE,
			OpenAIEndpointCapabilityChatCompletions,
			false,
			false,
			false,
		)
		if err != nil {
			return chatgptweb.AccountRef{}, err
		}
		if selection == nil || selection.Account == nil || !selection.Acquired {
			return chatgptweb.AccountRef{}, errors.New("chatgptweb: scheduler returned no acquired account")
		}
		if selection.Account.SupportsOpenAIUpstreamRoute(OpenAIUpstreamRouteChatGPTWeb) {
			return chatgptweb.NewAccountRef(selection.Account.ID, selection.ReleaseFunc), nil
		}
		if selection.ReleaseFunc != nil {
			selection.ReleaseFunc()
		}
		excluded[selection.Account.ID] = struct{}{}
	}
	return chatgptweb.AccountRef{}, errors.New("chatgptweb: no route-compatible account available")
}

func chatGPTWebAccountCompatible(account *Account, groupID *int64) bool {
	if account == nil || !account.IsSchedulable() || !account.SupportsOpenAIUpstreamRoute(OpenAIUpstreamRouteChatGPTWeb) {
		return false
	}
	if groupID == nil {
		return true
	}
	for _, id := range account.GroupIDs {
		if id == *groupID {
			return true
		}
	}
	for _, membership := range account.AccountGroups {
		if membership.GroupID == *groupID {
			return true
		}
	}
	return false
}

// ChatGPTWebHTTPClientProvider delegates every round trip to HTTPUpstream, so
// proxy handling, TLS fingerprinting and pooled transports retain their
// existing lifecycle.
type ChatGPTWebHTTPClientProvider struct {
	accountRepo AccountRepository
	proxy       *ProxyService
	upstream    HTTPUpstream
	tlsProfiles *TLSFingerprintProfileService
}

var _ chatgptweb.HTTPClientProvider = (*ChatGPTWebHTTPClientProvider)(nil)

func NewChatGPTWebHTTPClientProvider(accountRepo AccountRepository, proxy *ProxyService, upstream HTTPUpstream, tlsProfiles *TLSFingerprintProfileService) chatgptweb.HTTPClientProvider {
	return &ChatGPTWebHTTPClientProvider{accountRepo: accountRepo, proxy: proxy, upstream: upstream, tlsProfiles: tlsProfiles}
}

func (p *ChatGPTWebHTTPClientProvider) HTTPClient(ctx context.Context, ref chatgptweb.AccountRef) (*http.Client, error) {
	if p == nil || p.accountRepo == nil || p.upstream == nil {
		return nil, errors.New("chatgptweb: HTTP client provider is not initialized")
	}
	account, err := p.accountRepo.GetByID(ctx, ref.ID)
	if err != nil {
		return nil, fmt.Errorf("chatgptweb: load HTTP account %d: %w", ref.ID, err)
	}
	proxyURL := ""
	if account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	} else if account.ProxyID != nil {
		if p.proxy == nil {
			return nil, errors.New("chatgptweb: proxy service is not initialized")
		}
		proxyURL, err = p.proxy.GetURL(ctx, *account.ProxyID)
		if err != nil {
			return nil, fmt.Errorf("chatgptweb: resolve account proxy: %w", err)
		}
	}
	return &http.Client{Transport: chatGPTWebRoundTripper(func(req *http.Request) (*http.Response, error) {
		req = req.WithContext(WithHTTPUpstreamProfile(req.Context(), HTTPUpstreamProfileOpenAI))
		if p.tlsProfiles != nil {
			return p.upstream.DoWithTLS(req, proxyURL, account.ID, account.Concurrency, p.tlsProfiles.ResolveTLSProfile(account))
		}
		return p.upstream.Do(req, proxyURL, account.ID, account.Concurrency)
	})}, nil
}

type chatGPTWebRoundTripper func(*http.Request) (*http.Response, error)

func (f chatGPTWebRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
