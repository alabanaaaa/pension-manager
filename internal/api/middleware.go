package api

import (
	"context"
	"net/http"
	"strings"

	"pension-manager/internal/auth"
)

type contextKey string

const (
	ContextUserID    contextKey = "user_id"
	ContextSchemeID  contextKey = "scheme_id"
	ContextUserRole  contextKey = "user_role"
	ContextUserEmail contextKey = "user_email"
)

func AuthMiddleware(authSvc *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondError(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			token := parts[1]
			claims, err := authSvc.VerifyToken(token)
			if err != nil {
				respondError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextSchemeID, claims.SchemeID)
			ctx = context.WithValue(ctx, ContextUserRole, claims.Role)
			ctx = context.WithValue(ctx, ContextUserEmail, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(ContextUserRole).(string)
			if !ok {
				respondError(w, http.StatusUnauthorized, "missing role context")
				return
			}

			roleLevel := map[string]int{
				"member":          1,
				"claims_examiner": 2,
				"pension_officer": 3,
				"auditor":         3,
				"hospital_admin":  3,
				"admin":           4,
				"super_admin":     5,
			}

			userLevel := roleLevel[role]
			if userLevel == 0 {
				respondError(w, http.StatusForbidden, "unknown role")
				return
			}

			for _, allowed := range allowedRoles {
				if role == allowed || userLevel >= roleLevel[allowed] {
					next.ServeHTTP(w, r)
					return
				}
			}

			respondError(w, http.StatusForbidden, "insufficient permissions")
		})
	}
}

func GetSchemeID(r *http.Request) string {
	v, _ := r.Context().Value(ContextSchemeID).(string)
	return v
}

func GetUserID(r *http.Request) string {
	v, _ := r.Context().Value(ContextUserID).(string)
	return v
}

func GetUserRole(r *http.Request) string {
	v, _ := r.Context().Value(ContextUserRole).(string)
	return v
}

func GetUserEmail(r *http.Request) string {
	v, _ := r.Context().Value(ContextUserEmail).(string)
	return v
}
