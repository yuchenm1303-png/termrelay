package chatgptweb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestCredentialSessionProviderBuildsRedactedAccountSession(t *testing.T) {
	record := NewCredentialRecord(7, AccountPlatformChatGPTWeb, AccountTypeCookie, map[string]any{
		"access_token": "secret-access-token",
		"cookie":       "session=secret-cookie",
		"account_id":   "upstream-account",
		"device_id":    "browser-device",
		"user_agent":   "browser-ua",
	})
	provider, err := NewCredentialSessionProvider(CredentialSourceFunc(func(_ context.Context, accountID int64) (CredentialRecord, error) {
		if accountID != 7 {
			t.Fatalf("account id = %d", accountID)
		}
		return record, nil
	}))
	if err != nil {
		t.Fatal(err)
	}

	session, err := provider.LoadSession(context.Background(), AccountRef{ID: 7})
	if err != nil {
		t.Fatal(err)
	}
	if session.DeviceID() != "browser-device" || session.UserAgent() != "browser-ua" {
		t.Fatalf("session fingerprint = %q %q", session.DeviceID(), session.UserAgent())
	}

	rendered := fmt.Sprintf("%+v %#v %+v %#v", record, record, session, session)
	encodedRecord, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	encodedSession, err := json.Marshal(session)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"secret-access-token", "secret-cookie"} {
		if strings.Contains(rendered, secret) || strings.Contains(string(encodedRecord), secret) || strings.Contains(string(encodedSession), secret) {
			t.Fatalf("credential material leaked: %s", secret)
		}
	}
}

func TestCredentialSessionProviderRejectsInvalidRecordsWithoutLeakingValues(t *testing.T) {
	tests := []struct {
		name   string
		record CredentialRecord
	}{
		{
			name: "wrong platform",
			record: NewCredentialRecord(7, "openai", AccountTypeCookie, map[string]any{
				"access_token": "do-not-leak-platform-token",
			}),
		},
		{
			name: "wrong type",
			record: NewCredentialRecord(7, AccountPlatformChatGPTWeb, "oauth", map[string]any{
				"access_token": "do-not-leak-type-token",
			}),
		},
		{
			name:   "missing token",
			record: NewCredentialRecord(7, AccountPlatformChatGPTWeb, AccountTypeCookie, map[string]any{"cookie": "do-not-leak-cookie"}),
		},
		{
			name: "wrong token type",
			record: NewCredentialRecord(7, AccountPlatformChatGPTWeb, AccountTypeCookie, map[string]any{
				"access_token": 123,
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewCredentialSessionProvider(CredentialSourceFunc(func(context.Context, int64) (CredentialRecord, error) {
				return tt.record, nil
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = provider.LoadSession(context.Background(), AccountRef{ID: 7})
			if !errors.Is(err, ErrInvalidAccountCredentials) {
				t.Fatalf("err = %v", err)
			}
			for _, secret := range []string{"do-not-leak-platform-token", "do-not-leak-type-token", "do-not-leak-cookie"} {
				if strings.Contains(err.Error(), secret) {
					t.Fatalf("error leaked credential: %v", err)
				}
			}
		})
	}
}
