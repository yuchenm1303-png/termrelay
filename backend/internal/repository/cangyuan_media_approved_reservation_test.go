package repository

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type approvedPriceFixture struct {
	rules map[string]string
}
func (s *approvedPriceFixture) Set(_ context.Context, key, value string) error {
	s.rules[key] = value
	return nil
}
func (s *approvedPriceFixture) GetAll(_ context.Context) (map[string]string, error) {
	out := map[string]string{}
	for k, v := range s.rules { out[k] = v }
	return out, nil
}
func (s *approvedPriceFixture) Delete(_ context.Context, key string) error {
	delete(s.rules, key)
	return nil
}

func TestCangyuanMediaReserveUsesApprovedServerPriceAndModelWhitelist(t *testing.T) {
	ctx := context.Background()
	jobs, mr := newCangyuanMediaJobTestService(t)
	prices := service.NewCangyuanMediaPriceRegistry(&approvedPriceFixture{rules: map[string]string{}})
	rule := service.CangyuanMediaPriceRule{
		ID: "video-720-5s", Model: "video-model", Kind: service.CangyuanVideoGeneration,
		Options: map[string]json.RawMessage{
			"duration": json.RawMessage("5"),
			"resolution": json.RawMessage(`"720p"`),
		},
		Snapshot: service.CangyuanMediaPriceSnapshot{
			PricingVersion: "reviewed-v1", Currency: "USD",
			Tier: "720p-5s", Units: 1,
			UnitSalePriceUSD: "2.40", UnitProviderCostUSD: "1.20",
		},
		Enabled: true,
	}
	require.NoError(t, prices.Put(ctx, rule))
	account := &service.Account{
		ID: 321, Status: service.StatusActive,
		Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key": "fixture-secret", "base_url": "https://ai.cangyuansuanli.cn",
			"model_mapping": map[string]any{},
		},
		Extra: map[string]any{
			"upstream_vendor": "cangyuan",
			"cangyuan_media_only": true,
			"cangyuan_media_models": map[string]any{"image": []any{}, "video": []any{"video-model"}},
			"openai_passthrough": false,
			"openai_responses_mode": "force_chat_completions",
		},
	}
	owner := service.CangyuanMediaJobOwner{UserID: 4, APIKeyID: 12}
	request := []byte(`{"model":"video-model","prompt":"sunrise","duration":5,"resolution":"720p"}`)
	job, err := jobs.ReserveWithApprovedPricing(ctx, owner, account, service.CangyuanVideoGeneration, "video-model", request, "client-request-key", prices)
	require.NoError(t, err)
	require.NotNil(t, job.PriceSnapshot)
	require.True(t, job.PriceSnapshot.VerifiedByAdmin)
	require.Equal(t, "2.40", job.PriceSnapshot.UnitSalePriceUSD)
	require.Equal(t, int64(321), job.AccountID)
	require.Equal(t, "pending", job.BillingState)
	keysBefore := len(mr.Keys())

	again, err := jobs.ReserveWithApprovedPricing(ctx, owner, account, service.CangyuanVideoGeneration, "video-model", request, "client-request-key", prices)
	require.NoError(t, err)
	require.Equal(t, job.ID, again.ID)
	require.Len(t, mr.Keys(), keysBefore)

	for name, changed := range map[string][]byte{
		"costly high resolution": []byte(`{"model":"video-model","prompt":"sunrise","duration":5,"resolution":"1080p"}`),
		"longer video": []byte(`{"model":"video-model","prompt":"sunrise","duration":10,"resolution":"720p"}`),
		"missing dimension": []byte(`{"model":"video-model","prompt":"sunrise","duration":5}`),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := jobs.ReserveWithApprovedPricing(ctx, owner, account, service.CangyuanVideoGeneration, "video-model", changed, "new-key-"+name, prices)
			require.Error(t, err)
			require.Len(t, mr.Keys(), keysBefore, "unpriced requests must not create jobs")
		})
	}
	for name, mutate := range map[string]func(*service.Account){
		"wrong API key": func(a *service.Account){ a.Credentials["api_key"] = "" },
		"not active": func(a *service.Account){ a.Status = "disabled" },
		"missing model grant": func(a *service.Account){ a.Extra["cangyuan_media_models"] = map[string]any{"video":[]any{"different"}} },
		"not media-only": func(a *service.Account){ a.Extra["cangyuan_media_only"] = false },
	} {
		t.Run(name, func(t *testing.T) {
			cloned := *account
			cloned.Credentials = map[string]any{}
			for k,v := range account.Credentials { cloned.Credentials[k] = v }
			cloned.Extra = map[string]any{}
			for k,v := range account.Extra { cloned.Extra[k] = v }
			mutate(&cloned)
			_, err := jobs.ReserveWithApprovedPricing(ctx, owner, &cloned, service.CangyuanVideoGeneration, "video-model", request, "other-key-"+name, prices)
			require.Error(t, err)
		})
	}
	rule.Enabled = false
	require.NoError(t, prices.Put(ctx, rule))
	_, err = jobs.ReserveWithApprovedPricing(ctx, owner, account, service.CangyuanVideoGeneration, "video-model", request, "new-key-after-disable", prices)
	require.ErrorIs(t, err, service.ErrCangyuanMediaPriceNotFound)
}
