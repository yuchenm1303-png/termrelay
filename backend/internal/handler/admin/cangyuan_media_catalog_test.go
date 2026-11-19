package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestParseCangyuanMediaCatalogOnlyShowsRecognizedPublicMedia(t *testing.T) {
	input := `{"success":true,"pricing_version":"official-v1","data":[
		{"model_name":"image-model","quota_type":1,"billing_mode":"per_request","model_price":0.08,"request_unit":"image","supported_endpoint_types":["openai-image"]},
		{"model_name":"video-model","quota_type":1,"billing_mode":"tiered_expr","model_price":0.12,"supported_endpoint_types":["openai-video"]},
		{"model_name":"chat-model","quota_type":0,"billing_mode":"tiered_expr","model_price":12,"supported_endpoint_types":["openai"]},
		{"model_name":"unpriced-image","quota_type":1,"billing_mode":"per_request","supported_endpoint_types":["openai-image"]},
		{"model_name":"secret","quota_type":1,"billing_mode":"per_request","model_price":1,"supported_endpoint_types":["openai-chat"]},
		{"model_name":"image-model","quota_type":1,"billing_mode":"per_request","model_price":0.08,"supported_endpoint_types":["openai-image"]}
	]}`
	result, err := parseCangyuanMediaCatalog([]byte(input))
	require.NoError(t, err)
	require.Equal(t, "official-v1", result.PricingVersion)
	require.Len(t, result.Items, 2)
	require.Equal(t, "image-model", result.Items[0].ID)
	require.Equal(t, "image_generation", result.Items[0].Kind)
	require.Equal(t, "0.08", string(result.Items[0].ModelPrice))
	require.Equal(t, "video_generation", result.Items[1].Kind)
	require.True(t, result.Items[1].Tiered)
}

func TestParseCangyuanMediaCatalogFailsClosed(t *testing.T) {
	for _, input := range []string{
		`{}`,
		`{"success":true,"data":[]}`,
		`{"success":true,"pricing_version":"v1","data":[{"model_name":"image-model","quota_type":0,"billing_mode":"per_request","model_price":1,"supported_endpoint_types":["openai-image"]}]}`,
		`{"success":false,"pricing_version":"v1","data":[{"model_name":"image-model"}]}`,
	} {
		_, err := parseCangyuanMediaCatalog([]byte(input))
		require.Error(t, err)
	}
}

func TestCangyuanMediaCatalogHandlerNotConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/admin/cangyuan-media/catalog", (&CangyuanMediaPriceHandler{}).Catalog)
	req := httptest.NewRequest(http.MethodGet, "/admin/cangyuan-media/catalog", strings.NewReader(""))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusServiceUnavailable, w.Code)
}
