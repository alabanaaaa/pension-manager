package api

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"pension-manager/internal/signature"

	"github.com/go-chi/chi/v5"
)

func (s *Server) registerSignatureRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(RoleMiddleware("super_admin", "admin", "pension_officer"))
		r.Post("/sign", s.handleSignDocument)
		r.Get("/verify/{entityType}/{entityId}", s.handleVerifySignature)
		r.Get("/{entityType}/{entityId}", s.handleGetSignatures)
		r.Post("/merkle/generate", s.handleGenerateMerkleRoot)
		r.Get("/public-key", s.handleGetPublicKey)
	})
	r.Group(func(r chi.Router) {
		r.Use(RoleMiddleware("super_admin", "admin"))
		r.Post("/multisig/config", s.handleCreateMultiSigConfig)
		r.Get("/multisig/config/{entityType}", s.handleGetMultiSigConfig)
	})
}

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	return r.RemoteAddr
}

func (s *Server) handleSignDocument(w http.ResponseWriter, r *http.Request) {
	var req signature.SignRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.EntityType == "" || req.EntityID == "" {
		respondError(w, http.StatusBadRequest, "entity_type and entity_id are required")
		return
	}

	req.IPAddress = getClientIP(r)
	req.UserAgent = r.UserAgent()

	result, err := s.signatureService.Sign(r.Context(), &req)
	if err != nil {
		slog.Error("sign document failed", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to sign document")
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handleVerifySignature(w http.ResponseWriter, r *http.Request) {
	entityType := chi.URLParam(r, "entityType")
	entityID := chi.URLParam(r, "entityId")

	if entityType == "" || entityID == "" {
		respondError(w, http.StatusBadRequest, "entity_type and entity_id are required")
		return
	}

	signerID := r.URL.Query().Get("signer_id")
	sig := r.URL.Query().Get("signature")

	if signerID == "" || sig == "" {
		respondError(w, http.StatusBadRequest, "signer_id and signature are required")
		return
	}

	valid, err := s.signatureService.Verify(r.Context(), entityType, entityID, signerID, sig)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "verification failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"valid":       valid,
		"entity_type": entityType,
		"entity_id":   entityID,
	})
}

func (s *Server) handleGetSignatures(w http.ResponseWriter, r *http.Request) {
	entityType := chi.URLParam(r, "entityType")
	entityID := chi.URLParam(r, "entityId")

	signatures, err := s.signatureService.GetSignatures(r.Context(), entityType, entityID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get signatures")
		return
	}

	respondJSON(w, http.StatusOK, signatures)
}

func (s *Server) handleGenerateMerkleRoot(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid start_time format, use RFC3339")
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid end_time format, use RFC3339")
		return
	}

	merkleRoot, err := s.signatureService.GenerateMerkleRoot(r.Context(), startTime, endTime)
	if err != nil {
		slog.Error("generate merkle root failed", "error", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"merkle_root": merkleRoot,
	})
}

func (s *Server) handleGetPublicKey(w http.ResponseWriter, r *http.Request) {
	publicKey, err := s.signatureService.GetPublicKeyPEM()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get public key")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"public_key": publicKey,
		"algorithm":  "ECDSA-P256",
	})
}

func (s *Server) handleCreateMultiSigConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EntityType   string   `json:"entity_type"`
		RequiredSigs int      `json:"required_signatures"`
		SignerRoles  []string `json:"signer_roles"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.EntityType == "" || req.RequiredSigs < 1 || len(req.SignerRoles) == 0 {
		respondError(w, http.StatusBadRequest, "entity_type, required_signatures (>0), and signer_roles are required")
		return
	}

	err := s.signatureService.CreateMultiSigConfig(r.Context(), req.EntityType, req.RequiredSigs, req.SignerRoles)
	if err != nil {
		slog.Error("create multisig config failed", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create multi-sig config")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (s *Server) handleGetMultiSigConfig(w http.ResponseWriter, r *http.Request) {
	entityType := chi.URLParam(r, "entityType")

	config, err := s.signatureService.GetMultiSigConfig(r.Context(), entityType)
	if err != nil {
		respondError(w, http.StatusNotFound, "multi-sig config not found")
		return
	}

	respondJSON(w, http.StatusOK, config)
}
