package sponsor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"pension-manager/core/domain"
	"pension-manager/internal/db"
)

// Service manages sponsors and contribution schedules
type Service struct {
	db *db.DB
}

// NewService creates a new sponsor service
func NewService(db *db.DB) *Service {
	return &Service{db: db}
}

// CreateSponsor creates a new sponsor
func (s *Service) CreateSponsor(ctx context.Context, sponsor *domain.Sponsor) error {
	if err := sponsor.Validate(); err != nil {
		return err
	}

	now := time.Now()
	sponsor.CreatedAt = now
	sponsor.UpdatedAt = now
	if sponsor.Status == "" {
		sponsor.Status = "active"
	}

	query := `
		INSERT INTO sponsors (id, scheme_id, code, name, contact_person, phone, email, address,
		                      pay_mode, bank_name, bank_branch, bank_account, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := s.db.ExecContext(ctx, query,
		sponsor.ID, sponsor.SchemeID, sponsor.Code, sponsor.Name, sponsor.ContactPerson,
		sponsor.Phone, sponsor.Email, sponsor.Address, sponsor.PayMode,
		sponsor.BankName, sponsor.BankBranch, sponsor.BankAccount, sponsor.Status,
		sponsor.CreatedAt, sponsor.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create sponsor: %w", err)
	}
	return nil
}

// GetSponsor retrieves a sponsor by ID
func (s *Service) GetSponsor(ctx context.Context, sponsorID string) (*domain.Sponsor, error) {
	query := `
		SELECT id, scheme_id, code, name, contact_person, phone, email, address,
		       pay_mode, bank_name, bank_branch, bank_account, status, created_at, updated_at
		FROM sponsors WHERE id = $1
	`
	sp := &domain.Sponsor{}
	err := s.db.QueryRowContext(ctx, query, sponsorID).Scan(
		&sp.ID, &sp.SchemeID, &sp.Code, &sp.Name, &sp.ContactPerson, &sp.Phone, &sp.Email,
		&sp.Address, &sp.PayMode, &sp.BankName, &sp.BankBranch, &sp.BankAccount,
		&sp.Status, &sp.CreatedAt, &sp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get sponsor: %w", err)
	}
	return sp, nil
}

// UpdateSponsor updates an existing sponsor
func (s *Service) UpdateSponsor(ctx context.Context, sponsor *domain.Sponsor) error {
	if err := sponsor.Validate(); err != nil {
		return err
	}
	sponsor.UpdatedAt = time.Now()

	query := `
		UPDATE sponsors SET name = $1, contact_person = $2, phone = $3, email = $4,
		                    address = $5, pay_mode = $6, bank_name = $7, bank_branch = $8,
		                    bank_account = $9, status = $10, updated_at = $11
		WHERE id = $12
	`
	result, err := s.db.ExecContext(ctx, query,
		sponsor.Name, sponsor.ContactPerson, sponsor.Phone, sponsor.Email,
		sponsor.Address, sponsor.PayMode, sponsor.BankName, sponsor.BankBranch,
		sponsor.BankAccount, sponsor.Status, sponsor.UpdatedAt, sponsor.ID,
	)
	if err != nil {
		return fmt.Errorf("update sponsor: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("sponsor not found")
	}
	return nil
}

// ListSponsors returns all sponsors for a scheme
func (s *Service) ListSponsors(ctx context.Context, schemeID string) ([]*domain.Sponsor, error) {
	query := `
		SELECT s.id, s.scheme_id, s.code, s.name, s.contact_person, s.phone, s.email, s.address,
		       s.pay_mode, s.bank_name, s.bank_branch, s.bank_account, s.status, s.created_at, s.updated_at,
		       (SELECT COUNT(*) FROM members WHERE sponsor_id = s.id) as total_members
		FROM sponsors s WHERE s.scheme_id = $1 ORDER BY s.name
	`
	rows, err := s.db.QueryContext(ctx, query, schemeID)
	if err != nil {
		return nil, fmt.Errorf("list sponsors: %w", err)
	}
	defer rows.Close()

	var sponsors []*domain.Sponsor
	for rows.Next() {
		sp := &domain.Sponsor{}
		if err := rows.Scan(
			&sp.ID, &sp.SchemeID, &sp.Code, &sp.Name, &sp.ContactPerson, &sp.Phone, &sp.Email,
			&sp.Address, &sp.PayMode, &sp.BankName, &sp.BankBranch, &sp.BankAccount,
			&sp.Status, &sp.CreatedAt, &sp.UpdatedAt, &sp.TotalMembers,
		); err != nil {
			return nil, fmt.Errorf("scan sponsor: %w", err)
		}
		sponsors = append(sponsors, sp)
	}
	return sponsors, rows.Err()
}

// GetSponsorStats returns contribution stats for a sponsor
func (s *Service) GetSponsorStats(ctx context.Context, sponsorID string) (*SponsorStats, error) {
	stats := &SponsorStats{}
	query := `
		SELECT
			COUNT(*) as total_schedules,
			COUNT(*) FILTER (WHERE status = 'posted') as posted_schedules,
			COUNT(*) FILTER (WHERE status = 'pending') as pending_schedules,
			COALESCE(SUM(total_amount), 0) as total_contributions,
			COALESCE(SUM(total_employees), 0) as total_employees
		FROM contribution_schedules WHERE sponsor_id = $1
	`
	err := s.db.QueryRowContext(ctx, query, sponsorID).Scan(
		&stats.TotalSchedules, &stats.PostedSchedules, &stats.PendingSchedules,
		&stats.TotalContributions, &stats.TotalEmployees,
	)
	if err != nil {
		return nil, fmt.Errorf("get sponsor stats: %w", err)
	}
	return stats, nil
}

// CreateContributionSchedule creates a new contribution schedule
func (s *Service) CreateContributionSchedule(ctx context.Context, schedule *domain.ContributionSchedule) error {
	if err := schedule.Validate(); err != nil {
		return err
	}

	schedule.CreatedAt = time.Now()
	if schedule.Status == "" {
		schedule.Status = "pending"
	}

	// Calculate diffs if previous period exists
	var prevEmployees int
	var prevAmount int64
	err := s.db.QueryRowContext(ctx, `
		SELECT total_employees, total_amount FROM contribution_schedules
		WHERE sponsor_id = $1 AND period < $2
		ORDER BY period DESC LIMIT 1
	`, schedule.SponsorID, schedule.Period).Scan(&prevEmployees, &prevAmount)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("get previous schedule: %w", err)
	}

	schedule.PrevEmployees = prevEmployees
	schedule.PrevAmount = prevAmount
	schedule.EmployeeDiff = schedule.TotalEmployees - prevEmployees
	schedule.AmountDiff = schedule.TotalAmount - prevAmount

	query := `
		INSERT INTO contribution_schedules (id, sponsor_id, scheme_id, period, total_employees,
		                                    total_amount, prev_employees, prev_amount,
		                                    employee_diff, amount_diff, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err = s.db.ExecContext(ctx, query,
		schedule.ID, schedule.SponsorID, schedule.SchemeID, schedule.Period,
		schedule.TotalEmployees, schedule.TotalAmount, schedule.PrevEmployees,
		schedule.PrevAmount, schedule.EmployeeDiff, schedule.AmountDiff,
		schedule.Status, schedule.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create contribution schedule: %w", err)
	}
	return nil
}

// ListContributionSchedules returns schedules for a sponsor
func (s *Service) ListContributionSchedules(ctx context.Context, sponsorID string) ([]*domain.ContributionSchedule, error) {
	query := `
		SELECT id, sponsor_id, scheme_id, period, total_employees, total_amount,
		       prev_employees, prev_amount, employee_diff, amount_diff,
		       status, reconciliation_notes, created_at, posted_at
		FROM contribution_schedules WHERE sponsor_id = $1
		ORDER BY period DESC
	`
	rows, err := s.db.QueryContext(ctx, query, sponsorID)
	if err != nil {
		return nil, fmt.Errorf("list schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*domain.ContributionSchedule
	for rows.Next() {
		cs := &domain.ContributionSchedule{}
		var notes sql.NullString
		var postedAt sql.NullTime
		if err := rows.Scan(
			&cs.ID, &cs.SponsorID, &cs.SchemeID, &cs.Period, &cs.TotalEmployees,
			&cs.TotalAmount, &cs.PrevEmployees, &cs.PrevAmount, &cs.EmployeeDiff,
			&cs.AmountDiff, &cs.Status, &notes, &cs.CreatedAt, &postedAt,
		); err != nil {
			return nil, fmt.Errorf("scan schedule: %w", err)
		}
		if notes.Valid {
			cs.ReconciliationNotes = notes.String
		}
		if postedAt.Valid {
			cs.PostedAt = postedAt.Time
		}
		schedules = append(schedules, cs)
	}
	return schedules, rows.Err()
}

// PostSchedule marks a schedule as posted and allocates contributions
func (s *Service) PostSchedule(ctx context.Context, scheduleID string) error {
	return s.db.Transactional(ctx, func(tx *sql.Tx) error {
		var sponsorID string
		var period time.Time
		var totalAmount int64
		err := tx.QueryRowContext(ctx, `
			SELECT sponsor_id, period, total_amount FROM contribution_schedules WHERE id = $1 AND status = 'balanced'
		`, scheduleID).Scan(&sponsorID, &period, &totalAmount)
		if err == sql.ErrNoRows {
			return errors.New("schedule not found or not in balanced status")
		}
		if err != nil {
			return fmt.Errorf("get schedule: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE contribution_schedules SET status = 'posted', posted_at = NOW() WHERE id = $1
		`, scheduleID)
		if err != nil {
			return fmt.Errorf("post schedule: %w", err)
		}

		// Record audit event
		_, err = tx.ExecContext(ctx, `
			INSERT INTO audit_log (id, scheme_id, entity_type, entity_id, action, actor_id, created_at)
			VALUES (uuid_generate_v4(), (SELECT scheme_id FROM contribution_schedules WHERE id = $1),
			        'contribution_schedule', $1, 'posted', 'system', NOW())
		`, scheduleID)
		return err
	})
}

// SponsorStats holds aggregate statistics for a sponsor
type SponsorStats struct {
	TotalSchedules     int   `json:"total_schedules"`
	PostedSchedules    int   `json:"posted_schedules"`
	PendingSchedules   int   `json:"pending_schedules"`
	TotalContributions int64 `json:"total_contributions"`
	TotalEmployees     int64 `json:"total_employees"`
}
