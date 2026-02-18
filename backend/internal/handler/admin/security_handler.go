package admin

import (
	"archive/zip"
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type SecurityHandler struct {
	chatService   *service.SecurityChatService
	aiService     *service.SecurityChatAIService
	apiKeyService *service.APIKeyService
}

func NewSecurityHandler(chatService *service.SecurityChatService, aiService *service.SecurityChatAIService, apiKeyService *service.APIKeyService) *SecurityHandler {
	return &SecurityHandler{chatService: chatService, aiService: aiService, apiKeyService: apiKeyService}
}

// GET /api/v1/admin/security/sessions
func (h *SecurityHandler) ListChatSessions(c *gin.Context) {
	if h.chatService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Security service not available")
		return
	}

	page, pageSize := response.ParsePagination(c)
	if pageSize > 100 {
		pageSize = 100
	}

	startTime, endTime, err := parseOpsTimeRange(c, "24h")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	filter := &service.SecurityChatSessionFilter{
		Page:      page,
		PageSize:  pageSize,
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	filter.Query = strings.TrimSpace(c.Query("q"))
	filter.SessionID = strings.TrimSpace(c.Query("session_id"))
	filter.Platform = strings.TrimSpace(c.Query("platform"))
	filter.Model = strings.TrimSpace(c.Query("model"))

	if v := strings.TrimSpace(c.Query("user_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid user_id")
			return
		}
		filter.UserID = &id
	}
	if v := strings.TrimSpace(c.Query("api_key_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid api_key_id")
			return
		}
		filter.APIKeyID = &id
	}

	list, err := h.chatService.ListSessions(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, list)
}

// GET /api/v1/admin/security/messages
func (h *SecurityHandler) ListChatMessages(c *gin.Context) {
	if h.chatService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Security service not available")
		return
	}

	page, pageSize := response.ParsePagination(c)
	if pageSize > 500 {
		pageSize = 500
	}

	startTime, endTime, err := parseOpsTimeRange(c, "24h")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	filter := &service.SecurityChatMessageFilter{
		Page:      page,
		PageSize:  pageSize,
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	filter.SessionID = strings.TrimSpace(c.Query("session_id"))
	if filter.SessionID == "" {
		response.BadRequest(c, "session_id is required")
		return
	}
	filter.IgnoreTimeRange = true

	if v := strings.TrimSpace(c.Query("user_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid user_id")
			return
		}
		filter.UserID = &id
	}
	if v := strings.TrimSpace(c.Query("api_key_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid api_key_id")
			return
		}
		filter.APIKeyID = &id
	}

	list, err := h.chatService.ListMessages(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, list)
}

