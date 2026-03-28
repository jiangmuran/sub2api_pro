package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
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

func ConvertOpenAIResponsesRequestToLegacy(body []byte, protocol string) ([]byte, error) {
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

	legacy := map[string]any{}
	for key, value := range payload {
		switch key {
		case "input", "max_output_tokens", "store", "previous_response_id", "reasoning", "text", "include", "parallel_tool_calls", "prompt_cache_key":
			continue
		default:
			legacy[key] = value
		}
	}

	if value, ok := payload["max_output_tokens"]; ok {
		legacy["max_tokens"] = value
	}

	if stream, ok := payload["stream"].(bool); ok && stream {
		legacy["stream_options"] = map[string]any{"include_usage": true}
	}

	if protocol == OpenAILegacyProtocolCompletions {
		delete(legacy, "tools")
		delete(legacy, "tool_choice")
		delete(legacy, "parallel_tool_calls")
		legacy["prompt"] = extractResponsesPromptValue(payload["input"])
		return json.Marshal(legacy)
	}

	legacy["messages"] = extractResponsesMessagesValue(payload["input"])
	return json.Marshal(legacy)
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

	if normalizeOpenAITools(payload) {
		changed = true
	}

	return changed
}

func normalizeOpenAITools(payload map[string]any) bool {
	toolsRaw, ok := payload["tools"].([]any)
	if !ok {
		return false
	}
	cleaned := make([]any, 0, len(toolsRaw))
	for _, tool := range toolsRaw {
		toolMap, ok := tool.(map[string]any)
		if !ok {
			continue
		}
		toolType, _ := toolMap["type"].(string)
		if strings.EqualFold(strings.TrimSpace(toolType), "function") {
			// Check if it's ChatCompletions format with function wrapper
			functionMap, ok := toolMap["function"].(map[string]any)
			if ok {
				// ChatCompletions format: {type: "function", function: {...}}
				name, _ := functionMap["name"].(string)
				if strings.TrimSpace(name) == "" {
					continue
				}
				cleaned = append(cleaned, map[string]any{
					"type":     "function",
					"function": functionMap,
				})
				continue
			}
			// Responses format: {type: "function", name: "...", ...}
			// Just validate it has a name and keep as-is
			name, _ := toolMap["name"].(string)
			if strings.TrimSpace(name) != "" {
				cleaned = append(cleaned, toolMap)
				continue
			}
		}
		cleaned = append(cleaned, toolMap)
	}
	if len(cleaned) == 0 {
		delete(payload, "tools")
		if _, exists := payload["tool_choice"]; exists {
			delete(payload, "tool_choice")
		}
		return true
	}
	if len(cleaned) == len(toolsRaw) {
		return false
	}
	payload["tools"] = cleaned
	return true
}

// StripToolCallingFields removes tool-related fields for APIs that don't support them.
// Converts tool role messages to user role and removes tool_call_id.
// Returns true if any modifications were made.
func StripToolCallingFields(payload map[string]any) bool {
	if payload == nil {
		return false
	}

	modified := false

	// Remove tools array if present
	if _, hasTools := payload["tools"]; hasTools {
		delete(payload, "tools")
		modified = true
	}

	// Remove tool_choice if present
	if _, hasToolChoice := payload["tool_choice"]; hasToolChoice {
		delete(payload, "tool_choice")
		modified = true
	}

	// Process messages array
	messagesRaw, ok := payload["messages"].([]any)
	if !ok || len(messagesRaw) == 0 {
		return modified
	}

	for _, msgItem := range messagesRaw {
		msg, ok := msgItem.(map[string]any)
		if !ok {
			continue
		}

		role, _ := msg["role"].(string)
		role = strings.TrimSpace(strings.ToLower(role))

		// Convert tool role to user role
		if role == "tool" {
			msg["role"] = "user"
			modified = true

			// Prepend tool_call_id to content if present
			toolCallID, _ := msg["tool_call_id"].(string)
			toolCallID = strings.TrimSpace(toolCallID)
			if toolCallID != "" {
				if content, ok := msg["content"].(string); ok {
					msg["content"] = fmt.Sprintf("[Tool Response: %s] %s", toolCallID, content)
				}
			}

			// Remove tool_call_id field
			if _, hasID := msg["tool_call_id"]; hasID {
				delete(msg, "tool_call_id")
				modified = true
			}
		}

		// Remove tool_calls from assistant messages
		if role == "assistant" {
			if _, hasToolCalls := msg["tool_calls"]; hasToolCalls {
				delete(msg, "tool_calls")
				modified = true
			}
		}
	}

	return modified
}

