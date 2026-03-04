package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"log"
	"math/big"
	"net"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/proxyurl"
	"golang.org/x/net/proxy"
)

var (
	ErrEmailNotConfigured    = infraerrors.ServiceUnavailable("EMAIL_NOT_CONFIGURED", "email service not configured")
	ErrInvalidVerifyCode     = infraerrors.BadRequest("INVALID_VERIFY_CODE", "invalid or expired verification code")
	ErrVerifyCodeTooFrequent = infraerrors.TooManyRequests("VERIFY_CODE_TOO_FREQUENT", "please wait before requesting a new code")
	ErrVerifyCodeMaxAttempts = infraerrors.TooManyRequests("VERIFY_CODE_MAX_ATTEMPTS", "too many failed attempts, please request a new code")

	// Password reset errors
	ErrInvalidResetToken = infraerrors.BadRequest("INVALID_RESET_TOKEN", "invalid or expired password reset token")
)

// EmailCache defines cache operations for email service
type EmailCache interface {
	GetVerificationCode(ctx context.Context, email string) (*VerificationCodeData, error)
	SetVerificationCode(ctx context.Context, email string, data *VerificationCodeData, ttl time.Duration) error
	DeleteVerificationCode(ctx context.Context, email string) error

	// Password reset token methods
	GetPasswordResetToken(ctx context.Context, email string) (*PasswordResetTokenData, error)
	SetPasswordResetToken(ctx context.Context, email string, data *PasswordResetTokenData, ttl time.Duration) error
	DeletePasswordResetToken(ctx context.Context, email string) error

	// Password reset email cooldown methods
	// Returns true if in cooldown period (email was sent recently)
	IsPasswordResetEmailInCooldown(ctx context.Context, email string) bool
	SetPasswordResetEmailCooldown(ctx context.Context, email string, ttl time.Duration) error
}

// VerificationCodeData represents verification code data
type VerificationCodeData struct {
	Code      string
	Attempts  int
	CreatedAt time.Time
}

// PasswordResetTokenData represents password reset token data
type PasswordResetTokenData struct {
	Token     string
	CreatedAt time.Time
}

const (
	verifyCodeTTL         = 15 * time.Minute
	verifyCodeCooldown    = 1 * time.Minute
	maxVerifyCodeAttempts = 5

	// Password reset token settings
	passwordResetTokenTTL = 30 * time.Minute

	// Password reset email cooldown (prevent email bombing)
	passwordResetEmailCooldown = 30 * time.Second
)

// SMTPConfig SMTP配置
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
	UseTLS   bool
	ProxyURL string
}

// EmailService 邮件服务
type EmailService struct {
	settingRepo SettingRepository
	cache       EmailCache
}

// NewEmailService 创建邮件服务实例
func NewEmailService(settingRepo SettingRepository, cache EmailCache) *EmailService {
	return &EmailService{
		settingRepo: settingRepo,
		cache:       cache,
	}
}

// GetSMTPConfig 从数据库获取SMTP配置
func (s *EmailService) GetSMTPConfig(ctx context.Context) (*SMTPConfig, error) {
	keys := []string{
		SettingKeySMTPHost,
		SettingKeySMTPPort,
		SettingKeySMTPUsername,
		SettingKeySMTPPassword,
		SettingKeySMTPFrom,
		SettingKeySMTPFromName,
		SettingKeySMTPUseTLS,
		SettingKeySMTPProxyURL,
	}

	settings, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("get smtp settings: %w", err)
	}

	host := settings[SettingKeySMTPHost]
	if host == "" {
		return nil, ErrEmailNotConfigured
	}

	port := 587 // 默认端口
	if portStr := settings[SettingKeySMTPPort]; portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	useTLS := settings[SettingKeySMTPUseTLS] == "true"

	return &SMTPConfig{
		Host:     host,
		Port:     port,
		Username: settings[SettingKeySMTPUsername],
		Password: settings[SettingKeySMTPPassword],
		From:     settings[SettingKeySMTPFrom],
		FromName: settings[SettingKeySMTPFromName],
		UseTLS:   useTLS,
		ProxyURL: settings[SettingKeySMTPProxyURL],
	}, nil
}

// SendEmail 发送邮件（使用数据库中保存的配置）
func (s *EmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	config, err := s.GetSMTPConfig(ctx)
	if err != nil {
		return err
	}
	return s.SendEmailWithConfig(config, to, subject, body)
}

