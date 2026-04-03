package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"pension-manager/core/domain"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// registerVotingRoutes registers online voting routes
func (s *Server) registerVotingRoutes(r chi.Router) {
	// Admin routes (protected)
	r.Group(func(r chi.Router) {
		r.Use(RoleMiddleware("super_admin", "admin", "pension_officer"))

		r.Route("/voting/admin", func(r chi.Router) {
			// Elections
			r.Post("/elections", s.handleCreateElection)
			r.Get("/elections", s.handleListElections)
			r.Get("/elections/{id}", s.handleGetElection)
			r.Put("/elections/{id}/status", s.handleUpdateElectionStatus)

			// Candidates
			r.Post("/elections/{electionId}/candidates", s.handleAddCandidate)
			r.Get("/elections/{electionId}/candidates", s.handleListCandidates)

			// Voter Register
			r.Post("/elections/{electionId}/voters", s.handleAddVoter)
			r.Post("/elections/{electionId}/voters/bulk", s.handleBulkAddVoters)

			// Results & Reports
			r.Get("/elections/{electionId}/results", s.handleGetResults)
			r.Get("/elections/{electionId}/results/station/{station}", s.handleGetResultsByStation)
			r.Get("/elections/{electionId}/results/scheme/{schemeType}", s.handleGetResultsByScheme)
			r.Get("/elections/{electionId}/voted-members", s.handleGetVotedMembers)
			r.Get("/elections/{electionId}/not-voted-members", s.handleGetNotVotedMembers)
			r.Get("/elections/{electionId}/stats", s.handleGetVotingStats)
			r.Get("/elections/{electionId}/export/results", s.handleExportResults)
		})
	})

	// Member voting routes (web portal)
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(s.auth))
		r.Use(MemberPortalMiddleware(s.db))

		r.Route("/voting", func(r chi.Router) {
			r.Get("/elections", s.handleMemberListElections)
			r.Get("/elections/{id}", s.handleMemberGetElection)
			r.Get("/elections/{id}/candidates", s.handleMemberListCandidates)
			r.Post("/elections/{id}/vote", s.handleCastVote)
			r.Get("/elections/{id}/my-votes", s.handleGetMyVotes)
			r.Get("/elections/{id}/live-results", s.handleGetLiveResults)
		})
	})

	// USSD voting (no auth - uses phone number validation)
	r.Group(func(r chi.Router) {
		r.Post("/voting/ussd", s.handleUSSDVote)
	})

	// URL-based voting (token-based, no session required)
	r.Get("/voting/url/{token}", s.handleURLVote)
}

// handleCreateElection handles POST /voting/admin/elections
func (s *Server) handleCreateElection(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	userID := GetUserID(r)

	var req struct {
		Title         string `json:"title"`
		Description   string `json:"description,omitempty"`
		Type          string `json:"type"`
		MaxCandidates int    `json:"max_candidates"`
		StartDate     string `json:"start_date"`
		EndDate       string `json:"end_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	startDate, err := time.Parse("2006-01-02T15:04:05Z", req.StartDate)
	if err != nil {
		startDate, err = time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid start_date format")
			return
		}
	}

	endDate, err := time.Parse("2006-01-02T15:04:05Z", req.EndDate)
	if err != nil {
		endDate, err = time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid end_date format")
			return
		}
	}

	election := &domain.Election{
		ID:            uuid.New().String(),
		SchemeID:      schemeID,
		Title:         req.Title,
		Description:   req.Description,
		Type:          domain.ElectionType(req.Type),
		Status:        domain.ElectionDraft,
		MaxCandidates: req.MaxCandidates,
		StartDate:     startDate,
		EndDate:       endDate,
		CreatedBy:     userID,
	}

	if err := s.votingService.CreateElection(r.Context(), election); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, election)
}

// handleListElections handles GET /voting/admin/elections
func (s *Server) handleListElections(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)

	elections, err := s.votingService.ListElections(r.Context(), schemeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, elections)
}

// handleGetElection handles GET /voting/admin/elections/{id}
func (s *Server) handleGetElection(w http.ResponseWriter, r *http.Request) {
	electionID := chi.URLParam(r, "id")

	election, err := s.votingService.GetElection(r.Context(), electionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if election == nil {
		respondError(w, http.StatusNotFound, "election not found")
		return
	}

	respondJSON(w, http.StatusOK, election)
}

// handleUpdateElectionStatus handles PUT /voting/admin/elections/{id}/status
func (s *Server) handleUpdateElectionStatus(w http.ResponseWriter, r *http.Request) {
	electionID := chi.URLParam(r, "id")

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.votingService.UpdateElectionStatus(r.Context(), electionID, domain.ElectionStatus(req.Status)); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": req.Status})
}

// handleAddCandidate handles POST /voting/admin/elections/{electionId}/candidates
func (s *Server) handleAddCandidate(w http.ResponseWriter, r *http.Request) {
	electionID := chi.URLParam(r, "electionId")

	var req struct {
		Name           string `json:"name"`
		Position       string `json:"position,omitempty"`
		Manifesto      string `json:"manifesto,omitempty"`
		PhotoURL       string `json:"photo_url,omitempty"`
		PollingStation string `json:"polling_station,omitempty"`
		SchemeType     string `json:"scheme_type,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	candidate := &domain.Candidate{
		ID:             uuid.New().String(),
		ElectionID:     electionID,
		Name:           req.Name,
		Position:       req.Position,
		Manifesto:      req.Manifesto,
		PhotoURL:       req.PhotoURL,
		PollingStation: req.PollingStation,
		SchemeType:     req.SchemeType,
	}

	if err := s.votingService.AddCandidate(r.Context(), candidate); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, candidate)
}

