package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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

	svc := NewAccountTestService(nil, nil, nil, &probeHTTPUpstream{client: client}, &config.Config{})
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

	svc := NewAccountTestService(nil, nil, nil, &probeHTTPUpstream{client: client}, &config.Config{})
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
