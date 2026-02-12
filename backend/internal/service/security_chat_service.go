package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
)

type SecurityChatRepository interface {
	InsertChatLog(ctx context.Context, input *SecurityChatLogInput) (int64, error)
	ListSessions(ctx context.Context, filter *SecurityChatSessionFilter) ([]*SecurityChatSession, int64, error)
	ListMessages(ctx context.Context, filter *SecurityChatMessageFilter) ([]*SecurityChatLog, int64, error)
	DeleteExpired(ctx context.Context, cutoff time.Time) (int64, error)
}

type SecurityChatService struct {
	repo        SecurityChatRepository
	settingRepo SettingRepository
}

func NewSecurityChatService(repo SecurityChatRepository, settingRepo SettingRepository) *SecurityChatService {
	return &SecurityChatService{repo: repo, settingRepo: settingRepo}
}

func (s *SecurityChatService) GetRetentionDays(ctx context.Context) int {
	if s == nil || s.settingRepo == nil {
		return 7
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeySecurityChatRetentionDays)
	if err != nil {
		return 7
	}
	if v, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil {
		if v < 1 {
			return 1
		}
		if v > 365 {
			return 365
		}
		return v
	}
	return 7
}

func (s *SecurityChatService) RecordChat(ctx context.Context, input *SecurityChatCaptureInput) error {
	if s == nil || s.repo == nil || input == nil {
		return nil
	}
	if len(input.RequestBody) == 0 {
		return nil
	}

	messages := buildSecurityChatMessages(input.Platform, input.RequestBody, input.ResponseBody)
	if len(messages) == 0 {
		return nil
	}

	sessionID := extractSecuritySessionID(input.RequestBody, input.ClientRequestID, input.RequestID)
	if sessionID == "" {
		return nil
	}

	retentionDays := s.GetRetentionDays(ctx)
	createdAt := time.Now().UTC()
	expiresAt := createdAt.AddDate(0, 0, retentionDays)

	preview := ""
	if len(messages) > 0 {
		preview = messages[len(messages)-1].Content
		if len(preview) > 280 {
			preview = preview[:280]
		}
	}

	_, err := s.repo.InsertChatLog(ctx, &SecurityChatLogInput{
		SessionID:       sessionID,
		RequestID:       input.RequestID,
		ClientRequestID: input.ClientRequestID,
		UserID:          input.UserID,
		APIKeyID:        input.APIKeyID,
		AccountID:       input.AccountID,
		GroupID:         input.GroupID,
		Platform:        input.Platform,
		Model:           input.Model,
		RequestPath:     input.RequestPath,
		Stream:          input.Stream,
		StatusCode:      input.StatusCode,
		Messages:        messages,
		MessagePreview:  preview,
		CreatedAt:       createdAt,
		ExpiresAt:       expiresAt,
	})
	return err
}

