package admin

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateAccountRequestHonorsSkipDefaultGroupBind(t *testing.T) {
	var req CreateAccountRequest
	err := json.Unmarshal([]byte(`{
		"name": "Cangyuan",
		"platform": "openai",
		"type": "apikey",
		"credentials": {"base_url": "https://ai.cangyuansuanli.cn", "api_key": "test-only"},
		"extra": {"openai_passthrough": false, "openai_responses_mode": "force_chat_completions"},
		"group_ids": [],
		"skip_default_group_bind": true
	}`), &req)
	require.NoError(t, err)
	require.True(t, req.SkipDefaultGroupBind)
	require.Empty(t, req.GroupIDs)
}
