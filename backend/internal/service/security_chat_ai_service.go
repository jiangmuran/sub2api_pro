package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type SecurityChatAIService struct {
	settings     *SettingService
	apiKeyService *APIKeyService
	cfg          *config.Config
}

func NewSecurityChatAIService(settings *SettingService, apiKeyService *APIKeyService, cfg *config.Config) *SecurityChatAIService {
	return &SecurityChatAIService{
		settings:     settings,
		apiKeyService: apiKeyService,
		cfg:          cfg,
	}
}

type SecurityChatSummaryInput struct {
	Messages []SecurityChatMessage
	Meta     SecurityChatSummaryMeta
}

type SecurityChatSummaryMeta struct {
	SessionCount int
	MessageCount int
	StartTime    time.Time
	EndTime      time.Time
	UserID       *int64
	APIKeyID     *int64
}

type SecurityChatSummaryResult struct {
	Summary            string   `json:"summary"`
	SensitiveFindings  []string `json:"sensitive_findings"`
	RiskLevel          string   `json:"risk_level"`
	RecommendedActions []string `json:"recommended_actions"`
}

func (s *SecurityChatAIService) Summarize(ctx context.Context, input *SecurityChatSummaryInput, userID int64, apiKeyID *int64) (*SecurityChatSummaryResult, error) {
	if s == nil || s.settings == nil || s.apiKeyService == nil {
		return nil, errors.New("security ai service unavailable")
	}
	settings, err := s.settings.GetAllSettings(ctx)
	if err != nil {
		return nil, err
	}
	if !settings.SecurityChatAIEnabled {
		return nil, errors.New("security chat AI disabled")
	}
	if userID <= 0 {
		return nil, errors.New("invalid user")
	}
	model := strings.TrimSpace(settings.SecurityChatAIModel)
	if model == "" {
		return nil, errors.New("security chat AI config invalid")
	}

	apiKey, err := s.getUserAPIKey(ctx, userID, apiKeyID)
	if err != nil {
		return nil, err
	}

	payload := buildSecurityChatSummaryPrompt(input)

	requestBody := map[string]any{
		"model": model,
		"messages": []map[string]any{
			{"role": "system", "content": payload.System},
			{"role": "user", "content": payload.User},
		},
		"temperature": 0.2,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	endpoint, err := s.resolveGatewayEndpoint(settings)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey.Key))

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ai request failed: %s", string(respBytes))
	}

	content, usage := extractChatCompletionResult(respBytes)
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("ai response empty")
	}

	_ = usage
	var out SecurityChatSummaryResult
	if err := json.Unmarshal([]byte(content), &out); err == nil {
		return &out, nil
	}

	return &SecurityChatSummaryResult{Summary: content}, nil
}

func (s *SecurityChatAIService) Chat(ctx context.Context, userID int64, apiKeyID *int64, contextText string, model string, messages []SecurityChatMessage) (*SecurityChatSummaryResult, error) {
	if s == nil || s.settings == nil || s.apiKeyService == nil {
		return nil, errors.New("security ai service unavailable")
	}
	settings, err := s.settings.GetAllSettings(ctx)
	if err != nil {
		return nil, err
	}
	if !settings.SecurityChatAIEnabled {
		return nil, errors.New("security chat AI disabled")
	}
	if userID <= 0 {
		return nil, errors.New("invalid user")
	}
	model = strings.TrimSpace(model)
	if model == "" {
		model = strings.TrimSpace(settings.SecurityChatAIModel)
	}
	if model == "" {
		return nil, errors.New("security chat AI config invalid")
	}
	apiKey, err := s.getUserAPIKey(ctx, userID, apiKeyID)
	if err != nil {
		return nil, err
	}

	endpoint, err := s.resolveGatewayEndpoint(settings)
	if err != nil {
		return nil, err
	}

	chatMessages := make([]map[string]any, 0, len(messages)+2)
	chatMessages = append(chatMessages, map[string]any{
		"role":    "system",
		"content": "You are a security audit assistant. Provide concise, practical answers.",
	})
	if strings.TrimSpace(contextText) != "" {
		chatMessages = append(chatMessages, map[string]any{
			"role":    "system",
			"content": contextText,
		})
	}
	for _, msg := range messages {
		role := strings.TrimSpace(msg.Role)
		if role == "" {
			role = "user"
		}
		chatMessages = append(chatMessages, map[string]any{
			"role":    role,
			"content": msg.Content,
		})
	}

	requestBody := map[string]any{
		"model": model,
		"messages": chatMessages,
		"temperature": 0.2,
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey.Key))

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ai request failed: %s", string(respBytes))
	}

	content, _ := extractChatCompletionResult(respBytes)
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("ai response empty")
	}
	return &SecurityChatSummaryResult{Summary: content}, nil
}

