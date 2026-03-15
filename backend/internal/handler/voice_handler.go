package handler

import (
	"context"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type VoiceHandler struct {
	voiceService        *service.VoiceChatService
	apiKeyService       *service.APIKeyService
	subscriptionService *service.SubscriptionService
	billingCacheService *service.BillingCacheService
	billingService      *service.BillingService
	gatewayService      *service.OpenAIGatewayService
}

type voiceSessionRequest struct {
	APIKey      string  `json:"api_key"`
	Voice       string  `json:"voice"`
	Personality string  `json:"personality"`
	Speed       float64 `json:"speed"`
}

func NewVoiceHandler(
	voiceService *service.VoiceChatService,
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	billingCacheService *service.BillingCacheService,
	billingService *service.BillingService,
	gatewayService *service.OpenAIGatewayService,
) *VoiceHandler {
	return &VoiceHandler{
		voiceService:        voiceService,
		apiKeyService:       apiKeyService,
		subscriptionService: subscriptionService,
		billingCacheService: billingCacheService,
		billingService:      billingService,
		gatewayService:      gatewayService,
	}
}

func (h *VoiceHandler) Preflight(c *gin.Context) {
	ctx := c.Request.Context()
	var req voiceSessionRequest
	_ = c.ShouldBindJSON(&req)
	apiKeyValue := strings.TrimSpace(req.APIKey)
	if apiKeyValue == "" {
		apiKeyValue = strings.TrimSpace(c.Query("api_key"))
	}
	if apiKeyValue == "" {
		response.BadRequest(c, "api_key is required")
		return
	}
	apiKey, subscription, account, err := h.resolveVoiceRequest(ctx, c, apiKeyValue)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	check, err := h.voiceService.Preflight(ctx, account)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	basePrice := h.lookupVoiceSinglePrice(account)
	response.Success(c, gin.H{
		"model":                 service.OpenAIModelGrokLivechat,
		"single_price_per_call": basePrice,
		"account_id":            account.ID,
		"group_id":              apiKey.GroupID,
		"subscription_mode":     apiKey.Group != nil && apiKey.Group.IsSubscriptionType() && subscription != nil,
		"function_ready":        check.FunctionReady,
		"server_livekit_ready":  check.LivekitReady,
		"livekit_probe_url":     check.LivekitProbeURL,
	})
}

func (h *VoiceHandler) CreateSession(c *gin.Context) {
	ctx := c.Request.Context()
	var req voiceSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if strings.TrimSpace(req.APIKey) == "" {
		response.BadRequest(c, "api_key is required")
		return
	}
	apiKey, subscription, account, err := h.resolveVoiceRequest(ctx, c, req.APIKey)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	voice := strings.TrimSpace(req.Voice)
	if voice == "" {
		voice = "ara"
	}
	personality := strings.TrimSpace(req.Personality)
	if personality == "" {
		personality = "assistant"
	}
	speed := req.Speed
	if speed <= 0 {
		speed = 1.0
	}
	start := time.Now()
	session, err := h.voiceService.CreateSessionToken(ctx, account, voice, personality, speed)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if err := h.gatewayService.RecordUsage(ctx, &service.OpenAIRecordUsageInput{
		Result: &service.OpenAIForwardResult{
			Model:      service.OpenAIModelGrokLivechat,
			ImageCount: 1,
			MediaType:  "voice",
			Duration:   time.Since(start),
		},
		APIKey:        apiKey,
		User:          apiKey.User,
		Account:       account,
		Subscription:  subscription,
		UserAgent:     c.GetHeader("User-Agent"),
		IPAddress:     ip.GetClientIP(c),
		APIKeyService: h.apiKeyService,
	}); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"model":                 service.OpenAIModelGrokLivechat,
		"single_price_per_call": h.lookupVoiceSinglePrice(account),
		"token":                 session.Token,
		"url":                   session.URL,
		"participant_name":      session.ParticipantName,
		"room_name":             session.RoomName,
	})
}

func (h *VoiceHandler) resolveVoiceRequest(ctx context.Context, c *gin.Context, apiKeyValue string) (*service.APIKey, *service.UserSubscription, *service.Account, error) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		return nil, nil, nil, service.ErrInsufficientPerms
	}
	apiKeyValue = strings.TrimSpace(apiKeyValue)
	if apiKeyValue == "" {
		return nil, nil, nil, service.ErrAPIKeyNotFound
	}
	apiKey, err := h.apiKeyService.GetByKey(ctx, apiKeyValue)
	if err != nil {
		return nil, nil, nil, err
	}
	if apiKey.UserID != subject.UserID {
		return nil, nil, nil, service.ErrInsufficientPerms
	}
	if !apiKey.IsActive() || apiKey.IsExpired() || apiKey.IsQuotaExhausted() {
		return nil, nil, nil, service.ErrAPIKeyQuotaExhausted
	}
	var subscription *service.UserSubscription
	if apiKey.Group != nil && apiKey.Group.IsSubscriptionType() && h.subscriptionService != nil {
		subscription, err = h.subscriptionService.GetActiveSubscription(ctx, apiKey.User.ID, apiKey.Group.ID)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	if h.billingCacheService != nil {
		if err := h.billingCacheService.CheckBillingEligibility(ctx, apiKey.User, apiKey, apiKey.Group, subscription); err != nil {
			return nil, nil, nil, err
		}
	}
	account, err := h.gatewayService.SelectAccountForModel(ctx, apiKey.GroupID, "", service.OpenAIModelGrokLivechat)
	if err != nil {
		return nil, nil, nil, err
	}
	return apiKey, subscription, account, nil
}

func (h *VoiceHandler) lookupVoiceSinglePrice(account *service.Account) float64 {
	if pricing, ok := lookupManualModelPricing(account, service.OpenAIModelGrokLivechat); ok {
		return pricing.ImagePricePerImage
	}
	if h.billingService != nil {
		if pricing := h.billingService.GetPreviewModelPricing(service.OpenAIModelGrokLivechat); pricing != nil {
			return pricing.OutputPricePerImage
		}
	}
	return 0
}
