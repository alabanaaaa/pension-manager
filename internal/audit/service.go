package audit

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
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

type ActorContext struct {
	UserID       string `json:"user_id"`
	UserRole     string `json:"user_role"`
	IPAddress    string `json:"ip_address"`
	UserAgent    string `json:"user_agent"`
	Geolocation  string `json:"geolocation"`
	DeviceFinger string `json:"device_fingerprint,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
}

type AuditEvent struct {
	ID           string                 `json:"id"`
	SchemeID     string                 `json:"scheme_id"`
	EntityType   string                 `json:"entity_type"`
	EntityID     string                 `json:"entity_id"`
	Action       string                 `json:"action"`
	ActorID      string                 `json:"actor_id"`
	ActorRole    string                 `json:"actor_role,omitempty"`
	BeforeValue  map[string]interface{} `json:"before_value,omitempty"`
	AfterValue   map[string]interface{} `json:"after_value,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	Geolocation  string                 `json:"geolocation,omitempty"`
	DeviceFinger string                 `json:"device_fingerprint,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
	RequestID    string                 `json:"request_id,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	Hash         string                 `json:"hash"`
	PreviousHash string                 `json:"previous_hash,omitempty"`
	Signature    string                 `json:"signature,omitempty"`
}

type AuditEntry struct {
	ID          string    `json:"id"`
	SchemeID    string    `json:"scheme_id"`
	EntityType  string    `json:"entity_type"`
	EntityID    string    `json:"entity_id"`
	Action      string    `json:"action"`
	ActorID     string    `json:"actor_id"`
	ActorRole   string    `json:"actor_role,omitempty"`
	BeforeData  string    `json:"before_data,omitempty"`
	AfterData   string    `json:"after_data,omitempty"`
	Details     string    `json:"details,omitempty"`
	IPAddress   string    `json:"ip_address,omitempty"`
	UserAgent   string    `json:"user_agent,omitempty"`
	Geolocation string    `json:"geolocation,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type ImmutableAuditLog struct {
	ID          string     `json:"id"`
	AuditLogID  string     `json:"audit_log_id"`
	ChainHash   string     `json:"chain_hash"`
	MerkleProof string     `json:"merkle_proof,omitempty"`
	Published   bool       `json:"published"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (s *Service) LogEvent(ctx context.Context, actor *ActorContext, entry *AuditEntry) error {
	var previousHash string
	err := s.db.QueryRowContext(ctx, `
		SELECT chain_hash FROM immutable_audit_log 
		ORDER BY created_at DESC LIMIT 1
	`).Scan(&previousHash)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("get previous hash: %w", err)
	}

	entryJSON, _ := json.Marshal(entry)
	hash := s.computeChainHash(string(entryJSON), previousHash)

	var auditID string
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO audit_log (
			id, scheme_id, entity_type, entity_id, action, actor_id, actor_role,
			before_data, after_data, details, ip_address, user_agent, geolocation,
			created_at, hash, previous_hash
		) VALUES (
			uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), $13, $14
		) RETURNING id
	`, entry.SchemeID, entry.EntityType, entry.EntityID, entry.Action,
		actor.UserID, actor.UserRole, entry.BeforeData, entry.AfterData,
		entry.Details, actor.IPAddress, actor.UserAgent, actor.Geolocation,
		hash, previousHash,
	).Scan(&auditID)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO immutable_audit_log (id, audit_log_id, chain_hash, created_at)
		VALUES (uuid_generate_v4(), $1, $2, NOW())
	`, auditID, hash)

	return err
}

func (s *Service) LogChange(ctx context.Context, actor *ActorContext, schemeID, entityType, entityID, action string, before, after interface{}) error {
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)

	return s.LogEvent(ctx, actor, &AuditEntry{
		SchemeID:   schemeID,
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		BeforeData: string(beforeJSON),
		AfterData:  string(afterJSON),
	})
}

func (s *Service) LogAction(ctx context.Context, actor *ActorContext, schemeID, entityType, entityID, action, details string) error {
	return s.LogEvent(ctx, actor, &AuditEntry{
		SchemeID:   schemeID,
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		Details:    details,
	})
}

func (s *Service) computeChainHash(entryJSON, previousHash string) string {
	data := entryJSON + previousHash
	h := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(h[:])
}

