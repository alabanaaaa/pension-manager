package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// registerIPBlacklistRoutes registers IP blacklist management routes
func (s *Server) registerIPBlacklistRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(RoleMiddleware("super_admin", "admin"))

		r.Route("/api/security/ip-blacklist", func(r chi.Router) {
			r.Post("/", s.handleBlacklistIP)
			r.Delete("/{ip}", s.handleRemoveIP)
			r.Get("/", s.handleListBlacklistedIPs)
			r.Get("/check/{ip}", s.handleCheckIP)
		})
	})
}

func (s *Server) handleBlacklistIP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IPAddress string `json:"ip_address"`
		Reason    string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.IPAddress == "" || req.Reason == "" {
		respondError(w, http.StatusBadRequest, "ip_address and reason are required")
		return
	}

	userID := GetUserID(r)
	if err := s.ipBlacklistSvc.BlacklistIP(r.Context(), req.IPAddress, req.Reason, userID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"status": "blacklisted", "ip": req.IPAddress})
}

func (s *Server) handleRemoveIP(w http.ResponseWriter, r *http.Request) {
	ipAddress := chi.URLParam(r, "ip")
	if ipAddress == "" {
		respondError(w, http.StatusBadRequest, "IP address is required")
		return
	}

	if err := s.ipBlacklistSvc.RemoveIP(r.Context(), ipAddress); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "removed", "ip": ipAddress})
}

func (s *Server) handleListBlacklistedIPs(w http.ResponseWriter, r *http.Request) {
	ips, err := s.ipBlacklistSvc.ListBlacklistedIPs(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, ips)
}

func (s *Server) handleCheckIP(w http.ResponseWriter, r *http.Request) {
	ipAddress := chi.URLParam(r, "ip")
	if ipAddress == "" {
		respondError(w, http.StatusBadRequest, "IP address is required")
		return
	}

	isBlacklisted, reason, err := s.ipBlacklistSvc.IsBlacklisted(r.Context(), ipAddress)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"ip_address":     ipAddress,
		"is_blacklisted": isBlacklisted,
		"reason":         reason,
	})
}
