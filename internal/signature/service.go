package signature

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"pension-manager/internal/db"
)

type Service struct {
	db         *db.DB
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
}

func NewService(db *db.DB) *Service {
	return &Service{db: db}
}

type KeyPair struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

type DigitalSignature struct {
	ID          string    `json:"id"`
	EntityType  string    `json:"entity_type"`
	EntityID    string    `json:"entity_id"`
	SignerID    string    `json:"signer_id"`
	SignerRole  string    `json:"signer_role"`
	Signature   string    `json:"signature"`
	Hash        string    `json:"hash"`
	Algorithm   string    `json:"algorithm"`
	IPAddress   string    `json:"ip_address,omitempty"`
	UserAgent   string    `json:"user_agent,omitempty"`
	Geolocation string    `json:"geolocation,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type SignRequest struct {
	EntityType  string                 `json:"entity_type"`
	EntityID    string                 `json:"entity_id"`
	SignerID    string                 `json:"signer_id"`
	SignerRole  string                 `json:"signer_role"`
	Data        map[string]interface{} `json:"data"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Geolocation string                 `json:"geolocation,omitempty"`
}

type SignatureResult struct {
	Signature *DigitalSignature `json:"signature"`
	Verified  bool              `json:"verified"`
}

func (s *Service) Initialize() error {
	var exists bool
	s.db.QueryRowContext(context.Background(), `
		SELECT EXISTS(SELECT 1 FROM signing_keys WHERE key_id = 'system_master')
	`).Scan(&exists)

	if !exists {
		return s.generateAndStoreMasterKey()
	}

	return s.loadMasterKey()
}

func (s *Service) generateAndStoreMasterKey() error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("marshal private key: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	_, err = s.db.ExecContext(context.Background(), `
		INSERT INTO signing_keys (key_id, private_key, public_key, algorithm, created_at)
		VALUES ('system_master', $1, $2, 'ECDSA-P256', NOW())
	`, string(privateKeyPEM), string(publicKeyPEM))
	if err != nil {
		return fmt.Errorf("store keys: %w", err)
	}

	s.privateKey = privateKey
	s.publicKey = &privateKey.PublicKey

	return nil
}

func (s *Service) loadMasterKey() error {
	var privateKeyPEM string
	err := s.db.QueryRowContext(context.Background(), `
		SELECT private_key FROM signing_keys WHERE key_id = 'system_master'
	`).Scan(&privateKeyPEM)
	if err != nil {
		return fmt.Errorf("load private key: %w", err)
	}

	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return fmt.Errorf("decode PEM block")
	}

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse private key: %w", err)
	}

	s.privateKey = privateKey
	s.publicKey = &privateKey.PublicKey

	return nil
}

func (s *Service) Sign(ctx context.Context, req *SignRequest) (*SignatureResult, error) {
	if s.privateKey == nil {
		if err := s.Initialize(); err != nil {
			return nil, fmt.Errorf("initialize signing: %w", err)
		}
	}

	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		return nil, fmt.Errorf("marshal data: %w", err)
	}

	hash := sha256.Sum256(dataBytes)

	signature, err := ecdsa.SignASN1(rand.Reader, s.privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("sign data: %w", err)
	}

	var sigRecord DigitalSignature
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO digital_signatures (
			id, entity_type, entity_id, signer_id, signer_role, signature, hash,
			algorithm, ip_address, user_agent, geolocation, created_at
		) VALUES (
			uuid_generate_v4(), $1, $2, $3, $4, $5, $6, 'ECDSA-P256', $7, $8, $9, NOW()
		) RETURNING id, created_at
	`, req.EntityType, req.EntityID, req.SignerID, req.SignerRole,
		base64.StdEncoding.EncodeToString(signature),
		base64.StdEncoding.EncodeToString(hash[:]),
		req.IPAddress, req.UserAgent, req.Geolocation,
	).Scan(&sigRecord.ID, &sigRecord.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("store signature: %w", err)
	}

	sigRecord.EntityType = req.EntityType
	sigRecord.EntityID = req.EntityID
	sigRecord.SignerID = req.SignerID
	sigRecord.SignerRole = req.SignerRole
	sigRecord.Signature = base64.StdEncoding.EncodeToString(signature)
	sigRecord.Hash = base64.StdEncoding.EncodeToString(hash[:])
	sigRecord.Algorithm = "ECDSA-P256"
	sigRecord.IPAddress = req.IPAddress
	sigRecord.UserAgent = req.UserAgent
	sigRecord.Geolocation = req.Geolocation

	return &SignatureResult{
		Signature: &sigRecord,
		Verified:  true,
	}, nil
}