func (s *Service) QueryLogs(ctx context.Context, filters AuditFilters, limit, offset int) ([]AuditEvent, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM audit_log WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if filters.SchemeID != "" {
		countQuery += fmt.Sprintf(" AND scheme_id = $%d", argNum)
		args = append(args, filters.SchemeID)
		argNum++
	}
	if filters.EntityType != "" {
		countQuery += fmt.Sprintf(" AND entity_type = $%d", argNum)
		args = append(args, filters.EntityType)
		argNum++
	}
	if filters.EntityID != "" {
		countQuery += fmt.Sprintf(" AND entity_id = $%d", argNum)
		args = append(args, filters.EntityID)
		argNum++
	}
	if filters.ActorID != "" {
		countQuery += fmt.Sprintf(" AND actor_id = $%d", argNum)
		args = append(args, filters.ActorID)
		argNum++
	}
	if filters.Action != "" {
		countQuery += fmt.Sprintf(" AND action = $%d", argNum)
		args = append(args, filters.Action)
		argNum++
	}
	if !filters.StartDate.IsZero() {
		countQuery += fmt.Sprintf(" AND created_at >= $%d", argNum)
		args = append(args, filters.StartDate)
		argNum++
	}
	if !filters.EndDate.IsZero() {
		countQuery += fmt.Sprintf(" AND created_at <= $%d", argNum)
		args = append(args, filters.EndDate)
		argNum++
	}

	s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)

	query := `
		SELECT id, scheme_id, entity_type, entity_id, action, actor_id, actor_role,
		       before_data, after_data, details, ip_address, user_agent, geolocation,
		       created_at, hash, previous_hash
		FROM audit_log WHERE 1=1
	`
	args = []interface{}{}
	argNum = 1

	if filters.SchemeID != "" {
		query += fmt.Sprintf(" AND scheme_id = $%d", argNum)
		args = append(args, filters.SchemeID)
		argNum++
	}
	if filters.EntityType != "" {
		query += fmt.Sprintf(" AND entity_type = $%d", argNum)
		args = append(args, filters.EntityType)
		argNum++
	}
	if filters.EntityID != "" {
		query += fmt.Sprintf(" AND entity_id = $%d", argNum)
		args = append(args, filters.EntityID)
		argNum++
	}
	if filters.ActorID != "" {
		query += fmt.Sprintf(" AND actor_id = $%d", argNum)
		args = append(args, filters.ActorID)
		argNum++
	}
	if filters.Action != "" {
		query += fmt.Sprintf(" AND action = $%d", argNum)
		args = append(args, filters.Action)
		argNum++
	}
	if !filters.StartDate.IsZero() {
		query += fmt.Sprintf(" AND created_at >= $%d", argNum)
		args = append(args, filters.StartDate)
		argNum++
	}
	if !filters.EndDate.IsZero() {
		query += fmt.Sprintf(" AND created_at <= $%d", argNum)
		args = append(args, filters.EndDate)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var e AuditEvent
		var beforeData, afterData, details, actorRole sql.NullString
		var ipAddress, userAgent, geolocation sql.NullString

		if err := rows.Scan(&e.ID, &e.SchemeID, &e.EntityType, &e.EntityID, &e.Action,
			&e.ActorID, &actorRole, &beforeData, &afterData, &details,
			&ipAddress, &userAgent, &geolocation, &e.Timestamp, &e.Hash, &e.PreviousHash); err != nil {
			continue
		}

		if actorRole.Valid {
			e.ActorRole = actorRole.String
		}
		if ipAddress.Valid {
			e.IPAddress = ipAddress.String
		}
		if userAgent.Valid {
			e.UserAgent = userAgent.String
		}
		if geolocation.Valid {
			e.Geolocation = geolocation.String
		}

		if beforeData.Valid && beforeData.String != "" {
			json.Unmarshal([]byte(beforeData.String), &e.BeforeValue)
		}
		if afterData.Valid && afterData.String != "" {
			json.Unmarshal([]byte(afterData.String), &e.AfterValue)
		}

		events = append(events, e)
	}

	return events, total, rows.Err()
}

type AuditFilters struct {
	SchemeID   string
	EntityType string
	EntityID   string
	ActorID    string
	Action     string
	StartDate  time.Time
	EndDate    time.Time
}

