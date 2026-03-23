package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/tidwall/gjson"
)

const (
	OpenAICompatibleProbeStatusCompatible   = "compatible"
	OpenAICompatibleProbeStatusPartial      = "partial"
	OpenAICompatibleProbeStatusLegacyOnly   = "legacy_only"
	OpenAICompatibleProbeStatusIncompatible = "incompatible"

	OpenAICompatibleModeResponsesNative         = "responses_native"
	OpenAICompatibleModeResponsesPassthrough    = "responses_passthrough"
	OpenAICompatibleModeChatCompletionsFallback = "chat_completions_fallback"
	OpenAICompatibleModeCompletionsFallback     = "completions_fallback"
	OpenAICompatibleModeUnsupported             = "unsupported"
)

type OpenAICompatibleProbeInput struct {
	BaseURL   string
	APIKey    string
	ProxyURL  string
	UserAgent string
}

type OpenAICompatibleProbeCheck struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Status      string `json:"status"`
	HTTPStatus  int    `json:"http_status,omitempty"`
	Message     string `json:"message,omitempty"`
	EndpointURL string `json:"endpoint_url,omitempty"`
}

type OpenAICompatibleProbeCapabilities struct {
	Responses       bool `json:"responses"`
	ResponsesStream bool `json:"responses_stream"`
	ChatCompletions bool `json:"chat_completions"`
	Completions     bool `json:"completions"`
}

type OpenAICompatibleProbeResult struct {
	NormalizedBaseURL string                            `json:"normalized_base_url"`
	ProbeModel        string                            `json:"probe_model,omitempty"`
	DiscoveredModels  []string                          `json:"discovered_models,omitempty"`
	Status            string                            `json:"status"`
	RecommendedMode   string                            `json:"recommended_mode"`
	Checks            []OpenAICompatibleProbeCheck      `json:"checks"`
	Capabilities      OpenAICompatibleProbeCapabilities `json:"capabilities"`
	SuggestedExtra    map[string]any                    `json:"suggested_extra,omitempty"`
}