// ConvertResponsesToChatCompletionsTools converts Responses-format tools to ChatCompletions format
// Responses format: {type: "function", name: "...", parameters: {...}}
// ChatCompletions format: {type: "function", function: {name: "...", parameters: {...}}}
func ConvertResponsesToChatCompletionsTools(payload map[string]any) bool {
	toolsRaw, ok := payload["tools"].([]any)
	if !ok || len(toolsRaw) == 0 {
		return false
	}

	modified := false
	converted := make([]any, 0, len(toolsRaw))

	for _, tool := range toolsRaw {
		toolMap, ok := tool.(map[string]any)
		if !ok {
			converted = append(converted, tool)
			continue
		}

		toolType, _ := toolMap["type"].(string)
		if !strings.EqualFold(strings.TrimSpace(toolType), "function") {
			converted = append(converted, toolMap)
			continue
		}

		// If already has function wrapper, keep as-is
		if _, hasFn := toolMap["function"]; hasFn {
			converted = append(converted, toolMap)
			continue
		}

		// Check if it's Responses format (has name at top level)
		name, hasName := toolMap["name"].(string)
		if !hasName || strings.TrimSpace(name) == "" {
			// Invalid tool, skip
			continue
		}

		// Convert Responses format to ChatCompletions format
		functionObj := map[string]any{
			"name": name,
		}

		// Move description to function object
		if desc, ok := toolMap["description"]; ok {
			functionObj["description"] = desc
		}

		// Move parameters to function object
		if params, ok := toolMap["parameters"]; ok {
			functionObj["parameters"] = params
		}

		// Move strict to function object
		if strict, ok := toolMap["strict"]; ok {
			functionObj["strict"] = strict
		}

		converted = append(converted, map[string]any{
			"type":     "function",
			"function": functionObj,
		})
		modified = true
	}

	if modified {
		payload["tools"] = converted
	}

	return modified
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
	// DISABLED: This conversion breaks tool calling for APIs that support it natively
	// The original purpose was to make tool messages compatible with APIs that don't support them,
	// but it was applied universally and broke OpenAI-compatible APIs that DO support tool role.
	// TODO: Only apply this conversion when targeting APIs that don't support tool role
	return false

	/*
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
	*/
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
		if sanitized, changed := sanitizeOpenAIOutputText(content); changed {
			content = sanitized
			if strings.TrimSpace(content) == "" {
				return "", false
			}
		}
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

func ConvertOpenAILegacyResponseToResponses(body []byte, protocol string, fallbackModel string) ([]byte, error) {
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
	if created == 0 {
		created = time.Now().Unix()
	}

	content := ""
	toolCalls := []any(nil)
	if protocol == OpenAILegacyProtocolCompletions {
		content = gjson.GetBytes(body, "choices.0.text").String()
	} else {
		content = gjson.GetBytes(body, "choices.0.message.content").String()
		toolCalls = extractLegacyToolCalls(body)
	}
	if sanitized, changed := sanitizeOpenAIOutputText(content); changed {
		content = sanitized
	}

	usage := extractLegacyUsageAsResponses(body)

	output := make([]any, 0, 1+len(toolCalls))
	messageItem := map[string]any{
		"id":      responseIDWithSuffix(id, "msg_1"),
		"type":    "message",
		"role":    "assistant",
		"status":  "completed",
		"content": []any{},
	}
	if strings.TrimSpace(content) != "" {
		messageItem["content"] = []any{map[string]any{"type": "output_text", "text": content}}
		output = append(output, messageItem)
	}
	for _, toolCall := range toolCalls {
		output = append(output, toolCall)
	}
	if len(output) == 0 {
		output = append(output, messageItem)
	}

	response := map[string]any{
		"id":          id,
		"object":      "response",
		"created":     created,
		"model":       model,
		"output":      output,
		"output_text": content,
		"usage":       usage,
	}
	return json.Marshal(response)
}

func ConvertOpenAILegacySSEToResponses(data string, protocol string, fallbackModel string) (string, bool) {
	trimmed := strings.TrimSpace(data)
	if trimmed == "" {
		return "", false
	}
	if trimmed == "[DONE]" {
		payload := map[string]any{
			"type": "response.completed",
			"response": map[string]any{
				"id":          responseIDWithSuffix("", "resp_fallback"),
				"object":      "response",
				"created":     time.Now().Unix(),
				"model":       strings.TrimSpace(fallbackModel),
				"output":      []any{},
				"output_text": "",
				"usage":       map[string]any{},
			},
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return "", false
		}
		return string(encoded), true
	}
	if !gjson.Valid(trimmed) {
		return "", false
	}

	model := strings.TrimSpace(gjson.Get(trimmed, "model").String())
	if model == "" {
		model = strings.TrimSpace(fallbackModel)
	}
	id := strings.TrimSpace(gjson.Get(trimmed, "id").String())
	created := gjson.Get(trimmed, "created").Int()
	if created == 0 {
		created = time.Now().Unix()
	}

	if protocol == OpenAILegacyProtocolCompletions {
		text := gjson.Get(trimmed, "choices.0.text").String()
		if sanitized, changed := sanitizeOpenAIOutputText(text); changed {
			text = sanitized
		}
		if text == "" {
			return "", false
		}
		payload := map[string]any{
			"type":  "response.output_text.delta",
			"delta": text,
			"response": map[string]any{
				"id":      id,
				"created": created,
				"model":   model,
			},
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return "", false
		}
		return string(encoded), true
	}

	content := gjson.Get(trimmed, "choices.0.delta.content").String()
	if sanitized, changed := sanitizeOpenAIOutputText(content); changed {
		content = sanitized
	}
	if content != "" {
		payload := map[string]any{
			"type":  "response.output_text.delta",
			"delta": content,
			"response": map[string]any{
				"id":      id,
				"created": created,
				"model":   model,
			},
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return "", false
		}
		return string(encoded), true
	}
	return "", false
}

type openAILegacyResponsesStreamState struct {
	Protocol      string
	FallbackModel string
	ResponseID    string
	Model         string
	Created       int64
	OutputText    strings.Builder
	Usage         map[string]any
}

func (s *openAILegacyResponsesStreamState) Convert(data string) (string, bool) {
	trimmed := strings.TrimSpace(data)
	if trimmed == "" {
		return "", false
	}
	if trimmed == "[DONE]" {
		payload := map[string]any{
			"type": "response.completed",
			"response": map[string]any{
				"id":          s.responseID(),
				"object":      "response",
				"created":     s.createdAt(),
				"model":       s.modelName(),
				"output":      s.outputItems(),
				"output_text": s.OutputText.String(),
				"usage":       s.usageMap(),
			},
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return "", false
		}
		return string(encoded), true
	}
	if !gjson.Valid(trimmed) {
		return "", false
	}

	s.captureMetadata(trimmed)
	s.captureUsage(trimmed)
	content := ""
	if s.Protocol == OpenAILegacyProtocolCompletions {
		content = gjson.Get(trimmed, "choices.0.text").String()
	} else {
		content = gjson.Get(trimmed, "choices.0.delta.content").String()
	}
	if content != "" {
		s.OutputText.WriteString(content)
		payload := map[string]any{
			"type":  "response.output_text.delta",
			"delta": content,
			"response": map[string]any{
				"id":      s.responseID(),
				"created": s.createdAt(),
				"model":   s.modelName(),
			},
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return "", false
		}
		return string(encoded), true
	}
	finishReason := strings.TrimSpace(gjson.Get(trimmed, "choices.0.finish_reason").String())
	if finishReason != "" {
		payload := map[string]any{
			"type": "response.output_text.done",
			"text": s.OutputText.String(),
			"response": map[string]any{
				"id":      s.responseID(),
				"created": s.createdAt(),
				"model":   s.modelName(),
			},
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return "", false
		}
		return string(encoded), true
	}
	return "", false
}

func (s *openAILegacyResponsesStreamState) captureMetadata(payload string) {
	if id := strings.TrimSpace(gjson.Get(payload, "id").String()); id != "" {
		s.ResponseID = id
	}
	if model := strings.TrimSpace(gjson.Get(payload, "model").String()); model != "" {
		s.Model = model
	}
	if created := gjson.Get(payload, "created").Int(); created > 0 {
		s.Created = created
	}
}

func (s *openAILegacyResponsesStreamState) captureUsage(payload string) {
	usage := gjson.Get(payload, "usage")
	if !usage.Exists() || usage.Type != gjson.JSON || usage.Raw == "" || usage.Raw == "null" {
		return
	}
	inputTokens := gjson.Get(payload, "usage.prompt_tokens").Int()
	if inputTokens == 0 {
		inputTokens = gjson.Get(payload, "usage.input_tokens").Int()
	}
	outputTokens := gjson.Get(payload, "usage.completion_tokens").Int()
	if outputTokens == 0 {
		outputTokens = gjson.Get(payload, "usage.output_tokens").Int()
	}
	totalTokens := gjson.Get(payload, "usage.total_tokens").Int()
	usageMap := map[string]any{}
	if inputTokens > 0 {
		usageMap["input_tokens"] = inputTokens
	}
	if outputTokens > 0 {
		usageMap["output_tokens"] = outputTokens
	}
	if totalTokens > 0 {
		usageMap["total_tokens"] = totalTokens
	}
	if len(usageMap) > 0 {
		s.Usage = usageMap
	}
}

func extractLegacyUsageAsResponses(body []byte) map[string]any {
	inputTokens := gjson.GetBytes(body, "usage.prompt_tokens").Int()
	if inputTokens == 0 {
		inputTokens = gjson.GetBytes(body, "usage.input_tokens").Int()
	}
	outputTokens := gjson.GetBytes(body, "usage.completion_tokens").Int()
	if outputTokens == 0 {
		outputTokens = gjson.GetBytes(body, "usage.output_tokens").Int()
	}
	totalTokens := gjson.GetBytes(body, "usage.total_tokens").Int()
	usage := map[string]any{}
	if inputTokens > 0 {
		usage["input_tokens"] = inputTokens
	}
	if outputTokens > 0 {
		usage["output_tokens"] = outputTokens
	}
	if totalTokens > 0 {
		usage["total_tokens"] = totalTokens
	}
	return usage
}

func (s *openAILegacyResponsesStreamState) responseID() string {
	return responseIDWithSuffix(s.ResponseID, "resp_fallback")
}

func (s *openAILegacyResponsesStreamState) modelName() string {
	if strings.TrimSpace(s.Model) != "" {
		return strings.TrimSpace(s.Model)
	}
	return strings.TrimSpace(s.FallbackModel)
}

func (s *openAILegacyResponsesStreamState) createdAt() int64 {
	if s.Created > 0 {
		return s.Created
	}
	return time.Now().Unix()
}

func (s *openAILegacyResponsesStreamState) outputItems() []any {
	content := s.OutputText.String()
	return []any{map[string]any{
		"id":     responseIDWithSuffix(s.ResponseID+"_msg", "msg_1"),
		"type":   "message",
		"role":   "assistant",
		"status": "completed",
		"content": []any{map[string]any{
			"type": "output_text",
			"text": content,
		}},
	}}
}

func (s *openAILegacyResponsesStreamState) usageMap() map[string]any {
	if len(s.Usage) == 0 {
		return map[string]any{}
	}
	return s.Usage
}

func extractResponsesMessagesValue(input any) []map[string]any {
	items, ok := input.([]any)
	if !ok || len(items) == 0 {
		return []map[string]any{{"role": "user", "content": ""}}
	}
	messages := make([]map[string]any, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		role, _ := entry["role"].(string)
		role = strings.TrimSpace(role)
		if role == "" {
			role = "user"
		} else {
			switch strings.ToLower(role) {
			case "developer":
				role = "system"
			}
		}

		// Build message with role and content
		msg := map[string]any{
			"role":    role,
			"content": extractInputTextContent(entry["content"]),
		}

		// Preserve tool_call_id for tool role messages (required by some APIs like Volcengine)
		if strings.ToLower(role) == "tool" {
			if toolCallID, ok := entry["tool_call_id"]; ok {
				msg["tool_call_id"] = toolCallID
			}
		}

		// Preserve tool_calls for assistant messages
		if strings.ToLower(role) == "assistant" {
			if toolCalls, ok := entry["tool_calls"]; ok {
				msg["tool_calls"] = toolCalls
			}
		}

		// Preserve name field if present (for function/tool messages)
		if name, ok := entry["name"]; ok {
			msg["name"] = name
		}

		messages = append(messages, msg)
	}
	if len(messages) == 0 {
		return []map[string]any{{"role": "user", "content": ""}}
	}
	return messages
}

func extractResponsesPromptValue(input any) string {
	messages := extractResponsesMessagesValue(input)
	parts := make([]string, 0, len(messages))
	for _, msg := range messages {
		parts = append(parts, strings.TrimSpace(fmt.Sprintf("%v", msg["content"])))
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func extractInputTextContent(content any) string {
	if value, ok := content.(string); ok {
		return value
	}
	items, ok := content.([]any)
	if !ok {
		return ""
	}
	var builder strings.Builder
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		typeValue, _ := entry["type"].(string)
		if typeValue != "input_text" && typeValue != "text" && typeValue != "output_text" {
			continue
		}
		text, _ := entry["text"].(string)
		builder.WriteString(text)
	}
	return builder.String()
}

func extractLegacyToolCalls(body []byte) []any {
	toolCalls := gjson.GetBytes(body, "choices.0.message.tool_calls")
	if !toolCalls.Exists() || !toolCalls.IsArray() {
		return nil
	}
	result := make([]any, 0, len(toolCalls.Array()))
	for _, item := range toolCalls.Array() {
		name := strings.TrimSpace(item.Get("function.name").String())
		if name == "" {
			continue
		}
		arguments := item.Get("function.arguments").String()
		if arguments == "" {
			arguments = "{}"
		}
		result = append(result, map[string]any{
			"id":        responseIDWithSuffix(item.Get("id").String(), "call_"+name),
			"type":      "function_call",
			"call_id":   responseIDWithSuffix(item.Get("id").String(), "call_"+name),
			"name":      name,
			"arguments": arguments,
		})
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func responseIDWithSuffix(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return fallback
}

func extractResponsesOutputText(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	if v := gjson.GetBytes(body, "output_text"); v.Exists() {
		text := strings.TrimSpace(v.String())
		if sanitized, changed := sanitizeOpenAIOutputText(text); changed {
			// Only replace if sanitized text is not empty (avoid clearing valid content)
			if strings.TrimSpace(sanitized) != "" {
				text = sanitized
			}
		}
		return strings.TrimSpace(text)
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
			if sanitized, changed := sanitizeOpenAIOutputText(text); changed {
				// Only replace if sanitized text is not empty (avoid clearing valid content)
				if strings.TrimSpace(sanitized) != "" {
					text = sanitized
				}
			}
			if strings.TrimSpace(text) == "" {
				continue
			}
			builder.WriteString(text)
		}
	}
	return strings.TrimSpace(builder.String())
}

func sanitizeOpenAIOutputText(text string) (string, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return text, false
	}
	changed := false
	if strings.Contains(strings.ToLower(trimmed), "<system-reminder>") {
		sanitized := stripSystemReminderBlocks(trimmed)
		changed = sanitized != trimmed
		trimmed = sanitized
		if trimmed == "" {
			return trimmed, true
		}
	}
	if !looksLikePollutedOpenAIOutput(trimmed) {
		return trimmed, changed
	}
	sanitized := sanitizeUpstreamErrorMessage(trimmed)
	return sanitized, changed || sanitized != text
}

func looksLikePollutedOpenAIOutput(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}
	// Only detect OpenAI's specific error messages to avoid false positives with domestic APIs
	hasErrorPhrase := strings.Contains(lower, "an error occurred while processing your request")
	hasHelpCenter := strings.Contains(lower, "help.openai.com")
	hasRequestID := strings.Contains(lower, "request id") || strings.Contains(lower, "request_id")
	hasPastedPrefix := strings.HasPrefix(lower, "pasted") || strings.HasPrefix(lower, "[pasted]")

	// Case 1: Explicit OpenAI error message
	if hasErrorPhrase {
		return true
	}
	// Case 2: OpenAI help center link with request ID (both must be present)
	// This is more strict to avoid false positives with APIs that include request_id in normal responses
	if hasHelpCenter && hasRequestID {
		return true
	}
	// Case 3: Pasted error format with both help center AND request ID
	// Changed from OR to AND to be more conservative
	if hasPastedPrefix && hasHelpCenter && hasRequestID {
		return true
	}
	return false
}

func sanitizeOpenAIResponseBodyBytes(body []byte) []byte {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return body
	}
	updated := body
	if outputText := gjson.GetBytes(updated, "output_text"); outputText.Exists() {
		if sanitized, changed := sanitizeOpenAIOutputText(outputText.String()); changed {
			if next, err := sjson.SetBytes(updated, "output_text", sanitized); err == nil {
				updated = next
			}
		}
	}
	output := gjson.GetBytes(updated, "output")
	if output.Exists() && output.IsArray() {
		for itemIndex, item := range output.Array() {
			if strings.TrimSpace(item.Get("type").String()) != "message" {
				continue
			}
			content := item.Get("content")
			if !content.Exists() || !content.IsArray() {
				continue
			}
			for contentIndex, contentItem := range content.Array() {
				typeVal := strings.TrimSpace(contentItem.Get("type").String())
				if typeVal != "output_text" && typeVal != "text" {
					continue
				}
				text := contentItem.Get("text").String()
				if sanitized, changed := sanitizeOpenAIOutputText(text); changed {
					path := fmt.Sprintf("output.%d.content.%d.text", itemIndex, contentIndex)
					if next, err := sjson.SetBytes(updated, path, sanitized); err == nil {
						updated = next
					}
				}
			}
		}
	}
	return updated
}