// handleListCandidates handles GET /voting/admin/elections/{electionId}/candidates
func (s *Server) handleListCandidates(w http.ResponseWriter, r *http.Request) {
	electionID := chi.URLParam(r, "electionId")

	candidates, err := s.votingService.ListCandidates(r.Context(), electionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, candidates)
}

// handleAddVoter handles POST /voting/admin/elections/{electionId}/voters
func (s *Server) handleAddVoter(w http.ResponseWriter, r *http.Request) {
	electionID := chi.URLParam(r, "electionId")
	userID := GetUserID(r)

	var req struct {
		MemberID string `json:"member_id"`
		CheckNo  string `json:"check_no,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.votingService.AddVoter(r.Context(), electionID, req.MemberID, req.CheckNo, userID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"status": "voter_added"})
}

// handleBulkAddVoters handles POST /voting/admin/elections/{electionId}/voters/bulk
func (s *Server) handleBulkAddVoters(w http.ResponseWriter, r *http.Request) {
	electionID := chi.URLParam(r, "electionId")
	userID := GetUserID(r)

	var req struct {
		MemberIDs []string `json:"member_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.votingService.BulkAddVoters(r.Context(), electionID, req.MemberIDs, userID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"status": "voters_added",
		"count":  len(req.MemberIDs),
	})
}

// handleGetResults handles GET /voting/admin/elections/{electionId}/results
func (s *Server) handleGetResults(w http.ResponseWriter, r *http.Request) {
	electionID := chi.URLParam(r, "electionId")

	results, err := s.votingService.GetElectionResults(r.Context(), electionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, results)
}

// handleGetResultsByStation handles GET /voting/admin/elections/{electionId}/results/station/{station}
func (s *Server) handleGetResultsByStation(w http.ResponseWriter, r *http.Request) {
	electionID := chi.URLParam(r, "electionId")
	station := chi.URLParam(r, "station")

	results, err := s.votingService.GetResultsByPollingStation(r.Context(), electionID, station)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, results)
}

// handleGetResultsByScheme handles GET /voting/admin/elections/{electionId}/results/scheme/{schemeType}
func (s *Server) handleGetResultsByScheme(w http.ResponseWriter, r *http.Request) {
	electionID := chi.URLParam(r, "electionId")
	schemeType := chi.URLParam(r, "schemeType")

	results, err := s.votingService.GetResultsBySchemeType(r.Context(), electionID, schemeType)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, results)
}

// handleGetVotedMembers handles GET /voting/admin/elections/{electionId}/voted-members
func (s *Server) handleGetVotedMembers(w http.ResponseWriter, r *http.Request) {
	electionID := chi.URLParam(r, "electionId")

	members, err := s.votingService.GetVotedMembersReport(r.Context(), electionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, members)
}

