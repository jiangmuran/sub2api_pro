package service

import (
	"encoding/json"
	"testing"

	"github.com/tidwall/gjson"
)

func TestNormalizeOpenAIResponsesBody_WrapInputString(t *testing.T) {
	body := []byte(`{"model":"gpt-4o","input":"hello"}`)
	normalized, changed, err := NormalizeOpenAIResponsesBody(body)
	if err != nil {
		t.Fatalf("NormalizeOpenAIResponsesBody error: %v", err)
	}
	if !changed {
		t.Fatalf("expected change")
	}
	if !gjson.GetBytes(normalized, "input").IsArray() {
		t.Fatalf("expected input array")
	}
}

func TestNormalizeOpenAIResponsesBody_UsesMessages(t *testing.T) {
	body := []byte(`{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`)
	normalized, changed, err := NormalizeOpenAIResponsesBody(body)
	if err != nil {
		t.Fatalf("NormalizeOpenAIResponsesBody error: %v", err)
	}
	if !changed {
		t.Fatalf("expected change")
	}
	if !gjson.GetBytes(normalized, "input").IsArray() {
		t.Fatalf("expected input array")
	}
	if gjson.GetBytes(normalized, "messages").Exists() {
		t.Fatalf("expected messages removed")
	}
}

func TestNormalizeOpenAIResponsesBody_RemovesStreamOptions(t *testing.T) {
	body := []byte(`{"model":"gpt-4o","input":[{"type":"input_text","text":"hi"}],"stream_options":{"include_usage":true}}`)
	normalized, changed, err := NormalizeOpenAIResponsesBody(body)
	if err != nil {
		t.Fatalf("NormalizeOpenAIResponsesBody error: %v", err)
	}
	if !changed {
		t.Fatalf("expected change")
	}
	if gjson.GetBytes(normalized, "stream_options").Exists() {
		t.Fatalf("expected stream_options removed")
	}
}

func TestNormalizeOpenAIResponsesBody_RemovesUser(t *testing.T) {
	body := []byte(`{"model":"gpt-4o","input":[{"type":"input_text","text":"hi"}],"user":"u_123"}`)
	normalized, changed, err := NormalizeOpenAIResponsesBody(body)
	if err != nil {
		t.Fatalf("NormalizeOpenAIResponsesBody error: %v", err)
	}
	if !changed {
		t.Fatalf("expected change")
	}
	if gjson.GetBytes(normalized, "user").Exists() {
		t.Fatalf("expected user removed")
	}
}

func TestNormalizeOpenAIResponsesBody_ConvertsReasoningEffort(t *testing.T) {
	body := []byte(`{"model":"gpt-4o","input":[{"type":"input_text","text":"hi"}],"reasoning_effort":"high"}`)
	normalized, changed, err := NormalizeOpenAIResponsesBody(body)
	if err != nil {
		t.Fatalf("NormalizeOpenAIResponsesBody error: %v", err)
	}
	if !changed {
		t.Fatalf("expected change")
	}
	if gjson.GetBytes(normalized, "reasoning_effort").Exists() {
		t.Fatalf("expected reasoning_effort removed")
	}
	if gjson.GetBytes(normalized, "reasoning.effort").String() != "high" {
		t.Fatalf("expected reasoning.effort=high")
	}
}

func TestNormalizeOpenAIResponsesBody_ConvertsToolRole(t *testing.T) {
	body := []byte(`{"model":"gpt-4o","input":[{"role":"tool","tool_call_id":"call_123","content":"result"}]}`)
	normalized, changed, err := NormalizeOpenAIResponsesBody(body)
	if err != nil {
		t.Fatalf("NormalizeOpenAIResponsesBody error: %v", err)
	}
	if !changed {
		t.Fatalf("expected change")
	}
	if gjson.GetBytes(normalized, "input.0.role").String() != "user" {
		t.Fatalf("expected role converted to user")
	}
	if gjson.GetBytes(normalized, "input.0.tool_call_id").Exists() {
		t.Fatalf("expected tool_call_id removed")
	}
	if gjson.GetBytes(normalized, "input.0.content").String() != "[tool:call_123] result" {
		t.Fatalf("unexpected converted content")
	}
}

