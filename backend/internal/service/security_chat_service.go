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
	UpsertSession(ctx context.Context, input *SecurityChatSessionUpsertInput) error
	ListSessions(ctx context.Context, filter *SecurityChatSessionFilter) ([]*SecurityChatSession, int64, error)
	ListMessages(ctx context.Context, filter *SecurityChatMessageFilter) ([]*SecurityChatLog, int64, error)
	GetStats(ctx context.Context, filter *SecurityChatMessageFilter) (*SecurityChatStats, error)
	DeleteExpired(ctx context.Context, cutoff time.Time) (int64, error)
	DeleteExpiredSessions(ctx context.Context, cutoff time.Time) (int64, error)
	DeleteSession(ctx context.Context, sessionID string, userID *int64, apiKeyID *int64) (int64, int64, error)
	DeleteSessions(ctx context.Context, sessionIDs []string, userID *int64, apiKeyID *int64) (int64, int64, error)
	DeleteSessionsByFilter(ctx context.Context, filter *SecurityChatSessionFilter) (int64, int64, error)
	DeleteLogsByFilter(ctx context.Context, filter *SecurityChatMessageFilter) (int64, int64, error)
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

	retentionDays := s.GetRetentionDays(ctx)
	createdAt := time.Now().UTC()

	sessionID := extractSecuritySessionID(input.RequestBody, input.ClientRequestID, input.RequestID, input.UserID, input.APIKeyID, createdAt)
	if sessionID == "" {
		return nil
	}
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
	if err != nil {
		return err
	}
	if err := s.repo.UpsertSession(ctx, &SecurityChatSessionUpsertInput{
		SessionID:      sessionID,
		UserID:         input.UserID,
		APIKeyID:       input.APIKeyID,
		AccountID:      input.AccountID,
		GroupID:        input.GroupID,
		Platform:       input.Platform,
		Model:          input.Model,
		MessagePreview: preview,
		FirstAt:        createdAt,
		LastAt:         createdAt,
		ExpiresAt:      expiresAt,
	}); err != nil {
		return err
	}
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

func (s *SecurityChatService) GetStats(ctx context.Context, filter *SecurityChatMessageFilter) (*SecurityChatStats, error) {
	if s == nil || s.repo == nil {
		return &SecurityChatStats{PlatformBuckets: []SecurityChatPlatformBucket{}}, nil
	}
	if filter == nil {
		filter = &SecurityChatMessageFilter{}
	}
	return s.repo.GetStats(ctx, filter)
}

func (s *SecurityChatService) CleanupExpired(ctx context.Context) (int64, error) {
	if s == nil || s.repo == nil {
		return 0, nil
	}
	cutoff := time.Now().UTC()
	deletedLogs, err := s.repo.DeleteExpired(ctx, cutoff)
	if err != nil {
		return deletedLogs, err
	}
	deletedSessions, err := s.repo.DeleteExpiredSessions(ctx, cutoff)
	return deletedLogs + deletedSessions, err
}

func (s *SecurityChatService) DeleteSession(ctx context.Context, sessionID string, userID *int64, apiKeyID *int64) (int64, int64, error) {
	if s == nil || s.repo == nil {
		return 0, 0, nil
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return 0, 0, fmt.Errorf("session_id required")
	}
	return s.repo.DeleteSession(ctx, sessionID, userID, apiKeyID)
}

func (s *SecurityChatService) DeleteSessions(ctx context.Context, sessionIDs []string, userID *int64, apiKeyID *int64) (int64, int64, error) {
	if s == nil || s.repo == nil {
		return 0, 0, nil
	}
	cleaned := make([]string, 0, len(sessionIDs))
	for _, id := range sessionIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			cleaned = append(cleaned, id)
		}
	}
	if len(cleaned) == 0 {
		return 0, 0, fmt.Errorf("session_ids required")
	}
	return s.repo.DeleteSessions(ctx, cleaned, userID, apiKeyID)
}

func (s *SecurityChatService) DeleteSessionsByFilter(ctx context.Context, filter *SecurityChatSessionFilter) (int64, int64, error) {
	if s == nil || s.repo == nil {
		return 0, 0, nil
	}
	if filter == nil {
		return 0, 0, fmt.Errorf("filter required")
	}
	return s.repo.DeleteSessionsByFilter(ctx, filter)
}

func (s *SecurityChatService) DeleteLogsByFilter(ctx context.Context, filter *SecurityChatMessageFilter) (int64, int64, error) {
	if s == nil || s.repo == nil {
		return 0, 0, nil
	}
	if filter == nil {
		return 0, 0, fmt.Errorf("filter required")
	}
	return s.repo.DeleteLogsByFilter(ctx, filter)
}

func (s *SecurityChatService) ShouldCapture(ctx context.Context, userID *int64, userEmail string) bool {
	if s == nil || s.settingRepo == nil {
		return true
	}
	value, err := s.settingRepo.GetValue(ctx, SettingKeySecurityChatExcludedUsers)
	if err != nil {
		return true
	}
	list := parseExcludedUsers(value)
	if len(list.ids) == 0 && len(list.emails) == 0 {
		return true
	}
	if userID != nil {
		if _, ok := list.ids[*userID]; ok {
			return false
		}
	}
	if userEmail != "" {
		if _, ok := list.emails[strings.ToLower(strings.TrimSpace(userEmail))]; ok {
			return false
		}
	}
	return true
}

type excludedUserList struct {
	ids    map[int64]struct{}
	emails map[string]struct{}
}

