package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitizeOpenCodeText_RewritesCanonicalSentence(t *testing.T) {
	in := "You are OpenCode, the best coding agent on the planet."
	got := sanitizeSystemText(in)
	require.Equal(t, strings.TrimSpace(claudeCodeSystemPrompt), got)
}

func TestSanitizeUpstreamErrorMessage_StripsOpenAIHelpCenterTemplate(t *testing.T) {
	in := "PastedAn error occurred while processing your request. You can retry your request, or contact us through our help center at help.openai.com if the error persists. Please include the request ID 68a9b3ab-6a86-4240-8eaa-6b00521cd2b6 in your message."
	got := sanitizeUpstreamErrorMessage(in)
	require.Empty(t, got)
}

func TestSanitizeUpstreamErrorMessage_StripsSystemReminderAndKeepsUsefulText(t *testing.T) {
	in := `<system-reminder>Your operational mode has changed from plan to build.</system-reminder> previous_response_not_found`
	got := sanitizeUpstreamErrorMessage(in)
	require.Equal(t, "previous_response_not_found", got)
}

func TestSanitizeUpstreamErrorMessage_MasksSensitiveQueryParams(t *testing.T) {
	in := "request failed: https://example.com?a=1&access_token=secret-123&refresh_token=rt-456"
	got := sanitizeUpstreamErrorMessage(in)
	require.Equal(t, "request failed: https://example.com?a=1&access_token=***&refresh_token=***", got)
}
