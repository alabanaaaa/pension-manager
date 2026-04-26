package benefits

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"pension-manager/internal/db"
)

type Service struct {
	db *db.DB
}

func NewService(db *db.DB) *Service {
	return &Service{db: db}
}

type DeathInService struct {
	ID            string             `json:"id"`
	MemberID      string             `json:"member_id"`
	SchemeID      string             `json:"scheme_id"`
	DateOfDeath   time.Time          `json:"date_of_death"`
	CauseOfDeath  string             `json:"cause_of_death"`
	Status        string             `json:"status"`
	ClaimID       *string            `json:"claim_id,omitempty"`
	TotalBenefit  int64              `json:"total_benefit"`
	Documents     []DeathDocument    `json:"documents,omitempty"`
	Beneficiaries []DeathBeneficiary `json:"beneficiaries,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

type DeathDocument struct {
	ID         string     `json:"id"`
	DISID      string     `json:"dis_id"`
	DocType    string     `json:"doc_type"`
	FileName   string     `json:"file_name"`
	FilePath   string     `json:"file_path"`
	Verified   bool       `json:"verified"`
	VerifiedBy *string    `json:"verified_by,omitempty"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	UploadedAt time.Time  `json:"uploaded_at"`
}

type DeathBeneficiary struct {
	ID               string     `json:"id"`
	DISID            string     `json:"dis_id"`
	BeneficiaryID    string     `json:"beneficiary_id"`
	Name             string     `json:"name"`
	Relationship     string     `json:"relationship"`
	AllocationPct    float64    `json:"allocation_pct"`
	AllocatedAmount  int64      `json:"allocated_amount"`
	Balance          int64      `json:"balance"`
	BankName         string     `json:"bank_name,omitempty"`
	BankBranch       string     `json:"bank_branch,omitempty"`
	BankAccount      string     `json:"bank_account,omitempty"`
	Status           string     `json:"status"` // pending, active, exhausted
	LastDrawdownDate *time.Time `json:"last_drawdown_date,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type Drawdown struct {
	ID            string    `json:"id"`
	BeneficiaryID string    `json:"beneficiary_id"`
	Amount        int64     `json:"amount"`
	PaymentRef    string    `json:"payment_ref"`
	EndUser       string    `json:"end_user"` // school, guardian, hospital, etc.
	Purpose       string    `json:"purpose"`
	ApprovedBy    string    `json:"approved_by"`
	CreatedAt     time.Time `json:"created_at"`
}

type DeathBenefitSubmission struct {
	MemberID     string    `json:"member_id"`
	SchemeID     string    `json:"scheme_id"`
	DateOfDeath  time.Time `json:"date_of_death"`
	CauseOfDeath string    `json:"cause_of_death"`
	Documents    []string  `json:"documents,omitempty"`
}

func (s *Service) RegisterDeath(ctx context.Context, submission *DeathBenefitSubmission, registeredBy string) (string, error) {
	var disID string

	err := s.db.Transactional(ctx, func(tx *sql.Tx) error {
		query := `
			INSERT INTO death_in_service (
				id, member_id, scheme_id, date_of_death, cause_of_death, status, created_at, updated_at
			) VALUES (
				uuid_generate_v4(), $1, $2, $3, $4, 'registered', NOW(), NOW()
			) RETURNING id
		`
		err := tx.QueryRowContext(ctx, query,
			submission.MemberID, submission.SchemeID, submission.DateOfDeath, submission.CauseOfDeath,
		).Scan(&disID)
		if err != nil {
			return fmt.Errorf("register death: %w", err)
		}

		_, _ = tx.ExecContext(ctx, `
			UPDATE members SET 
				membership_status = 'deceased',
				date_of_death = $1,
				updated_at = NOW()
			WHERE id = $2
		`, submission.DateOfDeath, submission.MemberID)

		s.logAuditEvent(ctx, tx, submission.SchemeID, "death_registered", disID, registeredBy,
			map[string]interface{}{
				"action":        "death_in_service_registered",
				"member_id":     submission.MemberID,
				"date_of_death": submission.DateOfDeath,
				"cause":         submission.CauseOfDeath,
			})

		return nil
	})

	return disID, err
}

func (s *Service) UploadDeathDocument(ctx context.Context, disID, docType, fileName, filePath, uploadedBy string) error {
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO death_documents (id, dis_id, doc_type, file_name, file_path, uploaded_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, NOW())
	`, disID, docType, fileName, filePath)
	if err != nil {
		return fmt.Errorf("upload document: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("failed to upload document")
	}

	return nil
}

