package service

import (
	"testing"
	"time"

	"github.com/tidwall/gjson"
)

func TestIsNanoBananaModel(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  bool
	}{
		{name: "nano model", model: "nano-banana-fast", want: true},
		{name: "mixed case", model: "Nano-Banana-Pro", want: true},
		{name: "other model", model: "gpt-image-1", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNanoBananaModel(tt.model); got != tt.want {
				t.Fatalf("IsNanoBananaModel(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}

func TestLookupOpenAIAccountStoredImagePricing_UsesNanoBananaKey(t *testing.T) {
	account := &Account{
		Platform: PlatformNanoBanana,
		Extra: map[string]any{
			NanoBananaManualPricingExtraKey: map[string]any{
				"nano-banana-fast": map[string]any{
					"image_price_per_image": 0.1234,
				},
			},
		},
	}

	price := lookupOpenAIAccountStoredImagePricing(account, "nano-banana-fast")
	if price != 0.1234 {
		t.Fatalf("lookupOpenAIAccountStoredImagePricing() = %v, want 0.1234", price)
	}
}

func TestParseNanoBananaGenerationRequest(t *testing.T) {
	body := []byte(`{"model":"nano-banana-fast","prompt":"draw a cat","aspect_ratio":"16:9","image_size":"2K","urls":["https://example.com/ref.png"]}`)
	request, err := parseNanoBananaGenerationRequest(body)
	if err != nil {
		t.Fatalf("parseNanoBananaGenerationRequest() error = %v", err)
	}
	if request.Model != "nano-banana-fast" {
		t.Fatalf("request.Model = %q, want nano-banana-fast", request.Model)
	}
	if request.Prompt != "draw a cat" {
		t.Fatalf("request.Prompt = %q, want draw a cat", request.Prompt)
	}
	if request.AspectRatio != "16:9" {
		t.Fatalf("request.AspectRatio = %q, want 16:9", request.AspectRatio)
	}
	if request.ImageSize != "2K" {
		t.Fatalf("request.ImageSize = %q, want 2K", request.ImageSize)
	}
	if len(request.URLs) != 1 || request.URLs[0] != "https://example.com/ref.png" {
		t.Fatalf("request.URLs = %#v, want reference url", request.URLs)
	}
}

func TestNormalizeNanoBananaBaseURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "host only", in: "https://grsaiapi.com", want: "https://grsaiapi.com"},
		{name: "v1 draw endpoint", in: "https://grsaiapi.com/v1/draw/nano-banana", want: "https://grsaiapi.com"},
		{name: "result endpoint", in: "https://grsaiapi.com/v1/draw/result", want: "https://grsaiapi.com"},
		{name: "v1 base", in: "https://grsaiapi.com/v1", want: "https://grsaiapi.com/v1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeNanoBananaBaseURL(tt.in); got != tt.want {
				t.Fatalf("normalizeNanoBananaBaseURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNanoBananaTaskTimeout(t *testing.T) {
	tests := []struct {
		name      string
		imageSize string
		want      time.Duration
	}{
		{name: "default 1k", imageSize: "", want: nanoBananaRequestTimeout1K},
		{name: "1k", imageSize: "1K", want: nanoBananaRequestTimeout1K},
		{name: "2k", imageSize: "2K", want: nanoBananaRequestTimeout2K},
		{name: "4k", imageSize: "4K", want: nanoBananaRequestTimeout4K},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nanoBananaTaskTimeout(tt.imageSize); got != tt.want {
				t.Fatalf("nanoBananaTaskTimeout(%q) = %s, want %s", tt.imageSize, got, tt.want)
			}
		})
	}
}

func TestNormalizeNanoBananaReferenceValue(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "url passthrough", in: "https://example.com/a.png", want: "https://example.com/a.png"},
		{name: "raw base64 passthrough", in: "YWJjMTIz", want: "YWJjMTIz"},
		{name: "data url strips prefix", in: "data:image/png;base64,YWJjMTIz", want: "YWJjMTIz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeNanoBananaReferenceValue(tt.in); got != tt.want {
				t.Fatalf("normalizeNanoBananaReferenceValue(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRewriteNanoBananaProgress(t *testing.T) {
	body := []byte(`{"code":0,"data":{"status":"running","start_time":1000,"progress":3}}`)
	rewritten := RewriteNanoBananaProgress(body, "4K", time.Unix(1060, 0))
	progress := gjson.GetBytes(rewritten, "data.progress").Int()
	if progress <= 3 || progress >= 100 {
		t.Fatalf("rewritten progress = %d, want synthetic progress between current and 100", progress)
	}

	doneBody := []byte(`{"code":0,"data":{"status":"succeeded","start_time":1000,"progress":80}}`)
	doneRewritten := RewriteNanoBananaProgress(doneBody, "2K", time.Unix(1005, 0))
	if done := gjson.GetBytes(doneRewritten, "data.progress").Int(); done != 100 {
		t.Fatalf("succeeded progress = %d, want 100", done)
	}
}
