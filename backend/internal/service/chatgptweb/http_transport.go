package chatgptweb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HTTPClientProvider returns the account-scoped HTTP client used for ChatGPT
// Web. Proxy, TLS fingerprinting and connection pooling remain infrastructure
// concerns behind this boundary.
type HTTPClientProvider interface {
	HTTPClient(ctx context.Context, account AccountRef) (*http.Client, error)
}

type HTTPClientProviderFunc func(context.Context, AccountRef) (*http.Client, error)

func (f HTTPClientProviderFunc) HTTPClient(ctx context.Context, account AccountRef) (*http.Client, error) {
	return f(ctx, account)
}

// RequirementsTokenProviderFactory keeps the dynamic Sentinel requirements
// token out of persisted account credentials. A later browser-session layer can
// provide account-scoped tokens without changing the gateway or transport.
type RequirementsTokenProviderFactory interface {
	ForAccount(ctx context.Context, account AccountRef, session *AccountSession) (RequirementsTokenProvider, error)
}

type RequirementsTokenProviderFactoryFunc func(context.Context, AccountRef, *AccountSession) (RequirementsTokenProvider, error)

func (f RequirementsTokenProviderFactoryFunc) ForAccount(ctx context.Context, account AccountRef, session *AccountSession) (RequirementsTokenProvider, error) {
	return f(ctx, account, session)
}

type httpTransport struct {
	sessions     AccountSessionProvider
	httpClients  HTTPClientProvider
	requirements RequirementsTokenProviderFactory
	proofSolver  ProofOfWorkSolver
	baseURL      string
}

var _ Transport = (*httpTransport)(nil)
var _ StreamingTransport = (*httpTransport)(nil)

func NewHTTPTransport(
	sessions AccountSessionProvider,
	httpClients HTTPClientProvider,
	requirements RequirementsTokenProviderFactory,
	proofSolver ProofOfWorkSolver,
	baseURL string,
) (Transport, error) {
	if sessions == nil {
		return nil, errors.New("chatgptweb: account session provider is required")
	}
	if httpClients == nil {
		return nil, errors.New("chatgptweb: HTTP client provider is required")
	}
	if requirements == nil {
		return nil, errors.New("chatgptweb: requirements token provider factory is required")
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultBaseURL
	}
	return &httpTransport{
		sessions:     sessions,
		httpClients:  httpClients,
		requirements: requirements,
		proofSolver:  proofSolver,
		baseURL:      strings.TrimRight(baseURL, "/"),
	}, nil
}

func (t *httpTransport) Send(ctx context.Context, account AccountRef, request TransportRequest) (Snapshot, error) {
	return t.SendStream(ctx, account, request, nil)
}

func (t *httpTransport) SendStream(ctx context.Context, account AccountRef, request TransportRequest, sink StreamSink) (Snapshot, error) {
	if account.ID <= 0 {
		return Snapshot{}, errors.New("chatgptweb: transport account id must be positive")
	}
	session, err := t.sessions.LoadSession(ctx, account)
	if err != nil {
		return Snapshot{}, err
	}
	if session == nil || session.identity == nil {
		return Snapshot{}, errors.New("chatgptweb: account session has no identity")
	}
	httpClient, err := t.httpClients.HTTPClient(ctx, account)
	if err != nil {
		return Snapshot{}, fmt.Errorf("chatgptweb: resolve HTTP client: %w", err)
	}
	if httpClient == nil {
		return Snapshot{}, errors.New("chatgptweb: HTTP client provider returned nil")
	}
	client, err := NewClient(httpClient, t.baseURL, session.identity)
	if err != nil {
		return Snapshot{}, err
	}

	state := request.ClientState
	// Browser identity is account-scoped. Keep the gateway's logical session ID
	// and conversation continuity, but use the stored browser fingerprint when it
	// is available so credentials and device identity cannot drift apart.
	if deviceID := session.DeviceID(); deviceID != "" {
		state.DeviceID = deviceID
	}
	if userAgent := session.UserAgent(); userAgent != "" {
		state.UserAgent = userAgent
	}
	if request.Conversation.ConversationID != "" {
		state.ConversationID = request.Conversation.ConversationID
	}
	if request.Conversation.ParentMessageID != "" {
		state.ParentMessageID = request.Conversation.ParentMessageID
	}

	if strings.TrimSpace(request.Conversation.ConversationID) == "" {
		if _, err := client.InitConversation(ctx, &state); err != nil {
			return Snapshot{}, err
		}
	}

	requirementsProvider, err := t.requirements.ForAccount(ctx, account, session)
	if err != nil {
		return Snapshot{}, fmt.Errorf("chatgptweb: resolve requirements token provider: %w", err)
	}
	if requirementsProvider == nil {
		return Snapshot{}, errors.New("chatgptweb: requirements token provider factory returned nil")
	}
	sentinel, err := NewSentinelFlow(client, requirementsProvider, t.proofSolver)
	if err != nil {
		return Snapshot{}, err
	}
	tokens, err := sentinel.Run(ctx, &state)
	if err != nil {
		return Snapshot{}, err
	}

	turnTraceID := newUUID()
	conduitToken, err := client.PrepareConversation(ctx, request.Conversation, &state, tokens, turnTraceID)
	if err != nil {
		return Snapshot{}, err
	}
	resp, err := client.OpenConversation(ctx, request.Conversation, &state, tokens, conduitToken, turnTraceID)
	if err != nil {
		return Snapshot{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	parser := NewStreamParser(resp.Body)
	for {
		event, err := parser.Next()
		if errors.Is(err, io.EOF) {
			return Snapshot{}, &UpstreamError{Kind: ErrorKindProtocol, Message: "conversation stream ended before DONE"}
		}
		if err != nil {
			return Snapshot{}, fmt.Errorf("chatgptweb: parse conversation stream: %w", err)
		}
		if sink != nil {
			if err := sink(event); err != nil {
				return Snapshot{}, err
			}
		}
		if event.Type == StreamEventDone {
			return event.Snapshot, nil
		}
	}
}
