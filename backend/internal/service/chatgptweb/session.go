package chatgptweb

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	AccountPlatformChatGPTWeb = "chatgpt_web"
	AccountTypeCookie         = "cookie"
)

var ErrInvalidAccountCredentials = errors.New("chatgptweb: invalid account credentials")

// CredentialRecord is the persistence boundary for ChatGPT Web account
// credentials. Secret material stays private to this package and is copied on
// construction so callers cannot mutate a loaded record after the fact.
type CredentialRecord struct {
	accountID   int64
	platform    string
	accountType string
	credentials map[string]any
}

func NewCredentialRecord(accountID int64, platform, accountType string, credentials map[string]any) CredentialRecord {
	copied := make(map[string]any, len(credentials))
	for key, value := range credentials {
		copied[key] = value
	}
	return CredentialRecord{
		accountID:   accountID,
		platform:    strings.TrimSpace(platform),
		accountType: strings.TrimSpace(accountType),
		credentials: copied,
	}
}

func (r CredentialRecord) String() string {
	return fmt.Sprintf("chatgptweb.CredentialRecord{account_id:%d,redacted}", r.accountID)
}

func (r CredentialRecord) GoString() string { return r.String() }

func (r CredentialRecord) MarshalJSON() ([]byte, error) { return []byte("{}"), nil }

// CredentialSource owns persistence only. It deliberately returns an opaque
// record so repository code does not need to understand ChatGPT Web credential
// semantics.
type CredentialSource interface {
	LoadCredential(ctx context.Context, accountID int64) (CredentialRecord, error)
}

type CredentialSourceFunc func(context.Context, int64) (CredentialRecord, error)

func (f CredentialSourceFunc) LoadCredential(ctx context.Context, accountID int64) (CredentialRecord, error) {
	return f(ctx, accountID)
}

// AccountSession is the in-memory, account-scoped browser session used by the
// HTTP transport. Identity protects access token and cookie values from generic
// formatting and JSON serialization.
type AccountSession struct {
	identity *Identity
}

func (s *AccountSession) String() string   { return "chatgptweb.AccountSession{redacted}" }
func (s *AccountSession) GoString() string { return s.String() }
func (s *AccountSession) MarshalJSON() ([]byte, error) {
	return []byte("{}"), nil
}

func (s *AccountSession) DeviceID() string {
	if s == nil || s.identity == nil {
		return ""
	}
	return s.identity.DeviceID()
}

func (s *AccountSession) UserAgent() string {
	if s == nil || s.identity == nil {
		return ""
	}
	return s.identity.UserAgent()
}

type AccountSessionProvider interface {
	LoadSession(ctx context.Context, account AccountRef) (*AccountSession, error)
}

type credentialSessionProvider struct {
	source CredentialSource
}

func NewCredentialSessionProvider(source CredentialSource) (AccountSessionProvider, error) {
	if source == nil {
		return nil, errors.New("chatgptweb: credential source is required")
	}
	return &credentialSessionProvider{source: source}, nil
}

func (p *credentialSessionProvider) LoadSession(ctx context.Context, account AccountRef) (*AccountSession, error) {
	if account.ID <= 0 {
		return nil, errors.New("chatgptweb: account id must be positive")
	}
	record, err := p.source.LoadCredential(ctx, account.ID)
	if err != nil {
		return nil, fmt.Errorf("chatgptweb: load account credentials: %w", err)
	}
	if record.accountID != account.ID {
		return nil, fmt.Errorf("%w: credential record account mismatch", ErrInvalidAccountCredentials)
	}
	if strings.ToLower(record.platform) != AccountPlatformChatGPTWeb {
		return nil, fmt.Errorf("%w: account platform must be %s", ErrInvalidAccountCredentials, AccountPlatformChatGPTWeb)
	}
	if strings.ToLower(record.accountType) != AccountTypeCookie {
		return nil, fmt.Errorf("%w: account type must be %s", ErrInvalidAccountCredentials, AccountTypeCookie)
	}

	accessToken, err := credentialString(record.credentials, "access_token", true)
	if err != nil {
		return nil, err
	}
	upstreamAccountID, err := credentialString(record.credentials, "account_id", false)
	if err != nil {
		return nil, err
	}
	deviceID, err := credentialString(record.credentials, "device_id", false)
	if err != nil {
		return nil, err
	}
	userAgent, err := credentialString(record.credentials, "user_agent", false)
	if err != nil {
		return nil, err
	}
	cookie, err := credentialString(record.credentials, "cookie", false)
	if err != nil {
		return nil, err
	}

	identity, err := NewIdentity(accessToken, upstreamAccountID, deviceID, userAgent, cookie)
	if err != nil {
		return nil, fmt.Errorf("%w: access_token is required", ErrInvalidAccountCredentials)
	}
	return &AccountSession{identity: identity}, nil
}

func credentialString(credentials map[string]any, key string, required bool) (string, error) {
	value, found := credentials[key]
	if !found || value == nil {
		if required {
			return "", fmt.Errorf("%w: %s is required", ErrInvalidAccountCredentials, key)
		}
		return "", nil
	}
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%w: %s must be a string", ErrInvalidAccountCredentials, key)
	}
	text = strings.TrimSpace(text)
	if required && text == "" {
		return "", fmt.Errorf("%w: %s is required", ErrInvalidAccountCredentials, key)
	}
	return text, nil
}
