package api

import (
	"net/http"
)

// handleUSSDVoting handles POST /api/ussd/voting
func (s *Server) handleUSSDVoting(w http.ResponseWriter, r *http.Request) {
	if s.ussdService == nil {
		http.Error(w, "USSD service not configured", http.StatusServiceUnavailable)
		return
	}

	s.ussdService.HandleUSSD(w, r)
}
