from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SERVICE = ROOT / "backend" / "internal" / "service"


def replace_once(text: str, old: str, new: str, label: str) -> str:
    count = text.count(old)
    if count != 1:
        raise RuntimeError(f"{label}: expected exactly one match, got {count}")
    return text.replace(old, new, 1)


# 1. Add a preparation hook for provider-specific credential/token resolution
# that must happen before endpoint construction.
p = SERVICE / "provider_adapter.go"
text = p.read_text()
old = '''// ProviderRequestBuilder builds the provider-native upstream request.
type ProviderRequestBuilder interface {
\tBuildRequest(ctx context.Context, input ProviderRequestInput) (*http.Request, error)
}
'''
new = '''// ProviderRequestPreparer resolves provider/account state that must be known
// before endpoint construction. It may return a copied input with ephemeral
// auth material, but must not perform scheduling, retries or billing.
type ProviderRequestPreparer interface {
\tPrepareRequest(ctx context.Context, input ProviderRequestInput) (ProviderRequestInput, error)
}

// ProviderRequestBuilder builds the provider-native upstream request.
type ProviderRequestBuilder interface {
\tBuildRequest(ctx context.Context, input ProviderRequestInput) (*http.Request, error)
}
'''
text = replace_once(text, old, new, "provider request preparer interface")
p.write_text(text)


# 2. Expose the optional preparer capability.
p = SERVICE / "provider_adapter_capabilities.go"
text = p.read_text()
text = replace_once(
    text,
    '''const (
\tProviderCapabilityRequestBuilder       ProviderCapability = "request_builder"
''',
    '''const (
\tProviderCapabilityRequestPreparer      ProviderCapability = "request_preparer"
\tProviderCapabilityRequestBuilder       ProviderCapability = "request_builder"
''',
    "preparer capability constant",
)
text = replace_once(
    text,
    '''\tcapabilities := make([]ProviderCapability, 0, 9)
\tif _, ok := adapter.(ProviderRequestBuilder); ok {
''',
    '''\tcapabilities := make([]ProviderCapability, 0, 10)
\tif _, ok := adapter.(ProviderRequestPreparer); ok {
\t\tcapabilities = append(capabilities, ProviderCapabilityRequestPreparer)
\t}
\tif _, ok := adapter.(ProviderRequestBuilder); ok {
''',
    "preparer capability detection",
)
p.write_text(text)


# 3. Make Gemini resolve OAuth/service-account tokens before BuildRequest so
# project auto-detection can affect endpoint selection on the same attempt.
p = SERVICE / "gemini_provider_request.go"
text = p.read_text()
insert_before = '''func (a *geminiGatewayProviderAdapter) NormalizeError(resp *http.Response, body []byte) NormalizedProviderError {
\treturn normalizeGeminiProviderError(resp, body)
}

'''
preparer_impl = '''func (a *geminiGatewayProviderAdapter) PrepareRequest(ctx context.Context, input ProviderRequestInput) (ProviderRequestInput, error) {
\tif a == nil || a.service == nil {
\t\treturn input, errors.New("gemini provider adapter is not configured")
\t}
\taccount := input.Account
\tif account == nil {
\t\treturn input, errors.New("gemini account is required")
\t}
\tif !a.Supports(account) {
\t\treturn input, errors.New("account is not supported by gemini provider adapter")
\t}

\tswitch account.Type {
\tcase AccountTypeAPIKey:
\t\treturn input, nil
\tcase AccountTypeOAuth, AccountTypeServiceAccount:
\t\tif a.service.tokenProvider == nil {
\t\t\treturn input, errors.New("gemini token provider not configured")
\t\t}
\t\taccessToken, err := a.service.tokenProvider.GetAccessToken(ctx, account)
\t\tif err != nil {
\t\t\treturn input, err
\t\t}
\t\tinput.AuthToken = accessToken
\t\treturn input, nil
\tdefault:
\t\treturn input, fmt.Errorf("unsupported account type: %s", account.Type)
\t}
}

'''
text = replace_once(text, insert_before, preparer_impl + insert_before, "gemini preparer implementation")
old_auth = '''\tcase AccountTypeOAuth, AccountTypeServiceAccount:
\t\tif a.service.tokenProvider == nil {
\t\t\treturn errors.New("gemini token provider not configured")
\t\t}
\t\taccessToken, err := a.service.tokenProvider.GetAccessToken(ctx, account)
\t\tif err != nil {
\t\t\treturn err
\t\t}
\t\treq.Header.Set("Authorization", "Bearer "+accessToken)
'''
new_auth = '''\tcase AccountTypeOAuth, AccountTypeServiceAccount:
\t\taccessToken := input.AuthToken
\t\tif strings.TrimSpace(accessToken) == "" {
\t\t\tif a.service.tokenProvider == nil {
\t\t\t\treturn errors.New("gemini token provider not configured")
\t\t\t}
\t\t\tvar err error
\t\t\taccessToken, err = a.service.tokenProvider.GetAccessToken(ctx, account)
\t\t\tif err != nil {
\t\t\t\treturn err
\t\t\t}
\t\t}
\t\treq.Header.Set("Authorization", "Bearer "+accessToken)
'''
text = replace_once(text, old_auth, new_auth, "gemini auth uses prepared token")
old_factory = '''\treturn func(ctx context.Context) (*http.Request, string, error) {
\t\treq, err := builder.BuildRequest(ctx, input)
\t\tif err != nil {
\t\t\treturn nil, "", err
\t\t}
\t\tif err := auth.ApplyAuth(ctx, req, input); err != nil {
\t\t\treturn nil, "", err
\t\t}
\t\treturn req, geminiProviderRequestIDHeader, nil
\t}, geminiProviderRequestIDHeader, nil
'''
new_factory = '''\treturn func(ctx context.Context) (*http.Request, string, error) {
\t\tpreparedInput := input
\t\tif preparer, ok := adapter.(ProviderRequestPreparer); ok {
\t\t\tvar prepareErr error
\t\t\tpreparedInput, prepareErr = preparer.PrepareRequest(ctx, input)
\t\t\tif prepareErr != nil {
\t\t\t\treturn nil, "", prepareErr
\t\t\t}
\t\t}

\t\treq, err := builder.BuildRequest(ctx, preparedInput)
\t\tif err != nil {
\t\t\treturn nil, "", err
\t\t}
\t\tif err := auth.ApplyAuth(ctx, req, preparedInput); err != nil {
\t\t\treturn nil, "", err
\t\t}
\t\treturn req, geminiProviderRequestIDHeader, nil
\t}, geminiProviderRequestIDHeader, nil
'''
text = replace_once(text, old_factory, new_factory, "factory preparation order")
p.write_text(text)