func (s *AccountTestService) ProbeOpenAICompatible(ctx context.Context, input OpenAICompatibleProbeInput) (*OpenAICompatibleProbeResult, error) {
	baseURL, err := s.validateUpstreamBaseURL(strings.TrimSpace(input.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("invalid base_url: %w", err)
	}
	apiKey := strings.TrimSpace(input.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	responsesURLs := buildOpenAIProbeEndpointCandidates(baseURL, "responses")
	chatURLs := buildOpenAIProbeEndpointCandidates(baseURL, "chat/completions")
	completionURLs := buildOpenAIProbeEndpointCandidates(baseURL, "completions")

	checks := make([]OpenAICompatibleProbeCheck, 0, 3)
	capabilities := OpenAICompatibleProbeCapabilities{}

	discoveredModels := s.discoverOpenAICompatibleProbeModels(ctx, baseURL, apiKey, input.ProxyURL, input.UserAgent)
	probeModel := openai.DefaultTestModel
	if len(discoveredModels) > 0 {
		probeModel = discoveredModels[0]
	}

	responsesCheck := s.probeOpenAICompatibleJSON(ctx, openAIProbeRequest{
		URLs:      responsesURLs,
		APIKey:    apiKey,
		ProxyURL:  input.ProxyURL,
		UserAgent: input.UserAgent,
		Body: map[string]any{
			"model":             probeModel,
			"input":             []map[string]any{{"role": "user", "content": []map[string]any{{"type": "input_text", "text": "ping"}}}},
			"max_output_tokens": 8,
		},
		CheckKey:   "responses",
		CheckLabel: "Responses",
	})
	checks = append(checks, responsesCheck)
	capabilities.Responses = responsesCheck.Status == "success" || responsesCheck.Status == "partial"

	responsesStreamCheck := s.probeOpenAICompatibleStream(ctx, openAIProbeRequest{
		URLs:      responsesURLs,
		APIKey:    apiKey,
		ProxyURL:  input.ProxyURL,
		UserAgent: input.UserAgent,
		Body: map[string]any{
			"model":             probeModel,
			"input":             []map[string]any{{"role": "user", "content": []map[string]any{{"type": "input_text", "text": "ping"}}}},
			"max_output_tokens": 8,
			"stream":            true,
		},
		CheckKey:   "responses_stream",
		CheckLabel: "Responses Stream",
	})
	checks = append(checks, responsesStreamCheck)
	capabilities.ResponsesStream = responsesStreamCheck.Status == "success"

	chatCheck := s.probeOpenAICompatibleJSON(ctx, openAIProbeRequest{
		URLs:      chatURLs,
		APIKey:    apiKey,
		ProxyURL:  input.ProxyURL,
		UserAgent: input.UserAgent,
		Body: map[string]any{
			"model":      probeModel,
			"messages":   []map[string]any{{"role": "user", "content": "ping"}},
			"max_tokens": 8,
		},
		CheckKey:   "chat_completions",
		CheckLabel: "Chat Completions",
	})
	checks = append(checks, chatCheck)
	capabilities.ChatCompletions = chatCheck.Status == "success" || chatCheck.Status == "partial"

	completionCheck := s.probeOpenAICompatibleJSON(ctx, openAIProbeRequest{
		URLs:      completionURLs,
		APIKey:    apiKey,
		ProxyURL:  input.ProxyURL,
		UserAgent: input.UserAgent,
		Body: map[string]any{
			"model":      probeModel,
			"prompt":     "ping",
			"max_tokens": 8,
		},
		CheckKey:   "completions",
		CheckLabel: "Completions",
	})
	checks = append(checks, completionCheck)
	capabilities.Completions = completionCheck.Status == "success" || completionCheck.Status == "partial"

	status := OpenAICompatibleProbeStatusIncompatible
	recommendedMode := OpenAICompatibleModeUnsupported
	suggestedExtra := map[string]any{}

	switch {
	case capabilities.Responses && capabilities.ResponsesStream:
		status = OpenAICompatibleProbeStatusCompatible
		recommendedMode = OpenAICompatibleModeResponsesNative
		suggestedExtra["openai_compat_mode"] = recommendedMode
	case capabilities.Responses:
		status = OpenAICompatibleProbeStatusPartial
		recommendedMode = OpenAICompatibleModeResponsesPassthrough
		suggestedExtra["openai_compat_mode"] = recommendedMode
		suggestedExtra["openai_passthrough"] = true
	case capabilities.ChatCompletions:
		status = OpenAICompatibleProbeStatusLegacyOnly
		recommendedMode = OpenAICompatibleModeChatCompletionsFallback
		suggestedExtra["openai_compat_mode"] = recommendedMode
		suggestedExtra["openai_passthrough"] = true
	case capabilities.Completions:
		status = OpenAICompatibleProbeStatusLegacyOnly
		recommendedMode = OpenAICompatibleModeCompletionsFallback
		suggestedExtra["openai_compat_mode"] = recommendedMode
		suggestedExtra["openai_passthrough"] = true
	}
	if len(suggestedExtra) > 0 {
		suggestedExtra["openai_compat_capabilities"] = map[string]any{
			"responses":        capabilities.Responses,
			"responses_stream": capabilities.ResponsesStream,
			"chat_completions": capabilities.ChatCompletions,
			"completions":      capabilities.Completions,
		}
		if len(discoveredModels) > 0 {
			suggestedExtra["openai_compat_models"] = cloneStringSlice(discoveredModels)
		}
	}

	return &OpenAICompatibleProbeResult{
		NormalizedBaseURL: baseURL,
		ProbeModel:        probeModel,
		DiscoveredModels:  cloneStringSlice(discoveredModels),
		Status:            status,
		RecommendedMode:   recommendedMode,
		Checks:            checks,
		Capabilities:      capabilities,
		SuggestedExtra:    suggestedExtra,
	}, nil
}

type openAICompatibleModelInfo struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type openAIProbeRequest struct {
	URLs       []string
	APIKey     string
	ProxyURL   string
	UserAgent  string
	Body       map[string]any
	CheckKey   string
	CheckLabel string
}

func (s *AccountTestService) probeOpenAICompatibleJSON(ctx context.Context, probe openAIProbeRequest) OpenAICompatibleProbeCheck {
	if len(probe.URLs) == 0 {
		return OpenAICompatibleProbeCheck{Key: probe.CheckKey, Label: probe.CheckLabel, Status: "failed", Message: "no probe URLs configured"}
	}
	bodyBytes, _ := json.Marshal(probe.Body)
	var lastCheck OpenAICompatibleProbeCheck
	for _, targetURL := range probe.URLs {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(bodyBytes))
		if err != nil {
			lastCheck = OpenAICompatibleProbeCheck{Key: probe.CheckKey, Label: probe.CheckLabel, Status: "failed", Message: err.Error(), EndpointURL: targetURL}
			continue
		}
		applyOpenAIProbeHeaders(req, probe.APIKey, probe.UserAgent)
		resp, err := s.httpUpstream.Do(req, probe.ProxyURL, 0, 1)
		if err != nil {
			lastCheck = OpenAICompatibleProbeCheck{Key: probe.CheckKey, Label: probe.CheckLabel, Status: "failed", Message: err.Error(), EndpointURL: targetURL}
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		_ = resp.Body.Close()
		check := classifyOpenAICompatibleJSONCheck(probe, targetURL, resp, body)
		if check.Status != "unsupported" {
			return check
		}
		lastCheck = check
	}
	return lastCheck
}

func (s *AccountTestService) probeOpenAICompatibleStream(ctx context.Context, probe openAIProbeRequest) OpenAICompatibleProbeCheck {
	if len(probe.URLs) == 0 {
		return OpenAICompatibleProbeCheck{Key: probe.CheckKey, Label: probe.CheckLabel, Status: "failed", Message: "no probe URLs configured"}
	}
	bodyBytes, _ := json.Marshal(probe.Body)
	var lastCheck OpenAICompatibleProbeCheck
	for _, targetURL := range probe.URLs {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(bodyBytes))
		if err != nil {
			lastCheck = OpenAICompatibleProbeCheck{Key: probe.CheckKey, Label: probe.CheckLabel, Status: "failed", Message: err.Error(), EndpointURL: targetURL}
			continue
		}
		applyOpenAIProbeHeaders(req, probe.APIKey, probe.UserAgent)
		req.Header.Set("Accept", "text/event-stream")
		resp, err := s.httpUpstream.Do(req, probe.ProxyURL, 0, 1)
		if err != nil {
			lastCheck = OpenAICompatibleProbeCheck{Key: probe.CheckKey, Label: probe.CheckLabel, Status: "failed", Message: err.Error(), EndpointURL: targetURL}
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		_ = resp.Body.Close()

		check := classifyOpenAICompatibleStreamCheck(probe, targetURL, resp, body)
		if check.Status != "unsupported" {
			return check
		}
		lastCheck = check
	}
	return lastCheck
}

func classifyOpenAICompatibleStreamCheck(probe openAIProbeRequest, targetURL string, resp *http.Response, body []byte) OpenAICompatibleProbeCheck {
	check := OpenAICompatibleProbeCheck{
		Key:         probe.CheckKey,
		Label:       probe.CheckLabel,
		HTTPStatus:  resp.StatusCode,
		EndpointURL: targetURL,
	}
	if resp.StatusCode == http.StatusOK {
		contentType := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type")))
		bodyText := string(body)
		if strings.Contains(contentType, "text/event-stream") && (strings.Contains(bodyText, "data:") || strings.Contains(bodyText, "[DONE]")) {
			check.Status = "success"
			check.Message = "SSE stream looks compatible"
			return check
		}
		check.Status = "failed"
		check.Message = "stream response is not valid SSE"
		return check
	}
	if endpointReachableDespiteModelError(resp.StatusCode, body) {
		check.Status = "partial"
		check.Message = normalizeOpenAICompatibleErrorMessage(body)
		if check.Message == "" {
			check.Message = "endpoint reachable but stream could not be fully verified"
		}
		return check
	}
	if isLikelyUnsupportedEndpoint(resp.StatusCode, body) {
		check.Status = "unsupported"
		check.Message = normalizeOpenAICompatibleErrorMessage(body)
		if check.Message == "" {
			check.Message = "endpoint not supported"
		}
		return check
	}
	check.Status = "failed"
	check.Message = normalizeOpenAICompatibleErrorMessage(body)
	if check.Message == "" {
		check.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	return check
}

func classifyOpenAICompatibleJSONCheck(probe openAIProbeRequest, targetURL string, resp *http.Response, body []byte) OpenAICompatibleProbeCheck {
	check := OpenAICompatibleProbeCheck{
		Key:         probe.CheckKey,
		Label:       probe.CheckLabel,
		HTTPStatus:  resp.StatusCode,
		EndpointURL: targetURL,
	}
	if resp.StatusCode == http.StatusOK {
		check.Status = "success"
		check.Message = "endpoint responded successfully"
		return check
	}
	if endpointReachableDespiteModelError(resp.StatusCode, body) {
		check.Status = "partial"
		check.Message = normalizeOpenAICompatibleErrorMessage(body)
		if check.Message == "" {
			check.Message = "endpoint reachable but rejected the probe model"
		}
		return check
	}
	if isLikelyUnsupportedEndpoint(resp.StatusCode, body) {
		check.Status = "unsupported"
		check.Message = normalizeOpenAICompatibleErrorMessage(body)
		if check.Message == "" {
			check.Message = "endpoint not supported"
		}
		return check
	}
	check.Status = "failed"
	check.Message = normalizeOpenAICompatibleErrorMessage(body)
	if check.Message == "" {
		check.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	return check
}

func buildOpenAIProbeEndpointCandidates(base, endpoint string) []string {
	normalizedBase := strings.TrimRight(strings.TrimSpace(base), "/")
	endpoint = strings.TrimLeft(strings.TrimSpace(endpoint), "/")
	if normalizedBase == "" || endpoint == "" {
		return nil
	}

	candidates := make([]string, 0, 3)
	seen := make(map[string]struct{}, 3)
	appendCandidate := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		candidates = append(candidates, value)
	}

	appendCandidate(normalizedBase + "/" + endpoint)
	if !strings.HasSuffix(normalizedBase, "/v1") && !strings.HasSuffix(normalizedBase, "/"+endpoint) {
		appendCandidate(normalizedBase + "/v1/" + endpoint)
	}
	return candidates
}

func (s *AccountTestService) discoverOpenAICompatibleProbeModels(ctx context.Context, baseURL, apiKey, proxyURL, userAgent string) []string {
	modelURLs := buildOpenAIProbeEndpointCandidates(baseURL, "models")
	for _, targetURL := range modelURLs {
		modelIDs := s.fetchOpenAICompatibleModels(ctx, targetURL, apiKey, proxyURL, userAgent)
		if len(modelIDs) == 0 {
			continue
		}
		return modelIDs
	}
	return nil
}

func (s *AccountTestService) DiscoverOpenAICompatibleModels(ctx context.Context, baseURL, apiKey, proxyURL, userAgent string) []string {
	validatedBaseURL, err := s.validateUpstreamBaseURL(strings.TrimSpace(baseURL))
	if err != nil {
		return nil
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil
	}
	return s.discoverOpenAICompatibleProbeModels(ctx, validatedBaseURL, apiKey, proxyURL, userAgent)
}

func (s *AccountTestService) fetchOpenAICompatibleModels(ctx context.Context, targetURL, apiKey, proxyURL, userAgent string) []string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil
	}
	applyOpenAIProbeHeaders(req, apiKey, userAgent)
	resp, err := s.httpUpstream.Do(req, proxyURL, 0, 1)
	if err != nil {
		return nil
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512<<10))
	if err != nil || !gjson.ValidBytes(body) {
		return nil
	}
	models := gjson.GetBytes(body, "data")
	if !models.Exists() || !models.IsArray() {
		return nil
	}

	preferred := make([]string, 0, len(models.Array()))
	fallback := make([]string, 0, len(models.Array()))
	for _, item := range models.Array() {
		modelID := strings.TrimSpace(item.Get("id").String())
		if modelID == "" {
			continue
		}
		status := strings.ToLower(strings.TrimSpace(item.Get("status").String()))
		if status == "shutdown" || status == "retiring" || status == "deprecated" {
			continue
		}
		if isPreferredOpenAICompatibleProbeModel(modelID) {
			preferred = append(preferred, modelID)
			continue
		}
		fallback = append(fallback, modelID)
	}
	sort.Strings(preferred)
	sort.Strings(fallback)
	return append(preferred, fallback...)
}