func openAIResponseBodyHasVisibleOutput(body []byte) bool {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return false
	}
	if strings.TrimSpace(gjson.GetBytes(body, "output_text").String()) != "" {
		return true
	}
	output := gjson.GetBytes(body, "output")
	if !output.Exists() || !output.IsArray() {
		return false
	}
	for _, item := range output.Array() {
		if strings.TrimSpace(item.Get("type").String()) != "message" {
			continue
		}
		contents := item.Get("content")
		if !contents.Exists() || !contents.IsArray() {
			continue
		}
		for _, content := range contents.Array() {
			typeVal := strings.TrimSpace(content.Get("type").String())
			if typeVal != "output_text" && typeVal != "text" {
				continue
			}
			if strings.TrimSpace(content.Get("text").String()) != "" {
				return true
			}
		}
	}
	return false
}

func openAIResponseBodyHasToolLikeOutput(body []byte) bool {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return false
	}
	output := gjson.GetBytes(body, "output")
	if !output.Exists() || !output.IsArray() {
		return false
	}
	for _, item := range output.Array() {
		typeVal := strings.TrimSpace(item.Get("type").String())
		switch typeVal {
		case "function_call", "computer_call", "custom_tool_call", "code_interpreter_call", "file_search_call", "web_search_call", "tool_search_call", "local_shell_call", "apply_patch_call":
			return true
		}
	}
	return false
}