func parseExcludedUsers(raw string) excludedUserList {
	result := excludedUserList{ids: map[int64]struct{}{}, emails: map[string]struct{}{}}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return result
	}
	separators := []rune{',', '\n', '\t', ' '}
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		for _, sep := range separators {
			if r == sep {
				return true
			}
		}
		return false
	})
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if id, err := strconv.ParseInt(field, 10, 64); err == nil {
			result.ids[id] = struct{}{}
			continue
		}
		result.emails[strings.ToLower(field)] = struct{}{}
	}
	return result
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

type SecurityChatSessionUpsertInput struct {
	SessionID      string
	UserID         *int64
	APIKeyID       *int64
	AccountID      *int64
	GroupID        *int64
	Platform       string
	Model          string
	MessagePreview string
	FirstAt        time.Time
	LastAt         time.Time
	ExpiresAt      time.Time
}

type SecurityChatLog struct {
	ID              int64                 `json:"id"`
	SessionID       string                `json:"session_id"`
	RequestID       *string               `json:"request_id,omitempty"`
	ClientRequestID *string               `json:"client_request_id,omitempty"`
	UserID          *int64                `json:"user_id,omitempty"`
	UserEmail       *string               `json:"user_email,omitempty"`
	APIKeyID        *int64                `json:"api_key_id,omitempty"`
	AccountID       *int64                `json:"account_id,omitempty"`
	GroupID         *int64                `json:"group_id,omitempty"`
	Platform        *string               `json:"platform,omitempty"`
	Model           *string               `json:"model,omitempty"`
	RequestPath     *string               `json:"request_path,omitempty"`
	Stream          bool                  `json:"stream"`
	StatusCode      *int                  `json:"status_code,omitempty"`
	Messages        []SecurityChatMessage `json:"messages"`
	CreatedAt       time.Time             `json:"created_at"`
}

type SecurityChatSession struct {
	SessionID      string    `json:"session_id"`
	UserID         *int64    `json:"user_id,omitempty"`
	UserEmail      *string   `json:"user_email,omitempty"`
	APIKeyID       *int64    `json:"api_key_id,omitempty"`
	AccountID      *int64    `json:"account_id,omitempty"`
	GroupID        *int64    `json:"group_id,omitempty"`
	Platform       *string   `json:"platform,omitempty"`
	Model          *string   `json:"model,omitempty"`
	MessagePreview *string   `json:"message_preview,omitempty"`
	LastAt         time.Time `json:"last_at"`
	RequestCount   int64     `json:"request_count"`
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
	AccountID *int64
	GroupID   *int64
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

	SessionID         string
	UserID            *int64
	APIKeyID          *int64
	AccountID         *int64
	GroupID           *int64
	Platform          string
	Model             string
	RequestPath       string
	AllowEmptySession bool
	IgnoreTimeRange   bool

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

type SecurityChatPlatformBucket struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

type SecurityChatStats struct {
	RequestCount    int64                        `json:"request_count"`
	SessionCount    int64                        `json:"session_count"`
	EstimatedBytes  int64                        `json:"estimated_bytes"`
	TableBytes      int64                        `json:"table_bytes"`
	PlatformBuckets []SecurityChatPlatformBucket `json:"platform_buckets"`
}

func SecurityChatSessionFromRow(sessionID string, userID, apiKeyID, accountID, groupID sql.NullInt64, userEmail, platform, model, preview sql.NullString, lastAt time.Time, requestCount int64) *SecurityChatSession {
	item := &SecurityChatSession{
		SessionID:    sessionID,
		LastAt:       lastAt,
		RequestCount: requestCount,
	}
	if userID.Valid {
		v := userID.Int64
		item.UserID = &v
	}
	if userEmail.Valid {
		v := userEmail.String
		item.UserEmail = &v
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

func extractSecuritySessionID(requestBody []byte, clientRequestID string, requestID string, userID *int64, apiKeyID *int64, createdAt time.Time) string {
	var req map[string]any
	if err := json.Unmarshal(requestBody, &req); err == nil {
		for _, key := range []string{"session_id", "conversation_id", "thread_id"} {
			if v, ok := req[key].(string); ok {
				if s := strings.TrimSpace(v); s != "" {
					return buildSessionKey(s, userID, apiKeyID)
				}
			}
		}
		if meta, ok := req["metadata"].(map[string]any); ok {
			if v, ok := meta["session_id"].(string); ok {
				if s := strings.TrimSpace(v); s != "" {
					return buildSessionKey(s, userID, apiKeyID)
				}
			}
		}
	}
	if s := strings.TrimSpace(clientRequestID); s != "" {
		return buildSessionKey(s, userID, apiKeyID)
	}
	if s := strings.TrimSpace(requestID); s != "" {
		return buildSessionKey(s, userID, apiKeyID)
	}
	if !createdAt.IsZero() {
		uid := int64(0)
		kid := int64(0)
		if userID != nil {
			uid = *userID
		}
		if apiKeyID != nil {
			kid = *apiKeyID
		}
		window := createdAt.UTC().Truncate(15 * time.Minute)
		return fmt.Sprintf("auto:%d:%d:%s", uid, kid, window.Format("20060102T1504"))
	}
	return ""
}

func buildSessionKey(base string, userID *int64, apiKeyID *int64) string {
	base = strings.TrimSpace(base)
	uid := int64(0)
	kid := int64(0)
	if userID != nil {
		uid = *userID
	}
	if apiKeyID != nil {
		kid = *apiKeyID
	}
	if base == "" {
		return fmt.Sprintf("auto:%d:%d", uid, kid)
	}
	return fmt.Sprintf("%s:%d:%d", base, uid, kid)
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
