package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// registerNewsRoutes registers news API routes
func (s *Server) registerNewsRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(s.auth))

		r.Route("/news", func(r chi.Router) {
			r.Get("/", s.handleGetNews)
			r.Get("/categories", s.handleGetNewsCategories)
			r.Get("/refresh", s.handleRefreshNews)
		})
	})

	// Public news endpoint (no auth required)
	r.Get("/api/news/public", s.handleGetPublicNews)
}

// handleGetNews handles GET /news
func (s *Server) handleGetNews(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	pageSize := 10

	news, err := s.newsService.GetKenyaNews(r.Context(), category, pageSize)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, news)
}

// handleGetNewsCategories handles GET /news/categories
func (s *Server) handleGetNewsCategories(w http.ResponseWriter, r *http.Request) {
	categories := []map[string]string{
		{"id": "general", "name": "General", "description": "All government news"},
		{"id": "business", "name": "Business & Economy", "description": "Economic policy, treasury, CBK"},
		{"id": "politics", "name": "Politics & Legislation", "description": "Parliament, bills, regulations"},
		{"id": "health", "name": "Health", "description": "Healthcare policy, NHIF/SHA"},
		{"id": "technology", "name": "Technology", "description": "Digital government, e-services"},
	}

	respondJSON(w, http.StatusOK, categories)
}

// handleRefreshNews handles GET /news/refresh
func (s *Server) handleRefreshNews(w http.ResponseWriter, r *http.Request) {
	s.newsService.ClearCache()

	news, err := s.newsService.GetKenyaNews(r.Context(), "", 10)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "refreshed",
		"articles":   news.TotalResults,
		"fetched_at": news.FetchedAt,
		"provider":   s.newsService.ProviderName(),
	})
}

// handleGetPublicNews handles GET /api/news/public
func (s *Server) handleGetPublicNews(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	pageSize := 5

	news, err := s.newsService.GetKenyaNews(r.Context(), category, pageSize)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, news)
}
