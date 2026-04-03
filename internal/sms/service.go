package sms

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Provider is the interface for SMS gateway providers
type Provider interface {
	Send(ctx context.Context, to, message string) error
	SendBulk(ctx context.Context, messages []Message) ([]SendResult, error)
	CheckBalance(ctx context.Context) (float64, error)
	Name() string
}

// Message represents an SMS message
type Message struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

// SendResult holds the result of a single SMS send
type SendResult struct {
	To        string `json:"to"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	MessageID string `json:"message_id,omitempty"`
}

// DeliveryStatus holds SMS delivery status
type DeliveryStatus struct {
	MessageID   string    `json:"message_id"`
	Status      string    `json:"status"` // delivered, failed, pending
	DeliveredAt time.Time `json:"delivered_at,omitempty"`
}

// AfricaTalkingProvider implements Provider using Africa's Talking API
type AfricaTalkingProvider struct {
	APIKey    string
	Username  string
	ShortCode string
	BaseURL   string
}

// NewAfricaTalkingProvider creates a new Africa's Talking SMS provider
func NewAfricaTalkingProvider(apiKey, username, shortCode string) *AfricaTalkingProvider {
	return &AfricaTalkingProvider{
		APIKey:    apiKey,
		Username:  username,
		ShortCode: shortCode,
		BaseURL:   "https://api.africastalking.com/version1/messaging",
	}
}

func (p *AfricaTalkingProvider) Name() string {
	return "africastalking"
}

func (p *AfricaTalkingProvider) Send(ctx context.Context, to, message string) error {
	results, err := p.SendBulk(ctx, []Message{{To: to, Message: message}})
	if err != nil {
		return err
	}
	if len(results) > 0 && !results[0].Success {
		return fmt.Errorf("sms send failed: %s", results[0].Error)
	}
	return nil
}

func (p *AfricaTalkingProvider) SendBulk(ctx context.Context, messages []Message) ([]SendResult, error) {
	if len(messages) == 0 {
		return nil, nil
	}

	// Build recipients and messages
	var recipients []string
	var messageTexts []string
	for _, m := range messages {
		recipients = append(recipients, m.To)
		messageTexts = append(messageTexts, m.Message)
	}

	// Africa's Talking supports multiple recipients in one request
	// For simplicity, we'll send them one by one
	var results []SendResult
	for _, m := range messages {
		formData := url.Values{}
		formData.Set("username", p.Username)
		formData.Set("to", m.To)
		formData.Set("message", m.Message)
		if p.ShortCode != "" {
			formData.Set("from", p.ShortCode)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL, strings.NewReader(formData.Encode()))
		if err != nil {
			results = append(results, SendResult{To: m.To, Success: false, Error: err.Error()})
			continue
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("apiKey", p.APIKey)
		req.Header.Set("Accept", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			results = append(results, SendResult{To: m.To, Success: false, Error: err.Error()})
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			results = append(results, SendResult{To: m.To, Success: true, MessageID: fmt.Sprintf("at-%d", time.Now().UnixNano())})
		} else {
			results = append(results, SendResult{To: m.To, Success: false, Error: fmt.Sprintf("HTTP %d", resp.StatusCode)})
		}
	}

	return results, nil
}

func (p *AfricaTalkingProvider) CheckBalance(ctx context.Context) (float64, error) {
	// Africa's Talking balance check endpoint
	return 0, nil // Simplified - would call actual balance API
}

// MockProvider implements Provider for testing/sandbox
type MockProvider struct {
	SentMessages []Message
}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) Name() string {
	return "mock"
}

func (p *MockProvider) Send(ctx context.Context, to, message string) error {
	p.SentMessages = append(p.SentMessages, Message{To: to, Message: message})
	return nil
}

func (p *MockProvider) SendBulk(ctx context.Context, messages []Message) ([]SendResult, error) {
	var results []SendResult
	for _, m := range messages {
		p.SentMessages = append(p.SentMessages, m)
		results = append(results, SendResult{To: m.To, Success: true, MessageID: fmt.Sprintf("mock-%d", time.Now().UnixNano())})
	}
	return results, nil
}

func (p *MockProvider) CheckBalance(ctx context.Context) (float64, error) {
	return 9999.99, nil
}

// Service manages SMS operations
type Service struct {
	provider Provider
}

// NewService creates a new SMS service
func NewService(provider Provider) *Service {
	return &Service{provider: provider}
}

// SendOTP sends an OTP code to a phone number
func (s *Service) SendOTPSMS(ctx context.Context, phone, otp string) error {
	message := fmt.Sprintf("Your verification code is: %s. Valid for 5 minutes. Do not share this code.", otp)
	return s.provider.Send(ctx, phone, message)
}

// SendSMS sends a single SMS message
func (s *Service) SendSMS(ctx context.Context, to, message string) error {
	return s.provider.Send(ctx, to, message)
}

// SendBulkSMS sends messages to multiple recipients
func (s *Service) SendBulkSMS(ctx context.Context, messages []Message) ([]SendResult, error) {
	return s.provider.SendBulk(ctx, messages)
}

// SendMemberNotification sends a notification to a member
func (s *Service) SendMemberNotification(ctx context.Context, phone, subject, message string) error {
	fullMessage := fmt.Sprintf("[%s] %s", subject, message)
	return s.provider.Send(ctx, phone, fullMessage)
}

// SendContributionAlert sends a contribution alert
func (s *Service) SendContributionAlert(ctx context.Context, phone, memberNo string, amount int64) error {
	message := fmt.Sprintf("Dear member %s, your contribution of KES %d has been received. Thank you.", memberNo, amount)
	return s.provider.Send(ctx, phone, message)
}

// SendClaimStatusUpdate sends a claim status update
func (s *Service) SendClaimStatusUpdate(ctx context.Context, phone, claimNo, status string) error {
	message := fmt.Sprintf("Your claim %s status has been updated to: %s. Login to portal for details.", claimNo, status)
	return s.provider.Send(ctx, phone, message)
}

// SendElectionReminder sends an election voting reminder
func (s *Service) SendElectionReminder(ctx context.Context, phone, electionTitle string) error {
	message := fmt.Sprintf("Reminder: Voting for %s is now open. Login to the member portal to cast your vote.", electionTitle)
	return s.provider.Send(ctx, phone, message)
}

// SendTaxExemptionReminder sends a KRA tax exemption renewal reminder
func (s *Service) SendTaxExemptionReminder(ctx context.Context, phone, memberNo string, daysUntilExpiry int) error {
	message := fmt.Sprintf("Dear member %s, your KRA tax exemption certificate expires in %d days. Please renew it to avoid tax computation on your benefits.", memberNo, daysUntilExpiry)
	return s.provider.Send(ctx, phone, message)
}

// CheckBalance checks the SMS provider balance
func (s *Service) CheckBalance(ctx context.Context) (float64, error) {
	return s.provider.CheckBalance(ctx)
}

// ProviderName returns the name of the current SMS provider
func (s *Service) ProviderName() string {
	return s.provider.Name()
}