func (s *Service) VerifyDocument(ctx context.Context, docID, verifiedBy string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE death_documents SET verified = true, verified_by = $1, verified_at = NOW()
		WHERE id = $2
	`, verifiedBy, docID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

func (s *Service) CalculateDeathBenefit(ctx context.Context, disID string, calculatedBy string) (int64, error) {
	var memberID, schemeID string
	var dateOfDeath time.Time
	err := s.db.QueryRowContext(ctx, `
		SELECT member_id, scheme_id, date_of_death FROM death_in_service WHERE id = $1
	`, disID).Scan(&memberID, &schemeID, &dateOfDeath)
	if err != nil {
		return 0, fmt.Errorf("get DIS: %w", err)
	}

	var accountBalance int64
	s.db.QueryRowContext(ctx, `
		SELECT account_balance FROM members WHERE id = $1
	`, memberID).Scan(&accountBalance)

	var totalContributions int64
	s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total_amount), 0) FROM contributions 
		WHERE member_id = $1
	`, memberID).Scan(&totalContributions)

	var totalWithdrawals int64
	s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0) FROM claims 
		WHERE member_id = $1 AND status = 'paid'
	`, memberID).Scan(&totalWithdrawals)

	benefitAmount := accountBalance + totalContributions - totalWithdrawals

	_, err = s.db.ExecContext(ctx, `
		UPDATE death_in_service SET total_benefit = $1 WHERE id = $2
	`, benefitAmount, disID)
	if err != nil {
		return 0, err
	}

	s.logAuditEvent(ctx, nil, schemeID, "death_benefit_calculated", disID, calculatedBy,
		map[string]interface{}{
			"action":         "death_benefit_calculated",
			"member_id":      memberID,
			"benefit_amount": benefitAmount,
		})

	return benefitAmount, nil
}

func (s *Service) DistributeToBeneficiaries(ctx context.Context, disID string, approvedBy string) error {
	return s.db.Transactional(ctx, func(tx *sql.Tx) error {
		var totalBenefit int64
		err := tx.QueryRowContext(ctx, `
			SELECT total_benefit FROM death_in_service WHERE id = $1
		`, disID).Scan(&totalBenefit)
		if err != nil {
			return err
		}

		if totalBenefit <= 0 {
			return fmt.Errorf("no benefit to distribute")
		}

		rows, err := tx.QueryContext(ctx, `
			SELECT id, name, relationship, allocation_pct
			FROM beneficiaries WHERE member_id = (
				SELECT member_id FROM death_in_service WHERE id = $1
			)
		`, disID)
		if err != nil {
			return err
		}
		defer rows.Close()

		var distributions []struct {
			beneficiaryID string
			name          string
			relationship  string
			allocation    float64
		}

		for rows.Next() {
			var d struct {
				beneficiaryID string
				name          string
				relationship  string
				allocation    float64
			}
			if err := rows.Scan(&d.beneficiaryID, &d.name, &d.relationship, &d.allocation); err != nil {
				continue
			}
			distributions = append(distributions, d)
		}

		for _, dist := range distributions {
			allocatedAmount := int64(float64(totalBenefit) * dist.allocation / 100)

			_, err := tx.ExecContext(ctx, `
				INSERT INTO death_beneficiaries (
					id, dis_id, beneficiary_id, name, relationship, allocation_pct,
					allocated_amount, balance, status, created_at
				) VALUES (
					uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $6, 'active', NOW()
				)
			`, disID, dist.beneficiaryID, dist.name, dist.relationship, dist.allocation, allocatedAmount)
			if err != nil {
				return fmt.Errorf("distribute to %s: %w", dist.name, err)
			}
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE death_in_service SET status = 'distributed' WHERE id = $1
		`, disID)

		s.logAuditEvent(ctx, tx, "", "death_benefit_distributed", disID, approvedBy,
			map[string]interface{}{
				"action": "benefits_distributed",
				"total":  totalBenefit,
				"count":  len(distributions),
			})

		return err
	})
}

