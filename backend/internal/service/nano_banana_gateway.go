package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	nanoBananaDrawPath          = "v1/draw/nano-banana"
	nanoBananaResultPath        = "v1/draw/result"
	nanoBananaPollInterval      = 2 * time.Second
	nanoBananaRequestTimeout1K  = 4 * time.Minute
	nanoBananaRequestTimeout2K  = 6 * time.Minute
	nanoBananaRequestTimeout4K  = 10 * time.Minute
	nanoBananaDefaultAspectRate = "auto"
	nanoBananaDefaultImageSize  = "1K"
)

type nanoBananaSubmitRequest struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	AspectRatio string   `json:"aspectRatio,omitempty"`
	ImageSize   string   `json:"imageSize,omitempty"`
	URLs        []string `json:"urls,omitempty"`
	WebHook     string   `json:"webHook,omitempty"`
}

type nanoBananaSubmitEnvelope struct {
	Code int `json:"code"`
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type nanoBananaResultEnvelope struct {
	Code int                  `json:"code"`
	Data nanoBananaTaskResult `json:"data"`
	Msg  string               `json:"msg"`
}

type nanoBananaTaskResult struct {
	CallbackURL   string                  `json:"callback_url"`
	StartTime     int64                   `json:"start_time"`
	EndTime       int64                   `json:"end_time"`
	ID            string                  `json:"id"`
	Results       []nanoBananaImageResult `json:"results"`
	Progress      int                     `json:"progress"`
	Status        string                  `json:"status"`
	FailureReason string                  `json:"failure_reason"`
	Error         string                  `json:"error"`
}

type nanoBananaImageResult struct {
	URL     string `json:"url"`
	Content string `json:"content"`
}

type NanoBananaRawResponse struct {
	StatusCode  int
	ContentType string
	Headers     http.Header
	Body        []byte
}

func (s *OpenAIGatewayService) ForwardNanoBananaImageGeneration(ctx context.Context, c *gin.Context, account *Account, body []byte) (*OpenAIForwardResult, error) {
	request, err := parseNanoBananaGenerationRequest(body)
	if err != nil {
		writeNanoBananaGatewayError(c, http.StatusBadRequest, err.Error())
		return nil, err
	}
	return s.forwardNanoBananaRequest(ctx, c, account, request)
}

func (s *OpenAIGatewayService) ForwardNanoBananaImageEdits(ctx context.Context, c *gin.Context, account *Account) (*OpenAIForwardResult, error) {
	request, err := parseNanoBananaEditRequest(c)
	if err != nil {
		writeNanoBananaGatewayError(c, http.StatusBadRequest, err.Error())
		return nil, err
	}
	return s.forwardNanoBananaRequest(ctx, c, account, request)
}

func (s *OpenAIGatewayService) forwardNanoBananaRequest(ctx context.Context, c *gin.Context, account *Account, request nanoBananaSubmitRequest) (*OpenAIForwardResult, error) {
	start := time.Now()
	if account == nil || !account.IsNanoBanana() {
		err := fmt.Errorf("nano banana account is required")
		writeNanoBananaGatewayError(c, http.StatusBadRequest, err.Error())
		return nil, err
	}

	taskTimeout := nanoBananaTaskTimeout(request.ImageSize)
	requestCtx, cancel := context.WithTimeout(ctx, taskTimeout)
	defer cancel()

	result, err := s.submitAndPollNanoBanana(requestCtx, account, request, taskTimeout)
	if err != nil {
		var failoverErr *UpstreamFailoverError
		if errorsAsUpstreamFailover(err, &failoverErr) {
			return nil, failoverErr
		}
		if errors.Is(err, context.DeadlineExceeded) {
			err = fmt.Errorf("nano banana task timed out after %s", taskTimeout.Round(time.Second))
		}
		writeNanoBananaGatewayError(c, http.StatusBadGateway, err.Error())
		return nil, err
	}

	urls := make([]gin.H, 0, len(result.Results))
	for _, item := range result.Results {
		url := strings.TrimSpace(item.URL)
		if url == "" {
			continue
		}
		urls = append(urls, gin.H{"url": url})
	}
	if len(urls) == 0 {
		err := fmt.Errorf("nano banana returned no images")
		writeNanoBananaGatewayError(c, http.StatusBadGateway, err.Error())
		return nil, err
	}

	c.JSON(http.StatusOK, gin.H{
		"created": time.Now().Unix(),
		"data":    urls,
	})

	return &OpenAIForwardResult{
		Model:                request.Model,
		Duration:             time.Since(start),
		EstimatedInputTokens: len(request.Prompt) / 4,
		ImageCount:           len(urls),
		ImageSize:            normalizeNanoBananaImageSize(request.ImageSize),
		MediaType:            "image",
	}, nil
}

func (s *OpenAIGatewayService) submitAndPollNanoBanana(ctx context.Context, account *Account, request nanoBananaSubmitRequest, taskTimeout time.Duration) (*nanoBananaTaskResult, error) {
	request.WebHook = "-1"
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	baseURL, err := s.validateUpstreamBaseURL(account.GetNanoBananaBaseURL())
	if err != nil {
		return nil, err
	}
	baseURL = normalizeNanoBananaBaseURL(baseURL)
	token := account.GetNanoBananaAPIKey()
	if token == "" {
		return nil, fmt.Errorf("api_key not found in credentials")
	}

	submitURL := buildOpenAIEndpointURL(baseURL, nanoBananaDrawPath)
	submitRespBody, err := s.doNanoBananaJSONRequest(ctx, account, submitURL, token, body)
	if err != nil {
		return nil, err
	}

	var submitResp nanoBananaSubmitEnvelope
	if err := json.Unmarshal(submitRespBody, &submitResp); err != nil {
		return nil, fmt.Errorf("parse nano banana submit response: %w", err)
	}
	if submitResp.Code != 0 || strings.TrimSpace(submitResp.Data.ID) == "" {
		return nil, fmt.Errorf("%s", strings.TrimSpace(firstNonEmpty(submitResp.Msg, extractNanoBananaErrorMessage(submitRespBody), "nano banana task creation failed")))
	}

	resultURL := buildOpenAIEndpointURL(baseURL, nanoBananaResultPath)
	deadline := time.Now().Add(taskTimeout)
	for {
		resultBody, pollErr := s.doNanoBananaJSONRequest(ctx, account, resultURL, token, []byte(`{"id":"`+submitResp.Data.ID+`"}`))
		if pollErr != nil {
			return nil, pollErr
		}
		var resultResp nanoBananaResultEnvelope
		if err := json.Unmarshal(resultBody, &resultResp); err != nil {
			return nil, fmt.Errorf("parse nano banana result response: %w", err)
		}
		if resultResp.Code != 0 {
			return nil, fmt.Errorf("%s", strings.TrimSpace(firstNonEmpty(resultResp.Msg, extractNanoBananaErrorMessage(resultBody), "nano banana result query failed")))
		}
		switch strings.ToLower(strings.TrimSpace(resultResp.Data.Status)) {
		case "succeeded":
			return &resultResp.Data, nil
		case "failed":
			return nil, fmt.Errorf("%s", strings.TrimSpace(firstNonEmpty(resultResp.Data.Error, resultResp.Data.FailureReason, "nano banana task failed")))
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("nano banana task timed out")
		}
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return nil, fmt.Errorf("nano banana task timed out after %s", taskTimeout.Round(time.Second))
			}
			return nil, ctx.Err()
		case <-time.After(nanoBananaPollInterval):
		}
	}
}