// POST /api/v1/admin/security/summarize
func (h *SecurityHandler) SummarizeChat(c *gin.Context) {
	if h.chatService == nil || h.aiService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Security service not available")
		return
	}

	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		StartTime  string  `json:"start_time"`
		EndTime    string  `json:"end_time"`
		UserID     *int64  `json:"user_id"`
		APIKeyID   *int64  `json:"api_key_id"`
		SessionID  *string `json:"session_id"`
		AIAPIKeyID *int64  `json:"ai_api_key_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	startTime, endTime, err := parseOpsTimeRange(c, "24h")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if strings.TrimSpace(req.StartTime) != "" || strings.TrimSpace(req.EndTime) != "" {
		values := c.Request.URL.Query()
		values.Set("start_time", strings.TrimSpace(req.StartTime))
		values.Set("end_time", strings.TrimSpace(req.EndTime))
		c.Request.URL.RawQuery = values.Encode()
		startTime, endTime, err = parseOpsTimeRange(c, "24h")
		if err != nil {
			response.BadRequest(c, err.Error())
			return
		}
	}

	filter := &service.SecurityChatMessageFilter{
		StartTime:         &startTime,
		EndTime:           &endTime,
		PageSize:          500,
		AllowEmptySession: true,
	}
	if req.SessionID != nil && strings.TrimSpace(*req.SessionID) != "" {
		filter.SessionID = strings.TrimSpace(*req.SessionID)
		filter.IgnoreTimeRange = true
	} else {
		filter.UserID = req.UserID
		filter.APIKeyID = req.APIKeyID
	}

	logs, err := h.chatService.ListMessages(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if len(logs.Items) == 0 {
		if req.SessionID != nil {
			base := stripSessionSuffix(*req.SessionID)
			if base != "" && base != strings.TrimSpace(*req.SessionID) {
				filter.SessionID = base
				filter.IgnoreTimeRange = true
				logs, err = h.chatService.ListMessages(c.Request.Context(), filter)
				if err != nil {
					response.ErrorFrom(c, err)
					return
				}
			}
		}
		if len(logs.Items) == 0 {
			response.Success(c, &service.SecurityChatSummaryResult{
				Summary:            "No chat logs found",
				SensitiveFindings:  []string{},
				RecommendedActions: []string{},
				RiskLevel:          "low",
			})
			return
		}
	}
	messages := make([]service.SecurityChatMessage, 0)
	for _, log := range logs.Items {
		messages = append(messages, log.Messages...)
	}

	result, err := h.aiService.Summarize(c.Request.Context(), &service.SecurityChatSummaryInput{
		Messages: messages,
		Meta: service.SecurityChatSummaryMeta{
			SessionCount: 0,
			MessageCount: len(messages),
			StartTime:    startTime,
			EndTime:      endTime,
			UserID:       req.UserID,
			APIKeyID:     req.APIKeyID,
		},
	}, subject.UserID, req.AIAPIKeyID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, result)
}

// GET /api/v1/admin/security/api-keys
func (h *SecurityHandler) ListAPIKeys(c *gin.Context) {
	if h.apiKeyService == nil {
		response.Error(c, http.StatusServiceUnavailable, "API key service not available")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	keys, _, err := h.apiKeyService.List(c.Request.Context(), subject.UserID, pagination.PaginationParams{Page: 1, PageSize: 200})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	items := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		items = append(items, map[string]any{
			"id":       key.ID,
			"name":     key.Name,
			"group_id": key.GroupID,
			"status":   key.Status,
		})
	}
	response.Success(c, items)
}

// GET /api/v1/admin/security/sessions/export
func (h *SecurityHandler) ExportSessions(c *gin.Context) {
	if h.chatService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Security service not available")
		return
	}

	startTime, endTime, err := parseOpsTimeRange(c, "24h")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	filter := &service.SecurityChatSessionFilter{
		StartTime: &startTime,
		EndTime:   &endTime,
		Query:     strings.TrimSpace(c.Query("q")),
		SessionID: strings.TrimSpace(c.Query("session_id")),
		Platform:  strings.TrimSpace(c.Query("platform")),
		Model:     strings.TrimSpace(c.Query("model")),
		Page:      1,
		PageSize:  1000,
	}

	if v := strings.TrimSpace(c.Query("user_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid user_id")
			return
		}
		filter.UserID = &id
	}
	if v := strings.TrimSpace(c.Query("api_key_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid api_key_id")
			return
		}
		filter.APIKeyID = &id
	}

	const maxRows = 20000
	var rowsWritten int
	var truncated bool

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=security_sessions.csv")

	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"session_id", "user_email", "api_key_id", "platform", "model", "last_at", "request_count", "message_preview"})

	for {
		list, err := h.chatService.ListSessions(c.Request.Context(), filter)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		for _, item := range list.Items {
			row := []string{
				item.SessionID,
				stringOrEmpty(item.UserEmail),
				int64OrEmpty(item.APIKeyID),
				stringOrEmpty(item.Platform),
				stringOrEmpty(item.Model),
				item.LastAt.Format(time.RFC3339),
				strconv.FormatInt(item.RequestCount, 10),
				stringOrEmpty(item.MessagePreview),
			}
			_ = w.Write(row)
			rowsWritten++
			if rowsWritten >= maxRows {
				truncated = true
				break
			}
		}
		w.Flush()
		if truncated || len(list.Items) < filter.PageSize {
			break
		}
		filter.Page++
	}
	if truncated {
		c.Header("X-Export-Truncated", "true")
	}
}

// GET /api/v1/admin/security/stats
func (h *SecurityHandler) GetChatStats(c *gin.Context) {
	if h.chatService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Security service not available")
		return
	}

	filter, startTime, endTime, err := parseSecurityChatLogFilterFromQuery(c, "24h")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	stats, err := h.chatService.GetStats(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	days := math.Ceil(endTime.Sub(startTime).Hours() / 24)
	if days < 1 {
		days = 1
	}

	platformShare := map[string]map[string]any{
		"opencode": {"count": int64(0), "ratio": 0.0},
		"codex":    {"count": int64(0), "ratio": 0.0},
		"other":    {"count": int64(0), "ratio": 0.0},
	}
	for _, bucket := range stats.PlatformBuckets {
		if _, ok := platformShare[bucket.Key]; ok {
			platformShare[bucket.Key]["count"] = bucket.Count
		}
	}
	if stats.RequestCount > 0 {
		for key, item := range platformShare {
			count, _ := item["count"].(int64)
			platformShare[key]["ratio"] = float64(count) / float64(stats.RequestCount)
		}
	}

	response.Success(c, gin.H{
		"start_time":           startTime,
		"end_time":             endTime,
		"request_count":        stats.RequestCount,
		"session_count":        stats.SessionCount,
		"avg_requests_per_day": float64(stats.RequestCount) / days,
		"avg_sessions_per_day": float64(stats.SessionCount) / days,
		"estimated_bytes":      stats.EstimatedBytes,
		"table_bytes":          stats.TableBytes,
		"platform_share":       platformShare,
		"platform_share_basis": "request",
	})
}

// GET /api/v1/admin/security/logs/export
func (h *SecurityHandler) ExportChatLogs(c *gin.Context) {
	if h.chatService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Security service not available")
		return
	}

	filter, startTime, endTime, err := parseSecurityChatLogFilterFromQuery(c, "24h")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	filter.Page = 1
	filter.PageSize = 500
	filter.AllowEmptySession = true

	firstPage, err := h.chatService.ListMessages(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	startLabel := startTime.UTC().Format("20060102")
	endLabel := endTime.UTC().Format("20060102")
	baseName := fmt.Sprintf("security_logs_%s_%s.txt", startLabel, endLabel)
	archiveName := baseName + ".zip"

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename="+archiveName)

	zipWriter := zip.NewWriter(c.Writer)
	defer func() { _ = zipWriter.Close() }()

	fileWriter, err := zipWriter.Create(baseName)
	if err != nil {
		return
	}
	bufWriter := bufio.NewWriter(fileWriter)
	defer func() { _ = bufWriter.Flush() }()

	const maxRows = 200000
	rowsWritten := 0

	writeLogs := func(items []*service.SecurityChatLog) {
		for _, item := range items {
			if rowsWritten >= maxRows {
				return
			}
			writeSecurityChatLogText(bufWriter, item)
			rowsWritten++
		}
	}

	writeLogs(firstPage.Items)
	for rowsWritten < maxRows && len(firstPage.Items) == filter.PageSize {
		filter.Page++
		page, err := h.chatService.ListMessages(c.Request.Context(), filter)
		if err != nil {
			return
		}
		writeLogs(page.Items)
		firstPage = page
	}

	if rowsWritten >= maxRows {
		c.Header("X-Export-Truncated", "true")
	}
}

func stringOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func int64OrEmpty(v *int64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatInt(*v, 10)
}

// DELETE /api/v1/admin/security/sessions/:session_id
func (h *SecurityHandler) DeleteSession(c *gin.Context) {
	if h.chatService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Security service not available")
		return
	}
	requestID := strings.TrimSpace(c.Param("session_id"))
	if requestID == "" {
		response.BadRequest(c, "session_id is required")
		return
	}

	var userID *int64
	var apiKeyID *int64
	if v := strings.TrimSpace(c.Query("user_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid user_id")
			return
		}
		userID = &id
	}
	if v := strings.TrimSpace(c.Query("api_key_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid api_key_id")
			return
		}
		apiKeyID = &id
	}

	logsDeleted, sessionsDeleted, err := h.chatService.DeleteSession(c.Request.Context(), requestID, userID, apiKeyID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, map[string]any{
		"logs_deleted":     logsDeleted,
		"sessions_deleted": sessionsDeleted,
	})
}

// POST /api/v1/admin/security/sessions/bulk-delete
func (h *SecurityHandler) BulkDeleteSessions(c *gin.Context) {
	if h.chatService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Security service not available")
		return
	}
	var req struct {
		SessionIDs []string `json:"session_ids"`
		UserID     *int64   `json:"user_id"`
		APIKeyID   *int64   `json:"api_key_id"`
		AccountID  *int64   `json:"account_id"`
		GroupID    *int64   `json:"group_id"`
		SelectAll  bool     `json:"select_all"`
		StartTime  string   `json:"start_time"`
		EndTime    string   `json:"end_time"`
		Query      string   `json:"q"`
		Platform   string   `json:"platform"`
		Model      string   `json:"model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if req.SelectAll {
		startTime, endTime, err := parseOpsTimeRange(c, "24h")
		if err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		if strings.TrimSpace(req.StartTime) != "" || strings.TrimSpace(req.EndTime) != "" {
			values := c.Request.URL.Query()
			values.Set("start_time", strings.TrimSpace(req.StartTime))
			values.Set("end_time", strings.TrimSpace(req.EndTime))
			c.Request.URL.RawQuery = values.Encode()
			startTime, endTime, err = parseOpsTimeRange(c, "24h")
			if err != nil {
				response.BadRequest(c, err.Error())
				return
			}
		}
		filter := &service.SecurityChatSessionFilter{
			StartTime: &startTime,
			EndTime:   &endTime,
			UserID:    req.UserID,
			APIKeyID:  req.APIKeyID,
			AccountID: req.AccountID,
			GroupID:   req.GroupID,
			Query:     strings.TrimSpace(req.Query),
			Platform:  strings.TrimSpace(req.Platform),
			Model:     strings.TrimSpace(req.Model),
		}
		logsDeleted, sessionsDeleted, err := h.chatService.DeleteSessionsByFilter(c.Request.Context(), filter)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		response.Success(c, map[string]any{
			"logs_deleted":     logsDeleted,
			"sessions_deleted": sessionsDeleted,
		})
		return
	}
	logsDeleted, sessionsDeleted, err := h.chatService.DeleteSessions(c.Request.Context(), req.SessionIDs, req.UserID, req.APIKeyID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, map[string]any{
		"logs_deleted":     logsDeleted,
		"sessions_deleted": sessionsDeleted,
	})
}

