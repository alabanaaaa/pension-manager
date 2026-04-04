package api

import (
	"encoding/json"
	"net/http"

	"pension-manager/internal/sms"

	"github.com/go-chi/chi/v5"
)

// registerSMSRoutes registers SMS gateway routes
func (s *Server) registerSMSRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(RoleMiddleware("super_admin", "admin", "pension_officer"))

		r.Route("/api/sms", func(r chi.Router) {
			r.Post("/send", s.handleSendSMS)
			r.Post("/send/bulk", s.handleSendBulkSMS)
			r.Post("/send/otp", s.handleSendOTPSMS)
			r.Post("/send/member-notification", s.handleSendMemberNotification)
			r.Post("/send/contribution-alert", s.handleSendContributionAlert)
			r.Post("/send/claim-update", s.handleSendClaimUpdate)
			r.Post("/send/election-reminder", s.handleSendElectionReminder)
			r.Get("/balance", s.handleSMSBalance)
			r.Get("/provider", s.handleSMSProvider)
		})
	})
}

func (s *Server) handleSendSMS(w http.ResponseWriter, r *http.Request) {
	var req struct {
		To      string `json:"to"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.To == "" || req.Message == "" {
		respondError(w, http.StatusBadRequest, "to and message are required")
		return
	}

	if err := s.smsService.SendSMS(r.Context(), req.To, req.Message); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "sent", "to": req.To})
}

func (s *Server) handleSendBulkSMS(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Messages []sms.Message `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.Messages) == 0 {
		respondError(w, http.StatusBadRequest, "messages array is required")
		return
	}

	results, err := s.smsService.SendBulkSMS(r.Context(), req.Messages)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"total":   len(results),
		"success": successCount,
		"failed":  len(results) - successCount,
		"results": results,
	})
}

func (s *Server) handleSendOTPSMS(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
		OTP   string `json:"otp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.smsService.SendOTPSMS(r.Context(), req.Phone, req.OTP); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "otp_sent"})
}

func (s *Server) handleSendMemberNotification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone   string `json:"phone"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.smsService.SendMemberNotification(r.Context(), req.Phone, req.Subject, req.Message); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "notification_sent"})
}

func (s *Server) handleSendContributionAlert(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone    string `json:"phone"`
		MemberNo string `json:"member_no"`
		Amount   int64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.smsService.SendContributionAlert(r.Context(), req.Phone, req.MemberNo, req.Amount); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "alert_sent"})
}

func (s *Server) handleSendClaimUpdate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone   string `json:"phone"`
		ClaimNo string `json:"claim_no"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.smsService.SendClaimStatusUpdate(r.Context(), req.Phone, req.ClaimNo, req.Status); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "update_sent"})
}

func (s *Server) handleSendElectionReminder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone         string `json:"phone"`
		ElectionTitle string `json:"election_title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.smsService.SendElectionReminder(r.Context(), req.Phone, req.ElectionTitle); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "reminder_sent"})
}

func (s *Server) handleSMSBalance(w http.ResponseWriter, r *http.Request) {
	balance, err := s.smsService.CheckBalance(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"balance":  balance,
		"provider": s.smsService.ProviderName(),
	})
}

func (s *Server) handleSMSProvider(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"provider": s.smsService.ProviderName(),
	})
}
