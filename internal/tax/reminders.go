package tax

import (
	"context"
	"fmt"
	"time"

	"pension-manager/internal/db"
)

// Reminder represents a tax exemption reminder
type Reminder struct {
	ID         string     `json:"id"`
	MemberID   string     `json:"member_id"`
	MemberNo   string     `json:"member_no"`
	FullName   string     `json:"full_name"`
	Email      string     `json:"email"`
	Phone      string     `json:"phone"`
	ExpiryDate time.Time  `json:"expiry_date"`
	DaysLeft   int        `json:"days_left"`
	SentAt     *time.Time `json:"sent_at,omitempty"`
}

// Service manages tax exemption reminders
type Service struct {
	db *db.DB
}

// NewReminderService creates a new tax exemption reminder service
func NewReminderService(db *db.DB) *Service {
	return &Service{db: db}
}

// GetExpiringExemptions returns members whose tax exemption is expiring soon
func (s *Service) GetExpiringExemptions(ctx context.Context, schemeID string, daysAhead int) ([]Reminder, error) {
	query := `
		SELECT m.id, m.member_no, m.first_name || ' ' || m.last_name, m.email, m.phone,
		       m.tax_exempt_cutoff_date,
		       (m.tax_exempt_cutoff_date - NOW())::int as days_left
		FROM members m
		WHERE m.scheme_id = $1
		  AND m.tax_exempt_cutoff_date IS NOT NULL
		  AND m.tax_exempt_cutoff_date > NOW()
		  AND m.tax_exempt_cutoff_date <= NOW() + ($2 || ' days')::interval
		ORDER BY m.tax_exempt_cutoff_date ASC
	`
	rows, err := s.db.QueryContext(ctx, query, schemeID, daysAhead)
	if err != nil {
		return nil, fmt.Errorf("query expiring exemptions: %w", err)
	}
	defer rows.Close()

	var reminders []Reminder
	for rows.Next() {
		var r Reminder
		var email, phone *string
		if err := rows.Scan(&r.MemberID, &r.MemberNo, &r.FullName, &email, &phone, &r.ExpiryDate, &r.DaysLeft); err != nil {
			continue
		}
		if email != nil {
			r.Email = *email
		}
		if phone != nil {
			r.Phone = *phone
		}
		reminders = append(reminders, r)
	}
	return reminders, rows.Err()
}

// GetOverdueExemptions returns members whose tax exemption has expired
func (s *Service) GetOverdueExemptions(ctx context.Context, schemeID string) ([]Reminder, error) {
	query := `
		SELECT m.id, m.member_no, m.first_name || ' ' || m.last_name, m.email, m.phone,
		       m.tax_exempt_cutoff_date,
		       (NOW() - m.tax_exempt_cutoff_date)::int as days_left
		FROM members m
		WHERE m.scheme_id = $1
		  AND m.tax_exempt_cutoff_date IS NOT NULL
		  AND m.tax_exempt_cutoff_date < NOW()
		ORDER BY m.tax_exempt_cutoff_date ASC
	`
	rows, err := s.db.QueryContext(ctx, query, schemeID)
	if err != nil {
		return nil, fmt.Errorf("query overdue exemptions: %w", err)
	}
	defer rows.Close()

	var reminders []Reminder
	for rows.Next() {
		var r Reminder
		var email, phone *string
		if err := rows.Scan(&r.MemberID, &r.MemberNo, &r.FullName, &email, &phone, &r.ExpiryDate, &r.DaysLeft); err != nil {
			continue
		}
		if email != nil {
			r.Email = *email
		}
		if phone != nil {
			r.Phone = *phone
		}
		reminders = append(reminders, r)
	}
	return reminders, rows.Err()
}

// RecordReminderSent records that a reminder was sent
func (s *Service) RecordReminderSent(ctx context.Context, memberID string, reminderType string) error {
	query := `
		INSERT INTO tax_exemption_reminders (id, member_id, scheme_id, reminder_type, due_date, sent_at, created_at)
		SELECT uuid_generate_v4(), $1, m.scheme_id, $2, m.tax_exempt_cutoff_date, NOW(), NOW()
		FROM members m WHERE m.id = $1
	`
	_, err := s.db.ExecContext(ctx, query, memberID, reminderType)
	return err
}

// GetPendingReminders returns reminders that haven't been sent yet
func (s *Service) GetPendingReminders(ctx context.Context, schemeID string) ([]Reminder, error) {
	query := `
		SELECT m.id, m.member_no, m.first_name || ' ' || m.last_name, m.email, m.phone,
		       m.tax_exempt_cutoff_date,
		       (m.tax_exempt_cutoff_date - NOW())::int as days_left
		FROM members m
		WHERE m.scheme_id = $1
		  AND m.tax_exempt_cutoff_date IS NOT NULL
		  AND m.tax_exempt_cutoff_date <= NOW() + INTERVAL '30 days'
		  AND NOT EXISTS (
		    SELECT 1 FROM tax_exemption_reminders ter
		    WHERE ter.member_id = m.id
		      AND ter.sent_at > NOW() - INTERVAL '7 days'
		  )
		ORDER BY m.tax_exempt_cutoff_date ASC
	`
	rows, err := s.db.QueryContext(ctx, query, schemeID)
	if err != nil {
		return nil, fmt.Errorf("query pending reminders: %w", err)
	}
	defer rows.Close()

	var reminders []Reminder
	for rows.Next() {
		var r Reminder
		var email, phone *string
		if err := rows.Scan(&r.MemberID, &r.MemberNo, &r.FullName, &email, &phone, &r.ExpiryDate, &r.DaysLeft); err != nil {
			continue
		}
		if email != nil {
			r.Email = *email
		}
		if phone != nil {
			r.Phone = *phone
		}
		reminders = append(reminders, r)
	}
	return reminders, rows.Err()
}
