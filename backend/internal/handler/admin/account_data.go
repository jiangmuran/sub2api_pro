package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	dataType                 = "sub2api-data"
	legacyDataType           = "sub2api-bundle"
	openAICompatibleDataType = "openai-compatible-import"
	dataVersion              = 1
	dataPageCap              = 1000
)

type DataPayload struct {
	Type       string        `json:"type,omitempty"`
	Version    int           `json:"version,omitempty"`
	ExportedAt string        `json:"exported_at"`
	Proxies    []DataProxy   `json:"proxies"`
	Accounts   []DataAccount `json:"accounts"`
}

type DataProxy struct {
	ProxyKey string `json:"proxy_key"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Status   string `json:"status"`
}

type DataAccount struct {
	Name               string         `json:"name"`
	Notes              *string        `json:"notes,omitempty"`
	Platform           string         `json:"platform"`
	Type               string         `json:"type"`
	Credentials        map[string]any `json:"credentials"`
	Extra              map[string]any `json:"extra,omitempty"`
	ProxyKey           *string        `json:"proxy_key,omitempty"`
	Concurrency        int            `json:"concurrency"`
	Priority           int            `json:"priority"`
	RateMultiplier     *float64       `json:"rate_multiplier,omitempty"`
	ExpiresAt          *int64         `json:"expires_at,omitempty"`
	AutoPauseOnExpired *bool          `json:"auto_pause_on_expired,omitempty"`
}

type DataImportRequest struct {
	Data                 json.RawMessage `json:"data"`
	SkipDefaultGroupBind *bool           `json:"skip_default_group_bind"`
}

type normalizedDataImportRequest struct {
	Data                 DataPayload
	SkipDefaultGroupBind *bool
}

type DataImportResult struct {
	ProxyCreated   int               `json:"proxy_created"`
	ProxyReused    int               `json:"proxy_reused"`
	ProxyFailed    int               `json:"proxy_failed"`
	AccountCreated int               `json:"account_created"`
	AccountFailed  int               `json:"account_failed"`
	Errors         []DataImportError `json:"errors,omitempty"`
}

type DataImportError struct {
	Kind     string `json:"kind"`
	Name     string `json:"name,omitempty"`
	ProxyKey string `json:"proxy_key,omitempty"`
	Message  string `json:"message"`
}

func buildProxyKey(protocol, host string, port int, username, password string) string {
	return fmt.Sprintf("%s|%s|%d|%s|%s", strings.TrimSpace(protocol), strings.TrimSpace(host), port, strings.TrimSpace(username), strings.TrimSpace(password))
}

func (h *AccountHandler) ExportData(c *gin.Context) {
	ctx := c.Request.Context()

	selectedIDs, err := parseAccountIDs(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	accounts, err := h.resolveExportAccounts(ctx, selectedIDs, c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	includeProxies, err := parseIncludeProxies(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var proxies []service.Proxy
	if includeProxies {
		proxies, err = h.resolveExportProxies(ctx, accounts)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
	} else {
		proxies = []service.Proxy{}
	}

	proxyKeyByID := make(map[int64]string, len(proxies))
	dataProxies := make([]DataProxy, 0, len(proxies))
	for i := range proxies {
		p := proxies[i]
		key := buildProxyKey(p.Protocol, p.Host, p.Port, p.Username, p.Password)
		proxyKeyByID[p.ID] = key
		dataProxies = append(dataProxies, DataProxy{
			ProxyKey: key,
			Name:     p.Name,
			Protocol: p.Protocol,
			Host:     p.Host,
			Port:     p.Port,
			Username: p.Username,
			Password: p.Password,
			Status:   p.Status,
		})
	}

	dataAccounts := make([]DataAccount, 0, len(accounts))
	for i := range accounts {
		acc := accounts[i]
		var proxyKey *string
		if acc.ProxyID != nil {
			if key, ok := proxyKeyByID[*acc.ProxyID]; ok {
				proxyKey = &key
			}
		}
		var expiresAt *int64
		if acc.ExpiresAt != nil {
			v := acc.ExpiresAt.Unix()
			expiresAt = &v
		}
		dataAccounts = append(dataAccounts, DataAccount{
			Name:               acc.Name,
			Notes:              acc.Notes,
			Platform:           acc.Platform,
			Type:               acc.Type,
			Credentials:        acc.Credentials,
			Extra:              acc.Extra,
			ProxyKey:           proxyKey,
			Concurrency:        acc.Concurrency,
			Priority:           acc.Priority,
			RateMultiplier:     acc.RateMultiplier,
			ExpiresAt:          expiresAt,
			AutoPauseOnExpired: &acc.AutoPauseOnExpired,
		})
	}

	payload := DataPayload{
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Proxies:    dataProxies,
		Accounts:   dataAccounts,
	}

	response.Success(c, payload)
}

func (h *AccountHandler) ImportData(c *gin.Context) {
	var req DataImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	normalizedReq, err := normalizeDataImportRequest(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	executeAdminIdempotentJSON(c, "admin.accounts.import_data", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		return h.importData(ctx, normalizedReq)
	})
}

func (h *AccountHandler) importData(ctx context.Context, req normalizedDataImportRequest) (DataImportResult, error) {
	skipDefaultGroupBind := true
	if req.SkipDefaultGroupBind != nil {
		skipDefaultGroupBind = *req.SkipDefaultGroupBind
	}

	dataPayload := req.Data
	result := DataImportResult{}

	existingProxies, err := h.listAllProxies(ctx)
	if err != nil {
		return result, err
	}

	proxyKeyToID := make(map[string]int64, len(existingProxies))
	for i := range existingProxies {
		p := existingProxies[i]
		key := buildProxyKey(p.Protocol, p.Host, p.Port, p.Username, p.Password)
		proxyKeyToID[key] = p.ID
	}

	for i := range dataPayload.Proxies {
		item := dataPayload.Proxies[i]
		key := item.ProxyKey
		if key == "" {
			key = buildProxyKey(item.Protocol, item.Host, item.Port, item.Username, item.Password)
		}
		if err := validateDataProxy(item); err != nil {
			result.ProxyFailed++
			result.Errors = append(result.Errors, DataImportError{
				Kind:     "proxy",
				Name:     item.Name,
				ProxyKey: key,
				Message:  err.Error(),
			})
			continue
		}
		normalizedStatus := normalizeProxyStatus(item.Status)
		if existingID, ok := proxyKeyToID[key]; ok {
			proxyKeyToID[key] = existingID
			result.ProxyReused++
			if normalizedStatus != "" {
				if proxy, getErr := h.adminService.GetProxy(ctx, existingID); getErr == nil && proxy != nil && proxy.Status != normalizedStatus {
					_, _ = h.adminService.UpdateProxy(ctx, existingID, &service.UpdateProxyInput{
						Status: normalizedStatus,
					})
				}
			}
			continue
		}

		created, createErr := h.adminService.CreateProxy(ctx, &service.CreateProxyInput{
			Name:     defaultProxyName(item.Name),
			Protocol: item.Protocol,
			Host:     item.Host,
			Port:     item.Port,
			Username: item.Username,
			Password: item.Password,
		})
		if createErr != nil {
			result.ProxyFailed++
			result.Errors = append(result.Errors, DataImportError{
				Kind:     "proxy",
				Name:     item.Name,
				ProxyKey: key,
				Message:  createErr.Error(),
			})
			continue
		}
		proxyKeyToID[key] = created.ID
		result.ProxyCreated++

		if normalizedStatus != "" && normalizedStatus != created.Status {
			_, _ = h.adminService.UpdateProxy(ctx, created.ID, &service.UpdateProxyInput{
				Status: normalizedStatus,
			})
		}
	}

	for i := range dataPayload.Accounts {
		item := dataPayload.Accounts[i]
		if err := validateDataAccount(item); err != nil {
			result.AccountFailed++
			result.Errors = append(result.Errors, DataImportError{
				Kind:    "account",
				Name:    item.Name,
				Message: err.Error(),
			})
			continue
		}

		var proxyID *int64
		if item.ProxyKey != nil && *item.ProxyKey != "" {
			if id, ok := proxyKeyToID[*item.ProxyKey]; ok {
				proxyID = &id
			} else {
				result.AccountFailed++
				result.Errors = append(result.Errors, DataImportError{
					Kind:     "account",
					Name:     item.Name,
					ProxyKey: *item.ProxyKey,
					Message:  "proxy_key not found",
				})
				continue
			}
		}

		accountInput := &service.CreateAccountInput{
			Name:                 item.Name,
			Notes:                item.Notes,
			Platform:             item.Platform,
			Type:                 item.Type,
			Credentials:          item.Credentials,
			Extra:                item.Extra,
			ProxyID:              proxyID,
			Concurrency:          item.Concurrency,
			Priority:             item.Priority,
			RateMultiplier:       item.RateMultiplier,
			GroupIDs:             nil,
			ExpiresAt:            item.ExpiresAt,
			AutoPauseOnExpired:   item.AutoPauseOnExpired,
			SkipDefaultGroupBind: skipDefaultGroupBind,
		}

		if _, err := h.adminService.CreateAccount(ctx, accountInput); err != nil {
			result.AccountFailed++
			result.Errors = append(result.Errors, DataImportError{
				Kind:    "account",
				Name:    item.Name,
				Message: err.Error(),
			})
			continue
		}
		result.AccountCreated++
	}

	return result, nil
}

func normalizeDataImportRequest(req DataImportRequest) (normalizedDataImportRequest, error) {
	dataPayload, err := normalizeImportedDataPayload(req.Data)
	if err != nil {
		return normalizedDataImportRequest{}, err
	}
	return normalizedDataImportRequest{
		Data:                 dataPayload,
		SkipDefaultGroupBind: req.SkipDefaultGroupBind,
	}, nil
}

func normalizeImportedDataPayload(raw json.RawMessage) (DataPayload, error) {
	var header struct {
		Type string `json:"type"`
	}
	if len(raw) == 0 {
		return DataPayload{}, errors.New("data is required")
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return DataPayload{}, fmt.Errorf("invalid data payload: %w", err)
	}

	dataTypeValue := strings.TrimSpace(header.Type)
	switch dataTypeValue {
	case "", dataType, legacyDataType:
		var payload DataPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return DataPayload{}, fmt.Errorf("invalid data payload: %w", err)
		}
		if err := validateDataHeader(payload); err != nil {
			return DataPayload{}, err
		}
		return payload, nil
	case openAICompatibleDataType:
		payload, err := normalizeOpenAICompatibleDataPayload(raw)
		if err != nil {
			return DataPayload{}, err
		}
		if err := validateDataHeader(payload); err != nil {
			return DataPayload{}, err
		}
		return payload, nil
	default:
		return DataPayload{}, fmt.Errorf("unsupported data type: %s", dataTypeValue)
	}
}

func normalizeOpenAICompatibleDataPayload(raw json.RawMessage) (DataPayload, error) {
	var payload struct {
		Type       string           `json:"type"`
		Version    int              `json:"version"`
		ExportedAt string           `json:"exported_at"`
		Proxies    []DataProxy      `json:"proxies"`
		Accounts   []map[string]any `json:"accounts"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return DataPayload{}, fmt.Errorf("invalid compatible import payload: %w", err)
	}
	if payload.Version != 0 && payload.Version != dataVersion {
		return DataPayload{}, fmt.Errorf("unsupported data version: %d", payload.Version)
	}
	if payload.Accounts == nil {
		return DataPayload{}, errors.New("accounts is required")
	}

	normalizedAccounts := make([]DataAccount, 0, len(payload.Accounts))
	for idx := range payload.Accounts {
		item, err := normalizeOpenAICompatibleAccount(payload.Accounts[idx])
		if err != nil {
			return DataPayload{}, fmt.Errorf("accounts[%d]: %w", idx, err)
		}
		normalizedAccounts = append(normalizedAccounts, item)
	}

	proxies := payload.Proxies
	if proxies == nil {
		proxies = []DataProxy{}
	}

	return DataPayload{
		Type:       openAICompatibleDataType,
		Version:    dataVersion,
		ExportedAt: payload.ExportedAt,
		Proxies:    proxies,
		Accounts:   normalizedAccounts,
	}, nil
}

