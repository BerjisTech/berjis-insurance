package clients

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/insurance-broker/backend/internal/config"
)

// SMSClient defines behavior for sending SMS messages
type SMSClient interface {
	Send(ctx context.Context, to, message string) error
}

const africasTalkingAPI = "https://api.africastalking.com/version1/messaging"

// NewSMSClient creates an SMS client for the configured provider
func NewSMSClient(cfg *config.SMSConfig) SMSClient {
	if cfg == nil || cfg.APIKey == "" {
		return &ConsoleSMSClient{}
	}

	return &AfricasTalkingClient{
		username: cfg.Username,
		apiKey:   cfg.APIKey,
		senderID: cfg.SenderID,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// AfricasTalkingClient implements SMS sending via Africa's Talking API
type AfricasTalkingClient struct {
	username   string
	apiKey     string
	senderID   string
	httpClient *http.Client
}

func (c *AfricasTalkingClient) Send(ctx context.Context, to, message string) error {
	data := url.Values{}
	data.Set("username", c.username)
	data.Set("to", to)
	data.Set("message", message)
	if c.senderID != "" {
		data.Set("from", c.senderID)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, africasTalkingAPI, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apiKey", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("africasTalking status %d", resp.StatusCode)
	}

	return nil
}

// ConsoleSMSClient logs sms messages (development fallback)
type ConsoleSMSClient struct{}

func (c *ConsoleSMSClient) Send(ctx context.Context, to, message string) error {
	log.Printf("[sms] -> %s : %s", to, message)
	return nil
}