func (s *SecurityChatService) ListSessions(ctx context.Context, filter *SecurityChatSessionFilter) (*SecurityChatSessionList, error) {
	if s == nil || s.repo == nil {
		return &SecurityChatSessionList{Items: []*SecurityChatSession{}, Total: 0, Page: 1, PageSize: 50}, nil
	}
	if filter == nil {
		filter = &SecurityChatSessionFilter{}
	}
	items, total, err := s.repo.ListSessions(ctx, filter)
	if err != nil {
		return nil, err
	}
	page, pageSize, _, _ := filter.Normalize()
	return &SecurityChatSessionList{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *SecurityChatService) ListMessages(ctx context.Context, filter *SecurityChatMessageFilter) (*SecurityChatLogList, error) {
	if s == nil || s.repo == nil {
		return &SecurityChatLogList{Items: []*SecurityChatLog{}, Total: 0, Page: 1, PageSize: 50}, nil
	}
	if filter == nil {
		filter = &SecurityChatMessageFilter{}
	}
	items, total, err := s.repo.ListMessages(ctx, filter)
	if err != nil {
		return nil, err
	}
	page, pageSize, _, _ := filter.Normalize()
	return &SecurityChatLogList{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *SecurityChatService) CleanupExpired(ctx context.Context) (int64, error) {
	if s == nil || s.repo == nil {
		return 0, nil
	}
	return s.repo.DeleteExpired(ctx, time.Now().UTC())
}

type SecurityChatCaptureInput struct {
	RequestID       string
	ClientRequestID string
	UserID          *int64
	APIKeyID        *int64
	AccountID       *int64
	GroupID         *int64
	Platform        string
	Model           string
	RequestPath     string
	Stream          bool
	StatusCode      int
	RequestBody     []byte
	ResponseBody    []byte
}

type SecurityChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Source  string `json:"source"`
	Index   int    `json:"index"`
}

type SecurityChatLogInput struct {
	SessionID       string
	RequestID       string
	ClientRequestID string
	UserID          *int64
	APIKeyID        *int64
	AccountID       *int64
	GroupID         *int64
	Platform        string
	Model           string
	RequestPath     string
	Stream          bool
	StatusCode      int
	Messages        []SecurityChatMessage
	MessagePreview  string
	CreatedAt       time.Time
	ExpiresAt       time.Time
}

type SecurityChatLog struct {
	ID              int64                `json:"id"`
	SessionID       string               `json:"session_id"`
	RequestID       *string              `json:"request_id,omitempty"`
	ClientRequestID *string              `json:"client_request_id,omitempty"`
	UserID          *int64               `json:"user_id,omitempty"`
	APIKeyID        *int64               `json:"api_key_id,omitempty"`
	AccountID       *int64               `json:"account_id,omitempty"`
	GroupID         *int64               `json:"group_id,omitempty"`
	Platform        *string              `json:"platform,omitempty"`
	Model           *string              `json:"model,omitempty"`
	RequestPath     *string              `json:"request_path,omitempty"`
	Stream          bool                 `json:"stream"`
	StatusCode      *int                 `json:"status_code,omitempty"`
	Messages        []SecurityChatMessage `json:"messages"`
	CreatedAt       time.Time            `json:"created_at"`
}

type SecurityChatSession struct {
	SessionID     string    `json:"session_id"`
	UserID        *int64    `json:"user_id,omitempty"`
	APIKeyID      *int64    `json:"api_key_id,omitempty"`
	AccountID     *int64    `json:"account_id,omitempty"`
	GroupID       *int64    `json:"group_id,omitempty"`
	Platform      *string   `json:"platform,omitempty"`
	Model         *string   `json:"model,omitempty"`
	MessagePreview *string   `json:"message_preview,omitempty"`
	LastAt        time.Time `json:"last_at"`
	RequestCount  int64     `json:"request_count"`
}

type SecurityChatSessionList struct {
	Items    []*SecurityChatSession `json:"items"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

type SecurityChatLogList struct {
	Items    []*SecurityChatLog `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

type SecurityChatSessionFilter struct {
	StartTime *time.Time
	EndTime   *time.Time

	SessionID string
	UserID    *int64
	APIKeyID  *int64
	Query     string
	Platform  string
	Model     string

	Page     int
	PageSize int
}

func (f *SecurityChatSessionFilter) Normalize() (page, pageSize int, startTime, endTime time.Time) {
	page = 1
	pageSize = 50
	endTime = time.Now()
	startTime = endTime.Add(-24 * time.Hour)

	if f == nil {
		return page, pageSize, startTime, endTime
	}
	if f.Page > 0 {
		page = f.Page
	}
	if f.PageSize > 0 {
		pageSize = f.PageSize
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if f.EndTime != nil {
		endTime = *f.EndTime
	}
	if f.StartTime != nil {
		startTime = *f.StartTime
	} else if f.EndTime != nil {
		startTime = endTime.Add(-24 * time.Hour)
	}
	if startTime.After(endTime) {
		startTime, endTime = endTime, startTime
	}
	return page, pageSize, startTime, endTime
}

type SecurityChatMessageFilter struct {
	StartTime *time.Time
	EndTime   *time.Time

	SessionID string
	UserID    *int64
	APIKeyID  *int64
	AllowEmptySession bool

	Page     int
	PageSize int
}

func (f *SecurityChatMessageFilter) Normalize() (page, pageSize int, startTime, endTime time.Time) {
	page = 1
	pageSize = 200
	endTime = time.Now()
	startTime = endTime.Add(-24 * time.Hour)

	if f == nil {
		return page, pageSize, startTime, endTime
	}
	if f.Page > 0 {
		page = f.Page
	}
	if f.PageSize > 0 {
		pageSize = f.PageSize
	}
	if pageSize > 500 {
		pageSize = 500
	}
	if f.EndTime != nil {
		endTime = *f.EndTime
	}
	if f.StartTime != nil {
		startTime = *f.StartTime
	} else if f.EndTime != nil {
		startTime = endTime.Add(-24 * time.Hour)
	}
	if startTime.After(endTime) {
		startTime, endTime = endTime, startTime
	}
	return page, pageSize, startTime, endTime
}

func SecurityChatSessionFromRow(sessionID string, userID, apiKeyID, accountID, groupID sql.NullInt64, platform, model, preview sql.NullString, lastAt time.Time, requestCount int64) *SecurityChatSession {
	item := &SecurityChatSession{
		SessionID:    sessionID,
		LastAt:       lastAt,
		RequestCount: requestCount,
	}
	if userID.Valid {
		v := userID.Int64
		item.UserID = &v
	}
	if apiKeyID.Valid {
		v := apiKeyID.Int64
		item.APIKeyID = &v
	}
	if accountID.Valid {
		v := accountID.Int64
		item.AccountID = &v
	}
	if groupID.Valid {
		v := groupID.Int64
		item.GroupID = &v
	}
	if platform.Valid {
		item.Platform = &platform.String
	}
	if model.Valid {
		item.Model = &model.String
	}
	if preview.Valid {
		item.MessagePreview = &preview.String
	}
	return item
}

func buildSecurityChatMessages(platform string, requestBody []byte, responseBody []byte) []SecurityChatMessage {
	var req map[string]any
	if err := json.Unmarshal(requestBody, &req); err != nil {
		return nil
	}

	msgIndex := 0
	msgs := make([]SecurityChatMessage, 0, 12)

	appendMsg := func(role, content, source string) {
		content = strings.TrimSpace(content)
		if content == "" {
			return
		}
		msgs = append(msgs, SecurityChatMessage{
			Role:    role,
			Content: content,
			Source:  source,
			Index:   msgIndex,
		})
		msgIndex++
	}

	protocol := platform
	if protocol == "" {
		protocol = domain.PlatformOpenAI
	}
	protocol = strings.ToLower(protocol)

	parsed, err := ParseGatewayRequest(requestBody, protocol)
	if err == nil {
		if parsed.HasSystem && parsed.System != nil {
			appendMsg("system", stringifyChatContent(parsed.System), "request")
		}
		for _, msg := range parsed.Messages {
			role, content := normalizeMessage(msg)
			appendMsg(role, content, "request")
		}
	}

	if len(msgs) == 0 {
		if input, ok := req["input"]; ok {
			appendMsg("user", stringifyChatContent(input), "request")
		}
	}

	respMsgs := extractResponseMessages(platform, responseBody)
	for _, msg := range respMsgs {
		appendMsg(msg.Role, msg.Content, "response")
	}

	return msgs
}

func extractSecuritySessionID(requestBody []byte, clientRequestID string, requestID string) string {
	var req map[string]any
	if err := json.Unmarshal(requestBody, &req); err == nil {
		for _, key := range []string{"session_id", "conversation_id", "thread_id"} {
			if v, ok := req[key].(string); ok {
				if s := strings.TrimSpace(v); s != "" {
					return s
				}
			}
		}
		if meta, ok := req["metadata"].(map[string]any); ok {
			if v, ok := meta["session_id"].(string); ok {
				if s := strings.TrimSpace(v); s != "" {
					return s
				}
			}
		}
	}
	if s := strings.TrimSpace(clientRequestID); s != "" {
		return s
	}
	return strings.TrimSpace(requestID)
}

func normalizeMessage(msg any) (role string, content string) {
	if m, ok := msg.(map[string]any); ok {
		if r, ok := m["role"].(string); ok {
			role = r
		}
		if c, ok := m["content"]; ok {
			content = stringifyChatContent(c)
		}
	}
	if role == "" {
		role = "user"
	}
	return role, content
}

func stringifyChatContent(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []any:
		parts := make([]string, 0, len(t))
		for _, item := range t {
			parts = append(parts, stringifyChatContent(item))
		}
		return strings.TrimSpace(strings.Join(parts, "\n"))
	case map[string]any:
		if text, ok := t["text"].(string); ok {
			return text
		}
		if content, ok := t["content"]; ok {
			return stringifyChatContent(content)
		}
		bytes, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprintf("%v", t)
		}
		return string(bytes)
	default:
		bytes, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprintf("%v", t)
		}
		return string(bytes)
	}
}

