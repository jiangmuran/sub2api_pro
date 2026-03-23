package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

type probeHTTPUpstream struct {
	client *http.Client
}

func (u *probeHTTPUpstream) Do(req *http.Request, proxyURL string, accountID int64, accountConcurrency int) (*http.Response, error) {
	return u.client.Do(req)
}

func (u *probeHTTPUpstream) DoWithTLS(req *http.Request, proxyURL string, accountID int64, accountConcurrency int, enableTLSFingerprint bool) (*http.Response, error) {
	return u.client.Do(req)
}

func TestProbeOpenAICompatibleResponsesNative(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/responses":
			if r.Header.Get("Accept") == "text/event-stream" {
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = w.Write([]byte("data: {\"type\":\"response.completed\",\"response\":{\"usage\":{\"input_tokens\":1,\"output_tokens\":1}}}\n\ndata: [DONE]\n\n"))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"resp_1","object":"response","usage":{"input_tokens":1,"output_tokens":1}}`))
		case "/v1/chat/completions":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"chatcmpl_1","choices":[{"message":{"role":"assistant","content":"pong"}}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := server.Client()

	svc := NewAccountTestService(nil, nil, nil, &probeHTTPUpstream{client: client}, nil, &config.Config{})
	result, err := svc.ProbeOpenAICompatible(context.Background(), OpenAICompatibleProbeInput{
		BaseURL: server.URL,
		APIKey:  "sk-test",
	})
	if err != nil {
		t.Fatalf("ProbeOpenAICompatible error = %v", err)
	}
	if result.Status != OpenAICompatibleProbeStatusCompatible {
		t.Fatalf("status = %s, want %s", result.Status, OpenAICompatibleProbeStatusCompatible)
	}
	if result.RecommendedMode != OpenAICompatibleModeResponsesNative {
		t.Fatalf("recommended mode = %s", result.RecommendedMode)
	}
	if !result.Capabilities.Responses || !result.Capabilities.ResponsesStream || !result.Capabilities.ChatCompletions {
		t.Fatalf("unexpected capabilities: %+v", result.Capabilities)
	}
}

func TestProbeOpenAICompatibleChatFallback(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/responses":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"message":"responses endpoint not found"}}`))
		case "/v1/chat/completions":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"chatcmpl_1","choices":[{"message":{"role":"assistant","content":"pong"}}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := server.Client()

	svc := NewAccountTestService(nil, nil, nil, &probeHTTPUpstream{client: client}, nil, &config.Config{})
	result, err := svc.ProbeOpenAICompatible(context.Background(), OpenAICompatibleProbeInput{
		BaseURL: server.URL,
		APIKey:  "sk-test",
	})
	if err != nil {
		t.Fatalf("ProbeOpenAICompatible error = %v", err)
	}
	if result.Status != OpenAICompatibleProbeStatusLegacyOnly {
		t.Fatalf("status = %s, want %s", result.Status, OpenAICompatibleProbeStatusLegacyOnly)
	}
	if result.RecommendedMode != OpenAICompatibleModeChatCompletionsFallback {
		t.Fatalf("recommended mode = %s", result.RecommendedMode)
	}
	if result.SuggestedExtra["openai_passthrough"] != true {
		t.Fatalf("expected passthrough suggestion, got %#v", result.SuggestedExtra)
	}
	if !result.Capabilities.ChatCompletions || result.Capabilities.Responses {
		t.Fatalf("unexpected capabilities: %+v", result.Capabilities)
	}
}