// handleGetNotVotedMembers handles GET /voting/admin/elections/{electionId}/not-voted-members
func (s *Server) handleGetNotVotedMembers(w http.ResponseWriter, r *http.Request) {
	electionID := chi.URLParam(r, "electionId")

	members, err := s.votingService.GetNotVotedMembers(r.Context(), electionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, members)
}

// handleGetVotingStats handles GET /voting/admin/elections/{electionId}/stats
func (s *Server) handleGetVotingStats(w http.ResponseWriter, r *http.Request) {
	electionID := chi.URLParam(r, "electionId")

	stats, err := s.votingService.GetVotingStats(r.Context(), electionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// handleExportResults handles GET /voting/admin/elections/{electionId}/export/results
func (s *Server) handleExportResults(w http.ResponseWriter, r *http.Request) {
	electionID := chi.URLParam(r, "electionId")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}

	results, err := s.votingService.GetElectionResults(r.Context(), electionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=election_results_%s.csv", electionID))

		// Write CSV
		w.Write([]byte("Candidate,Position,Votes,Percentage\n"))
		for _, r := range results.Results {
			w.Write([]byte(fmt.Sprintf("%s,%s,%d,%.2f%%\n", r.CandidateName, r.Position, r.VoteCount, r.VotePercentage)))
		}
	} else {
		respondJSON(w, http.StatusOK, results)
	}
}

// Member Portal Voting Handlers

// handleMemberListElections handles GET /voting/elections
func (s *Server) handleMemberListElections(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)
	schemeID := GetSchemeID(r)

	elections, err := s.votingService.ListElections(r.Context(), schemeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Filter to only show elections the member is registered for
	var eligible []*domain.Election
	for _, e := range elections {
		isEligible, _ := s.votingService.IsEligibleToVote(r.Context(), e.ID, memberID)
		if isEligible {
			eligible = append(eligible, e)
		}
	}

	respondJSON(w, http.StatusOK, eligible)
}

// handleMemberGetElection handles GET /voting/elections/{id}
func (s *Server) handleMemberGetElection(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)
	electionID := chi.URLParam(r, "id")

	election, err := s.votingService.GetElection(r.Context(), electionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if election == nil {
		respondError(w, http.StatusNotFound, "election not found")
		return
	}

	// Check eligibility
	isEligible, _ := s.votingService.IsEligibleToVote(r.Context(), electionID, memberID)
	hasVoted, votedMethod, _ := s.votingService.HasVotedAny(r.Context(), electionID, memberID)
	voteCount, _ := s.votingService.GetVoteCountForMember(r.Context(), electionID, memberID)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"election":     election,
		"is_eligible":  isEligible,
		"has_voted":    hasVoted,
		"voted_method": votedMethod,
		"votes_cast":   voteCount,
		"max_votes":    election.MaxCandidates,
	})
}

// handleMemberListCandidates handles GET /voting/elections/{id}/candidates
func (s *Server) handleMemberListCandidates(w http.ResponseWriter, r *http.Request) {
	electionID := chi.URLParam(r, "id")

	candidates, err := s.votingService.ListCandidates(r.Context(), electionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, candidates)
}

