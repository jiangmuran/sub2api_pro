package service

import (
	"strings"

	openaitypes "github.com/Wei-Shaw/sub2api/internal/pkg/openai"
)

const (
	NanoBananaManualPricingExtraKey = "nano_banana_manual_model_pricing"
	NanoBananaDefaultBaseURL        = "https://grsaiapi.com"
)

var defaultNanoBananaModels = []openaitypes.Model{
	{ID: "nano-banana-2", Object: "model", Type: "model", DisplayName: "Nano Banana 2"},
	{ID: "nano-banana-2-cl", Object: "model", Type: "model", DisplayName: "Nano Banana 2 CL"},
	{ID: "nano-banana-2-4k-cl", Object: "model", Type: "model", DisplayName: "Nano Banana 2 4K CL"},
	{ID: "nano-banana-fast", Object: "model", Type: "model", DisplayName: "Nano Banana Fast"},
	{ID: "nano-banana", Object: "model", Type: "model", DisplayName: "Nano Banana"},
	{ID: "nano-banana-pro", Object: "model", Type: "model", DisplayName: "Nano Banana Pro"},
	{ID: "nano-banana-pro-vt", Object: "model", Type: "model", DisplayName: "Nano Banana Pro VT"},
	{ID: "nano-banana-pro-cl", Object: "model", Type: "model", DisplayName: "Nano Banana Pro CL"},
	{ID: "nano-banana-pro-vip", Object: "model", Type: "model", DisplayName: "Nano Banana Pro VIP"},
	{ID: "nano-banana-pro-4k-vip", Object: "model", Type: "model", DisplayName: "Nano Banana Pro 4K VIP"},
}

func DefaultNanoBananaModels() []openaitypes.Model {
	out := make([]openaitypes.Model, len(defaultNanoBananaModels))
	copy(out, defaultNanoBananaModels)
	return out
}

func DefaultNanoBananaModelIDs() []string {
	out := make([]string, 0, len(defaultNanoBananaModels))
	for _, model := range defaultNanoBananaModels {
		out = append(out, model.ID)
	}
	return out
}

func IsNanoBananaModel(model string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(model)), "nano-banana")
}

func manualPricingExtraKeyForPlatform(platform string) string {
	switch strings.TrimSpace(platform) {
	case PlatformNanoBanana:
		return NanoBananaManualPricingExtraKey
	default:
		return "openai_manual_model_pricing"
	}
}