// SendEmailWithConfig 使用指定配置发送邮件
func (s *EmailService) SendEmailWithConfig(config *SMTPConfig, to, subject, body string) error {
	if config == nil {
		return fmt.Errorf("smtp config is nil")
	}
	host := normalizeSMTPHost(config.Host)
	if host == "" {
		return fmt.Errorf("smtp host is empty")
	}
	if config.Port <= 0 {
		config.Port = 587
	}

	from := config.From
	if config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", config.FromName, config.From)
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body)

	addr := fmt.Sprintf("%s:%d", host, config.Port)

	if config.UseTLS {
		return s.sendMailTLS(addr, config.From, to, []byte(msg), host, config.Username, config.Password, config.ProxyURL)
	}

	return s.sendMailWithSTARTTLS(addr, config.From, to, []byte(msg), host, config.Username, config.Password, config.ProxyURL)
}

// sendMailTLS 使用TLS发送邮件
func (s *EmailService) sendMailTLS(addr, from, to string, msg []byte, host, username, password, proxyURL string) error {
	tlsConfig := &tls.Config{
		ServerName: host,
		// 强制 TLS 1.2+，避免协议降级导致的弱加密风险。
		MinVersion: tls.VersionTLS12,
	}

	rawConn, err := dialSMTPConnection(addr, proxyURL)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}

	conn := tls.Client(rawConn, tlsConfig)
	if err := conn.Handshake(); err != nil {
		_ = rawConn.Close()
		return fmt.Errorf("tls handshake: %w", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("new smtp client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if err = smtpAuthenticate(client, host, username, password); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}

	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("write msg: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("close writer: %w", err)
	}

	// Email is sent successfully after w.Close(), ignore Quit errors
	// Some SMTP servers return non-standard responses on QUIT
	_ = client.Quit()
	return nil
}

func (s *EmailService) sendMailWithSTARTTLS(addr, from, to string, msg []byte, host, username, password, proxyURL string) error {
	conn, err := dialSMTPConnection(addr, proxyURL)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("new smtp client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if err := smtpAuthenticate(client, host, username, password); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("write msg: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("close writer: %w", err)
	}
	_ = client.Quit()
	return nil
}

func smtpAuthenticate(client *smtp.Client, host, username, password string) error {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return nil
	}

	plainErr := client.Auth(smtp.PlainAuth("", username, password, host))
	if plainErr == nil {
		return nil
	}
	loginErr := client.Auth(LoginAuth(username, password))
	if loginErr == nil {
		return nil
	}
	return errors.Join(plainErr, loginErr)
}

func dialSMTPConnection(addr, proxyRaw string) (net.Conn, error) {
	const dialTimeout = 10 * time.Second
	trimmed, parsed, err := proxyurl.Parse(proxyRaw)
	if err != nil {
		return nil, fmt.Errorf("parse smtp proxy: %w", err)
	}

	dialer := &net.Dialer{Timeout: dialTimeout}
	if trimmed == "" {
		return dialer.Dial("tcp", addr)
	}

	if parsed.Scheme != "socks5" && parsed.Scheme != "socks5h" {
		return nil, fmt.Errorf("smtp proxy must use socks5 or socks5h")
	}

	proxyDialer, err := proxy.FromURL(parsed, dialer)
	if err != nil {
		return nil, fmt.Errorf("create socks5 dialer: %w", err)
	}

	if contextDialer, ok := proxyDialer.(proxy.ContextDialer); ok {
		ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
		defer cancel()
		return contextDialer.DialContext(ctx, "tcp", addr)
	}

	return proxyDialer.Dial("tcp", addr)
}

type loginAuth struct {
	username string
	password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username: username, password: password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	prompt := strings.ToLower(strings.TrimSpace(string(fromServer)))
	if strings.Contains(prompt, "username") {
		return []byte(a.username), nil
	}
	if strings.Contains(prompt, "password") {
		return []byte(a.password), nil
	}
	return nil, fmt.Errorf("unexpected smtp login challenge: %s", prompt)
}

func normalizeSMTPHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		return strings.TrimSpace(parsedHost)
	}
	if strings.Count(host, ":") == 1 {
		parts := strings.SplitN(host, ":", 2)
		if strings.TrimSpace(parts[0]) != "" {
			return strings.TrimSpace(parts[0])
		}
	}
	return host
}

// GenerateVerifyCode 生成6位数字验证码
func (s *EmailService) GenerateVerifyCode() (string, error) {
	const digits = "0123456789"
	code := make([]byte, 6)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[num.Int64()]
	}
	return string(code), nil
}

