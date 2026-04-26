package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// registerNewsRoutes registers news API routes
func (s *Server) registerNewsRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(s.auth))

		r.Route("/api/news", func(r chi.Router) {
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
	// No cache - always fetch fresh
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	category := r.URL.Query().Get("category")
	pageSize := 10

	// Log which provider is being used
	slog.Info("Fetching news", "category", category, "provider", s.newsService.ProviderName())

	news, err := s.newsService.GetKenyaNews(r.Context(), category, pageSize)
	if err != nil {
		slog.Error("Failed to fetch news", "error", err, "provider", s.newsService.ProviderName())
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	slog.Info("News fetched successfully", "count", len(news.Articles), "provider", s.newsService.ProviderName())

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
	slog.Info("Cache cleared, fetching fresh news", "provider", s.newsService.ProviderName())

	news, err := s.newsService.GetKenyaNews(r.Context(), "", 10)
	if err != nil {
		slog.Error("Failed to refresh news", "error", err, "provider", s.newsService.ProviderName())
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	slog.Info("News refreshed", "articles", news.TotalResults, "provider", s.newsService.ProviderName())

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "refreshed",
		"articles":   news.TotalResults,
		"fetched_at": news.FetchedAt,
		"provider":   s.newsService.ProviderName(),
	})
}

// handleGetPublicNews handles GET /api/news/public
func (s *Server) handleGetPublicNews(w http.ResponseWriter, r *http.Request) {
	// No cache headers - always fetch fresh
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	category := r.URL.Query().Get("category")
	pageSize := 10

	slog.Info("Fetching public news", "category", category, "provider", s.newsService.ProviderName())

	news, err := s.newsService.GetKenyaNews(r.Context(), category, pageSize)
	if err != nil {
		slog.Error("Failed to fetch public news", "error", err, "provider", s.newsService.ProviderName())
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	slog.Info("Public news fetched", "count", len(news.Articles), "provider", s.newsService.ProviderName())

	respondJSON(w, http.StatusOK, news)
}
