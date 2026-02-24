package service

import (
	"net/url"
	"strings"
)

const redeemCategoryPrefix = "[category:"

func EncodeRedeemNotes(note, category string) string {
	note = strings.TrimSpace(note)
	category = strings.TrimSpace(category)
	if category == "" {
		return note
	}
	prefix := redeemCategoryPrefix + url.QueryEscape(category) + "]"
	if note == "" {
		return prefix
	}
	return prefix + "\n" + note
}

func DecodeRedeemNotes(raw string) (note, category string) {
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.HasPrefix(raw, redeemCategoryPrefix) {
		return raw, ""
	}
	end := strings.Index(raw, "]")
	if end <= len(redeemCategoryPrefix) {
		return raw, ""
	}
	encodedCategory := raw[len(redeemCategoryPrefix):end]
	decodedCategory, err := url.QueryUnescape(encodedCategory)
	if err != nil {
		return raw, ""
	}
	remaining := strings.TrimSpace(raw[end+1:])
	return remaining, strings.TrimSpace(decodedCategory)
}