func (s *Service) ProcessDrawdown(ctx context.Context, beneficiaryID string, amount int64, paymentRef, endUser, purpose, approvedBy string) error {
	return s.db.Transactional(ctx, func(tx *sql.Tx) error {
		var balance int64
		var disID string
		err := tx.QueryRowContext(ctx, `
			SELECT balance, dis_id FROM death_beneficiaries WHERE id = $1 AND status = 'active'
		`, beneficiaryID).Scan(&balance, &disID)
		if err != nil {
			return fmt.Errorf("get beneficiary: %w", err)
		}

		if amount > balance {
			return fmt.Errorf("drawdown amount (%d) exceeds balance (%d)", amount, balance)
		}

		drawdownID, err := s.insertDrawdown(ctx, tx, beneficiaryID, amount, paymentRef, endUser, purpose, approvedBy)
		if err != nil {
			return err
		}

		newBalance := balance - amount
		status := "active"
		var lastDrawdown *time.Time
		now := time.Now()
		if newBalance <= 0 {
			status = "exhausted"
			newBalance = 0
			lastDrawdown = &now
		} else {
			lastDrawdown = &now
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE death_beneficiaries SET balance = $1, status = $2, last_drawdown_date = $3
			WHERE id = $4
		`, newBalance, status, lastDrawdown, beneficiaryID)
		if err != nil {
			return fmt.Errorf("update balance: %w", err)
		}

		s.logAuditEvent(ctx, tx, "", "drawdown_processed", drawdownID, approvedBy,
			map[string]interface{}{
				"action":         "drawdown_processed",
				"beneficiary_id": beneficiaryID,
				"amount":         amount,
				"end_user":       endUser,
				"new_balance":    newBalance,
			})

		return nil
	})
}

func (s *Service) insertDrawdown(ctx context.Context, tx *sql.Tx, beneficiaryID string, amount int64, paymentRef, endUser, purpose, approvedBy string) (string, error) {
	var drawdownID string
	err := tx.QueryRowContext(ctx, `
		INSERT INTO beneficiary_drawdowns (id, beneficiary_id, amount, payment_ref, end_user, purpose, approved_by, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, NOW())
		RETURNING id
	`, beneficiaryID, amount, paymentRef, endUser, purpose, approvedBy).Scan(&drawdownID)
	return drawdownID, err
}

func (s *Service) GetDeathBeneficiary(ctx context.Context, disID string) ([]DeathBeneficiary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, dis_id, beneficiary_id, name, relationship, allocation_pct,
		       allocated_amount, balance, bank_name, bank_branch, bank_account, 
		       status, last_drawdown_date, created_at
		FROM death_beneficiaries WHERE dis_id = $1
	`, disID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var beneficiaries []DeathBeneficiary
	for rows.Next() {
		var b DeathBeneficiary
		var bankName, bankBranch, bankAccount sql.NullString
		var lastDrawdown sql.NullTime
		if err := rows.Scan(&b.ID, &b.DISID, &b.BeneficiaryID, &b.Name, &b.Relationship,
			&b.AllocationPct, &b.AllocatedAmount, &b.Balance, &bankName, &bankBranch,
			&bankAccount, &b.Status, &lastDrawdown, &b.CreatedAt); err != nil {
			continue
		}
		if bankName.Valid {
			b.BankName = bankName.String
		}
		if bankBranch.Valid {
			b.BankBranch = bankBranch.String
		}
		if bankAccount.Valid {
			b.BankAccount = bankAccount.String
		}
		if lastDrawdown.Valid {
			b.LastDrawdownDate = &lastDrawdown.Time
		}
		beneficiaries = append(beneficiaries, b)
	}
	return beneficiaries, rows.Err()
}

func (s *Service) AllocateInterest(ctx context.Context, disID string, annualRate float64, approvedBy string) error {
	return s.db.Transactional(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT id, balance FROM death_beneficiaries WHERE dis_id = $1 AND status = 'active'
		`, disID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			var balance int64
			if err := rows.Scan(&id, &balance); err != nil {
				continue
			}

			interest := int64(float64(balance) * annualRate / 100)
			newBalance := balance + interest

			_, err := tx.ExecContext(ctx, `
				UPDATE death_beneficiaries SET balance = $1 WHERE id = $2
			`, newBalance, id)
			if err != nil {
				return err
			}

			s.logAuditEvent(ctx, tx, "", "interest_allocated", id, approvedBy,
				map[string]interface{}{
					"action":      "interest_allocated",
					"dis_id":      disID,
					"rate":        annualRate,
					"interest":    interest,
					"new_balance": newBalance,
				})
		}

		return nil
	})
}

func (s *Service) GenerateBeneficiaryStatement(ctx context.Context, beneficiaryID string) (*BeneficiaryStatement, error) {
	var b DeathBeneficiary
	var disID string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, dis_id, beneficiary_id, name, relationship, allocation_pct,
		       allocated_amount, balance, status, created_at
		FROM death_beneficiaries WHERE id = $1
	`, beneficiaryID).Scan(&b.ID, &disID, &b.BeneficiaryID, &b.Name, &b.Relationship,
		&b.AllocationPct, &b.AllocatedAmount, &b.Balance, &b.Status, &b.CreatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, amount, payment_ref, end_user, purpose, approved_by, created_at
		FROM beneficiary_drawdowns WHERE beneficiary_id = $1 ORDER BY created_at DESC
	`, beneficiaryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drawdowns []Drawdown
	for rows.Next() {
		var d Drawdown
		if err := rows.Scan(&d.ID, &d.Amount, &d.PaymentRef, &d.EndUser, &d.Purpose, &d.ApprovedBy, &d.CreatedAt); err != nil {
			continue
		}
		drawdowns = append(drawdowns, d)
	}

	return &BeneficiaryStatement{
		Beneficiary: b,
		Drawdowns:   drawdowns,
		GeneratedAt: time.Now(),
	}, nil
}

type BeneficiaryStatement struct {
	Beneficiary DeathBeneficiary `json:"beneficiary"`
	Drawdowns   []Drawdown       `json:"drawdowns"`
	GeneratedAt time.Time        `json:"generated_at"`
}

func (s *Service) logAuditEvent(ctx context.Context, tx *sql.Tx, schemeID, entityType, entityID, actorID string, details map[string]interface{}) {
	detailsJSON, _ := json.Marshal(details)
	if tx != nil {
		_, _ = tx.ExecContext(ctx, `
			INSERT INTO audit_log (id, scheme_id, entity_type, entity_id, action, actor_id, details, created_at)
			VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, NOW())
		`, schemeID, entityType, entityID, entityType, actorID, detailsJSON)
	} else {
		_, _ = s.db.ExecContext(ctx, `
			INSERT INTO audit_log (id, scheme_id, entity_type, entity_id, action, actor_id, details, created_at)
			VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, NOW())
		`, schemeID, entityType, entityID, entityType, actorID, detailsJSON)
	}
}
