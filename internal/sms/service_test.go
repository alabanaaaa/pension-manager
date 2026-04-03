package sms

import (
	"context"
	"testing"
)

func TestMockProvider_Send(t *testing.T) {
	provider := NewMockProvider()

	err := provider.Send(context.Background(), "+254712345678", "Hello World")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(provider.SentMessages) != 1 {
		t.Errorf("Expected 1 sent message, got: %d", len(provider.SentMessages))
	}
	if provider.SentMessages[0].To != "+254712345678" {
		t.Errorf("Expected to +254712345678, got: %s", provider.SentMessages[0].To)
	}
	if provider.SentMessages[0].Message != "Hello World" {
		t.Errorf("Expected message 'Hello World', got: %s", provider.SentMessages[0].Message)
	}
}

func TestMockProvider_SendBulk(t *testing.T) {
	provider := NewMockProvider()

	messages := []Message{
		{To: "+254712345678", Message: "Hello 1"},
		{To: "+254712345679", Message: "Hello 2"},
		{To: "+254712345680", Message: "Hello 3"},
	}

	results, err := provider.SendBulk(context.Background(), messages)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got: %d", len(results))
	}

	for i, r := range results {
		if !r.Success {
			t.Errorf("Expected result %d to be successful", i)
		}
		if r.MessageID == "" {
			t.Errorf("Expected result %d to have message ID", i)
		}
	}

	if len(provider.SentMessages) != 3 {
		t.Errorf("Expected 3 sent messages, got: %d", len(provider.SentMessages))
	}
}

func TestMockProvider_SendEmpty(t *testing.T) {
	provider := NewMockProvider()

	results, err := provider.SendBulk(context.Background(), []Message{})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty messages, got: %d", len(results))
	}
}

func TestMockProvider_CheckBalance(t *testing.T) {
	provider := NewMockProvider()

	balance, err := provider.CheckBalance(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if balance != 9999.99 {
		t.Errorf("Expected balance 9999.99, got: %f", balance)
	}
}

func TestMockProvider_Name(t *testing.T) {
	provider := NewMockProvider()

	if provider.Name() != "mock" {
		t.Errorf("Expected name 'mock', got: %s", provider.Name())
	}
}

func TestService_SendOTP(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider)

	err := service.SendOTPSMS(context.Background(), "+254712345678", "123456")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(provider.SentMessages) != 1 {
		t.Errorf("Expected 1 sent message, got: %d", len(provider.SentMessages))
	}

	expectedPrefix := "Your verification code is: 123456"
	if len(provider.SentMessages[0].Message) < len(expectedPrefix) ||
		provider.SentMessages[0].Message[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Expected OTP message prefix, got: %s", provider.SentMessages[0].Message)
	}
}

func TestService_SendBulkMessages(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider)

	messages := []Message{
		{To: "+254712345678", Message: "Message 1"},
		{To: "+254712345679", Message: "Message 2"},
	}

	results, err := service.SendBulkSMS(context.Background(), messages)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got: %d", len(results))
	}
	for _, r := range results {
		if !r.Success {
			t.Errorf("Expected success, got failure for %s", r.To)
		}
	}
}

func TestService_SendMemberNotification(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider)

	err := service.SendMemberNotification(context.Background(), "+254712345678", "Alert", "Your account has been updated")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(provider.SentMessages) != 1 {
		t.Errorf("Expected 1 sent message, got: %d", len(provider.SentMessages))
	}

	expected := "[Alert] Your account has been updated"
	if provider.SentMessages[0].Message != expected {
		t.Errorf("Expected message '%s', got: %s", expected, provider.SentMessages[0].Message)
	}
}

func TestService_SendContributionAlert(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider)

	err := service.SendContributionAlert(context.Background(), "+254712345678", "M001", 500000)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expected := "Dear member M001, your contribution of KES 500000 has been received. Thank you."
	if provider.SentMessages[0].Message != expected {
		t.Errorf("Expected message '%s', got: %s", expected, provider.SentMessages[0].Message)
	}
}

func TestService_SendClaimStatusUpdate(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider)

	err := service.SendClaimStatusUpdate(context.Background(), "+254712345678", "CLM-001", "Approved")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expected := "Your claim CLM-001 status has been updated to: Approved. Login to portal for details."
	if provider.SentMessages[0].Message != expected {
		t.Errorf("Expected message '%s', got: %s", expected, provider.SentMessages[0].Message)
	}
}

func TestService_SendElectionReminder(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider)

	err := service.SendElectionReminder(context.Background(), "+254712345678", "Trustee Election 2026")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expected := "Reminder: Voting for Trustee Election 2026 is now open. Login to the member portal to cast your vote."
	if provider.SentMessages[0].Message != expected {
		t.Errorf("Expected message '%s', got: %s", expected, provider.SentMessages[0].Message)
	}
}

func TestService_SendTaxExemptionReminder(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider)

	err := service.SendTaxExemptionReminder(context.Background(), "+254712345678", "M001", 15)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expected := "Dear member M001, your KRA tax exemption certificate expires in 15 days. Please renew it to avoid tax computation on your benefits."
	if provider.SentMessages[0].Message != expected {
		t.Errorf("Expected message '%s', got: %s", expected, provider.SentMessages[0].Message)
	}
}

func TestService_ProviderName(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider)

	if service.ProviderName() != "mock" {
		t.Errorf("Expected provider name 'mock', got: %s", service.ProviderName())
	}
}

func TestService_CheckBalance(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider)

	balance, err := service.CheckBalance(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if balance != 9999.99 {
		t.Errorf("Expected balance 9999.99, got: %f", balance)
	}
}