// SendVerifyCode 发送验证码邮件
func (s *EmailService) SendVerifyCode(ctx context.Context, email, siteName string) error {
	// 检查是否在冷却期内
	existing, err := s.cache.GetVerificationCode(ctx, email)
	if err == nil && existing != nil {
		if time.Since(existing.CreatedAt) < verifyCodeCooldown {
			return ErrVerifyCodeTooFrequent
		}
	}

	// 生成验证码
	code, err := s.GenerateVerifyCode()
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	// 保存验证码到 Redis
	data := &VerificationCodeData{
		Code:      code,
		Attempts:  0,
		CreatedAt: time.Now(),
	}
	if err := s.cache.SetVerificationCode(ctx, email, data, verifyCodeTTL); err != nil {
		return fmt.Errorf("save verify code: %w", err)
	}

	// 构建邮件内容
	subject := fmt.Sprintf("[%s] 邮箱验证码", siteName)
	body := s.buildVerifyCodeEmailBody(code, siteName)

	// 发送邮件
	if err := s.SendEmail(ctx, email, subject, body); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}

// VerifyCode 验证验证码
func (s *EmailService) VerifyCode(ctx context.Context, email, code string) error {
	data, err := s.cache.GetVerificationCode(ctx, email)
	if err != nil || data == nil {
		return ErrInvalidVerifyCode
	}

	// 检查是否已达到最大尝试次数
	if data.Attempts >= maxVerifyCodeAttempts {
		return ErrVerifyCodeMaxAttempts
	}

	// 验证码不匹配 (constant-time comparison to prevent timing attacks)
	if subtle.ConstantTimeCompare([]byte(data.Code), []byte(code)) != 1 {
		data.Attempts++
		if err := s.cache.SetVerificationCode(ctx, email, data, verifyCodeTTL); err != nil {
			log.Printf("[Email] Failed to update verification attempt count: %v", err)
		}
		if data.Attempts >= maxVerifyCodeAttempts {
			return ErrVerifyCodeMaxAttempts
		}
		return ErrInvalidVerifyCode
	}

	// 验证成功，删除验证码
	if err := s.cache.DeleteVerificationCode(ctx, email); err != nil {
		log.Printf("[Email] Failed to delete verification code after success: %v", err)
	}
	return nil
}

// buildVerifyCodeEmailBody 构建验证码邮件HTML内容
func (s *EmailService) buildVerifyCodeEmailBody(code, siteName string) string {
	safeSiteName := html.EscapeString(siteName)
	safeCode := html.EscapeString(code)
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - 邮箱验证码</title>
</head>
<body style="margin:0;padding:0;background:#f3f6fb;">
    <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%%" style="background:#f3f6fb;padding:24px 12px;">
        <tr>
            <td align="center">
                <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%%" style="max-width:640px;background:#ffffff;border:1px solid #e6edf5;border-radius:16px;overflow:hidden;">
                    <tr>
                        <td style="background:linear-gradient(135deg,#0284c7 0%%,#0369a1 100%%);padding:28px 28px 24px;">
                            <div style="font-family:'PingFang SC','Hiragino Sans GB','Microsoft YaHei',sans-serif;color:#d7efff;font-size:12px;letter-spacing:1px;text-transform:uppercase;">%s</div>
                            <div style="margin-top:8px;font-family:'PingFang SC','Hiragino Sans GB','Microsoft YaHei',sans-serif;color:#ffffff;font-size:22px;line-height:30px;font-weight:600;">邮箱验证码</div>
                            <div style="margin-top:8px;font-family:'PingFang SC','Hiragino Sans GB','Microsoft YaHei',sans-serif;color:#d7efff;font-size:14px;line-height:22px;">请使用下方验证码完成邮箱验证。</div>
                        </td>
                    </tr>
                    <tr>
                        <td style="padding:28px;">
                            <div style="font-family:'PingFang SC','Hiragino Sans GB','Microsoft YaHei',sans-serif;color:#1f2937;font-size:14px;line-height:22px;">验证码（15 分钟内有效）</div>
                            <div style="margin-top:12px;background:#f0f9ff;border:1px dashed #7dd3fc;border-radius:12px;padding:18px 16px;text-align:center;font-family:Menlo,Consolas,'SFMono-Regular',monospace;font-size:34px;line-height:40px;letter-spacing:8px;font-weight:700;color:#0c4a6e;">%s</div>
                            <div style="margin-top:18px;font-family:'PingFang SC','Hiragino Sans GB','Microsoft YaHei',sans-serif;color:#4b5563;font-size:14px;line-height:22px;">如果这不是您本人操作，请忽略本邮件。为保障安全，请勿将验证码透露给任何人。</div>
                        </td>
                    </tr>
                    <tr>
                        <td style="padding:18px 28px;background:#f8fafc;border-top:1px solid #e6edf5;font-family:'PingFang SC','Hiragino Sans GB','Microsoft YaHei',sans-serif;color:#94a3b8;font-size:12px;line-height:20px;">
                            此邮件由系统自动发送，请勿直接回复。
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, safeSiteName, safeSiteName, safeCode)
}

