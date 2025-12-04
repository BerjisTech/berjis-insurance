package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/insurance-broker/backend/internal/config"
)

// EmailClient defines behavior for sending transactional emails
type EmailClient interface {
	SendOTP(ctx context.Context, to, code string) error
	SendPasswordReset(ctx context.Context, to, token string) error
}

const sendGridAPI = "https://api.sendgrid.com/v3/mail/send"

// NewEmailClient selects an email client based on configuration
func NewEmailClient(cfg *config.EmailConfig) EmailClient {
	if cfg == nil {
		return &ConsoleEmailClient{}
	}

	if cfg.APIKey == "" {
		return &ConsoleEmailClient{}
	}

	switch cfg.Provider {
	case "sendgrid":
		return &SendGridEmailClient{
			apiKey:      cfg.APIKey,
			fromAddress: cfg.FromAddress,
			fromName:    cfg.FromName,
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}
	default:
		return &ConsoleEmailClient{}
	}
}

// SendGridEmailClient implements EmailClient using SendGrid REST API
type SendGridEmailClient struct {
	apiKey      string
	fromAddress string
	fromName    string
	httpClient  *http.Client
}

func (c *SendGridEmailClient) SendOTP(ctx context.Context, to, code string) error {
	subject := "Your Insurance OTP"
	body := fmt.Sprintf("Your verification code is %s. It expires in 10 minutes.", code)
	return c.send(ctx, to, subject, body)
}

func (c *SendGridEmailClient) SendPasswordReset(ctx context.Context, to, token string) error {
	subject := "Password Reset Instructions"
	body := fmt.Sprintf("Use this token to reset your password: %s", token)
	return c.send(ctx, to, subject, body)
}

func (c *SendGridEmailClient) send(ctx context.Context, to, subject, body string) error {
	payload := map[string]interface{}{
		"personalizations": []map[string]interface{}{{
			"to":      []map[string]string{{"email": to}},
			"subject": subject,
		}},
		"from": map[string]string{
			"email": c.fromAddress,
			"name":  c.fromName,
		},
		"content": []map[string]string{{
			"type":  "text/plain",
			"value": body,
		}},
	}

	buf, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sendGridAPI, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sendgrid error status: %d", resp.StatusCode)
	}

	return nil
}

// ConsoleEmailClient logs emails to stdout (development fallback)
type ConsoleEmailClient struct{}

func (c *ConsoleEmailClient) SendOTP(ctx context.Context, to, code string) error {
	log.Printf("[email] OTP -> %s : %s", to, code)
	return nil
}

func (c *ConsoleEmailClient) SendPasswordReset(ctx context.Context, to, token string) error {
	log.Printf("[email] Password reset token -> %s : %s", to, token)
	return nil
}
