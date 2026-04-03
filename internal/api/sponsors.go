package api

import (
	"encoding/json"
	"net/http"
	"time"

	"pension-manager/core/domain"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// registerSponsorRoutes registers sponsor management routes
func (s *Server) registerSponsorRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(RoleMiddleware("admin", "pension_officer", "super_admin"))

		r.Route("/sponsors", func(r chi.Router) {
			r.Post("/", s.handleCreateSponsor)
			r.Get("/", s.handleListSponsors)
			r.Get("/{id}", s.handleGetSponsor)
			r.Put("/{id}", s.handleUpdateSponsor)
			r.Get("/{id}/stats", s.handleGetSponsorStats)
			r.Get("/{id}/schedules", s.handleListSponsorSchedules)
			r.Post("/{id}/schedules", s.handleCreateContributionSchedule)
			r.Post("/schedules/{scheduleId}/post", s.handlePostSchedule)
		})
	})
}

// handleCreateSponsor handles POST /sponsors
func (s *Server) handleCreateSponsor(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	var sp domain.Sponsor
	if err := json.NewDecoder(r.Body).Decode(&sp); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	sp.SchemeID = schemeID
	sp.ID = uuid.New().String()

	if err := s.sponsorService.CreateSponsor(r.Context(), &sp); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, sp)
}

// handleGetSponsor handles GET /sponsors/{id}
func (s *Server) handleGetSponsor(w http.ResponseWriter, r *http.Request) {
	sponsorID := chi.URLParam(r, "id")
	if sponsorID == "" {
		respondError(w, http.StatusBadRequest, "sponsor ID is required")
		return
	}

	sp, err := s.sponsorService.GetSponsor(r.Context(), sponsorID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sp == nil {
		respondError(w, http.StatusNotFound, "sponsor not found")
		return
	}

	respondJSON(w, http.StatusOK, sp)
}

// handleListSponsors handles GET /sponsors
func (s *Server) handleListSponsors(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	sponsors, err := s.sponsorService.ListSponsors(r.Context(), schemeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, sponsors)
}

// handleUpdateSponsor handles PUT /sponsors/{id}
func (s *Server) handleUpdateSponsor(w http.ResponseWriter, r *http.Request) {
	sponsorID := chi.URLParam(r, "id")
	if sponsorID == "" {
		respondError(w, http.StatusBadRequest, "sponsor ID is required")
		return
	}

	var sp domain.Sponsor
	if err := json.NewDecoder(r.Body).Decode(&sp); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	sp.ID = sponsorID

	if err := s.sponsorService.UpdateSponsor(r.Context(), &sp); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, sp)
}

// handleGetSponsorStats handles GET /sponsors/{id}/stats
func (s *Server) handleGetSponsorStats(w http.ResponseWriter, r *http.Request) {
	sponsorID := chi.URLParam(r, "id")
	if sponsorID == "" {
		respondError(w, http.StatusBadRequest, "sponsor ID is required")
		return
	}

	stats, err := s.sponsorService.GetSponsorStats(r.Context(), sponsorID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// handleListSponsorSchedules handles GET /sponsors/{id}/schedules
func (s *Server) handleListSponsorSchedules(w http.ResponseWriter, r *http.Request) {
	sponsorID := chi.URLParam(r, "id")
	if sponsorID == "" {
		respondError(w, http.StatusBadRequest, "sponsor ID is required")
		return
	}

	schedules, err := s.sponsorService.ListContributionSchedules(r.Context(), sponsorID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, schedules)
}

// handleCreateContributionSchedule handles POST /sponsors/{id}/schedules
func (s *Server) handleCreateContributionSchedule(w http.ResponseWriter, r *http.Request) {
	sponsorID := chi.URLParam(r, "id")
	schemeID := GetSchemeID(r)

	var req struct {
		Period         string `json:"period"`
		TotalEmployees int    `json:"total_employees"`
		TotalAmount    int64  `json:"total_amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	period, err := time.Parse("2006-01-02", req.Period)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid period format (use YYYY-MM-DD)")
		return
	}

	schedule := &domain.ContributionSchedule{
		ID:             uuid.New().String(),
		SponsorID:      sponsorID,
		SchemeID:       schemeID,
		Period:         period,
		TotalEmployees: req.TotalEmployees,
		TotalAmount:    req.TotalAmount,
	}

	if err := s.sponsorService.CreateContributionSchedule(r.Context(), schedule); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, schedule)
}

// handlePostSchedule handles POST /sponsors/schedules/{scheduleId}/post
func (s *Server) handlePostSchedule(w http.ResponseWriter, r *http.Request) {
	scheduleID := chi.URLParam(r, "scheduleId")
	if scheduleID == "" {
		respondError(w, http.StatusBadRequest, "schedule ID is required")
		return
	}

	if err := s.sponsorService.PostSchedule(r.Context(), scheduleID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "posted", "schedule_id": scheduleID})
}
