package repository

import (
	"context"
	"errors"
	"fmt"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbaccount "github.com/Wei-Shaw/sub2api/ent/account"
	"github.com/Wei-Shaw/sub2api/internal/service/chatgptweb"
)

type chatGPTWebCredentialSource struct {
	client *dbent.Client
}

var _ chatgptweb.CredentialSource = (*chatGPTWebCredentialSource)(nil)

// NewChatGPTWebCredentialSource reuses the application's existing Ent client and
// accounts.credentials JSONB storage. It owns persistence mapping only; parsing
// and validation of ChatGPT Web credential semantics stay in the service layer.
func NewChatGPTWebCredentialSource(client *dbent.Client) chatgptweb.CredentialSource {
	return &chatGPTWebCredentialSource{client: client}
}

func (s *chatGPTWebCredentialSource) LoadCredential(ctx context.Context, accountID int64) (chatgptweb.CredentialRecord, error) {
	if accountID <= 0 {
		return chatgptweb.CredentialRecord{}, errors.New("chatgptweb credential source: account id must be positive")
	}
	if s == nil || s.client == nil {
		return chatgptweb.CredentialRecord{}, errors.New("chatgptweb credential source: ent client is required")
	}
	account, err := s.client.Account.Query().Where(dbaccount.IDEQ(accountID)).Only(ctx)
	if err != nil {
		return chatgptweb.CredentialRecord{}, fmt.Errorf("chatgptweb credential source: load account %d: %w", accountID, err)
	}
	credentials := make(map[string]any, len(account.Credentials)+2)
	for key, value := range account.Credentials {
		credentials[key] = value
	}
	if _, exists := credentials["device_id"]; !exists {
		if value, ok := account.Extra["openai_device_id"]; ok {
			credentials["device_id"] = value
		}
	}
	if _, exists := credentials["user_agent"]; !exists {
		if value, ok := account.Extra["openai_user_agent"]; ok {
			credentials["user_agent"] = value
		}
	}
	return chatgptweb.NewCredentialRecord(account.ID, account.Platform, account.Type, credentials), nil
}
