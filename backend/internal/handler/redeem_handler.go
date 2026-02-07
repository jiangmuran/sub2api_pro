package handler

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RedeemHandler handles redeem code-related requests
type RedeemHandler struct {
	redeemService *service.RedeemService
}

// NewRedeemHandler creates a new RedeemHandler
func NewRedeemHandler(redeemService *service.RedeemService) *RedeemHandler {
	return &RedeemHandler{
		redeemService: redeemService,
	}
}

// RedeemRequest represents the redeem code request payload
type RedeemRequest struct {
	Code string `json:"code" binding:"required"`
}

// RedeemResponse represents the redeem response
type RedeemResponse struct {
	Message        string   `json:"message"`
	Type           string   `json:"type"`
	Value          float64  `json:"value"`
	NewBalance     *float64 `json:"new_balance,omitempty"`
	NewConcurrency *int     `json:"new_concurrency,omitempty"`
}

// DailyCheckinStatusResponse represents daily check-in status payload.
type DailyCheckinStatusResponse struct {
	Enabled        bool     `json:"enabled"`
	CheckedInToday bool     `json:"checked_in_today"`
	RewardMin      float64  `json:"reward_min"`
	RewardMax      float64  `json:"reward_max"`
	RewardAmount   *float64 `json:"reward_amount,omitempty"`
}

// DailyCheckinResponse represents daily check-in result payload.
type DailyCheckinResponse struct {
	Message      string  `json:"message"`
	RewardAmount float64 `json:"reward_amount"`
	NewBalance   float64 `json:"new_balance"`
	CheckedInAt  string  `json:"checked_in_at"`
}

// Redeem handles redeeming a code
// POST /api/v1/redeem
func (h *RedeemHandler) Redeem(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req RedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	result, err := h.redeemService.Redeem(c.Request.Context(), subject.UserID, req.Code)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.RedeemCodeFromService(result))
}

// GetDailyCheckinStatus 获取每日签到状态
// GET /api/v1/redeem/checkin/status
func (h *RedeemHandler) GetDailyCheckinStatus(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	status, err := h.redeemService.GetDailyCheckinStatus(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, DailyCheckinStatusResponse{
		Enabled:        status.Enabled,
		CheckedInToday: status.CheckedInToday,
		RewardMin:      status.RewardMin,
		RewardMax:      status.RewardMax,
		RewardAmount:   status.RewardAmount,
	})
}

// DailyCheckin 执行每日签到
// POST /api/v1/redeem/checkin
func (h *RedeemHandler) DailyCheckin(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	result, err := h.redeemService.DailyCheckin(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, DailyCheckinResponse{
		Message:      result.Message,
		RewardAmount: result.RewardAmount,
		NewBalance:   result.NewBalance,
		CheckedInAt:  result.CheckedInAt.Format(time.RFC3339),
	})
}

// GetHistory returns the user's redemption history
// GET /api/v1/redeem/history
func (h *RedeemHandler) GetHistory(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Default limit is 25
	limit := 25

	codes, err := h.redeemService.GetUserHistory(c.Request.Context(), subject.UserID, limit)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.RedeemCode, 0, len(codes))
	for i := range codes {
		out = append(out, *dto.RedeemCodeFromService(&codes[i]))
	}
	response.Success(c, out)
}
