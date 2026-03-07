package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const (
	OpenAILegacyProtocolKey         = "openai_legacy_protocol"
	OpenAILegacyProtocolChat        = "chat"
	OpenAILegacyProtocolCompletions = "completions"
)

func SetOpenAILegacyProtocol(c *gin.Context, protocol string) {
	if c == nil {
		return
	}
	protocol = strings.TrimSpace(protocol)
	if protocol == "" {
		return
	}
	c.Set(OpenAILegacyProtocolKey, protocol)
}

func GetOpenAILegacyProtocol(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Get(OpenAILegacyProtocolKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func NormalizeOpenAIResponsesBody(body []byte) ([]byte, bool, error) {
	if len(body) == 0 {
		return body, false, nil
	}
	if !gjson.ValidBytes(body) {
		return body, false, fmt.Errorf("invalid json body")
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return body, false, fmt.Errorf("parse request: %w", err)
	}

	changed := false

	if _, ok := payload["input"]; !ok {
		if v, ok := payload["messages"]; ok {
			payload["input"] = v
			delete(payload, "messages")
			changed = true
		} else if v, ok := payload["prompt"]; ok {
			payload["input"] = v
			delete(payload, "prompt")
			changed = true
		}
	}

	if v, ok := payload["input"]; ok {
		if items, coerced := coerceOpenAIInputToItems(v); coerced {
			payload["input"] = items
			changed = true
		} else {
			switch v.(type) {
			case []any:
				// ok
			default:
				payload["input"] = []any{v}
				changed = true
			}
		}
	}

	if _, exists := payload["stream_options"]; exists {
		delete(payload, "stream_options")
		changed = true
	}

	if _, exists := payload["user"]; exists {
		delete(payload, "user")
		changed = true
	}

	if normalizeOpenAICompatibilityPayload(payload) {
		changed = true
	}

	if !changed {
		return body, false, nil
	}

	normalized, err := json.Marshal(payload)
	if err != nil {
		return body, false, fmt.Errorf("serialize request: %w", err)
	}
	return normalized, true, nil
}

func ConvertOpenAILegacyRequestBody(body []byte, protocol string) ([]byte, error) {
	if len(body) == 0 {
		return body, nil
	}
	if !gjson.ValidBytes(body) {
		return body, fmt.Errorf("invalid json body")
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return body, fmt.Errorf("parse request: %w", err)
	}

	switch protocol {
	case OpenAILegacyProtocolChat:
		if v, ok := payload["messages"]; ok {
			payload["input"] = v
			delete(payload, "messages")
		}
	case OpenAILegacyProtocolCompletions:
		if v, ok := payload["prompt"]; ok {
			payload["input"] = v
			delete(payload, "prompt")
		}
	}

	if v, ok := payload["input"]; ok {
		if items, coerced := coerceOpenAIInputToItems(v); coerced {
			payload["input"] = items
		} else {
			switch v.(type) {
			case []any:
				// ok
			default:
				payload["input"] = []any{v}
			}
		}
	}

	if v, ok := payload["max_tokens"]; ok {
		if _, exists := payload["max_output_tokens"]; !exists {
			payload["max_output_tokens"] = v
		}
		delete(payload, "max_tokens")
	}

	if _, exists := payload["stream_options"]; exists {
		delete(payload, "stream_options")
	}
	if _, exists := payload["user"]; exists {
		delete(payload, "user")
	}

	_ = normalizeOpenAICompatibilityPayload(payload)

	converted, err := json.Marshal(payload)
	if err != nil {
		return body, fmt.Errorf("serialize request: %w", err)
	}
	return converted, nil
}

func normalizeOpenAICompatibilityPayload(payload map[string]any) bool {
	if payload == nil {
		return false
	}

	changed := false

	if _, exists := payload["stream_options"]; exists {
		delete(payload, "stream_options")
		changed = true
	}

	if rawEffort, exists := payload["reasoning_effort"]; exists {
		effort := ""
		if s, ok := rawEffort.(string); ok {
			effort = normalizeOpenAIReasoningEffort(s)
		}
		if effort != "" {
			reasoning, _ := payload["reasoning"].(map[string]any)
			if reasoning == nil {
				reasoning = map[string]any{}
				payload["reasoning"] = reasoning
			}
			if _, has := reasoning["effort"]; !has {
				reasoning["effort"] = effort
				changed = true
			}
		}
		delete(payload, "reasoning_effort")
		changed = true
	}

	if _, exists := payload["reasoningSummary"]; exists {
		delete(payload, "reasoningSummary")
		changed = true
	}

	if input, ok := payload["input"].([]any); ok {
		if normalizeOpenAIInputRoles(input) {
			changed = true
		}
		if normalizeOpenAIInputContentTypes(input) {
			changed = true
		}
		if normalizeOpenAIInputMessageNames(input) {
			changed = true
		}
	}

	return changed
}

// coerceOpenAIInputToItems converts legacy string input into Responses-compatible
// input items without changing existing map/array structures.
// It only handles string values (or arrays of strings) and leaves other shapes
// to the existing fallback logic.
func coerceOpenAIInputToItems(value any) ([]any, bool) {
	switch v := value.(type) {
	case []any:
		changed := false
		items := make([]any, len(v))
		for i, item := range v {
			if s, ok := item.(string); ok {
				items[i] = map[string]any{
					"type": "input_text",
					"text": s,
				}
				changed = true
				continue
			}
			items[i] = item
		}
		if !changed {
			return nil, false
		}
		return items, true
	case string:
		return []any{map[string]any{
			"type": "input_text",
			"text": v,
		}}, true
	default:
		return nil, false
	}
}

func normalizeOpenAIInputRoles(input []any) bool {
	changed := false
	for _, item := range input {
		msg, ok := item.(map[string]any)
		if !ok {
			continue
		}
		role, ok := msg["role"].(string)
		if !ok || !strings.EqualFold(strings.TrimSpace(role), "tool") {
			continue
		}

		msg["role"] = "user"
		changed = true

		toolCallID, _ := msg["tool_call_id"].(string)
		if content, ok := msg["content"].(string); ok {
			trimmedID := strings.TrimSpace(toolCallID)
			if trimmedID != "" {
				msg["content"] = fmt.Sprintf("[tool:%s] %s", trimmedID, content)
				changed = true
			}
		}

		if _, has := msg["tool_call_id"]; has {
			delete(msg, "tool_call_id")
			changed = true
		}
	}
	return changed
}

func normalizeOpenAIInputContentTypes(input []any) bool {
	changed := false
	for _, item := range input {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if _, exists := entry["cache_control"]; exists {
			delete(entry, "cache_control")
			changed = true
		}
		assistantRole := isAssistantInputRole(entry)

		if value, ok := entry["type"].(string); ok {
			normalizedValue := strings.ToLower(strings.TrimSpace(value))
			if normalizedValue == "text" {
				entry["type"] = normalizedInputContentType(assistantRole)
				changed = true
			} else if assistantRole && normalizedValue == "input_text" {
				entry["type"] = "output_text"
				changed = true
			} else if !assistantRole && normalizedValue == "output_text" {
				entry["type"] = "input_text"
				changed = true
			}
		}

		contents, ok := entry["content"].([]any)
		if !ok {
			continue
		}
		for _, contentItem := range contents {
			contentMap, ok := contentItem.(map[string]any)
			if !ok {
				continue
			}
			if _, exists := contentMap["cache_control"]; exists {
				delete(contentMap, "cache_control")
				changed = true
			}
			contentType, ok := contentMap["type"].(string)
			if !ok {
				continue
			}
			normalizedType := strings.ToLower(strings.TrimSpace(contentType))
			if normalizedType == "text" {
				contentMap["type"] = normalizedInputContentType(assistantRole)
				changed = true
			} else if assistantRole && normalizedType == "input_text" {
				contentMap["type"] = "output_text"
				changed = true
			} else if !assistantRole && normalizedType == "output_text" {
				contentMap["type"] = "input_text"
				changed = true
			}
		}
	}

	return changed
}

func isAssistantInputRole(entry map[string]any) bool {
	if entry == nil {
		return false
	}
	role, _ := entry["role"].(string)
	return strings.EqualFold(strings.TrimSpace(role), "assistant")
}

func normalizedInputContentType(assistantRole bool) string {
	if assistantRole {
		return "output_text"
	}
	return "input_text"
}

func normalizeOpenAIInputMessageNames(input []any) bool {
	changed := false
	for _, item := range input {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if _, hasName := entry["name"]; !hasName {
			continue
		}

		if shouldStripInputMessageName(entry) {
			delete(entry, "name")
			changed = true
		}
	}

	return changed
}

func shouldStripInputMessageName(entry map[string]any) bool {
	if entry == nil {
		return false
	}
	if _, hasRole := entry["role"]; hasRole {
		return true
	}
	itemType, _ := entry["type"].(string)
	normalizedType := strings.ToLower(strings.TrimSpace(itemType))
	return normalizedType == "message"
}

func ConvertOpenAIResponsesToLegacy(body []byte, protocol string, fallbackModel string) ([]byte, error) {
	if len(body) == 0 {
		return body, nil
	}
	if !gjson.ValidBytes(body) {
		return body, fmt.Errorf("invalid json response")
	}

	model := strings.TrimSpace(gjson.GetBytes(body, "model").String())
	if model == "" {
		model = strings.TrimSpace(fallbackModel)
	}
	id := strings.TrimSpace(gjson.GetBytes(body, "id").String())
	created := gjson.GetBytes(body, "created").Int()
	content := extractResponsesOutputText(body)
	if content == "" {
		content = ""
	}
	usage := buildLegacyUsage(body)

	switch protocol {
	case OpenAILegacyProtocolCompletions:
		payload := map[string]any{
			"id":      id,
			"object":  "text_completion",
			"created": created,
			"model":   model,
			"choices": []any{
				map[string]any{
					"index":         0,
					"text":          content,
					"finish_reason": "stop",
				},
			},
			"usage": usage,
		}
		return json.Marshal(payload)
	default:
		toolCalls := extractResponsesToolCalls(body)
		finishReason := "stop"
		if len(toolCalls) > 0 && content == "" {
			finishReason = "tool_calls"
		}
		message := map[string]any{
			"role":    "assistant",
			"content": content,
		}
		if len(toolCalls) > 0 {
			message["tool_calls"] = toolCalls
		}
		payload := map[string]any{
			"id":      id,
			"object":  "chat.completion",
			"created": created,
			"model":   model,
			"choices": []any{
				map[string]any{
					"index":         0,
					"message":       message,
					"finish_reason": finishReason,
				},
			},
			"usage": usage,
		}
		return json.Marshal(payload)
	}
}

func ConvertOpenAIResponsesSSEToLegacy(data string, protocol string, fallbackModel string) (string, bool) {
	trimmed := strings.TrimSpace(data)
	if trimmed == "" {
		return "", false
	}
	if trimmed == "[DONE]" {
		return "[DONE]", true
	}
	if !gjson.Valid(trimmed) {
		return "", false
	}

	model := strings.TrimSpace(gjson.Get(trimmed, "response.model").String())
	if model == "" {
		model = strings.TrimSpace(gjson.Get(trimmed, "model").String())
	}
	if model == "" {
		model = strings.TrimSpace(fallbackModel)
	}
	id := strings.TrimSpace(gjson.Get(trimmed, "response.id").String())
	if id == "" {
		id = strings.TrimSpace(gjson.Get(trimmed, "id").String())
	}
	created := gjson.Get(trimmed, "response.created").Int()
	if created == 0 {
		created = gjson.Get(trimmed, "created").Int()
	}

	eventType := strings.TrimSpace(gjson.Get(trimmed, "type").String())
	if eventType == "" {
		return "", false
	}

	if protocol == OpenAILegacyProtocolChat && eventType == "response.output_item.added" {
		itemType := strings.TrimSpace(gjson.Get(trimmed, "item.type").String())
		if itemType == "message" {
			choice := map[string]any{
				"index":         0,
				"finish_reason": "",
				"delta":         map[string]any{"role": "assistant"},
			}
			payload := map[string]any{
				"id":      id,
				"object":  "chat.completion.chunk",
				"created": created,
				"model":   model,
				"choices": []any{choice},
			}
			encoded, err := json.Marshal(payload)
			if err != nil {
				return "", false
			}
			return string(encoded), true
		}
	}

	var content string
	finishReason := ""
	if eventType == "response.output_text.delta" {
		content = gjson.Get(trimmed, "delta").String()
	} else if eventType == "response.output_text.done" {
		finishReason = "stop"
	} else if eventType == "response.completed" || eventType == "response.done" {
		return "[DONE]", true
	} else {
		return "", false
	}

	if protocol == OpenAILegacyProtocolCompletions {
		payload := map[string]any{
			"id":      id,
			"object":  "text_completion",
			"created": created,
			"model":   model,
			"choices": []any{
				map[string]any{
					"index":         0,
					"text":          content,
					"finish_reason": finishReason,
				},
			},
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return "", false
		}
		return string(encoded), true
	}

	choice := map[string]any{
		"index":         0,
		"finish_reason": finishReason,
		"delta":         map[string]any{},
	}
	if content != "" {
		choice["delta"] = map[string]any{"content": content}
	}
	payload := map[string]any{
		"id":      id,
		"object":  "chat.completion.chunk",
		"created": created,
		"model":   model,
		"choices": []any{choice},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", false
	}
	return string(encoded), true
}

func extractResponsesOutputText(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	if v := gjson.GetBytes(body, "output_text"); v.Exists() {
		return strings.TrimSpace(v.String())
	}
	output := gjson.GetBytes(body, "output")
	if !output.Exists() || !output.IsArray() {
		return ""
	}
	var builder strings.Builder
	for _, item := range output.Array() {
		if item.Get("type").String() != "message" {
			continue
		}
		contents := item.Get("content")
		if !contents.Exists() || !contents.IsArray() {
			continue
		}
		for _, c := range contents.Array() {
			typeVal := strings.TrimSpace(c.Get("type").String())
			if typeVal != "output_text" && typeVal != "text" {
				continue
			}
			text := c.Get("text").String()
			if text == "" {
				continue
			}
			builder.WriteString(text)
		}
	}
	return strings.TrimSpace(builder.String())
}

func extractResponsesToolCalls(body []byte) []any {
	if len(body) == 0 {
		return nil
	}
	output := gjson.GetBytes(body, "output")
	if !output.Exists() || !output.IsArray() {
		return nil
	}

	toolCalls := make([]any, 0)
	for _, item := range output.Array() {
		if strings.TrimSpace(item.Get("type").String()) != "function_call" {
			continue
		}
		name := strings.TrimSpace(item.Get("name").String())
		if name == "" {
			continue
		}
		arguments := item.Get("arguments").String()
		if strings.TrimSpace(arguments) == "" {
			arguments = "{}"
		}
		toolCallID := strings.TrimSpace(item.Get("call_id").String())
		if toolCallID == "" {
			toolCallID = strings.TrimSpace(item.Get("id").String())
		}
		if toolCallID == "" {
			toolCallID = "call_" + name
		}
		toolCalls = append(toolCalls, map[string]any{
			"id":   toolCallID,
			"type": "function",
			"function": map[string]any{
				"name":      name,
				"arguments": arguments,
			},
		})
	}

	if len(toolCalls) == 0 {
		return nil
	}
	return toolCalls
}

func buildLegacyUsage(body []byte) map[string]any {
	inputTokens := gjson.GetBytes(body, "usage.input_tokens").Int()
	outputTokens := gjson.GetBytes(body, "usage.output_tokens").Int()
	totalTokens := gjson.GetBytes(body, "usage.total_tokens").Int()
	if totalTokens == 0 {
		totalTokens = inputTokens + outputTokens
	}
	usage := map[string]any{}
	if inputTokens > 0 {
		usage["prompt_tokens"] = inputTokens
	}
	if outputTokens > 0 {
		usage["completion_tokens"] = outputTokens
	}
	if totalTokens > 0 {
		usage["total_tokens"] = totalTokens
	}
	return usage
}