// POST /api/v1/admin/security/logs/delete
func (h *SecurityHandler) DeleteChatLogs(c *gin.Context) {
	if h.chatService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Security service not available")
		return
	}

	var req securityChatLogFilterPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := validateSecurityChatLogFilter(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	applySecurityChatLogTimeRange(c, req.StartTime, req.EndTime)
	startTime, endTime, err := parseOpsTimeRange(c, "24h")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	filter := buildSecurityChatLogFilter(req, startTime, endTime)
	logsDeleted, sessionsDeleted, err := h.chatService.DeleteLogsByFilter(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, map[string]any{
		"logs_deleted":     logsDeleted,
		"sessions_deleted": sessionsDeleted,
	})
}

// POST /api/v1/admin/security/ai-chat
func (h *SecurityHandler) ChatWithAI(c *gin.Context) {
	if h.aiService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Security service not available")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		AIAPIKeyID *int64                        `json:"ai_api_key_id"`
		Context    string                        `json:"context"`
		Messages   []service.SecurityChatMessage `json:"messages"`
		Model      string                        `json:"model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if len(req.Messages) == 0 {
		response.BadRequest(c, "messages required")
		return
	}

	result, err := h.aiService.Chat(c.Request.Context(), subject.UserID, req.AIAPIKeyID, req.Context, req.Model, req.Messages)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, result)
}

func stripSessionSuffix(sessionID string) string {
	value := strings.TrimSpace(sessionID)
	if value == "" {
		return ""
	}
	parts := strings.Split(value, ":")
	if len(parts) < 3 {
		return value
	}
	last := parts[len(parts)-1]
	prev := parts[len(parts)-2]
	if _, err := strconv.ParseInt(last, 10, 64); err != nil {
		return value
	}
	if _, err := strconv.ParseInt(prev, 10, 64); err != nil {
		return value
	}
	return strings.Join(parts[:len(parts)-2], ":")
}

type securityChatLogFilterPayload struct {
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	UserID      *int64 `json:"user_id"`
	APIKeyID    *int64 `json:"api_key_id"`
	AccountID   *int64 `json:"account_id"`
	GroupID     *int64 `json:"group_id"`
	SessionID   string `json:"session_id"`
	Platform    string `json:"platform"`
	Model       string `json:"model"`
	RequestPath string `json:"request_path"`
}

func parseSecurityChatLogFilterFromQuery(c *gin.Context, defaultRange string) (*service.SecurityChatMessageFilter, time.Time, time.Time, error) {
	startTime, endTime, err := parseOpsTimeRange(c, defaultRange)
	if err != nil {
		return nil, time.Time{}, time.Time{}, err
	}

	filter := &service.SecurityChatMessageFilter{
		StartTime:         &startTime,
		EndTime:           &endTime,
		AllowEmptySession: true,
		SessionID:         strings.TrimSpace(c.Query("session_id")),
		Platform:          strings.TrimSpace(c.Query("platform")),
		Model:             strings.TrimSpace(c.Query("model")),
		RequestPath:       strings.TrimSpace(c.Query("request_path")),
	}

	if v := strings.TrimSpace(c.Query("user_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			return nil, time.Time{}, time.Time{}, fmt.Errorf("Invalid user_id")
		}
		filter.UserID = &id
	}
	if v := strings.TrimSpace(c.Query("api_key_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			return nil, time.Time{}, time.Time{}, fmt.Errorf("Invalid api_key_id")
		}
		filter.APIKeyID = &id
	}
	if v := strings.TrimSpace(c.Query("account_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			return nil, time.Time{}, time.Time{}, fmt.Errorf("Invalid account_id")
		}
		filter.AccountID = &id
	}
	if v := strings.TrimSpace(c.Query("group_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			return nil, time.Time{}, time.Time{}, fmt.Errorf("Invalid group_id")
		}
		filter.GroupID = &id
	}

	return filter, startTime, endTime, nil
}

func validateSecurityChatLogFilter(req *securityChatLogFilterPayload) error {
	if req == nil {
		return fmt.Errorf("Invalid request")
	}
	for label, value := range map[string]*int64{
		"user_id":    req.UserID,
		"api_key_id": req.APIKeyID,
		"account_id": req.AccountID,
		"group_id":   req.GroupID,
	} {
		if value != nil && *value <= 0 {
			return fmt.Errorf("Invalid %s", label)
		}
	}
	return nil
}

func applySecurityChatLogTimeRange(c *gin.Context, startTime, endTime string) {
	values := c.Request.URL.Query()
	if strings.TrimSpace(startTime) != "" {
		values.Set("start_time", strings.TrimSpace(startTime))
	}
	if strings.TrimSpace(endTime) != "" {
		values.Set("end_time", strings.TrimSpace(endTime))
	}
	c.Request.URL.RawQuery = values.Encode()
}

func buildSecurityChatLogFilter(req securityChatLogFilterPayload, startTime, endTime time.Time) *service.SecurityChatMessageFilter {
	filter := &service.SecurityChatMessageFilter{
		StartTime:         &startTime,
		EndTime:           &endTime,
		AllowEmptySession: true,
		SessionID:         strings.TrimSpace(req.SessionID),
		UserID:            req.UserID,
		APIKeyID:          req.APIKeyID,
		AccountID:         req.AccountID,
		GroupID:           req.GroupID,
		Platform:          strings.TrimSpace(req.Platform),
		Model:             strings.TrimSpace(req.Model),
		RequestPath:       strings.TrimSpace(req.RequestPath),
	}
	return filter
}

func writeSecurityChatLogText(w *bufio.Writer, log *service.SecurityChatLog) {
	if log == nil {
		return
	}
	_, _ = fmt.Fprintln(w, "----")
	_, _ = fmt.Fprintf(w, "id: %d\n", log.ID)
	_, _ = fmt.Fprintf(w, "created_at: %s\n", log.CreatedAt.UTC().Format(time.RFC3339))
	_, _ = fmt.Fprintf(w, "session_id: %s\n", log.SessionID)
	_, _ = fmt.Fprintf(w, "request_id: %s\n", stringOrBlank(log.RequestID))
	_, _ = fmt.Fprintf(w, "client_request_id: %s\n", stringOrBlank(log.ClientRequestID))
	_, _ = fmt.Fprintf(w, "user_id: %s\n", int64OrBlank(log.UserID))
	_, _ = fmt.Fprintf(w, "api_key_id: %s\n", int64OrBlank(log.APIKeyID))
	_, _ = fmt.Fprintf(w, "account_id: %s\n", int64OrBlank(log.AccountID))
	_, _ = fmt.Fprintf(w, "group_id: %s\n", int64OrBlank(log.GroupID))
	_, _ = fmt.Fprintf(w, "platform: %s\n", stringOrBlank(log.Platform))
	_, _ = fmt.Fprintf(w, "model: %s\n", stringOrBlank(log.Model))
	_, _ = fmt.Fprintf(w, "request_path: %s\n", stringOrBlank(log.RequestPath))
	_, _ = fmt.Fprintf(w, "status_code: %s\n", intOrBlank(log.StatusCode))
	_, _ = fmt.Fprintf(w, "stream: %v\n", log.Stream)
	_, _ = fmt.Fprintln(w, "messages:")
	for _, msg := range log.Messages {
		_, _ = fmt.Fprintf(w, "[%d][%s][%s]\n", msg.Index, msg.Source, msg.Role)
		if msg.Content != "" {
			_, _ = fmt.Fprintln(w, msg.Content)
		}
		_, _ = fmt.Fprintln(w)
	}
	_, _ = fmt.Fprintln(w)
}

func stringOrBlank(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func int64OrBlank(v *int64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatInt(*v, 10)
}

func intOrBlank(v *int) string {
	if v == nil {
		return ""
	}
	return strconv.Itoa(*v)
}
