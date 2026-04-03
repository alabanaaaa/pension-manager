package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// registerTaxReminderRoutes registers tax exemption reminder routes
func (s *Server) registerTaxReminderRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(RoleMiddleware("super_admin", "admin", "pension_officer"))

		r.Route("/tax/reminders", func(r chi.Router) {
			r.Get("/expiring", s.handleGetExpiringExemptions)
			r.Get("/overdue", s.handleGetOverdueExemptions)
			r.Get("/pending", s.handleGetPendingReminders)
			r.Post("/send", s.handleSendTaxReminders)
		})
	})
}

func (s *Server) handleGetExpiringExemptions(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	daysStr := r.URL.Query().Get("days")
	if daysStr == "" {
		daysStr = "30"
	}
	days, _ := strconv.Atoi(daysStr)

	reminders, err := s.taxReminderSvc.GetExpiringExemptions(r.Context(), schemeID, days)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, reminders)
}

func (s *Server) handleGetOverdueExemptions(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	reminders, err := s.taxReminderSvc.GetOverdueExemptions(r.Context(), schemeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, reminders)
}

func (s *Server) handleGetPendingReminders(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	reminders, err := s.taxReminderSvc.GetPendingReminders(r.Context(), schemeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, reminders)
}

func (s *Server) handleSendTaxReminders(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	// Get pending reminders
	reminders, err := s.taxReminderSvc.GetPendingReminders(r.Context(), schemeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Send SMS to each member
	var sent, failed int
	for _, reminder := range reminders {
		if reminder.Phone != "" {
			err := s.smsService.SendTaxExemptionReminder(r.Context(), reminder.Phone, reminder.MemberNo, reminder.DaysLeft)
			if err != nil {
				failed++
				continue
			}
		}
		// Record that reminder was sent
		_ = s.taxReminderSvc.RecordReminderSent(r.Context(), reminder.MemberID, "kra_renewal")
		sent++
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"sent":   sent,
		"failed": failed,
		"total":  len(reminders),
	})
}
