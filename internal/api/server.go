package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"pension-manager/internal/auth"
	"pension-manager/internal/bulk"
	"pension-manager/internal/config"
	"pension-manager/internal/db"
	"pension-manager/internal/documents"
	"pension-manager/internal/hospital"
	"pension-manager/internal/mpesa"
	"pension-manager/internal/news"
	"pension-manager/internal/portal"
	"pension-manager/internal/reports"
	"pension-manager/internal/security"
	"pension-manager/internal/sms"
	"pension-manager/internal/sponsor"
	"pension-manager/internal/tax"
	"pension-manager/internal/ussd"
	"pension-manager/internal/voting"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	router          *chi.Mux
	db              *db.DB
	auth            *auth.Service
	cfg             *config.Config
	mpesaClient     *mpesa.Client
	hospitalService *hospital.HospitalService
	docService      *documents.Service
	sponsorService  *sponsor.Service
	portalService   *portal.Service
	votingService   *voting.Service
	bulkService     *bulk.Service
	reportsService  *reports.Service
	smsService      *sms.Service
	taxReminderSvc  *tax.Service
	ipBlacklistSvc  *security.Service
	newsService     *news.Service
	ussdService     *ussd.Service
}

func New(database *db.DB, cfg *config.Config) *Server {
	authSvc := auth.NewService(cfg.JWTSecret)

	var mpesaClient *mpesa.Client
	if cfg.Mpesa.ConsumerKey != "" && cfg.Mpesa.ConsumerSecret != "" {
		mpesaClient = mpesa.NewClient(mpesa.Config{
			ConsumerKey:    cfg.Mpesa.ConsumerKey,
			ConsumerSecret: cfg.Mpesa.ConsumerSecret,
			ShortCode:      cfg.Mpesa.ShortCode,
			Passkey:        cfg.Mpesa.Passkey,
			Environment:    cfg.Env,
			CallbackURL:    cfg.Mpesa.CallbackURL,
			APIVersion:     cfg.Mpesa.APIVersion,
		})
		slog.Info("M-Pesa client initialized", "environment", cfg.Env)
	} else {
		slog.Warn("M-Pesa not configured — STK Push will be simulated")
	}

	hospitalService := hospital.NewHospitalService(database)
	docService := documents.NewService(database, documents.NewLocalStorage("/tmp/pension-docs"))
	sponsorService := sponsor.NewService(database)
	portalService := portal.NewService(database)
	votingService := voting.NewService(database)
	bulkService := bulk.NewService(database)
	reportsService := reports.NewService(database)
	smsService := sms.NewService(sms.NewMockProvider())
	taxReminderSvc := tax.NewReminderService(database)
	ipBlacklistSvc := security.NewIPBlacklistService(database)
	newsService := func() *news.Service {
		if cfg.NewsAPI.APIKey != "" {
			slog.Info("NewsAPI configured — fetching live Kenya government news")
			return news.NewService(news.NewNewsAPIProvider(cfg.NewsAPI.APIKey), 15*time.Minute)
		}
		slog.Warn("NewsAPI not configured — using mock provider with sample Kenya news")
		return news.NewService(news.NewMockProvider(), 15*time.Minute)
	}()

	// USSD Service (Africa's Talking)
	ussdService := func() *ussd.Service {
		ussdProvider := ussd.NewAfricaTalkingProvider(
			"atsk_12644551b2bfab30e7ea666ec36e7c6e564df2ee750320e5402e59574a758dd1fb37d214",
			"*384*28346#",
			"sandbox",
		)
		ussdVotingAdapter := ussd.NewVotingServiceAdapter(database)
		return ussd.NewService(ussdProvider, ussdVotingAdapter)
	}()

	s := &Server{
		router:          chi.NewRouter(),
		db:              database,
		auth:            authSvc,
		cfg:             cfg,
		mpesaClient:     mpesaClient,
		hospitalService: hospitalService,
		docService:      docService,
		sponsorService:  sponsorService,
		portalService:   portalService,
		votingService:   votingService,
		bulkService:     bulkService,
		reportsService:  reportsService,
		smsService:      smsService,
		taxReminderSvc:  taxReminderSvc,
		ipBlacklistSvc:  ipBlacklistSvc,
		newsService:     newsService,
		ussdService:     ussdService,
	}

	s.mountMiddleware()
	s.mountRoutes()

	return s
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) mountMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(30 * time.Second))
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	s.router.Use(RequestLogger)
}