func (s *OpenAIGatewayService) ForwardNanoBananaDrawPassthrough(ctx context.Context, c *gin.Context, account *Account, body []byte) error {
	if account == nil || !account.IsNanoBanana() {
		return fmt.Errorf("nano banana account is required")
	}
	baseURL, err := s.validateUpstreamBaseURL(account.GetNanoBananaBaseURL())
	if err != nil {
		return err
	}
	baseURL = normalizeNanoBananaBaseURL(baseURL)
	token := account.GetNanoBananaAPIKey()
	if token == "" {
		return fmt.Errorf("api_key not found in credentials")
	}
	targetURL := buildOpenAIEndpointURL(baseURL, nanoBananaDrawPath)
	resp, err := s.doNanoBananaHTTPRequest(ctx, c, account, targetURL, token, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	writeNanoBananaStreamResponse(c, resp)
	return nil
}

func (s *OpenAIGatewayService) CreateNanoBananaTask(ctx context.Context, account *Account, body []byte) (*NanoBananaRawResponse, error) {
	if account == nil || !account.IsNanoBanana() {
		return nil, fmt.Errorf("nano banana account is required")
	}
	baseURL, err := s.validateUpstreamBaseURL(account.GetNanoBananaBaseURL())
	if err != nil {
		return nil, err
	}
	baseURL = normalizeNanoBananaBaseURL(baseURL)
	token := account.GetNanoBananaAPIKey()
	if token == "" {
		return nil, fmt.Errorf("api_key not found in credentials")
	}
	targetURL := buildOpenAIEndpointURL(baseURL, nanoBananaDrawPath)
	return s.doNanoBananaRawJSONRequest(ctx, account, targetURL, token, body)
}

func (s *OpenAIGatewayService) ForwardNanoBananaResultPassthrough(ctx context.Context, c *gin.Context, groupID *int64, body []byte) error {
	resp, err := s.QueryNanoBananaResult(ctx, groupID, body)
	if err != nil {
		return err
	}
	WriteNanoBananaRawResponse(c, resp)
	return nil
}

func (s *OpenAIGatewayService) QueryNanoBananaResult(ctx context.Context, groupID *int64, body []byte) (*NanoBananaRawResponse, error) {
	accounts, err := s.listSchedulableAccountsByPlatform(ctx, groupID, PlatformNanoBanana)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, fmt.Errorf("no available %s accounts", PlatformNanoBanana)
	}
	var lastResp *NanoBananaRawResponse
	var lastErr error
	for i := range accounts {
		account := &accounts[i]
		if !account.IsSchedulable() {
			continue
		}
		baseURL, baseErr := s.validateUpstreamBaseURL(account.GetNanoBananaBaseURL())
		if baseErr != nil {
			lastErr = baseErr
			continue
		}
		baseURL = normalizeNanoBananaBaseURL(baseURL)
		token := account.GetNanoBananaAPIKey()
		if token == "" {
			lastErr = fmt.Errorf("api_key not found in credentials")
			continue
		}
		targetURL := buildOpenAIEndpointURL(baseURL, nanoBananaResultPath)
		resp, reqErr := s.doNanoBananaRawJSONRequest(ctx, account, targetURL, token, body)
		if reqErr != nil {
			lastErr = reqErr
			continue
		}
		lastResp = resp
		if !isNanoBananaTaskNotFound(resp.Body) {
			return resp, nil
		}
	}
	if lastResp != nil {
		return lastResp, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("nano banana task not found")
}

func nanoBananaTaskTimeout(imageSize string) time.Duration {
	switch normalizeNanoBananaImageSize(imageSize) {
	case "4K":
		return nanoBananaRequestTimeout4K
	case "2K":
		return nanoBananaRequestTimeout2K
	default:
		return nanoBananaRequestTimeout1K
	}
}

func nanoBananaProgressEstimate(imageSize string) time.Duration {
	switch normalizeNanoBananaImageSize(imageSize) {
	case "4K":
		return 2 * time.Minute
	default:
		return time.Minute
	}
}

func RewriteNanoBananaProgress(body []byte, imageSize string, now time.Time) []byte {
	if len(body) == 0 {
		return body
	}
	status := strings.ToLower(strings.TrimSpace(gjson.GetBytes(body, "data.status").String()))
	if status == "succeeded" {
		if next, err := sjson.SetBytes(body, "data.progress", 100); err == nil {
			return next
		}
		return body
	}
	if status == "failed" {
		return body
	}
	startTime := gjson.GetBytes(body, "data.start_time").Int()
	if startTime <= 0 {
		return body
	}
	elapsed := now.Sub(time.Unix(startTime, 0))
	if elapsed <= 0 {
		return body
	}
	estimated := nanoBananaProgressEstimate(imageSize)
	if estimated <= 0 {
		return body
	}
	synthetic := int(math.Round(math.Min(95, (elapsed.Seconds()/estimated.Seconds())*100)))
	current := int(gjson.GetBytes(body, "data.progress").Int())
	if synthetic < current {
		synthetic = current
	}
	if synthetic <= current {
		return body
	}
	next, err := sjson.SetBytes(body, "data.progress", synthetic)
	if err != nil {
		return body
	}
	return next
}

func normalizeNanoBananaBaseURL(baseURL string) string {
	normalized := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	for _, suffix := range []string{"/v1/draw/nano-banana", "/v1/draw/result"} {
		if strings.HasSuffix(normalized, suffix) {
			return strings.TrimSuffix(normalized, suffix)
		}
	}
	return normalized
}

func (s *OpenAIGatewayService) doNanoBananaHTTPRequest(ctx context.Context, c *gin.Context, account *Account, targetURL, token string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Set("content-type", "application/json")
	if c != nil {
		for key, values := range c.Request.Header {
			lowerKey := strings.ToLower(key)
			if !openaiAllowedHeaders[lowerKey] {
				continue
			}
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	return resp, nil
}

func (s *OpenAIGatewayService) doNanoBananaJSONRequest(ctx context.Context, account *Account, targetURL, token string, body []byte) ([]byte, error) {
	resp, err := s.doNanoBananaRawJSONRequest(ctx, account, targetURL, token, body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if s.shouldFailoverUpstreamError(resp.StatusCode) {
			return nil, &UpstreamFailoverError{StatusCode: resp.StatusCode, ResponseBody: resp.Body}
		}
		return nil, fmt.Errorf("%s", strings.TrimSpace(firstNonEmpty(extractNanoBananaErrorMessage(resp.Body), fmt.Sprintf("nano banana upstream error: %d", resp.StatusCode))))
	}
	return resp.Body, nil
}

func (s *OpenAIGatewayService) doNanoBananaRawJSONRequest(ctx context.Context, account *Account, targetURL, token string, body []byte) (*NanoBananaRawResponse, error) {
	resp, err := s.doNanoBananaHTTPRequest(ctx, nil, account, targetURL, token, body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := readUpstreamResponseBodyLimited(resp.Body, resolveUpstreamResponseReadLimit(s.cfg))
	if err != nil {
		return nil, err
	}
	return &NanoBananaRawResponse{
		StatusCode:  resp.StatusCode,
		ContentType: strings.TrimSpace(resp.Header.Get("Content-Type")),
		Headers:     resp.Header.Clone(),
		Body:        respBody,
	}, nil
}

func parseNanoBananaGenerationRequest(body []byte) (nanoBananaSubmitRequest, error) {
	model := strings.TrimSpace(gjson.GetBytes(body, "model").String())
	prompt := strings.TrimSpace(gjson.GetBytes(body, "prompt").String())
	if model == "" {
		return nanoBananaSubmitRequest{}, fmt.Errorf("model is required")
	}
	if prompt == "" {
		return nanoBananaSubmitRequest{}, fmt.Errorf("prompt is required")
	}
	request := nanoBananaSubmitRequest{
		Model:       model,
		Prompt:      prompt,
		AspectRatio: normalizeNanoBananaAspectRatio(firstNonEmpty(gjson.GetBytes(body, "aspect_ratio").String(), gjson.GetBytes(body, "aspectRatio").String())),
		ImageSize:   normalizeNanoBananaImageSize(firstNonEmpty(gjson.GetBytes(body, "image_size").String(), gjson.GetBytes(body, "imageSize").String(), gjson.GetBytes(body, "size").String())),
		URLs:        parseNanoBananaURLsFromJSON(body),
	}
	return request, nil
}

func parseNanoBananaEditRequest(c *gin.Context) (nanoBananaSubmitRequest, error) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return nanoBananaSubmitRequest{}, fmt.Errorf("parse multipart form: %w", err)
	}
	model := strings.TrimSpace(c.Request.FormValue("model"))
	prompt := strings.TrimSpace(c.Request.FormValue("prompt"))
	if model == "" {
		return nanoBananaSubmitRequest{}, fmt.Errorf("model is required")
	}
	if prompt == "" {
		return nanoBananaSubmitRequest{}, fmt.Errorf("prompt is required")
	}
	urls, err := parseNanoBananaURLsFromMultipart(c)
	if err != nil {
		return nanoBananaSubmitRequest{}, err
	}
	if len(urls) == 0 {
		return nanoBananaSubmitRequest{}, fmt.Errorf("image is required")
	}
	return nanoBananaSubmitRequest{
		Model:       model,
		Prompt:      prompt,
		AspectRatio: normalizeNanoBananaAspectRatio(firstNonEmpty(c.Request.FormValue("aspect_ratio"), c.Request.FormValue("aspectRatio"))),
		ImageSize:   normalizeNanoBananaImageSize(firstNonEmpty(c.Request.FormValue("image_size"), c.Request.FormValue("imageSize"), c.Request.FormValue("size"))),
		URLs:        urls,
	}, nil
}

func parseNanoBananaURLsFromJSON(body []byte) []string {
	result := gjson.GetBytes(body, "urls")
	if !result.Exists() || !result.IsArray() {
		return nil
	}
	urls := make([]string, 0, len(result.Array()))
	for _, item := range result.Array() {
		if value := normalizeNanoBananaReferenceValue(item.String()); value != "" {
			urls = append(urls, value)
		}
	}
	return urls
}

func parseNanoBananaURLsFromMultipart(c *gin.Context) ([]string, error) {
	if c == nil || c.Request == nil || c.Request.MultipartForm == nil {
		return nil, nil
	}
	fileHeaders := c.Request.MultipartForm.File["image"]
	urls := make([]string, 0, len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		encoded, err := fileHeaderToBase64(fileHeader)
		if err != nil {
			return nil, err
		}
		urls = append(urls, encoded)
	}
	return urls, nil
}

func fileHeaderToBase64(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil {
		return "", nil
	}
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("open uploaded file: %w", err)
	}
	defer func() { _ = file.Close() }()
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("read uploaded file: %w", err)
	}
	return base64.StdEncoding.EncodeToString(content), nil
}