func isPreferredOpenAICompatibleProbeModel(modelID string) bool {
	normalized := strings.ToLower(strings.TrimSpace(modelID))
	preferredPrefixes := []string{"kimi-", "glm-", "deepseek-", "minimax-"}
	for _, prefix := range preferredPrefixes {
		if strings.HasPrefix(normalized, prefix) {
			return true
		}
	}
	return false
}

func applyOpenAIProbeHeaders(req *http.Request, apiKey, userAgent string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if strings.TrimSpace(userAgent) != "" {
		req.Header.Set("User-Agent", strings.TrimSpace(userAgent))
	}
}

func buildOpenAIChatCompletionsURL(base string) string {
	return buildOpenAIEndpointURL(base, "chat/completions")
}

func buildOpenAICompletionsURL(base string) string {
	return buildOpenAIEndpointURL(base, "completions")
}

func buildOpenAIEndpointURL(base, endpoint string) string {
	normalized := strings.TrimRight(strings.TrimSpace(base), "/")
	endpoint = strings.TrimLeft(strings.TrimSpace(endpoint), "/")
	if normalized == "" || endpoint == "" {
		return normalized
	}
	if strings.HasSuffix(normalized, "/"+endpoint) {
		return normalized
	}
	if strings.HasPrefix(endpoint, "v1/") && strings.HasSuffix(normalized, "/v1") {
		return normalized + "/" + strings.TrimPrefix(endpoint, "v1/")
	}
	if strings.HasPrefix(endpoint, "v1/") {
		return normalized + "/" + endpoint
	}
	if strings.HasSuffix(normalized, "/v1") {
		return normalized + "/" + endpoint
	}
	parsed, err := url.Parse(normalized)
	if err == nil {
		path := strings.TrimSpace(parsed.Path)
		if path != "" && path != "/" {
			return normalized + "/" + endpoint
		}
	}
	return normalized + "/v1/" + endpoint
}