func (s *Service) GetEntityHistory(ctx context.Context, entityType, entityID string) ([]AuditEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, scheme_id, entity_type, entity_id, action, actor_id, actor_role,
		       before_data, after_data, details, ip_address, user_agent, geolocation,
		       created_at, hash, previous_hash
		FROM audit_log
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at ASC
	`, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var e AuditEvent
		var beforeData, afterData, details, actorRole sql.NullString
		var ipAddress, userAgent, geolocation sql.NullString

		if err := rows.Scan(&e.ID, &e.SchemeID, &e.EntityType, &e.EntityID, &e.Action,
			&e.ActorID, &actorRole, &beforeData, &afterData, &details,
			&ipAddress, &userAgent, &geolocation, &e.Timestamp, &e.Hash, &e.PreviousHash); err != nil {
			continue
		}

		if actorRole.Valid {
			e.ActorRole = actorRole.String
		}
		if ipAddress.Valid {
			e.IPAddress = ipAddress.String
		}
		if userAgent.Valid {
			e.UserAgent = userAgent.String
		}
		if geolocation.Valid {
			e.Geolocation = geolocation.String
		}

		if beforeData.Valid && beforeData.String != "" {
			json.Unmarshal([]byte(beforeData.String), &e.BeforeValue)
		}
		if afterData.Valid && afterData.String != "" {
			json.Unmarshal([]byte(afterData.String), &e.AfterValue)
		}

		events = append(events, e)
	}

	return events, rows.Err()
}

func (s *Service) VerifyChainIntegrity(ctx context.Context, startTime, endTime time.Time) (bool, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, hash, previous_hash FROM audit_log
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at ASC
	`, startTime, endTime)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var previousHash string
	for rows.Next() {
		var id, currentHash string
		var previousHashFromDB sql.NullString

		if err := rows.Scan(&id, &currentHash, &previousHashFromDB); err != nil {
			continue
		}

		if previousHashFromDB.Valid {
			if previousHash != "" && previousHashFromDB.String != previousHash {
				return false, fmt.Errorf("chain broken at %s: expected %s, got %s", id, previousHash, previousHashFromDB.String)
			}
		}

		previousHash = currentHash
	}

	return true, nil
}

func (s *Service) GenerateDailyMerkleRoot(ctx context.Context, date time.Time) (string, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	rows, err := s.db.QueryContext(ctx, `
		SELECT chain_hash FROM immutable_audit_log
		WHERE created_at >= $1 AND created_at < $2
		ORDER BY created_at ASC
	`, startOfDay, endOfDay)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var hashes []string
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			continue
		}
		hashes = append(hashes, hash)
	}

	if len(hashes) == 0 {
		return "", fmt.Errorf("no audit logs for %s", date.Format("2006-01-02"))
	}

	merkleRoot := s.computeMerkleRoot(hashes)

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO daily_merkle_roots (root_hash, date, signature_count, created_at)
		VALUES ($1, $2, $3, NOW())
	`, merkleRoot, startOfDay, len(hashes))
	if err != nil {
		return "", err
	}

	return merkleRoot, nil
}

func (s *Service) computeMerkleRoot(hashes []string) string {
	if len(hashes) == 0 {
		return ""
	}

	if len(hashes) == 1 {
		return hashes[0]
	}

	var pairs []string
	for i := 0; i < len(hashes); i += 2 {
		var pair string
		if i+1 < len(hashes) {
			pair = hashes[i] + hashes[i+1]
		} else {
			pair = hashes[i] + hashes[i]
		}
		h := sha256.Sum256([]byte(pair))
		pairs = append(pairs, base64.StdEncoding.EncodeToString(h[:]))
	}

	return s.computeMerkleRoot(pairs)
}

func (s *Service) GetActivitySummary(ctx context.Context, schemeID string, startDate, endDate time.Time) (*ActivitySummary, error) {
	summary := &ActivitySummary{}

	s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT actor_id) FROM audit_log
		WHERE scheme_id = $1 AND created_at >= $2 AND created_at <= $3
	`, schemeID, startDate, endDate).Scan(&summary.UniqueActors)

	s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM audit_log
		WHERE scheme_id = $1 AND created_at >= $2 AND created_at <= $3
	`, schemeID, startDate, endDate).Scan(&summary.TotalActions)

	s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM audit_log
		WHERE scheme_id = $1 AND created_at >= $2 AND created_at <= $3
		AND action LIKE '%reject%' OR action LIKE '%fail%'
	`, schemeID, startDate, endDate).Scan(&summary.FailedActions)

	s.db.QueryRowContext(ctx, `
		SELECT action, COUNT(*) as count FROM audit_log
		WHERE scheme_id = $1 AND created_at >= $2 AND created_at <= $3
		GROUP BY action ORDER BY count DESC LIMIT 10
	`, schemeID, startDate, endDate).Scan(&summary.TopActions)

	return summary, nil
}

type ActivitySummary struct {
	UniqueActors  int            `json:"unique_actors"`
	TotalActions  int            `json:"total_actions"`
	FailedActions int            `json:"failed_actions"`
	TopActions    map[string]int `json:"top_actions"`
}

func (s *Service) CreateActorContext(ctx context.Context, userID, userRole, ipAddress, userAgent, geolocation string) *ActorContext {
	return &ActorContext{
		UserID:      userID,
		UserRole:    userRole,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Geolocation: geolocation,
		RequestID:   generateRequestID(),
	}
}

func generateRequestID() string {
	b := make([]byte, 16)
	now := time.Now().UnixNano()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d-%d", now, time.Now().Unix())))
	copy(b, hash[:16])
	return base64.StdEncoding.EncodeToString(b)
}
