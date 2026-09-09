package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

// This file wires the CQU browser bridge into the existing Chat Completions
// forwarding path. It follows the same shape as the Gemini/Antigravity compat
// branches: the account decides the transport, everything after the upstream
// response (client writes, usage extraction, failover, billing) stays on the
// shared gateway lifecycle.

// cquBrowserUpstreamEndpoint is what usage records show for CQU traffic.
const cquBrowserUpstreamEndpoint = cquSendChatAPIPath

// cquProviderAdapterRegistry lazily builds the production registry that owns
// the CQU adapter. Registration happens here rather than in a global so the
// adapter always sees the running process's config, and so no other provider's
// registry or resolution behavior is touched.
func (s *OpenAIGatewayService) cquProviderAdapterRegistry() *ProviderAdapterRegistry {
	s.cquAdaptersOnce.Do(func() {
		s.cquConversations = newCQUConversationStore(0, 0)
		s.cquAdapters = MustNewProviderAdapterRegistry(
			NewCQUBrowserProviderAdapter(CQUBrowserAdapterDeps{
				Config:        s.cfg,
				Conversations: s.cquConversations,
			}),
		)
	})
	return s.cquAdapters
}

// forwardCQUBrowserChatCompletions runs one Chat Completions request through
// the CQU browser bridge.
//
// The adapter returns an already-OpenAI-shaped response, so the streaming and
// buffering helpers used by every other OpenAI-compatible upstream are reused
// verbatim — including client-disconnect handling, usage extraction and the
// first-token timing that feeds ops metrics.
func (s *OpenAIGatewayService) forwardCQUBrowserChatCompletions(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	defaultMappedModel string,
) (*OpenAIForwardResult, error) {
	startTime := time.Now()

	originalModel := gjson.GetBytes(body, "model").String()
	if originalModel == "" {
		writeChatCompletionsError(c, http.StatusBadRequest, "invalid_request_error", "model is required")
		return nil, errors.New("missing model in request")
	}
	clientStream := gjson.GetBytes(body, "stream").Bool()

	billingModel := resolveOpenAIForwardModel(account, originalModel, defaultMappedModel)
	upstreamModel := billingModel

	registry := s.cquProviderAdapterRegistry()
	adapter, err := registry.Resolve(account)
	if err != nil {
		return nil, err
	}

	resp, err := cquSendThroughAdapter(ctx, adapter, ProviderRequestInput{
		Account:       account,
		Protocol:      ProviderProtocolChatCompletions,
		Model:         upstreamModel,
		Body:          body,
		Stream:        clientStream,
		ClientHeaders: c.Request.Header,
	})
	if err != nil {
		if kind := CQUBridgeErrorKind(err); kind != "" && !c.Writer.Written() {
			status, errType, message := cquBridgeClientError(kind, err)
			SetOpsUpstreamError(c, status, message, "")
			writeChatCompletionsError(c, status, errType, message)
		}
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	SetActualOpenAIUpstreamEndpoint(c, cquBrowserUpstreamEndpoint)
	logger.L().Debug("cqu chat_completions: forwarded through browser bridge",
		zap.String("adapter", cquBrowserAdapterName),
		zap.Int64("account_id", account.ID),
		zap.String("model", originalModel),
		zap.Bool("stream", clientStream),
		zap.Int("upstream_status", resp.StatusCode),
	)

	if resp.StatusCode >= 400 {
		respBody, upstreamMsg := s.readOpenAIUpstreamError(resp)
		normalized := adapter.NormalizeError(resp, respBody)
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: resp.StatusCode,
			Kind:               normalized.Type,
			Message:            upstreamMsg,
		})
		if normalized.Failover || normalized.Retryable {
			return nil, &UpstreamFailoverError{
				StatusCode:      resp.StatusCode,
				ResponseBody:    respBody,
				ResponseHeaders: resp.Header.Clone(),
			}
		}
		return s.handleChatCompletionsErrorResponse(resp, c, account, billingModel)
	}

	var result *OpenAIForwardResult
	var forwardErr error
	if clientStream {
		result, forwardErr = s.streamRawChatCompletions(c, resp, account, originalModel, billingModel, upstreamModel, nil, nil, startTime, len(body))
	} else {
		result, forwardErr = s.bufferRawChatCompletions(c, resp, originalModel, billingModel, upstreamModel, nil, nil, startTime)
	}
	if result != nil {
		result.UpstreamEndpoint = cquBrowserUpstreamEndpoint
	}
	return result, forwardErr
}

// cquSendThroughAdapter runs the provider-adapter pipeline: prepare, build,
// apply auth, send. CQU is the one provider whose transport is not the shared
// HTTP upstream, so the send step goes through the adapter's own
// ProviderRequestSender.
func cquSendThroughAdapter(ctx context.Context, adapter ProviderAdapter, input ProviderRequestInput) (*http.Response, error) {
	builder, ok := adapter.(ProviderRequestBuilder)
	if !ok {
		return nil, fmt.Errorf("provider %s does not implement request builder", adapter.Name())
	}
	sender, ok := adapter.(ProviderRequestSender)
	if !ok {
		return nil, fmt.Errorf("provider %s does not implement request sender", adapter.Name())
	}

	prepared := input
	if preparer, ok := adapter.(ProviderRequestPreparer); ok {
		var err error
		prepared, err = preparer.PrepareRequest(ctx, input)
		if err != nil {
			return nil, err
		}
	}

	req, err := builder.BuildRequest(ctx, prepared)
	if err != nil {
		return nil, err
	}
	if auth, ok := adapter.(ProviderAuthApplier); ok {
		if err := auth.ApplyAuth(ctx, req, prepared); err != nil {
			return nil, err
		}
	}
	return sender.SendRequest(ctx, req, prepared.Account)
}

// cquBridgeClientError maps a bridge failure onto the client-facing error.
// The messages describe operator state (browser not running, no CQU tab) and
// never echo cookies, tokens or the dynamic security parameter, none of which
// this process ever holds — which is also why err itself is never rendered
// into the client response.
func cquBridgeClientError(kind string, _ error) (int, string, string) {
	switch kind {
	case CQUErrorKindDisabled:
		return http.StatusServiceUnavailable, "upstream_error",
			"CQU browser bridge is disabled; set CQU_ENABLED=true and point CQU_BROWSER_DEBUG_URL at the signed-in browser"
	case CQUErrorKindBrowserUnavailable:
		return http.StatusServiceUnavailable, CQUErrorKindBrowserUnavailable,
			"CQU browser is not reachable over CDP; start the persistent browser with --remote-debugging-port and sign in to " + CQUDefaultOrigin
	case CQUErrorKindPageMissing:
		return http.StatusServiceUnavailable, CQUErrorKindPageMissing,
			"No signed-in CQU page found in the browser; open " + CQUDefaultOrigin + " in the persistent browser profile"
	default:
		return http.StatusBadGateway, "upstream_error", "CQU browser bridge request failed"
	}
}