func endpointReachableDespiteModelError(statusCode int, body []byte) bool {
	if statusCode != http.StatusBadRequest && statusCode != http.StatusNotFound && statusCode != http.StatusUnprocessableEntity {
		return false
	}
	message := strings.ToLower(normalizeOpenAICompatibleErrorMessage(body))
	if message == "" {
		return false
	}
	if !strings.Contains(message, "model") {
		return false
	}
	keywords := []string{"does not exist", "unknown model", "unsupported model", "invalid model", "not found"}
	for _, keyword := range keywords {
		if strings.Contains(message, keyword) {
			return true
		}
	}
	return false
}

func isLikelyUnsupportedEndpoint(statusCode int, body []byte) bool {
	if statusCode == http.StatusMethodNotAllowed || statusCode == http.StatusNotImplemented {
		return true
	}
	if statusCode != http.StatusNotFound {
		return false
	}
	message := strings.ToLower(normalizeOpenAICompatibleErrorMessage(body))
	if message == "" {
		return true
	}
	return !strings.Contains(message, "model")
}

func normalizeOpenAICompatibleErrorMessage(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	if !gjson.ValidBytes(body) {
		return strings.TrimSpace(string(body))
	}
	for _, path := range []string{"error.message", "message", "detail", "error", "msg"} {
		if value := strings.TrimSpace(gjson.GetBytes(body, path).String()); value != "" {
			return value
		}
	}
	return strings.TrimSpace(string(body))
}