// handleCastVote handles POST /voting/elections/{id}/vote
func (s *Server) handleCastVote(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)
	electionID := chi.URLParam(r, "id")

	var req struct {
		CandidateID  string  `json:"candidate_id"`
		GPSLatitude  float64 `json:"gps_latitude,omitempty"`
		GPSLongitude float64 `json:"gps_longitude,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	vote := &domain.Vote{
		ElectionID:   electionID,
		MemberID:     memberID,
		CandidateID:  req.CandidateID,
		VotingMethod: "web",
		GPSLatitude:  req.GPSLatitude,
		GPSLongitude: req.GPSLongitude,
		IPAddress:    r.RemoteAddr,
		UserAgent:    r.UserAgent(),
	}

	if err := s.votingService.CastVote(r.Context(), vote); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"status": "vote_recorded"})
}

// handleGetMyVotes handles GET /voting/elections/{id}/my-votes
func (s *Server) handleGetMyVotes(w http.ResponseWriter, r *http.Request) {
	memberID := r.Context().Value("member_id").(string)
	electionID := chi.URLParam(r, "id")

	query := `
		SELECT v.id, v.candidate_id, c.name, v.voting_method, v.voted_at,
		       v.gps_latitude, v.gps_longitude
		FROM votes v
		JOIN candidates c ON c.id = v.candidate_id
		WHERE v.election_id = $1 AND v.member_id = $2
		ORDER BY v.voted_at DESC
	`
	rows, err := s.db.QueryContext(r.Context(), query, electionID, memberID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type MyVote struct {
		ID            string    `json:"id"`
		CandidateID   string    `json:"candidate_id"`
		CandidateName string    `json:"candidate_name"`
		VotingMethod  string    `json:"voting_method"`
		VotedAt       time.Time `json:"voted_at"`
		GPSLatitude   float64   `json:"gps_latitude,omitempty"`
		GPSLongitude  float64   `json:"gps_longitude,omitempty"`
	}

	var votes []MyVote
	for rows.Next() {
		var v MyVote
		if err := rows.Scan(&v.ID, &v.CandidateID, &v.CandidateName, &v.VotingMethod,
			&v.VotedAt, &v.GPSLatitude, &v.GPSLongitude); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		votes = append(votes, v)
	}

	respondJSON(w, http.StatusOK, votes)
}

// handleGetLiveResults handles GET /voting/elections/{id}/live-results
func (s *Server) handleGetLiveResults(w http.ResponseWriter, r *http.Request) {
	electionID := chi.URLParam(r, "id")

	results, err := s.votingService.GetElectionResults(r.Context(), electionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	stats, _ := s.votingService.GetVotingStats(r.Context(), electionID)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
		"stats":   stats,
	})
}

// handleUSSDVote handles POST /voting/ussd
func (s *Server) handleUSSDVote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PhoneNumber string `json:"phone_number"`
		ElectionID  string `json:"election_id"`
		CandidateID string `json:"candidate_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Look up member by phone
	var memberID string
	err := s.db.QueryRowContext(r.Context(), `
		SELECT id FROM members WHERE phone = $1
	`, req.PhoneNumber).Scan(&memberID)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusUnauthorized, "phone number not registered")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vote := &domain.Vote{
		ElectionID:   req.ElectionID,
		MemberID:     memberID,
		CandidateID:  req.CandidateID,
		VotingMethod: "ussd",
		MobileNumber: req.PhoneNumber,
	}

	if err := s.votingService.CastVote(r.Context(), vote); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{
		"status":  "vote_recorded",
		"message": "Your vote has been recorded successfully.",
	})
}

// handleURLVote handles GET /voting/url/{token}
func (s *Server) handleURLVote(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "token") // Token validation would go here in production

	// In production, this would validate a signed URL token
	// For now, we'll use a simple query param approach
	electionID := r.URL.Query().Get("election")
	candidateID := r.URL.Query().Get("candidate")
	memberID := r.URL.Query().Get("member")
	gpsLat := r.URL.Query().Get("lat")
	gpsLon := r.URL.Query().Get("lon")

	if electionID == "" || candidateID == "" || memberID == "" {
		respondError(w, http.StatusBadRequest, "election, candidate, and member are required")
		return
	}

	var lat, lon float64
	if gpsLat != "" {
		lat, _ = strconv.ParseFloat(gpsLat, 64)
	}
	if gpsLon != "" {
		lon, _ = strconv.ParseFloat(gpsLon, 64)
	}

	vote := &domain.Vote{
		ElectionID:   electionID,
		MemberID:     memberID,
		CandidateID:  candidateID,
		VotingMethod: "url",
		GPSLatitude:  lat,
		GPSLongitude: lon,
		IPAddress:    r.RemoteAddr,
		UserAgent:    r.UserAgent(),
	}

	if err := s.votingService.CastVote(r.Context(), vote); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Return a simple HTML confirmation
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<html>
		<head><title>Vote Recorded</title></head>
		<body style="font-family: Arial, sans-serif; text-align: center; padding: 50px;">
			<h1 style="color: #28a745;">✓ Vote Recorded Successfully</h1>
			<p>Your vote has been recorded for Election: ` + electionID + `</p>
			<p>Timestamp: ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
		</body>
		</html>
	`))
}