// TestSMTPConnectionWithConfig 使用指定配置测试SMTP连接
func (s *EmailService) TestSMTPConnectionWithConfig(config *SMTPConfig) error {
	if config == nil {
		return fmt.Errorf("smtp config is nil")
	}
	host := normalizeSMTPHost(config.Host)
	if host == "" {
		return fmt.Errorf("smtp host is empty")
	}
	if config.Port <= 0 {
		config.Port = 587
	}
	addr := fmt.Sprintf("%s:%d", host, config.Port)

	if config.UseTLS {
		tlsConfig := &tls.Config{
			ServerName: host,
			// 与发送逻辑一致，显式要求 TLS 1.2+。
			MinVersion: tls.VersionTLS12,
		}
		rawConn, err := dialSMTPConnection(addr, config.ProxyURL)
		if err != nil {
			return fmt.Errorf("smtp dial failed: %w", err)
		}

		conn := tls.Client(rawConn, tlsConfig)
		if err := conn.Handshake(); err != nil {
			_ = rawConn.Close()
			return fmt.Errorf("tls handshake failed: %w", err)
		}
		defer func() { _ = conn.Close() }()

		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("smtp client creation failed: %w", err)
		}
		defer func() { _ = client.Close() }()

		if err = smtpAuthenticate(client, host, config.Username, config.Password); err != nil {
			return fmt.Errorf("smtp authentication failed: %w", err)
		}

		return client.Quit()
	}

	// 非TLS连接测试
	conn, err := dialSMTPConnection(addr, config.ProxyURL)
	if err != nil {
		return fmt.Errorf("smtp connection failed: %w", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client creation failed: %w", err)
	}
	defer func() { _ = client.Close() }()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("smtp starttls failed: %w", err)
		}
	}

	if err = smtpAuthenticate(client, host, config.Username, config.Password); err != nil {
		return fmt.Errorf("smtp authentication failed: %w", err)
	}

	return client.Quit()
}

// GeneratePasswordResetToken generates a secure 32-byte random token (64 hex characters)
func (s *EmailService) GeneratePasswordResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// SendPasswordResetEmail sends a password reset email with a reset link
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, email, siteName, resetURL string) error {
	var token string
	var needSaveToken bool

	// Check if token already exists
	existing, err := s.cache.GetPasswordResetToken(ctx, email)
	if err == nil && existing != nil {
		// Token exists, reuse it (allows resending email without generating new token)
		token = existing.Token
		needSaveToken = false
	} else {
		// Generate new token
		token, err = s.GeneratePasswordResetToken()
		if err != nil {
			return fmt.Errorf("generate token: %w", err)
		}
		needSaveToken = true
	}

	// Save token to Redis (only if new token generated)
	if needSaveToken {
		data := &PasswordResetTokenData{
			Token:     token,
			CreatedAt: time.Now(),
		}
		if err := s.cache.SetPasswordResetToken(ctx, email, data, passwordResetTokenTTL); err != nil {
			return fmt.Errorf("save reset token: %w", err)
		}
	}

	// Build full reset URL with URL-encoded token and email
	fullResetURL := fmt.Sprintf("%s?email=%s&token=%s", resetURL, url.QueryEscape(email), url.QueryEscape(token))

	// Build email content
	subject := fmt.Sprintf("[%s] 密码重置请求", siteName)
	body := s.buildPasswordResetEmailBody(fullResetURL, siteName)

	// Send email
	if err := s.SendEmail(ctx, email, subject, body); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}

// SendPasswordResetEmailWithCooldown sends password reset email with cooldown check (called by queue worker)
// This method wraps SendPasswordResetEmail with email cooldown to prevent email bombing
func (s *EmailService) SendPasswordResetEmailWithCooldown(ctx context.Context, email, siteName, resetURL string) error {
	// Check email cooldown to prevent email bombing
	if s.cache.IsPasswordResetEmailInCooldown(ctx, email) {
		log.Printf("[Email] Password reset email skipped (cooldown): %s", email)
		return nil // Silent success to prevent revealing cooldown to attackers
	}

	// Send email using core method
	if err := s.SendPasswordResetEmail(ctx, email, siteName, resetURL); err != nil {
		return err
	}

	// Set cooldown marker (Redis TTL handles expiration)
	if err := s.cache.SetPasswordResetEmailCooldown(ctx, email, passwordResetEmailCooldown); err != nil {
		log.Printf("[Email] Failed to set password reset cooldown for %s: %v", email, err)
	}

	return nil
}

