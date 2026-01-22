package session

import (
	"context"
	"net/http"
	"strings"
)

/* SessionMiddleware provides session-based authentication middleware */
func (m *Manager) SessionMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			/* Skip auth for public endpoints */
			publicPaths := []string{
				"/health",
				"/api/v1/health",
				"/api/v1/auth/register",
				"/api/v1/auth/login",
				"/api/v1/auth/oidc/",
				"/api/v1/database/test",
			}

			for _, path := range publicPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			/* Try to get access token from cookie */
			accessToken := m.GetAccessTokenFromRequest(r)
			if accessToken == "" {
				/* Try refresh token */
				refreshToken := m.GetRefreshTokenFromRequest(r)
				if refreshToken == "" {
					/* No session cookie, check for API key in header (backward compatibility) */
					authHeader := r.Header.Get("Authorization")
					if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
						/* Allow API key auth to pass through */
						next.ServeHTTP(w, r)
						return
					}
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				/* Attempt refresh */
				session, newAccessToken, newRefreshToken, err := m.RefreshSession(r.Context(), refreshToken)
				if err != nil {
					/* Refresh failed, check for API key */
					authHeader := r.Header.Get("Authorization")
					if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
						next.ServeHTTP(w, r)
						return
					}
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				/* Set new cookies */
				m.SetCookies(w, newAccessToken, newRefreshToken)
				accessToken = newAccessToken

				/* Add session to context */
				ctx := r.Context()
				ctx = context.WithValue(ctx, sessionKey, session)
				ctx = context.WithValue(ctx, userIDKey, session.UserID)
				ctx = context.WithValue(ctx, databaseKey, session.Database)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			/* Validate session */
			session, err := m.ValidateSession(r.Context(), accessToken)
			if err != nil {
				/* Session invalid, check for API key */
				authHeader := r.Header.Get("Authorization")
				if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
					next.ServeHTTP(w, r)
					return
				}
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			/* Add session to context */
			ctx := r.Context()
			ctx = context.WithValue(ctx, sessionKey, session)
			ctx = context.WithValue(ctx, userIDKey, session.UserID)
			ctx = context.WithValue(ctx, databaseKey, session.Database)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
