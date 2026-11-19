package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCangyuanMediaPaidRequestQuantityMustMatchFrozenPrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		kind      CangyuanMediaKind
		quantity  int
		payload   string
		wantError bool
	}{
		{"default image one", CangyuanImageGeneration, 1, `{"model":"image","prompt":"cat"}`, false},
		{"default image two underquoted", CangyuanImageGeneration, 2, `{"model":"image","prompt":"cat"}`, true},
		{"two images quoted", CangyuanImageGeneration, 2, `{"model":"image","prompt":"cat","n":2}`, false},
		{"two images underquoted", CangyuanImageGeneration, 1, `{"model":"image","prompt":"cat","n":2}`, true},
		{"image edit mismatch", CangyuanImageEdit, 1, `{"model":"image","prompt":"cat","n":2}`, true},
		{"image count string", CangyuanImageGeneration, 1, `{"n":"1"}`, true},
		{"image count float", CangyuanImageGeneration, 1, `{"n":1.5}`, true},
		{"image count zero", CangyuanImageGeneration, 1, `{"n":0}`, true},
		{"image count negative", CangyuanImageGeneration, 1, `{"n":-1}`, true},
		{"video one output", CangyuanVideoGeneration, 1, `{"model":"vid","prompt":"scene","duration":5}`, false},
		{"video cannot specify image count", CangyuanVideoGeneration, 1, `{"n":1}`, true},
		{"video cannot sell multiple outputs as one job", CangyuanVideoGeneration, 2, `{}`, true},
		{"unrecognized capability", "music", 1, `{}`, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var request map[string]json.RawMessage
			require.NoError(t, json.Unmarshal([]byte(tc.payload), &request))
			err := validateCangyuanMediaRequestQuantity(tc.kind, request, &CangyuanMediaPriceSnapshot{Units: tc.quantity})
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