# 4. Signature fallback mutates geminiReq; rebuild the request factory so the
# next retry actually sends the downgraded payload instead of the old slice.
p = SERVICE / "gemini_messages_compat_service.go"
text = p.read_text()
old = '''\t\t\t\tif txErr == nil {
\t\t\t\t\tlogger.LegacyPrintf("service.gemini_messages_compat", "Gemini account %d: detected signature-related 400, retrying with downgraded Claude blocks (%s)", account.ID, stageName)
\t\t\t\t\tgeminiReq = retryGeminiReq
\t\t\t\t\t// Consume one retry budget attempt and continue with the updated request payload.
\t\t\t\t\tsleepGeminiBackoff(1)
\t\t\t\t\tcontinue
\t\t\t\t}
'''
new = '''\t\t\t\tif txErr == nil {
\t\t\t\t\tlogger.LegacyPrintf("service.gemini_messages_compat", "Gemini account %d: detected signature-related 400, retrying with downgraded Claude blocks (%s)", account.ID, stageName)
\t\t\t\t\tgeminiReq = retryGeminiReq
\t\t\t\t\tbuildReq, requestIDHeader, buildFactoryErr = s.newGeminiProviderRequestFactory(ProviderRequestInput{
\t\t\t\t\t\tAccount:  account,
\t\t\t\t\t\tProtocol: ProviderProtocolAnthropic,
\t\t\t\t\t\tEndpoint: action,
\t\t\t\t\t\tModel:    mappedModel,
\t\t\t\t\t\tBody:     geminiReq,
\t\t\t\t\t\tStream:   useUpstreamStream,
\t\t\t\t\t})
\t\t\t\t\tif buildFactoryErr != nil {
\t\t\t\t\t\treturn nil, s.writeClaudeError(c, http.StatusBadGateway, "upstream_error", buildFactoryErr.Error())
\t\t\t\t\t}
\t\t\t\t\t// Consume one retry budget attempt and continue with the updated request payload.
\t\t\t\t\tsleepGeminiBackoff(1)
\t\t\t\t\tcontinue
\t\t\t\t}
'''
text = replace_once(text, old, new, "signature retry request factory rebuild")
p.write_text(text)


# 5. Extend focused tests with pre-build preparation semantics.
p = SERVICE / "gemini_provider_request_test.go"
text = p.read_text()
text = replace_once(
    text,
    '''\trequire.Contains(t, caps, ProviderCapabilityRequestBuilder)
\trequire.Contains(t, caps, ProviderCapabilityAuthApplier)
}
''',
    '''\trequire.Contains(t, caps, ProviderCapabilityRequestPreparer)
\trequire.Contains(t, caps, ProviderCapabilityRequestBuilder)
\trequire.Contains(t, caps, ProviderCapabilityAuthApplier)
}

func TestGeminiGatewayProviderAdapterPrepareRequestResolvesOAuthToken(t *testing.T) {
\tsvc := newGeminiAdapterTestService()
\tadapter := newGeminiGatewayProviderAdapter(svc).(*geminiGatewayProviderAdapter)
\taccount := &Account{
\t\tPlatform: PlatformGemini,
\t\tType:     AccountTypeOAuth,
\t\tCredentials: map[string]any{
\t\t\t"access_token": "ya29.prepared-token",
\t\t},
\t}
\tinput := ProviderRequestInput{
\t\tAccount:  account,
\t\tProtocol: ProviderProtocolGemini,
\t\tEndpoint: "generateContent",
\t\tModel:    "gemini-2.5-flash",
\t\tBody:     []byte(`{"contents":[]}`),
\t}

\tprepared, err := adapter.PrepareRequest(context.Background(), input)
\trequire.NoError(t, err)
\trequire.Equal(t, "ya29.prepared-token", prepared.AuthToken)
}
''',
    "gemini preparer tests",
)
p.write_text(text)

print("provider phase3 semantic hardening applied")