func extractResponseMessages(platform string, responseBody []byte) []SecurityChatMessage {
	var resp map[string]any
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		return nil
	}

	msgs := make([]SecurityChatMessage, 0, 4)
	appendMsg := func(role, content string) {
		content = strings.TrimSpace(content)
		if content == "" {
			return
		}
		msgs = append(msgs, SecurityChatMessage{Role: role, Content: content})
	}

	if choices, ok := resp["choices"].([]any); ok {
		for _, c := range choices {
			if m, ok := c.(map[string]any); ok {
				if msg, ok := m["message"].(map[string]any); ok {
					role := "assistant"
					if r, ok := msg["role"].(string); ok && r != "" {
						role = r
					}
					appendMsg(role, stringifyChatContent(msg["content"]))
				}
				if text, ok := m["text"].(string); ok {
					appendMsg("assistant", text)
				}
			}
		}
		return msgs
	}

	if output, ok := resp["output"].([]any); ok {
		for _, item := range output {
			if m, ok := item.(map[string]any); ok {
				role := "assistant"
				if r, ok := m["role"].(string); ok && r != "" {
					role = r
				}
				appendMsg(role, stringifyChatContent(m["content"]))
			}
		}
		return msgs
	}

	if content, ok := resp["content"]; ok {
		appendMsg("assistant", stringifyChatContent(content))
		return msgs
	}

	if candidates, ok := resp["candidates"].([]any); ok {
		for _, item := range candidates {
			if m, ok := item.(map[string]any); ok {
				if c, ok := m["content"].(map[string]any); ok {
					appendMsg("assistant", stringifyChatContent(c["parts"]))
				}
			}
		}
		return msgs
	}

	return nil
}