type securityChatPrompt struct {
	System string
	User   string
}

func buildSecurityChatSummaryPrompt(input *SecurityChatSummaryInput) securityChatPrompt {
	meta := input.Meta
	lines := make([]string, 0, len(input.Messages)+8)
	lines = append(lines, fmt.Sprintf("session_count=%d message_count=%d", meta.SessionCount, meta.MessageCount))
	if meta.UserID != nil {
		lines = append(lines, fmt.Sprintf("user_id=%d", *meta.UserID))
	}
	if meta.APIKeyID != nil {
		lines = append(lines, fmt.Sprintf("api_key_id=%d", *meta.APIKeyID))
	}
	if !meta.StartTime.IsZero() && !meta.EndTime.IsZero() {
		lines = append(lines, fmt.Sprintf("range=%s..%s", meta.StartTime.Format(time.RFC3339), meta.EndTime.Format(time.RFC3339)))
	}
	lines = append(lines, "")

	for _, msg := range input.Messages {
		role := msg.Role
		if role == "" {
			role = "user"
		}
		content := strings.ReplaceAll(msg.Content, "\n", " ")
		if len(content) > 800 {
			content = content[:800]
		}
		lines = append(lines, fmt.Sprintf("[%s/%s] %s", role, msg.Source, content))
	}

	return securityChatPrompt{
		System: "You are a security auditor. Summarize the activity, identify any sensitive data exposure, and output JSON with keys: summary, sensitive_findings (array), risk_level (low|medium|high), recommended_actions (array).",
		User:   strings.Join(lines, "\n"),
	}
}

func extractChatCompletionResult(raw []byte) (string, OpenAIUsage) {
	var resp struct {
		Model  string `json:"model"`
		Usage  struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			InputTokens      int `json:"input_tokens"`
			OutputTokens     int `json:"output_tokens"`
		} `json:"usage"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Text string `json:"text"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return "", OpenAIUsage{}
	}
	content := ""
	if len(resp.Choices) > 0 {
		if c := strings.TrimSpace(resp.Choices[0].Message.Content); c != "" {
			content = c
		} else if c := strings.TrimSpace(resp.Choices[0].Text); c != "" {
			content = c
		}
	}
	usage := OpenAIUsage{}
	if resp.Usage.PromptTokens > 0 || resp.Usage.CompletionTokens > 0 {
		usage.InputTokens = resp.Usage.PromptTokens
		usage.OutputTokens = resp.Usage.CompletionTokens
	} else if resp.Usage.InputTokens > 0 || resp.Usage.OutputTokens > 0 {
		usage.InputTokens = resp.Usage.InputTokens
		usage.OutputTokens = resp.Usage.OutputTokens
	}
	return content, usage
}

func (s *SecurityChatAIService) getUserAPIKey(ctx context.Context, userID int64, apiKeyID *int64) (*APIKey, error) {
	if apiKeyID != nil && *apiKeyID > 0 {
		key, err := s.apiKeyService.GetByID(ctx, *apiKeyID)
		if err != nil {
			return nil, err
		}
		if key.UserID != userID {
			return nil, errors.New("api key does not belong to user")
		}
		return key, nil
	}
	keys, _, err := s.apiKeyService.List(ctx, userID, pagination.PaginationParams{Page: 1, PageSize: 50})
	if err != nil {
		return nil, err
	}
	for i := range keys {
		if keys[i].Status == StatusActive {
			return s.apiKeyService.GetByID(ctx, keys[i].ID)
		}
	}
	return nil, errors.New("no available api keys")
}

func (s *SecurityChatAIService) resolveGatewayEndpoint(settings *SystemSettings) (string, error) {
	baseURL := strings.TrimSpace(settings.APIBaseURL)
	if baseURL == "" && s.cfg != nil {
		port := s.cfg.Server.Port
		if port == 0 {
			port = 80
		}
		baseURL = fmt.Sprintf("http://127.0.0.1:%d", port)
	}
	if baseURL == "" {
		return "", errors.New("api_base_url not configured")
	}
	baseURL = strings.TrimRight(baseURL, "/")
	endpoint := baseURL + "/v1/chat/completions"
	if strings.HasSuffix(baseURL, "/v1") {
		endpoint = baseURL + "/chat/completions"
	}
	return endpoint, nil
}

func requestIDFromHeaders(headers http.Header) string {
	if v := strings.TrimSpace(headers.Get("x-request-id")); v != "" {
		return v
	}
	if v := strings.TrimSpace(headers.Get("request-id")); v != "" {
		return v
	}
	return fmt.Sprintf("secai-%d", time.Now().UnixNano())
}
