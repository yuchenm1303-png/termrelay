package chatgptweb

import (
	"fmt"
	"strings"
	"time"
)

type ErrorKind string

const (
	ErrorKindAuthentication ErrorKind = "authentication"
	ErrorKindChallenge      ErrorKind = "challenge_required"
	ErrorKindRateLimit      ErrorKind = "rate_limit"
	ErrorKindTransient      ErrorKind = "transient"
	ErrorKindProtocol       ErrorKind = "protocol"
	ErrorKindClientClosed   ErrorKind = "client_closed"
)

type UpstreamError struct {
	Kind              ErrorKind
	StatusCode        int
	Message           string
	RetryAfter        time.Duration
	RetryNextAccount  bool
	RequiresReauth    bool
	ChallengeRequired bool
}

func (e *UpstreamError) Error() string {
	if e == nil {
		return ""
	}
	if e.StatusCode > 0 {
		return fmt.Sprintf("chatgpt web upstream %s (status %d)", e.Kind, e.StatusCode)
	}
	return fmt.Sprintf("chatgpt web upstream %s", e.Kind)
}

func ClassifyHTTPError(status int, responseBody []byte, retryAfter time.Duration) *UpstreamError {
	body := strings.ToLower(string(responseBody))
	switch {
	case status == 401:
		return &UpstreamError{Kind: ErrorKindAuthentication, StatusCode: status, Message: "upstream authentication failed", RequiresReauth: true}
	case status == 403 && containsHumanChallenge(body):
		return &UpstreamError{Kind: ErrorKindChallenge, StatusCode: status, Message: "upstream requires interactive verification", ChallengeRequired: true}
	case status == 403:
		return &UpstreamError{Kind: ErrorKindAuthentication, StatusCode: status, Message: "upstream access forbidden", RequiresReauth: true}
	case status == 429:
		return &UpstreamError{Kind: ErrorKindRateLimit, StatusCode: status, Message: "upstream rate limited", RetryAfter: retryAfter, RetryNextAccount: true}
	case status >= 500 && status <= 599:
		return &UpstreamError{Kind: ErrorKindTransient, StatusCode: status, Message: "upstream temporary failure", RetryNextAccount: true}
	default:
		return &UpstreamError{Kind: ErrorKindProtocol, StatusCode: status, Message: "unexpected upstream response"}
	}
}

func containsHumanChallenge(body string) bool {
	for _, marker := range []string{"turnstile", "captcha", "challenge_required", "interactive verification", "verify you are human"} {
		if strings.Contains(body, marker) {
			return true
		}
	}
	return false
}
