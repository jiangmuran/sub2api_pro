package handler

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type DistributorHandler struct {
	svc *service.DistributorService
}

func NewDistributorHandler(svc *service.DistributorService) *DistributorHandler {
	return &DistributorHandler{svc: svc}
}

type createDistributorOrderRequest struct {
	OfferID      int64  `json:"offer_id" binding:"required,gt=0"`
	SellPriceCNY int64  `json:"sell_price_cny" binding:"omitempty,gte=0"`
	Memo         string `json:"memo" binding:"omitempty,max=500"`
}

type revokeDistributorOrderRequest struct {
	Notes string `json:"notes" binding:"omitempty,max=500"`
}

func (h *DistributorHandler) GetProfile(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Authentication required")
		return
	}
	item, err := h.svc.GetProfileByUserID(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *DistributorHandler) ListOffers(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Authentication required")
		return
	}
	items, err := h.svc.ListOffers(c.Request.Context(), subject.UserID, true)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *DistributorHandler) CreateOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Authentication required")
		return
	}
	var req createDistributorOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.svc.PurchaseOrder(c.Request.Context(), subject.UserID, req.OfferID, req.SellPriceCNY, req.Memo)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *DistributorHandler) ListOrders(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Authentication required")
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, pageResult, err := h.svc.ListOrdersForDistributor(
		c.Request.Context(),
		subject.UserID,
		pagination.PaginationParams{Page: page, PageSize: pageSize},
		strings.TrimSpace(c.Query("status")),
		strings.TrimSpace(c.Query("search")),
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, pageResult.Total, pageResult.Page, pageResult.PageSize)
}

func (h *DistributorHandler) RevokeOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Authentication required")
		return
	}
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || orderID <= 0 {
		response.BadRequest(c, "Invalid order id")
		return
	}
	var req revokeDistributorOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.svc.RevokeOrder(c.Request.Context(), subject.UserID, false, orderID, req.Notes)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}