func TestProbeOpenAICompatibleCompletionsFallback(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/responses":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"message":"responses endpoint not found"}}`))
		case "/v1/chat/completions":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"message":"chat completions endpoint not found"}}`))
		case "/v1/completions":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"cmpl_1","choices":[{"text":"pong"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := server.Client()

	svc := NewAccountTestService(nil, nil, nil, &probeHTTPUpstream{client: client}, nil, &config.Config{})
	result, err := svc.ProbeOpenAICompatible(context.Background(), OpenAICompatibleProbeInput{
		BaseURL: server.URL,
		APIKey:  "sk-test",
	})
	if err != nil {
		t.Fatalf("ProbeOpenAICompatible error = %v", err)
	}
	if result.Status != OpenAICompatibleProbeStatusLegacyOnly {
		t.Fatalf("status = %s, want %s", result.Status, OpenAICompatibleProbeStatusLegacyOnly)
	}
	if result.RecommendedMode != OpenAICompatibleModeCompletionsFallback {
		t.Fatalf("recommended mode = %s", result.RecommendedMode)
	}
	if !result.Capabilities.Completions || result.Capabilities.ChatCompletions || result.Capabilities.Responses {
		t.Fatalf("unexpected capabilities: %+v", result.Capabilities)
	}
}

func TestProbeOpenAICompatibleCustomVersionedBasePath(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/coding/v3/models":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"id":"kimi-k2.5","status":"Active"},{"id":"old-model","status":"Shutdown"}]}`))
		case "/api/coding/v3/responses":
			if r.Header.Get("Accept") == "text/event-stream" {
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = w.Write([]byte("data: {\"type\":\"response.completed\",\"response\":{\"usage\":{\"input_tokens\":1,\"output_tokens\":1}}}\n\ndata: [DONE]\n\n"))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"resp_1","object":"response","usage":{"input_tokens":1,"output_tokens":1}}`))
		case "/api/coding/v3/chat/completions":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"chatcmpl_1","choices":[{"message":{"role":"assistant","content":"pong"}}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := server.Client()

	svc := NewAccountTestService(nil, nil, nil, &probeHTTPUpstream{client: client}, nil, &config.Config{})
	result, err := svc.ProbeOpenAICompatible(context.Background(), OpenAICompatibleProbeInput{
		BaseURL: server.URL + "/api/coding/v3",
		APIKey:  "sk-test",
	})
	if err != nil {
		t.Fatalf("ProbeOpenAICompatible error = %v", err)
	}
	if result.Status != OpenAICompatibleProbeStatusCompatible {
		t.Fatalf("status = %s, want %s", result.Status, OpenAICompatibleProbeStatusCompatible)
	}
	if result.ProbeModel != "kimi-k2.5" {
		t.Fatalf("probe model = %s, want kimi-k2.5", result.ProbeModel)
	}
	if len(result.DiscoveredModels) == 0 || result.DiscoveredModels[0] != "kimi-k2.5" {
		t.Fatalf("unexpected discovered models: %+v", result.DiscoveredModels)
	}
	if result.Checks[0].EndpointURL != server.URL+"/api/coding/v3/responses" {
		t.Fatalf("unexpected responses endpoint: %s", result.Checks[0].EndpointURL)
	}
}

func TestPreviewOpenAICompatibleModels(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/coding/v3/models":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"id":"kimi-k2.5","status":"Active"},{"id":"glm-4.7","status":"Active"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := server.Client()
	billing := NewBillingService(&config.Config{}, nil)
	svc := NewAccountTestService(nil, nil, nil, &probeHTTPUpstream{client: client}, billing, &config.Config{})
	models, err := svc.PreviewOpenAICompatibleModels(context.Background(), OpenAICompatibleProbeInput{
		BaseURL: server.URL + "/api/coding/v3",
		APIKey:  "sk-test",
	}, 1.5, nil)
	if err != nil {
		t.Fatalf("PreviewOpenAICompatibleModels error = %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("models len = %d, want 2", len(models))
	}
	if models[0].ID != "glm-4.7" && models[0].ID != "kimi-k2.5" {
		t.Fatalf("unexpected model list: %+v", models)
	}
}

func TestPreviewOpenAICompatibleModels_IncludesImagePrice(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/coding/v3/models":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"id":"grok-imagine-1.0","status":"Active"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := server.Client()
	billing := NewBillingService(&config.Config{}, nil)
	billing.pricingService = NewPricingService(&config.Config{}, nil)
	billing.pricingService.pricingData = map[string]*LiteLLMModelPricing{
		"grok-imagine-1.0": {OutputCostPerImage: 0.12},
	}
	svc := NewAccountTestService(nil, nil, nil, &probeHTTPUpstream{client: client}, billing, &config.Config{})
	models, err := svc.PreviewOpenAICompatibleModels(context.Background(), OpenAICompatibleProbeInput{
		BaseURL: server.URL + "/api/coding/v3",
		APIKey:  "sk-test",
	}, 2, nil)
	if err != nil {
		t.Fatalf("PreviewOpenAICompatibleModels error = %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("models len = %d, want 1", len(models))
	}
	if models[0].ImagePricePerImage != 0.12 || models[0].AccountImagePricePerImage != 0.24 {
		t.Fatalf("unexpected image pricing: %+v", models[0])
	}
}

func TestAccountTestServiceLookupOpenRouterModelPricing(t *testing.T) {
	svc := NewAccountTestService(nil, nil, nil, &probeHTTPUpstream{client: http.DefaultClient}, nil, &config.Config{})
	svc.openRouterPricingCache = map[string]openRouterModelPricing{
		"kimi-k2.5":              {InputPerToken: 0.000001, OutputPerToken: 0.000002},
		"deepseek/deepseek-v3.2": {InputPerToken: 0.000003, OutputPerToken: 0.000004},
		"z-ai/glm-4.7":           {InputPerToken: 0.000005, OutputPerToken: 0.000006},
	}
	svc.openRouterPricingFetchedAt = time.Now()
	pricing, ok := svc.lookupOpenRouterModelPricing(context.Background(), "kimi-k2.5")
	if !ok {
		t.Fatal("expected pricing hit")
	}
	if pricing.InputPerToken != 0.000001 || pricing.OutputPerToken != 0.000002 {
		t.Fatalf("unexpected pricing: %+v", pricing)
	}
	pricing, ok = svc.lookupOpenRouterModelPricing(context.Background(), "deepseek-v3.2")
	if !ok || pricing.InputPerToken != 0.000003 {
		t.Fatalf("unexpected deepseek pricing: %+v ok=%v", pricing, ok)
	}
	pricing, ok = svc.lookupOpenRouterModelPricing(context.Background(), "glm-4.7")
	if !ok || pricing.InputPerToken != 0.000005 {
		t.Fatalf("unexpected glm pricing: %+v ok=%v", pricing, ok)
	}
}

func TestBuildOpenAIEndpointURL_PreservesCustomVersionedPath(t *testing.T) {
	if got := buildOpenAIEndpointURL("https://ark.cn-beijing.volces.com/api/coding/v3", "chat/completions"); got != "https://ark.cn-beijing.volces.com/api/coding/v3/chat/completions" {
		t.Fatalf("chat endpoint = %s", got)
	}
	if got := buildOpenAIEndpointURL("https://ark.cn-beijing.volces.com/api/coding/v3", "responses"); got != "https://ark.cn-beijing.volces.com/api/coding/v3/responses" {
		t.Fatalf("responses endpoint = %s", got)
	}
	if got := buildOpenAIEndpointURL("https://api.openai.com", "responses"); got != "https://api.openai.com/v1/responses" {
		t.Fatalf("default endpoint = %s", got)
	}
	if got := buildOpenAIEndpointURL("https://grsaiapi.com", "v1/draw/nano-banana"); got != "https://grsaiapi.com/v1/draw/nano-banana" {
		t.Fatalf("versioned endpoint = %s", got)
	}
	if got := buildOpenAIEndpointURL("https://grsaiapi.com/v1", "v1/draw/nano-banana"); got != "https://grsaiapi.com/v1/draw/nano-banana" {
		t.Fatalf("v1 base endpoint = %s", got)
	}
}
