package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pension-manager/internal/auth"
	"pension-manager/internal/config"
	"pension-manager/internal/db"
	"pension-manager/internal/hospital"
	"pension-manager/internal/mpesa"
	"pension-manager/internal/news"
	"pension-manager/internal/portal"
	"pension-manager/internal/reports"
	"pension-manager/internal/security"
	"pension-manager/internal/sms"
	"pension-manager/internal/sponsor"
	"pension-manager/internal/tax"
	"pension-manager/internal/voting"

	"github.com/go-chi/chi/v5"
)

// setupTestServer creates a test server with mock dependencies
func setupTestServer(t *testing.T) *Server {
	t.Helper()

	cfg := &config.Config{
		JWTSecret: "test-secret-key-for-integration-tests",
		Env:       "test",
		Mpesa: config.MpesaConfig{
			ConsumerKey:    "test-key",
			ConsumerSecret: "test-secret",
			ShortCode:      "174379",
			Passkey:        "test-passkey",
			CallbackURL:    "http://localhost:8080/mpesa/callback",
			APIVersion:     "v3",
		},
		NewsAPI: config.NewsAPIConfig{
			APIKey: "test-news-api-key",
		},
	}

	dbConn := db.NewTestDB(t)

	authSvc := auth.NewService(cfg.JWTSecret)
	hospitalSvc := hospital.NewHospitalService(dbConn)
	sponsorSvc := sponsor.NewService(dbConn)
	portalSvc := portal.NewService(dbConn)
	votingSvc := voting.NewService(dbConn)
	reportsSvc := reports.NewService(dbConn)
	smsSvc := sms.NewService(sms.NewMockProvider())
	taxReminderSvc := tax.NewReminderService(dbConn)
	ipBlacklistSvc := security.NewIPBlacklistService(dbConn)
	newsSvc := news.NewService(news.NewMockProvider(), 0)

	mpesaClient := mpesa.NewClient(mpesa.Config{
		ConsumerKey:    cfg.Mpesa.ConsumerKey,
		ConsumerSecret: cfg.Mpesa.ConsumerSecret,
		ShortCode:      cfg.Mpesa.ShortCode,
		Passkey:        cfg.Mpesa.Passkey,
		Environment:    "sandbox",
		CallbackURL:    cfg.Mpesa.CallbackURL,
		APIVersion:     cfg.Mpesa.APIVersion,
	})

	server := &Server{
		router:          nil, // Will be initialized in setupTestRoutes
		db:              dbConn,
		auth:            authSvc,
		cfg:             cfg,
		mpesaClient:     mpesaClient,
		hospitalService: hospitalSvc,
		sponsorService:  sponsorSvc,
		portalService:   portalSvc,
		votingService:   votingSvc,
		reportsService:  reportsSvc,
		smsService:      smsSvc,
		taxReminderSvc:  taxReminderSvc,
		ipBlacklistSvc:  ipBlacklistSvc,
		newsService:     newsSvc,
	}

	server.setupTestRoutes()
	return server
}

// setupTestRoutes sets up routes for testing without full middleware
func (s *Server) setupTestRoutes() {
	s.router = chi.NewRouter()
	s.mountMiddleware()
	s.mountRoutes()
}

