package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// ActivityHandler handles activity-related requests
type ActivityHandler struct {
	activityService *service.ActivityService
}

// NewActivityHandler creates a new ActivityHandler
func NewActivityHandler(activityService *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{
		activityService: activityService,
	}
}

// ListActivities handles getting user visible activities
// GET /api/v1/activities
func (h *ActivityHandler) ListActivities(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	activities, err := h.activityService.ListActivitiesForUser(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"activities": activities,
	})
}

// GetActivity handles getting activity detail
// GET /api/v1/activities/:id
func (h *ActivityHandler) GetActivity(c *gin.Context) {
	activityID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid activity ID")
		return
	}

	activity, err := h.activityService.GetActivityDetail(c.Request.Context(), activityID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, activity)
}

// ParticipateInActivity handles user participation in activity
// POST /api/v1/activities/:id/participate
func (h *ActivityHandler) ParticipateInActivity(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	activityID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid activity ID")
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	result, err := h.activityService.ParticipateInActivity(
		c.Request.Context(),
		subject.UserID,
		activityID,
		ipAddress,
		userAgent,
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, result)
}

// GetUserParticipations handles getting user participation history
// GET /api/v1/activities/participations
func (h *ActivityHandler) GetUserParticipations(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	participations, total, err := h.activityService.GetUserParticipations(
		c.Request.Context(),
		subject.UserID,
		limit,
		offset,
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"participations": participations,
		"total":          total,
		"limit":          limit,
		"offset":         offset,
	})
}