func normalizeOpenAICompatibleAccount(raw map[string]any) (DataAccount, error) {
	name := strings.TrimSpace(getStringAlias(raw, "name"))
	if name == "" {
		return DataAccount{}, errors.New("name is required")
	}

	baseURL := strings.TrimSpace(getStringAlias(raw, "base_url", "baseURL"))
	if baseURL == "" {
		return DataAccount{}, errors.New("base_url is required")
	}
	apiKey := strings.TrimSpace(getStringAlias(raw, "api_key", "apiKey"))
	if apiKey == "" {
		return DataAccount{}, errors.New("api_key is required")
	}

	credentials := cloneMapAny(getMapAlias(raw, "credentials"))
	if credentials == nil {
		credentials = make(map[string]any, 4)
	}
	credentials["base_url"] = baseURL
	credentials["api_key"] = apiKey
	if userAgent := strings.TrimSpace(getStringAlias(raw, "user_agent", "userAgent")); userAgent != "" {
		credentials["user_agent"] = userAgent
	}
	if modelMapping := normalizeStringMap(getMapAlias(raw, "model_mapping", "modelMap")); len(modelMapping) > 0 {
		credentials["model_mapping"] = modelMapping
	} else if modelList := normalizeStringSlice(getSliceAlias(raw, "models")); len(modelList) > 0 {
		credentials["model_mapping"] = buildIdentityModelMapping(modelList)
	}

	extra := cloneMapAny(getMapAlias(raw, "extra"))
	if extra == nil {
		extra = make(map[string]any, 4)
	}
	passthrough, ok := getBoolAlias(raw, "passthrough", "openai_passthrough")
	if !ok {
		passthrough = true
	}
	extra["openai_passthrough"] = passthrough
	if compatMode := strings.TrimSpace(getStringAlias(raw, "compat_mode", "compatMode")); compatMode != "" {
		extra["openai_compat_mode"] = compatMode
	}
	if capabilities := cloneMapAny(getMapAlias(raw, "compat_capabilities", "compatCapabilities")); len(capabilities) > 0 {
		extra["openai_compat_capabilities"] = capabilities
	}

	var notes *string
	if noteValue := strings.TrimSpace(getStringAlias(raw, "notes")); noteValue != "" {
		noteCopy := noteValue
		notes = &noteCopy
	}

	var proxyKey *string
	if proxyValue := strings.TrimSpace(getStringAlias(raw, "proxy_key", "proxyKey")); proxyValue != "" {
		proxyCopy := proxyValue
		proxyKey = &proxyCopy
	}

	concurrency, err := getOptionalIntAlias(raw, "concurrency")
	if err != nil {
		return DataAccount{}, fmt.Errorf("concurrency invalid: %w", err)
	}
	priority, err := getOptionalIntAlias(raw, "priority")
	if err != nil {
		return DataAccount{}, fmt.Errorf("priority invalid: %w", err)
	}
	rateMultiplier, err := getOptionalFloatAlias(raw, "rate_multiplier", "rateMultiplier")
	if err != nil {
		return DataAccount{}, fmt.Errorf("rate_multiplier invalid: %w", err)
	}
	expiresAt, err := getOptionalInt64Alias(raw, "expires_at", "expiresAt")
	if err != nil {
		return DataAccount{}, fmt.Errorf("expires_at invalid: %w", err)
	}
	autoPauseOnExpired, err := getOptionalBoolAlias(raw, "auto_pause_on_expired", "autoPauseOnExpired")
	if err != nil {
		return DataAccount{}, fmt.Errorf("auto_pause_on_expired invalid: %w", err)
	}

	return DataAccount{
		Name:               name,
		Notes:              notes,
		Platform:           service.PlatformOpenAI,
		Type:               service.AccountTypeAPIKey,
		Credentials:        credentials,
		Extra:              extra,
		ProxyKey:           proxyKey,
		Concurrency:        concurrency,
		Priority:           priority,
		RateMultiplier:     rateMultiplier,
		ExpiresAt:          expiresAt,
		AutoPauseOnExpired: autoPauseOnExpired,
	}, nil
}