func sanitizeOpenAIResponseEventBytes(message []byte) ([]byte, bool, bool) {
	if len(message) == 0 || !gjson.ValidBytes(message) {
		return message, false, false
	}
	eventType := strings.TrimSpace(gjson.GetBytes(message, "type").String())
	switch eventType {
	case "response.output_text.delta":
		delta := gjson.GetBytes(message, "delta").String()
		sanitized, changed := sanitizeOpenAIOutputText(delta)
		if !changed {
			return message, false, false
		}
		if strings.TrimSpace(sanitized) == "" {
			return nil, true, true
		}
		next, err := sjson.SetBytes(message, "delta", sanitized)
		if err != nil {
			return message, false, false
		}
		return next, true, false
	case "response.completed", "response.done":
		response := gjson.GetBytes(message, "response")
		if !response.Exists() || response.Type != gjson.JSON || response.Raw == "" {
			return message, false, false
		}
		originalResponse := []byte(response.Raw)
		sanitized := sanitizeOpenAIResponseBodyBytes(originalResponse)
		if !openAIResponseBodyHasVisibleOutput(sanitized) && openAIResponseBodyHasVisibleOutput(originalResponse) {
			return nil, true, true
		}
		if string(sanitized) == response.Raw {
			return message, false, false
		}
		next, err := sjson.SetRawBytes(message, "response", sanitized)
		if err != nil {
			return message, false, false
		}
		return next, true, false
	default:
		return message, false, false
	}
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
