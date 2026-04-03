package security

import (
	"context"
	"fmt"
	"net"
	"time"

	"pension-manager/internal/db"
)

// IPBlacklist manages IP address blacklisting for security
type Service struct {
	db *db.DB
}

// NewIPBlacklistService creates a new IP blacklist service
func NewIPBlacklistService(db *db.DB) *Service {
	return &Service{db: db}
}

// BlacklistIP adds an IP address to the blacklist
func (s *Service) BlacklistIP(ctx context.Context, ipAddress, reason, addedBy string) error {
	// Validate IP
	if net.ParseIP(ipAddress) == nil {
		return fmt.Errorf("invalid IP address: %s", ipAddress)
	}

	query := `
		INSERT INTO ip_blacklist (id, ip_address, reason, added_by, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, NOW())
		ON CONFLICT (ip_address) DO UPDATE SET reason = $2, added_by = $3, active = true, updated_at = NOW()
	`
	_, err := s.db.ExecContext(ctx, query, ipAddress, reason, addedBy)
	if err != nil {
		return fmt.Errorf("blacklist IP: %w", err)
	}
	return nil
}

// RemoveIP removes an IP address from the blacklist
func (s *Service) RemoveIP(ctx context.Context, ipAddress string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE ip_blacklist SET active = false, updated_at = NOW() WHERE ip_address = $1
	`, ipAddress)
	return err
}

// IsBlacklisted checks if an IP address is blacklisted
func (s *Service) IsBlacklisted(ctx context.Context, ipAddress string) (bool, string, error) {
	var reason string
	err := s.db.QueryRowContext(ctx, `
		SELECT reason FROM ip_blacklist WHERE ip_address = $1 AND active = true
	`, ipAddress).Scan(&reason)
	if err != nil {
		return false, "", nil
	}
	return true, reason, nil
}

// ListBlacklistedIPs returns all active blacklisted IPs
func (s *Service) ListBlacklistedIPs(ctx context.Context) ([]BlacklistedIP, error) {
	query := `
		SELECT id, ip_address, reason, added_by, created_at, updated_at
		FROM ip_blacklist WHERE active = true
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list blacklisted IPs: %w", err)
	}
	defer rows.Close()

	var ips []BlacklistedIP
	for rows.Next() {
		var ip BlacklistedIP
		if err := rows.Scan(&ip.ID, &ip.IPAddress, &ip.Reason, &ip.AddedBy, &ip.CreatedAt, &ip.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan blacklisted IP: %w", err)
		}
		ips = append(ips, ip)
	}
	return ips, rows.Err()
}

// GetLoginAttempts returns login attempts for an IP
func (s *Service) GetLoginAttempts(ctx context.Context, ipAddress string, hours int) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM login_attempts
		WHERE ip_address = $1 AND attempted_at > NOW() - ($2 || ' hours')::interval
	`
	err := s.db.QueryRowContext(ctx, query, ipAddress, hours).Scan(&count)
	return count, err
}

// RecordLoginAttempt records a login attempt
func (s *Service) RecordLoginAttempt(ctx context.Context, ipAddress, email string, success bool) error {
	query := `
		INSERT INTO login_attempts (id, ip_address, email, success, attempted_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, NOW())
	`
	_, err := s.db.ExecContext(ctx, query, ipAddress, email, success)
	return err
}

// BlacklistedIP represents a blacklisted IP address
type BlacklistedIP struct {
	ID        string    `json:"id"`
	IPAddress string    `json:"ip_address"`
	Reason    string    `json:"reason"`
	AddedBy   string    `json:"added_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
