package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type DistributorHandler struct {
	svc *service.DistributorService
}

func NewDistributorHandler(svc *service.DistributorService) *DistributorHandler {
	return &DistributorHandler{svc: svc}
}

type upsertDistributorProfileRequest struct {
	Email   string `json:"email" binding:"required,email"`
	Enabled bool   `json:"enabled"`
	Notes   string `json:"notes" binding:"omitempty,max=500"`
}

type adjustDistributorBalanceRequest struct {
	Operation string `json:"operation" binding:"required,oneof=topup refund"`
	AmountCNY int64  `json:"amount_cny" binding:"required,gt=0"`
	Notes     string `json:"notes" binding:"omitempty,max=500"`
}

type upsertDistributorOfferRequest struct {
	DistributorUserID int64  `json:"distributor_user_id" binding:"required,gt=0"`
	Name              string `json:"name" binding:"required,max=128"`
	TargetGroupID     int64  `json:"target_group_id" binding:"required,gt=0"`
	ValidityDays      int    `json:"validity_days" binding:"required,gt=0,lte=36500"`
	CostCNY           int64  `json:"cost_cny" binding:"required,gt=0"`
	Enabled           bool   `json:"enabled"`
	Notes             string `json:"notes" binding:"omitempty,max=500"`
}

type settleRequest struct {
	Notes string `json:"notes" binding:"omitempty,max=500"`
}

type revokeOrderRequest struct {
	Notes string `json:"notes" binding:"omitempty,max=500"`
}

func (h *DistributorHandler) ListProfiles(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, pageResult, err := h.svc.ListProfiles(c.Request.Context(), pagination.PaginationParams{Page: page, PageSize: pageSize}, strings.TrimSpace(c.Query("search")))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, pageResult.Total, pageResult.Page, pageResult.PageSize)
}

func (h *DistributorHandler) UpsertProfile(c *gin.Context) {
	var req upsertDistributorProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.svc.UpsertProfileByEmail(c.Request.Context(), req.Email, req.Enabled, req.Notes)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *DistributorHandler) AdjustBalance(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userID"), 10, 64)
	if err != nil || userID <= 0 {
		response.BadRequest(c, "Invalid user id")
		return
	}
	var req adjustDistributorBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	subject, _ := middleware.GetAuthSubjectFromContext(c)
	item, err := h.svc.AdjustBalance(c.Request.Context(), userID, subject.UserID, req.Operation, req.AmountCNY, req.Notes)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *DistributorHandler) ListOffers(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Query("distributor_user_id"), 10, 64)
	if err != nil || userID <= 0 {
		response.BadRequest(c, "Invalid distributor_user_id")
		return
	}
	items, err := h.svc.ListOffers(c.Request.Context(), userID, false)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *DistributorHandler) CreateOffer(c *gin.Context) {
	var req upsertDistributorOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.svc.CreateOffer(c.Request.Context(), req.DistributorUserID, req.Name, req.TargetGroupID, req.ValidityDays, req.CostCNY, req.Notes, req.Enabled)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *DistributorHandler) UpdateOffer(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid offer id")
		return
	}
	var req upsertDistributorOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.svc.UpdateOffer(c.Request.Context(), id, req.Name, req.TargetGroupID, req.ValidityDays, req.CostCNY, req.Notes, req.Enabled)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *DistributorHandler) DeleteOffer(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid offer id")
		return
	}
	if err := h.svc.DeleteOffer(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *DistributorHandler) ListOrders(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	var userID int64
	if raw := strings.TrimSpace(c.Query("distributor_user_id")); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || v <= 0 {
			response.BadRequest(c, "Invalid distributor_user_id")
			return
		}
		userID = v
	}
	items, pageResult, err := h.svc.ListOrdersAdmin(c.Request.Context(), pagination.PaginationParams{Page: page, PageSize: pageSize}, userID, strings.TrimSpace(c.Query("status")), strings.TrimSpace(c.Query("search")))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, pageResult.Total, pageResult.Page, pageResult.PageSize)
}

func (h *DistributorHandler) RevokeOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || orderID <= 0 {
		response.BadRequest(c, "Invalid order id")
		return
	}
	var req revokeOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	subject, _ := middleware.GetAuthSubjectFromContext(c)
	item, err := h.svc.RevokeOrder(c.Request.Context(), subject.UserID, true, orderID, req.Notes)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *DistributorHandler) Summary(c *gin.Context) {
	item, err := h.svc.GetAdminSummary(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *DistributorHandler) MarkSettled(c *gin.Context) {
	var req settleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	subject, _ := middleware.GetAuthSubjectFromContext(c)
	item, err := h.svc.MarkSettled(c.Request.Context(), subject.UserID, req.Notes)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}