func (s *Server) mountRoutes() {
	s.router.Get("/health", s.healthCheck)
	s.router.Get("/ready", s.readinessCheck)

	s.router.Post("/api/auth/login", s.login)
	s.router.Post("/api/auth/refresh", s.refreshToken)
	s.router.Get("/api/auth/otp", s.requestOTP)
	s.router.Post("/api/auth/unlock/{id}", s.unlockUser)

	// M-Pesa callback (no auth required - called by Safaricom)
	s.router.Post("/api/mpesa/callback", s.mpesaCallback)

	// USSD callback (no auth required - called by Africa's Talking)
	s.router.Post("/api/ussd/voting", s.handleUSSDVoting)

	s.router.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(s.auth))

		r.Get("/api/dashboard", s.dashboard)

		// Phase 1: Members
		r.Get("/api/members", s.listMembers)
		r.Post("/api/members", s.createMember)
		r.Get("/api/members/{id}", s.getMember)
		r.Put("/api/members/{id}", s.updateMember)
		r.Delete("/api/members/{id}", s.deactivateMember)

		// Phase 1: Beneficiaries
		r.Get("/api/members/{id}/beneficiaries", s.listBeneficiaries)
		r.Post("/api/members/{id}/beneficiaries", s.addBeneficiary)

		// Phase 1: Contributions
		r.Post("/api/contributions", s.recordContribution)
		r.Get("/api/contributions", s.listContributions)
		r.Get("/api/contributions/{member_id}", s.memberContributions)
		r.Post("/api/contributions/mpesa", s.mpesaContribution)
		r.Post("/api/contributions/reconcile", s.reconcileContributions)

		// Phase 1: Reports
		r.Get("/api/reports/quarterly", s.quarterlyReport)
		r.Get("/api/reports/contributions", s.contributionReport)
		r.Get("/api/reports/export", s.exportCSV)

		// Phase 1: Ghost Mode (fraud detection)
		r.Get("/api/ghost", s.ghostReport)

		// Admin-only routes
		r.Group(func(admin chi.Router) {
			admin.Use(RoleMiddleware("super_admin", "admin"))
			r.Get("/api/admin/users", s.listUsers)
			r.Post("/api/admin/users", s.createUser)
			r.Put("/api/admin/users/{id}/role", s.updateUserRole)
			r.Delete("/api/admin/users/{id}", s.disableUser)
		})

		// Hospital Management Routes
		s.registerHospitalRoutes(r)

		// Claims Management Routes
		s.registerClaimsRoutes(r)

		// Maker-Checker Workflow Routes
		s.registerMakerCheckerRoutes(r)

		// Document Management Routes
		s.registerDocumentRoutes(r)

		// Sponsor Management Routes
		s.registerSponsorRoutes(r)

		// Tax Computation Routes
		s.registerTaxRoutes(r)

		// Member Portal Routes
		s.registerPortalRoutes(r)

		// Online Voting Routes
		s.registerVotingRoutes(r)

		// Bulk Processing Routes
		s.registerBulkRoutes(r)

		// Contribution Report Routes
		s.registerReportRoutes(r)

		// SMS Gateway Routes
		s.registerSMSRoutes(r)

		// Tax Exemption Reminder Routes
		s.registerTaxReminderRoutes(r)

		// IP Blacklist Routes
		s.registerIPBlacklistRoutes(r)

		// Kenya Government News Routes
		s.registerNewsRoutes(r)
	})
}

func (s *Server) Start(addr string) error {
	server := &http.Server{
		Addr:              addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":" + fmt.Sprint(s.cfg.HTTPPort),
		Handler: s.Handler(),
	}
	return server.Shutdown(ctx)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func respondCreated(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		slog.Info("http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration", time.Since(start).Milliseconds(),
			"remote", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}