func normalizeNanoBananaAspectRatio(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nanoBananaDefaultAspectRate
	}
	return value
}

func normalizeNanoBananaImageSize(value string) string {
	value = strings.TrimSpace(strings.ToUpper(value))
	switch value {
	case "1K", "2K", "4K":
		return value
	default:
		return nanoBananaDefaultImageSize
	}
}

func extractNanoBananaErrorMessage(body []byte) string {
	return strings.TrimSpace(firstNonEmpty(
		gjson.GetBytes(body, "error").String(),
		gjson.GetBytes(body, "message").String(),
		gjson.GetBytes(body, "msg").String(),
		gjson.GetBytes(body, "data.error").String(),
	))
}

func normalizeNanoBananaReferenceValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if idx := strings.Index(trimmed, ","); idx > 0 && strings.Contains(trimmed[:idx], ";base64") {
		return strings.TrimSpace(trimmed[idx+1:])
	}
	return trimmed
}

func isNanoBananaTaskNotFound(body []byte) bool {
	code := gjson.GetBytes(body, "code")
	if code.Exists() && code.Int() == -22 {
		return true
	}
	msg := strings.ToLower(strings.TrimSpace(firstNonEmpty(
		gjson.GetBytes(body, "msg").String(),
		gjson.GetBytes(body, "message").String(),
		gjson.GetBytes(body, "error").String(),
	)))
	return strings.Contains(msg, "not exist") || strings.Contains(msg, "not found")
}

