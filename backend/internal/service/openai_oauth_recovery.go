package service

import "time"

const (
	OpenAIOAuthRefreshCooldown    = 10 * time.Minute
	OpenAIOAuthRefreshMaxAttempts = 2
)