func getStringAlias(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if values == nil {
			return ""
		}
		if raw, ok := values[key]; ok && raw != nil {
			switch v := raw.(type) {
			case string:
				return v
			case json.Number:
				return v.String()
			}
		}
	}
	return ""
}

func getMapAlias(values map[string]any, keys ...string) map[string]any {
	for _, key := range keys {
		if values == nil {
			return nil
		}
		if raw, ok := values[key]; ok && raw != nil {
			if mapped, ok := raw.(map[string]any); ok {
				return mapped
			}
		}
	}
	return nil
}

func getSliceAlias(values map[string]any, keys ...string) []any {
	for _, key := range keys {
		if values == nil {
			return nil
		}
		if raw, ok := values[key]; ok && raw != nil {
			if items, ok := raw.([]any); ok {
				return items
			}
		}
	}
	return nil
}

func getBoolAlias(values map[string]any, keys ...string) (bool, bool) {
	for _, key := range keys {
		if values == nil {
			return false, false
		}
		if raw, ok := values[key]; ok && raw != nil {
			if v, ok := raw.(bool); ok {
				return v, true
			}
		}
	}
	return false, false
}

func getOptionalBoolAlias(values map[string]any, keys ...string) (*bool, error) {
	for _, key := range keys {
		if values == nil {
			return nil, nil
		}
		raw, ok := values[key]
		if !ok || raw == nil {
			continue
		}
		v, ok := raw.(bool)
		if !ok {
			return nil, fmt.Errorf("expected boolean")
		}
		return &v, nil
	}
	return nil, nil
}

