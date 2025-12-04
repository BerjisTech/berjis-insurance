package services

import (
	"context"
	"fmt"

	"github.com/insurance-broker/backend/internal/clients"
)

// NotificationSender abstracts sending OTPs and transactional emails
type NotificationSender interface {
	SendOTP(ctx context.Context, channel, identifier, code string) error
	SendPasswordReset(ctx context.Context, email, token string) error
}

// NotificationService implements NotificationSender using email and sms clients
type NotificationService struct {
	emailClient clients.EmailClient
	smsClient   clients.SMSClient
}

// NewNotificationService creates a notification dispatcher
func NewNotificationService(email clients.EmailClient, sms clients.SMSClient) *NotificationService {
	return &NotificationService{
		emailClient: email,
		smsClient:   sms,
	}
}

// SendOTP delivers OTP over preferred channel
func (n *NotificationService) SendOTP(ctx context.Context, channel, identifier, code string) error {
	switch channel {
	case "sms":
		if n.smsClient != nil {
			return n.smsClient.Send(ctx, identifier, fmt.Sprintf("Your verification code is %s", code))
		}
	default:
		if n.emailClient != nil {
			return n.emailClient.SendOTP(ctx, identifier, code)
		}
	}
	return nil
}

// SendPasswordReset emails password reset token
func (n *NotificationService) SendPasswordReset(ctx context.Context, email, token string) error {
	if n.emailClient != nil {
		return n.emailClient.SendPasswordReset(ctx, email, token)
	}
	return nil
}
