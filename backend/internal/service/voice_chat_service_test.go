package service

import (
	"context"
	"testing"
)

func TestVoiceChatServicePreflight(t *testing.T) {
	service := NewVoiceChatService(nil, nil, nil)
	result, err := service.Preflight(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FunctionReady {
		t.Fatalf("expected functionReady=false for nil account")
	}
	if !result.LivekitReady {
		t.Fatalf("expected livekitReady=true")
	}
	if result.LivekitProbeURL == "" {
		t.Fatalf("expected livekit probe url")
	}
}