func WriteNanoBananaRawResponse(c *gin.Context, resp *NanoBananaRawResponse) {
	if c == nil || resp == nil || c.Writer == nil || c.Writer.Written() {
		return
	}
	for key, values := range resp.Headers {
		lower := strings.ToLower(key)
		if lower != "content-type" && lower != "cache-control" && lower != "transfer-encoding" && lower != "x-request-id" {
			continue
		}
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}
	if c.Writer.Header().Get("Content-Type") == "" {
		contentType := resp.ContentType
		if contentType == "" {
			contentType = "application/json"
		}
		c.Writer.Header().Set("Content-Type", contentType)
	}
	c.Status(resp.StatusCode)
	_, _ = c.Writer.Write(resp.Body)
}

func writeNanoBananaStreamResponse(c *gin.Context, resp *http.Response) {
	if c == nil || resp == nil || c.Writer == nil {
		return
	}
	for key, values := range resp.Header {
		lower := strings.ToLower(key)
		if lower != "content-type" && lower != "cache-control" && lower != "transfer-encoding" && lower != "x-request-id" {
			continue
		}
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}
	c.Status(resp.StatusCode)
	if flusher, ok := c.Writer.(http.Flusher); ok {
		buf := make([]byte, 4096)
		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				_, _ = c.Writer.Write(buf[:n])
				flusher.Flush()
			}
			if readErr != nil {
				break
			}
		}
		return
	}
	_, _ = io.Copy(c.Writer, resp.Body)
}

func writeNanoBananaGatewayError(c *gin.Context, status int, message string) {
	if c == nil || c.Writer == nil || c.Writer.Written() {
		return
	}
	c.JSON(status, gin.H{
		"error": gin.H{
			"type":    "api_error",
			"message": strings.TrimSpace(message),
		},
	})
}

func errorsAsUpstreamFailover(err error, target **UpstreamFailoverError) bool {
	if err == nil || target == nil {
		return false
	}
	failoverErr, ok := err.(*UpstreamFailoverError)
	if ok {
		*target = failoverErr
		return true
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
