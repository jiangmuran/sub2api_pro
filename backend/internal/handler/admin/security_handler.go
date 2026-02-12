package admin

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type SecurityHandler struct {
	chatService *service.SecurityChatService
	aiService   *service.SecurityChatAIService
}

func NewSecurityHandler(chatService *service.SecurityChatService, aiService *service.SecurityChatAIService) *SecurityHandler {
	return &SecurityHandler{chatService: chatService, aiService: aiService}
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
		StartTime string  `json:"start_time"`
		EndTime   string  `json:"end_time"`
		UserID    *int64  `json:"user_id"`
		APIKeyID  *int64  `json:"api_key_id"`
		SessionID *string `json:"session_id"`
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
		StartTime: &startTime,
		EndTime:   &endTime,
		UserID:    req.UserID,
		APIKeyID:  req.APIKeyID,
		PageSize:  500,
		AllowEmptySession: true,
	}
	if req.SessionID != nil {
		filter.SessionID = strings.TrimSpace(*req.SessionID)
	}

	logs, _, err := h.chatService.ListMessages(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if len(logs.Items) == 0 {
		response.BadRequest(c, "No chat logs found")
		return
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
	}, subject.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, result)
}
