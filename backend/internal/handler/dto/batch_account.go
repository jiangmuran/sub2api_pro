package dto

// BatchTestAccountsRequest represents the request to batch test accounts
type BatchTestAccountsRequest struct {
	AccountIDs     []int64 `json:"account_ids" binding:"required,min=1"`
	Model          string  `json:"model" binding:"required"`
	DelayMs        int     `json:"delay_ms" binding:"min=0,max=10000"`
	Concurrency    int     `json:"concurrency" binding:"required,min=1,max=20"`
	TimeoutSeconds int     `json:"timeout_seconds" binding:"required,min=5,max=60"`
}

// BatchTestResult represents the test result for a single account
type BatchTestResult struct {
	AccountID   int64  `json:"account_id"`
	AccountName string `json:"account_name"`
	Status      string `json:"status"` // "success" or "error"
	StatusCode  int    `json:"status_code,omitempty"`
	Error       string `json:"error,omitempty"`
	DurationMs  int64  `json:"duration_ms,omitempty"`
}

// BatchTestResponse represents the response for batch test
type BatchTestResponse struct {
	Results []BatchTestResult `json:"results"`
	Summary struct {
		Total   int `json:"total"`
		Success int `json:"success"`
		Failed  int `json:"failed"`
	} `json:"summary"`
}

// BatchTestProgressEvent represents a progress event (for SSE)
type BatchTestProgressEvent struct {
	Type      string           `json:"type"` // "progress" or "result"
	Completed int              `json:"completed,omitempty"`
	Total     int              `json:"total,omitempty"`
	Result    *BatchTestResult `json:"result,omitempty"`
}