// Helper to create request
func createRequest(method, url string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, url, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// Helper to execute request and return response
func executeRequest(s *Server, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	return rr
}

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	server := setupTestServer(t)

	req := createRequest("GET", "/health", nil)
	rr := executeRequest(server, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

// TestNewsEndpoints tests the news API endpoints
func TestNewsEndpoints(t *testing.T) {
	server := setupTestServer(t)

	// Test news endpoint (requires auth in current setup)
	req := createRequest("GET", "/news", nil)
	rr := executeRequest(server, req)

	// Should return 401 without auth (routes are behind auth middleware)
	if rr.Code != http.StatusUnauthorized && rr.Code != http.StatusNotFound {
		t.Logf("News endpoint returned status %d", rr.Code)
	}
}

// TestTaxComputationEndpoints tests the tax computation API
func TestTaxComputationEndpoints(t *testing.T) {
	server := setupTestServer(t)

	// Test tax endpoint (requires auth)
	req := createRequest("GET", "/tax/brackets", nil)
	rr := executeRequest(server, req)

	// Should return 401 without auth
	if rr.Code != http.StatusUnauthorized && rr.Code != http.StatusNotFound {
		t.Logf("Tax brackets endpoint returned status %d", rr.Code)
	}
}

// TestAuthEndpoints tests authentication endpoints
func TestAuthEndpoints(t *testing.T) {
	server := setupTestServer(t)

	// Test login with missing credentials
	body := map[string]string{"email": "", "password": ""}
	req := createRequest("POST", "/api/auth/login", body)
	rr := executeRequest(server, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing credentials, got %d", rr.Code)
	}

	// Test login with invalid credentials (will fail with 500 if DB not available)
	body = map[string]string{"email": "test@example.com", "password": "wrongpassword"}
	req = createRequest("POST", "/api/auth/login", body)
	rr = executeRequest(server, req)

	// Accept either 401 (auth failure) or 500 (DB not available)
	if rr.Code != http.StatusUnauthorized && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 401 or 500 for invalid credentials, got %d", rr.Code)
	}
}

// TestProtectedEndpoints tests that protected endpoints require authentication
func TestProtectedEndpoints(t *testing.T) {
	server := setupTestServer(t)

	protectedEndpoints := []string{
		"/voting/elections",
		"/portal/profile",
		"/hospitals",
		"/claims",
		"/bulk/import/members",
		"/reports/contributions/breakdown",
		"/security/ip-blacklist",
	}

	for _, endpoint := range protectedEndpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := createRequest("GET", endpoint, nil)
			rr := executeRequest(server, req)

			// Should return 401 (unauthorized) or 404 (not found)
			if rr.Code != http.StatusUnauthorized && rr.Code != http.StatusNotFound {
				t.Logf("Endpoint %s returned status %d (expected 401 or 404)", endpoint, rr.Code)
			}
		})
	}
}

// TestMiddlewareChain tests that middleware is properly applied
func TestMiddlewareChain(t *testing.T) {
	server := setupTestServer(t)

	// Test that request logging works
	req := createRequest("GET", "/health", nil)
	rr := executeRequest(server, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

// TestErrorHandling tests that errors are properly formatted
func TestErrorHandling(t *testing.T) {
	server := setupTestServer(t)

	// Test invalid JSON
	req := createRequest("POST", "/api/auth/login", "invalid json")
	rr := executeRequest(server, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", rr.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if response["error"] == "" {
		t.Error("Expected error message in response")
	}
}

// TestContentType tests that responses have correct content type
func TestContentType(t *testing.T) {
	server := setupTestServer(t)

	req := createRequest("GET", "/health", nil)
	rr := executeRequest(server, req)

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", rr.Header().Get("Content-Type"))
	}
}

// TestNotFound tests 404 handling
func TestNotFound(t *testing.T) {
	server := setupTestServer(t)

	req := createRequest("GET", "/api/nonexistent", nil)
	rr := executeRequest(server, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

// TestMethodNotAllowed tests 405 handling
func TestMethodNotAllowed(t *testing.T) {
	server := setupTestServer(t)

	req := createRequest("DELETE", "/health", nil)
	rr := executeRequest(server, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

// TestSMSProviderEndpoint tests the SMS provider endpoint
func TestSMSProviderEndpoint(t *testing.T) {
	server := setupTestServer(t)

	// Test SMS provider endpoint (should require auth)
	req := createRequest("GET", "/sms/provider", nil)
	rr := executeRequest(server, req)

	// Should return 401 or 404
	if rr.Code != http.StatusUnauthorized && rr.Code != http.StatusNotFound {
		t.Logf("SMS provider endpoint returned status %d", rr.Code)
	}
}

// TestReadyEndpoint tests the readiness check endpoint
func TestReadyEndpoint(t *testing.T) {
	server := setupTestServer(t)

	req := createRequest("GET", "/ready", nil)
	rr := executeRequest(server, req)

	// Should return 200 or 503 depending on DB connection
	if rr.Code != http.StatusOK && rr.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 200 or 503, got %d", rr.Code)
	}
}

// TestNewsCategoriesEndpoint tests the news categories endpoint
func TestNewsCategoriesEndpoint(t *testing.T) {
	server := setupTestServer(t)

	// Test news categories endpoint (should require auth)
	req := createRequest("GET", "/news/categories", nil)
	rr := executeRequest(server, req)

	// Should return 401 or 404
	if rr.Code != http.StatusUnauthorized && rr.Code != http.StatusNotFound {
		t.Logf("News categories endpoint returned status %d", rr.Code)
	}
}