func (s *Service) Verify(ctx context.Context, entityType, entityID, signerID string, providedSignature string) (bool, error) {
	if s.publicKey == nil {
		if err := s.loadMasterKey(); err != nil {
			return false, err
		}
	}

	var sigRecord DigitalSignature
	err := s.db.QueryRowContext(ctx, `
		SELECT id, signature, hash, signer_id FROM digital_signatures
		WHERE entity_type = $1 AND entity_id = $2 AND signer_id = $3
		ORDER BY created_at DESC LIMIT 1
	`, entityType, entityID, signerID).Scan(
		&sigRecord.ID, &sigRecord.Signature, &sigRecord.Hash, &sigRecord.SignerID)
	if err != nil {
		return false, fmt.Errorf("get signature: %w", err)
	}

	if providedSignature != sigRecord.Signature {
		return false, nil
	}

	return true, nil
}

func (s *Service) GetSignatures(ctx context.Context, entityType, entityID string) ([]DigitalSignature, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, entity_type, entity_id, signer_id, signer_role, signature, hash,
		       algorithm, ip_address, user_agent, geolocation, created_at
		FROM digital_signatures
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
	`, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var signatures []DigitalSignature
	for rows.Next() {
		var sig DigitalSignature
		var ipAddress, userAgent, geolocation sql.NullString
		if err := rows.Scan(&sig.ID, &sig.EntityType, &sig.EntityID, &sig.SignerID,
			&sig.SignerRole, &sig.Signature, &sig.Hash, &sig.Algorithm,
			&ipAddress, &userAgent, &geolocation, &sig.CreatedAt); err != nil {
			continue
		}
		if ipAddress.Valid {
			sig.IPAddress = ipAddress.String
		}
		if userAgent.Valid {
			sig.UserAgent = userAgent.String
		}
		if geolocation.Valid {
			sig.Geolocation = geolocation.String
		}
		signatures = append(signatures, sig)
	}
	return signatures, rows.Err()
}

func (s *Service) GenerateMerkleRoot(ctx context.Context, startTime, endTime time.Time) (string, error) {
	var hashes []string

	rows, err := s.db.QueryContext(ctx, `
		SELECT hash FROM digital_signatures
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at ASC
	`, startTime, endTime)
	if err != nil {
		return "", fmt.Errorf("query signatures: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			continue
		}
		hashes = append(hashes, hash)
	}

	if len(hashes) == 0 {
		return "", fmt.Errorf("no signatures in time range")
	}

	merkleRoot := s.computeMerkleRoot(hashes)

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO merkle_roots (root_hash, start_time, end_time, signature_count, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, merkleRoot, startTime, endTime, len(hashes))
	if err != nil {
		return "", fmt.Errorf("store merkle root: %w", err)
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

func (s *Service) VerifyMerkleProof(leafHash, proof []string, rootHash string) bool {
	currentHash := leafHash[0]

	for _, p := range proof {
		combined := currentHash + p
		h := sha256.Sum256([]byte(combined))
		currentHash = base64.StdEncoding.EncodeToString(h[:])
	}

	return currentHash == rootHash
}

func (s *Service) SignData(data []byte) (string, error) {
	if s.privateKey == nil {
		if err := s.Initialize(); err != nil {
			return "", err
		}
	}

	hash := sha256.Sum256(data)
	signature, err := ecdsa.SignASN1(rand.Reader, s.privateKey, hash[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func (s *Service) VerifySignature(data, signatureBase64 string) bool {
	if s.publicKey == nil {
		return false
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false
	}

	hash := sha256.Sum256([]byte(data))

	r := new(big.Int).SetBytes(signatureBytes[:32])
	ss := new(big.Int).SetBytes(signatureBytes[32:])

	return ecdsa.Verify(s.publicKey, hash[:], r, ss)
}

func (s *Service) GetPublicKeyPEM() (string, error) {
	if s.publicKey == nil {
		if err := s.loadMasterKey(); err != nil {
			return "", err
		}
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(s.publicKey)
	if err != nil {
		return "", err
	}

	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})), nil
}

type MultiSigConfig struct {
	ID           string   `json:"id"`
	EntityType   string   `json:"entity_type"`
	RequiredSigs int      `json:"required_signatures"`
	SignerRoles  []string `json:"signer_roles"`
	Active       bool     `json:"active"`
}

func (s *Service) CreateMultiSigConfig(ctx context.Context, entityType string, requiredSigs int, signerRoles []string) error {
	rolesJSON, _ := json.Marshal(signerRoles)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO multi_sig_configs (id, entity_type, required_signatures, signer_roles, active, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, true, NOW())
	`, entityType, requiredSigs, rolesJSON)

	return err
}

func (s *Service) GetMultiSigConfig(ctx context.Context, entityType string) (*MultiSigConfig, error) {
	var config MultiSigConfig
	var rolesJSON []byte

	err := s.db.QueryRowContext(ctx, `
		SELECT id, entity_type, required_signatures, signer_roles, active
		FROM multi_sig_configs WHERE entity_type = $1 AND active = true
	`, entityType).Scan(&config.ID, &config.EntityType, &config.RequiredSigs, &rolesJSON, &config.Active)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(rolesJSON, &config.SignerRoles)

	return &config, nil
}
