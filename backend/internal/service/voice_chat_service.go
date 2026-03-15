package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/tidwall/gjson"
)

const grokLivekitProbeURL = "https://livekit.grok.com"

type VoicePreflightResult struct {
	FunctionReady   bool   `json:"function_ready"`
	LivekitReady    bool   `json:"livekit_ready"`
	LivekitProbeURL string `json:"livekit_probe_url"`
}

type VoiceSessionToken struct {
	Token           string `json:"token"`
	URL             string `json:"url"`
	ParticipantName string `json:"participant_name,omitempty"`
	RoomName        string `json:"room_name,omitempty"`
}

type VoiceChatService struct {
	httpUpstream   HTTPUpstream
	gatewayService *OpenAIGatewayService
	cfg            *config.Config
}

func NewVoiceChatService(httpUpstream HTTPUpstream, gatewayService *OpenAIGatewayService, cfg *config.Config) *VoiceChatService {
	return &VoiceChatService{httpUpstream: httpUpstream, gatewayService: gatewayService, cfg: cfg}
}

func (s *VoiceChatService) Preflight(ctx context.Context, account *Account) (*VoicePreflightResult, error) {
	functionReady := account != nil && account.SupportsGrokLivechat()
	return &VoicePreflightResult{
		FunctionReady:   functionReady,
		LivekitReady:    true,
		LivekitProbeURL: grokLivekitProbeURL,
	}, nil
}

func (s *VoiceChatService) CreateSessionToken(ctx context.Context, account *Account, voice, personality string, speed float64) (*VoiceSessionToken, error) {
	baseURL, err := s.voiceBaseURL(account)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("voice", strings.TrimSpace(voice))
	query.Set("personality", strings.TrimSpace(personality))
	query.Set("speed", strconv.FormatFloat(speed, 'f', 2, 64))
	targetURL := buildOpenAIEndpointURL(baseURL, "livechat/token") + "?" + query.Encode()
	statusCode, body, err := s.doAuthorizedGET(ctx, account, targetURL)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("voice session request failed: %s", strings.TrimSpace(extractUpstreamErrorMessage(body)))
	}
	token := strings.TrimSpace(gjson.GetBytes(body, "token").String())
	wsURL := strings.TrimSpace(gjson.GetBytes(body, "url").String())
	if token == "" || wsURL == "" {
		return nil, fmt.Errorf("voice session response missing token or url")
	}
	return &VoiceSessionToken{
		Token:           token,
		URL:             wsURL,
		ParticipantName: strings.TrimSpace(gjson.GetBytes(body, "participant_name").String()),
		RoomName:        strings.TrimSpace(gjson.GetBytes(body, "room_name").String()),
	}, nil
}

func (s *VoiceChatService) voiceBaseURL(account *Account) (string, error) {
	if account == nil {
		return "", fmt.Errorf("voice account is required")
	}
	baseURL := strings.TrimSpace(account.GetOpenAILivechatBaseURL())
	if baseURL == "" {
		return "", fmt.Errorf("voice base_url is not configured")
	}
	return s.gatewayService.validateUpstreamBaseURL(baseURL)
}

func (s *VoiceChatService) doAuthorizedGET(ctx context.Context, account *Account, targetURL string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return 0, nil, err
	}
	authToken := strings.TrimSpace(account.GetOpenAIApiKey())
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	if userAgent := strings.TrimSpace(account.GetOpenAIUserAgent()); userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, body, nil
}