func getOptionalIntAlias(values map[string]any, keys ...string) (int, error) {
	for _, key := range keys {
		if values == nil {
			return 0, nil
		}
		raw, ok := values[key]
		if !ok || raw == nil {
			continue
		}
		switch v := raw.(type) {
		case float64:
			return int(v), nil
		case int:
			return v, nil
		case int64:
			return int(v), nil
		case json.Number:
			iv, err := v.Int64()
			if err != nil {
				return 0, err
			}
			return int(iv), nil
		default:
			return 0, fmt.Errorf("expected number")
		}
	}
	return 0, nil
}

func getOptionalInt64Alias(values map[string]any, keys ...string) (*int64, error) {
	for _, key := range keys {
		if values == nil {
			return nil, nil
		}
		raw, ok := values[key]
		if !ok || raw == nil {
			continue
		}
		var out int64
		switch v := raw.(type) {
		case float64:
			out = int64(v)
		case int:
			out = int64(v)
		case int64:
			out = v
		case json.Number:
			parsed, err := v.Int64()
			if err != nil {
				return nil, err
			}
			out = parsed
		default:
			return nil, fmt.Errorf("expected number")
		}
		return &out, nil
	}
	return nil, nil
}

func getOptionalFloatAlias(values map[string]any, keys ...string) (*float64, error) {
	for _, key := range keys {
		if values == nil {
			return nil, nil
		}
		raw, ok := values[key]
		if !ok || raw == nil {
			continue
		}
		var out float64
		switch v := raw.(type) {
		case float64:
			out = v
		case int:
			out = float64(v)
		case int64:
			out = float64(v)
		case json.Number:
			parsed, err := v.Float64()
			if err != nil {
				return nil, err
			}
			out = parsed
		default:
			return nil, fmt.Errorf("expected number")
		}
		return &out, nil
	}
	return nil, nil
}