func TestConvertOpenAILegacyRequestBody_Chat(t *testing.T) {
	body := []byte(`{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}],"max_tokens":64}`)
	converted, err := ConvertOpenAILegacyRequestBody(body, OpenAILegacyProtocolChat)
	if err != nil {
		t.Fatalf("ConvertOpenAILegacyRequestBody error: %v", err)
	}
	if !gjson.GetBytes(converted, "input").IsArray() {
		t.Fatalf("expected input array")
	}
	if gjson.GetBytes(converted, "messages").Exists() {
		t.Fatalf("expected messages removed")
	}
	if gjson.GetBytes(converted, "max_output_tokens").Int() != 64 {
		t.Fatalf("expected max_output_tokens=64")
	}
}

func TestConvertOpenAILegacyRequestBody_RemovesStreamOptions(t *testing.T) {
	body := []byte(`{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}],"stream_options":{"include_usage":true}}`)
	converted, err := ConvertOpenAILegacyRequestBody(body, OpenAILegacyProtocolChat)
	if err != nil {
		t.Fatalf("ConvertOpenAILegacyRequestBody error: %v", err)
	}
	if gjson.GetBytes(converted, "stream_options").Exists() {
		t.Fatalf("expected stream_options removed")
	}
}

func TestConvertOpenAILegacyRequestBody_ConvertsToolRoleAndReasoning(t *testing.T) {
	body := []byte(`{"model":"gpt-4o","messages":[{"role":"tool","content":"ok","tool_call_id":"call_9"}],"reasoning_effort":"medium"}`)
	converted, err := ConvertOpenAILegacyRequestBody(body, OpenAILegacyProtocolChat)
	if err != nil {
		t.Fatalf("ConvertOpenAILegacyRequestBody error: %v", err)
	}
	if gjson.GetBytes(converted, "input.0.role").String() != "user" {
		t.Fatalf("expected role converted to user")
	}
	if gjson.GetBytes(converted, "reasoning_effort").Exists() {
		t.Fatalf("expected reasoning_effort removed")
	}
	if gjson.GetBytes(converted, "reasoning.effort").String() != "medium" {
		t.Fatalf("expected reasoning.effort=medium")
	}
}

func TestConvertOpenAIResponsesToLegacy_Chat(t *testing.T) {
	body := []byte(`{
  "id":"resp_123",
  "model":"gpt-4o",
  "usage":{"input_tokens":10,"output_tokens":20},
  "output":[{"type":"message","content":[{"type":"output_text","text":"hello"}]}]
}`)
	converted, err := ConvertOpenAIResponsesToLegacy(body, OpenAILegacyProtocolChat, "gpt-4o")
	if err != nil {
		t.Fatalf("ConvertOpenAIResponsesToLegacy error: %v", err)
	}
	if gjson.GetBytes(converted, "object").String() != "chat.completion" {
		t.Fatalf("expected chat.completion object")
	}
	if gjson.GetBytes(converted, "choices.0.message.content").String() != "hello" {
		t.Fatalf("unexpected content")
	}
}

func TestConvertOpenAIResponsesSSEToLegacy_ChatDelta(t *testing.T) {
	data := `{"type":"response.output_text.delta","delta":"hi","response":{"id":"resp_1","model":"gpt-4o","created":123}}`
	converted, ok := ConvertOpenAIResponsesSSEToLegacy(data, OpenAILegacyProtocolChat, "gpt-4o")
	if !ok {
		t.Fatalf("expected conversion")
	}
	if !gjson.Valid(converted) {
		t.Fatalf("expected valid json")
	}
	if gjson.Get(converted, "object").String() != "chat.completion.chunk" {
		t.Fatalf("expected chat.completion.chunk")
	}
	if gjson.Get(converted, "choices.0.delta.content").String() != "hi" {
		t.Fatalf("unexpected delta content")
	}
}

func TestConvertOpenAIResponsesToLegacy_Completions(t *testing.T) {
	body := []byte(`{"id":"resp_123","model":"gpt-4o","usage":{"input_tokens":2,"output_tokens":3},"output_text":"ok"}`)
	converted, err := ConvertOpenAIResponsesToLegacy(body, OpenAILegacyProtocolCompletions, "gpt-4o")
	if err != nil {
		t.Fatalf("ConvertOpenAIResponsesToLegacy error: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(converted, &payload); err != nil {
		t.Fatalf("unmarshal converted: %v", err)
	}
	if payload["object"] != "text_completion" {
		t.Fatalf("expected text_completion object")
	}
}