// VerifyPasswordResetToken verifies the password reset token without consuming it
func (s *EmailService) VerifyPasswordResetToken(ctx context.Context, email, token string) error {
	data, err := s.cache.GetPasswordResetToken(ctx, email)
	if err != nil || data == nil {
		return ErrInvalidResetToken
	}

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(data.Token), []byte(token)) != 1 {
		return ErrInvalidResetToken
	}

	return nil
}

// ConsumePasswordResetToken verifies and deletes the token (one-time use)
func (s *EmailService) ConsumePasswordResetToken(ctx context.Context, email, token string) error {
	// Verify first
	if err := s.VerifyPasswordResetToken(ctx, email, token); err != nil {
		return err
	}

	// Delete after verification (one-time use)
	if err := s.cache.DeletePasswordResetToken(ctx, email); err != nil {
		log.Printf("[Email] Failed to delete password reset token after consumption: %v", err)
	}
	return nil
}

// buildPasswordResetEmailBody builds the HTML content for password reset email
func (s *EmailService) buildPasswordResetEmailBody(resetURL, siteName string) string {
	safeSiteName := html.EscapeString(siteName)
	safeResetURL := html.EscapeString(resetURL)
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - 密码重置</title>
</head>
<body style="margin:0;padding:0;background:#f3f6fb;">
    <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%%" style="background:#f3f6fb;padding:24px 12px;">
        <tr>
            <td align="center">
                <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%%" style="max-width:640px;background:#ffffff;border:1px solid #e6edf5;border-radius:16px;overflow:hidden;">
                    <tr>
                        <td style="background:linear-gradient(135deg,#0f766e 0%%,#0f766e 15%%,#0369a1 100%%);padding:28px 28px 24px;">
                            <div style="font-family:'PingFang SC','Hiragino Sans GB','Microsoft YaHei',sans-serif;color:#d7efff;font-size:12px;letter-spacing:1px;text-transform:uppercase;">%s</div>
                            <div style="margin-top:8px;font-family:'PingFang SC','Hiragino Sans GB','Microsoft YaHei',sans-serif;color:#ffffff;font-size:22px;line-height:30px;font-weight:600;">重置登录密码</div>
                            <div style="margin-top:8px;font-family:'PingFang SC','Hiragino Sans GB','Microsoft YaHei',sans-serif;color:#d7efff;font-size:14px;line-height:22px;">我们收到了一次密码重置请求。</div>
                        </td>
                    </tr>
                    <tr>
                        <td style="padding:28px;">
                            <div style="font-family:'PingFang SC','Hiragino Sans GB','Microsoft YaHei',sans-serif;color:#374151;font-size:14px;line-height:24px;">请点击下方按钮继续操作。该链接将在 <strong>30 分钟</strong>后失效。</div>
                            <div style="margin-top:20px;">
                                <a href="%s" style="display:inline-block;background:#0284c7;color:#ffffff;text-decoration:none;padding:12px 24px;border-radius:10px;font-family:'PingFang SC','Hiragino Sans GB','Microsoft YaHei',sans-serif;font-size:15px;font-weight:600;">立即重置密码</a>
                            </div>
                            <div style="margin-top:18px;padding:14px;border-radius:10px;background:#fff7ed;border:1px solid #fed7aa;font-family:'PingFang SC','Hiragino Sans GB','Microsoft YaHei',sans-serif;color:#9a3412;font-size:13px;line-height:20px;">
                                如果不是您本人操作，请忽略此邮件，您的账户密码不会发生变化。
                            </div>
                            <div style="margin-top:18px;font-family:'PingFang SC','Hiragino Sans GB','Microsoft YaHei',sans-serif;color:#6b7280;font-size:12px;line-height:20px;word-break:break-all;">
                                如果按钮无法点击，请复制以下链接到浏览器打开：<br>
                                <a href="%s" style="color:#0369a1;text-decoration:none;">%s</a>
                            </div>
                        </td>
                    </tr>
                    <tr>
                        <td style="padding:18px 28px;background:#f8fafc;border-top:1px solid #e6edf5;font-family:'PingFang SC','Hiragino Sans GB','Microsoft YaHei',sans-serif;color:#94a3b8;font-size:12px;line-height:20px;">
                            此邮件由系统自动发送，请勿直接回复。
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, safeSiteName, safeSiteName, safeResetURL, safeResetURL, safeResetURL)
}