func cloneMapAny(src map[string]any) map[string]any {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]any, len(src))
	for key, value := range src {
		out[key] = value
	}
	return out
}

func normalizeStringMap(src map[string]any) map[string]string {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]string, len(src))
	for key, value := range src {
		strValue, ok := value.(string)
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		strValue = strings.TrimSpace(strValue)
		if key == "" || strValue == "" {
			continue
		}
		out[key] = strValue
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeStringSlice(src []any) []string {
	if len(src) == 0 {
		return nil
	}
	out := make([]string, 0, len(src))
	for _, value := range src {
		strValue, ok := value.(string)
		if !ok {
			continue
		}
		strValue = strings.TrimSpace(strValue)
		if strValue == "" {
			continue
		}
		out = append(out, strValue)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func buildIdentityModelMapping(models []string) map[string]string {
	if len(models) == 0 {
		return nil
	}
	mapping := make(map[string]string, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		mapping[model] = model
	}
	if len(mapping) == 0 {
		return nil
	}
	return mapping
}

func (h *AccountHandler) listAllProxies(ctx context.Context) ([]service.Proxy, error) {
	page := 1
	pageSize := dataPageCap
	var out []service.Proxy
	for {
		items, total, err := h.adminService.ListProxies(ctx, page, pageSize, "", "", "")
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
		if len(out) >= int(total) || len(items) == 0 {
			break
		}
		page++
	}
	return out, nil
}

func (h *AccountHandler) listAccountsFiltered(ctx context.Context, platform, accountType, status, search string) ([]service.Account, error) {
	page := 1
	pageSize := dataPageCap
	var out []service.Account
	for {
		items, total, err := h.adminService.ListAccounts(ctx, page, pageSize, platform, accountType, status, search, 0)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
		if len(out) >= int(total) || len(items) == 0 {
			break
		}
		page++
	}
	return out, nil
}

func (h *AccountHandler) resolveExportAccounts(ctx context.Context, ids []int64, c *gin.Context) ([]service.Account, error) {
	if len(ids) > 0 {
		accounts, err := h.adminService.GetAccountsByIDs(ctx, ids)
		if err != nil {
			return nil, err
		}
		out := make([]service.Account, 0, len(accounts))
		for _, acc := range accounts {
			if acc == nil {
				continue
			}
			out = append(out, *acc)
		}
		return out, nil
	}

	platform := c.Query("platform")
	accountType := c.Query("type")
	status := c.Query("status")
	search := strings.TrimSpace(c.Query("search"))
	if len(search) > 100 {
		search = search[:100]
	}
	return h.listAccountsFiltered(ctx, platform, accountType, status, search)
}

func (h *AccountHandler) resolveExportProxies(ctx context.Context, accounts []service.Account) ([]service.Proxy, error) {
	if len(accounts) == 0 {
		return []service.Proxy{}, nil
	}

	seen := make(map[int64]struct{})
	ids := make([]int64, 0)
	for i := range accounts {
		if accounts[i].ProxyID == nil {
			continue
		}
		id := *accounts[i].ProxyID
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return []service.Proxy{}, nil
	}

	return h.adminService.GetProxiesByIDs(ctx, ids)
}

func parseAccountIDs(c *gin.Context) ([]int64, error) {
	values := c.QueryArray("ids")
	if len(values) == 0 {
		raw := strings.TrimSpace(c.Query("ids"))
		if raw != "" {
			values = []string{raw}
		}
	}
	if len(values) == 0 {
		return nil, nil
	}

	ids := make([]int64, 0, len(values))
	for _, item := range values {
		for _, part := range strings.Split(item, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := strconv.ParseInt(part, 10, 64)
			if err != nil || id <= 0 {
				return nil, fmt.Errorf("invalid account id: %s", part)
			}
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func parseIncludeProxies(c *gin.Context) (bool, error) {
	raw := strings.TrimSpace(strings.ToLower(c.Query("include_proxies")))
	if raw == "" {
		return true, nil
	}
	switch raw {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return true, fmt.Errorf("invalid include_proxies value: %s", raw)
	}
}

func validateDataHeader(payload DataPayload) error {
	if payload.Type != "" && payload.Type != dataType && payload.Type != legacyDataType && payload.Type != openAICompatibleDataType {
		return fmt.Errorf("unsupported data type: %s", payload.Type)
	}
	if payload.Version != 0 && payload.Version != dataVersion {
		return fmt.Errorf("unsupported data version: %d", payload.Version)
	}
	if payload.Proxies == nil {
		return errors.New("proxies is required")
	}
	if payload.Accounts == nil {
		return errors.New("accounts is required")
	}
	return nil
}

func validateDataProxy(item DataProxy) error {
	if strings.TrimSpace(item.Protocol) == "" {
		return errors.New("proxy protocol is required")
	}
	if strings.TrimSpace(item.Host) == "" {
		return errors.New("proxy host is required")
	}
	if item.Port <= 0 || item.Port > 65535 {
		return errors.New("proxy port is invalid")
	}
	switch item.Protocol {
	case "http", "https", "socks5", "socks5h":
	default:
		return fmt.Errorf("proxy protocol is invalid: %s", item.Protocol)
	}
	if item.Status != "" {
		normalizedStatus := normalizeProxyStatus(item.Status)
		if normalizedStatus != service.StatusActive && normalizedStatus != "inactive" {
			return fmt.Errorf("proxy status is invalid: %s", item.Status)
		}
	}
	return nil
}

func validateDataAccount(item DataAccount) error {
	if strings.TrimSpace(item.Name) == "" {
		return errors.New("account name is required")
	}
	if strings.TrimSpace(item.Platform) == "" {
		return errors.New("account platform is required")
	}
	if strings.TrimSpace(item.Type) == "" {
		return errors.New("account type is required")
	}
	if len(item.Credentials) == 0 {
		return errors.New("account credentials is required")
	}
	switch item.Type {
	case service.AccountTypeOAuth, service.AccountTypeSetupToken, service.AccountTypeAPIKey, service.AccountTypeUpstream:
	default:
		return fmt.Errorf("account type is invalid: %s", item.Type)
	}
	if item.RateMultiplier != nil && *item.RateMultiplier < 0 {
		return errors.New("rate_multiplier must be >= 0")
	}
	if item.Concurrency < 0 {
		return errors.New("concurrency must be >= 0")
	}
	if item.Priority < 0 {
		return errors.New("priority must be >= 0")
	}
	return nil
}

func defaultProxyName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "imported-proxy"
	}
	return name
}

func normalizeProxyStatus(status string) string {
	normalized := strings.TrimSpace(strings.ToLower(status))
	switch normalized {
	case "":
		return ""
	case service.StatusActive:
		return service.StatusActive
	case "inactive", service.StatusDisabled:
		return "inactive"
	default:
		return normalized
	}
}
