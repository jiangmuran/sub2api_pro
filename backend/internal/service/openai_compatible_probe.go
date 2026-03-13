package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
}

type OpenAICompatibleProbeResult struct {
	NormalizedBaseURL string                            `json:"normalized_base_url"`
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

	responsesURL := buildOpenAIResponsesURL(baseURL)
	chatURL := buildOpenAIChatCompletionsURL(baseURL)

	checks := make([]OpenAICompatibleProbeCheck, 0, 3)
	capabilities := OpenAICompatibleProbeCapabilities{}

	responsesCheck := s.probeOpenAICompatibleJSON(ctx, openAIProbeRequest{
		URL:       responsesURL,
		APIKey:    apiKey,
		ProxyURL:  input.ProxyURL,
		UserAgent: input.UserAgent,
		Body: map[string]any{
			"model":             openai.DefaultTestModel,
			"input":             []map[string]any{{"role": "user", "content": []map[string]any{{"type": "input_text", "text": "ping"}}}},
			"max_output_tokens": 8,
		},
		CheckKey:   "responses",
		CheckLabel: "Responses",
	})
	checks = append(checks, responsesCheck)
	capabilities.Responses = responsesCheck.Status == "success" || responsesCheck.Status == "partial"

	responsesStreamCheck := s.probeOpenAICompatibleStream(ctx, openAIProbeRequest{
		URL:       responsesURL,
		APIKey:    apiKey,
		ProxyURL:  input.ProxyURL,
		UserAgent: input.UserAgent,
		Body: map[string]any{
			"model":             openai.DefaultTestModel,
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
		URL:       chatURL,
		APIKey:    apiKey,
		ProxyURL:  input.ProxyURL,
		UserAgent: input.UserAgent,
		Body: map[string]any{
			"model":      openai.DefaultTestModel,
			"messages":   []map[string]any{{"role": "user", "content": "ping"}},
			"max_tokens": 8,
		},
		CheckKey:   "chat_completions",
		CheckLabel: "Chat Completions",
	})
	checks = append(checks, chatCheck)
	capabilities.ChatCompletions = chatCheck.Status == "success" || chatCheck.Status == "partial"

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
	}
	if len(suggestedExtra) > 0 {
		suggestedExtra["openai_compat_capabilities"] = map[string]any{
			"responses":        capabilities.Responses,
			"responses_stream": capabilities.ResponsesStream,
			"chat_completions": capabilities.ChatCompletions,
		}
	}

	return &OpenAICompatibleProbeResult{
		NormalizedBaseURL: baseURL,
		Status:            status,
		RecommendedMode:   recommendedMode,
		Checks:            checks,
		Capabilities:      capabilities,
		SuggestedExtra:    suggestedExtra,
	}, nil
}

type openAIProbeRequest struct {
	URL        string
	APIKey     string
	ProxyURL   string
	UserAgent  string
	Body       map[string]any
	CheckKey   string
	CheckLabel string
}

func (s *AccountTestService) probeOpenAICompatibleJSON(ctx context.Context, probe openAIProbeRequest) OpenAICompatibleProbeCheck {
	bodyBytes, _ := json.Marshal(probe.Body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, probe.URL, bytes.NewReader(bodyBytes))
	if err != nil {
		return OpenAICompatibleProbeCheck{Key: probe.CheckKey, Label: probe.CheckLabel, Status: "failed", Message: err.Error(), EndpointURL: probe.URL}
	}
	applyOpenAIProbeHeaders(req, probe.APIKey, probe.UserAgent)
	resp, err := s.httpUpstream.Do(req, probe.ProxyURL, 0, 1)
	if err != nil {
		return OpenAICompatibleProbeCheck{Key: probe.CheckKey, Label: probe.CheckLabel, Status: "failed", Message: err.Error(), EndpointURL: probe.URL}
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	return classifyOpenAICompatibleJSONCheck(probe, resp, body)
}

func (s *AccountTestService) probeOpenAICompatibleStream(ctx context.Context, probe openAIProbeRequest) OpenAICompatibleProbeCheck {
	bodyBytes, _ := json.Marshal(probe.Body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, probe.URL, bytes.NewReader(bodyBytes))
	if err != nil {
		return OpenAICompatibleProbeCheck{Key: probe.CheckKey, Label: probe.CheckLabel, Status: "failed", Message: err.Error(), EndpointURL: probe.URL}
	}
	applyOpenAIProbeHeaders(req, probe.APIKey, probe.UserAgent)
	req.Header.Set("Accept", "text/event-stream")
	resp, err := s.httpUpstream.Do(req, probe.ProxyURL, 0, 1)
	if err != nil {
		return OpenAICompatibleProbeCheck{Key: probe.CheckKey, Label: probe.CheckLabel, Status: "failed", Message: err.Error(), EndpointURL: probe.URL}
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))

	check := OpenAICompatibleProbeCheck{
		Key:         probe.CheckKey,
		Label:       probe.CheckLabel,
		HTTPStatus:  resp.StatusCode,
		EndpointURL: probe.URL,
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

func classifyOpenAICompatibleJSONCheck(probe openAIProbeRequest, resp *http.Response, body []byte) OpenAICompatibleProbeCheck {
	check := OpenAICompatibleProbeCheck{
		Key:         probe.CheckKey,
		Label:       probe.CheckLabel,
		HTTPStatus:  resp.StatusCode,
		EndpointURL: probe.URL,
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

func applyOpenAIProbeHeaders(req *http.Request, apiKey, userAgent string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if strings.TrimSpace(userAgent) != "" {
		req.Header.Set("User-Agent", strings.TrimSpace(userAgent))
	}
}

func buildOpenAIChatCompletionsURL(base string) string {
	normalized := strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(normalized, "/chat/completions") {
		return normalized
	}
	if strings.HasSuffix(normalized, "/v1") {
		return normalized + "/chat/completions"
	}
	return normalized + "/v1/chat/completions"
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
