package handler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizePasswordResetBaseURL_FromAPIBaseURL(t *testing.T) {
	base, ok := normalizePasswordResetBaseURL("https://example.com/api/v1", true)
	require.True(t, ok)
	require.Equal(t, "https://example.com", base)

	base, ok = normalizePasswordResetBaseURL("https://example.com/console/api", true)
	require.True(t, ok)
	require.Equal(t, "https://example.com/console", base)

	base, ok = normalizePasswordResetBaseURL("https://example.com/portal/api/v1?source=admin", true)
	require.True(t, ok)
	require.Equal(t, "https://example.com/portal", base)
}

func TestNormalizePasswordResetBaseURL_KeepPathWhenNotAPIBase(t *testing.T) {
	base, ok := normalizePasswordResetBaseURL("https://example.com/app", false)
	require.True(t, ok)
	require.Equal(t, "https://example.com/app", base)
}

func TestNormalizePasswordResetBaseURL_InvalidInput(t *testing.T) {
	_, ok := normalizePasswordResetBaseURL("", true)
	require.False(t, ok)

	_, ok = normalizePasswordResetBaseURL("/relative", true)
	require.False(t, ok)
}

func TestTrimAPIPathSuffix(t *testing.T) {
	require.Equal(t, "", trimAPIPathSuffix(""))
	require.Equal(t, "", trimAPIPathSuffix("/api"))
	require.Equal(t, "", trimAPIPathSuffix("/api/v1"))
	require.Equal(t, "/panel", trimAPIPathSuffix("/panel/api"))
	require.Equal(t, "/panel", trimAPIPathSuffix("/panel/api/v1"))
	require.Equal(t, "/panel/v2", trimAPIPathSuffix("/panel/v2"))
}
