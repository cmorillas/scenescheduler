// backend/webserver/http.go
//
// This file contains all HTTP-related logic for the WebServer, including
// route setup and middleware.
//
// Contents:
// - HTTP Router Setup
// - Authentication Middleware

package webserver

import (
	"crypto/subtle"
	"io/fs"
	"net/http"
)

// =============================================================================
// HTTP Router Setup
// =============================================================================

// setupRouter creates and configures the main HTTP request router (ServeMux).
// It registers handlers for API endpoints and the static file server, and wraps
// them with authentication middleware.
func (s *WebServer) setupRouter(staticFiles fs.FS) *http.ServeMux {
	mux := http.NewServeMux()
	auth := s.authMiddleware()

	// Register API endpoints.
	mux.Handle("/ws", auth(http.HandlerFunc(s.wsHandler.HandleConnection)))
	mux.Handle("/whep/", auth(http.HandlerFunc(s.whepHandler.HandleWhepRequest)))

	// Register HLS file server for dynamic preview streams.
	// Serves files from the configured HLS directory (e.g., ./hls).
	hlsHandler := http.StripPrefix("/hls/", http.FileServer(http.Dir(s.config.HlsPath)))
	mux.Handle("/hls/", auth(hlsHandler))

	// Register the static file server for the frontend application.
	// IMPORTANT: This must be registered LAST as it's a catch-all route.
	staticHandler := http.FileServer(http.FS(staticFiles))
	mux.Handle("/", auth(staticHandler))

	return mux
}

// =============================================================================
// Authentication Middleware
// =============================================================================

// authMiddleware creates a middleware that enforces Basic Authentication.
// If user or password is not set in the configuration, authentication is bypassed.
func (s *WebServer) authMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Bypass auth if credentials are not configured.
			if s.config.User == "" || s.config.Password == "" {
				next.ServeHTTP(w, r)
				return
			}

			user, pass, ok := r.BasicAuth()

			// Use constant-time comparison to prevent timing attacks.
			userMatch := subtle.ConstantTimeCompare([]byte(user), []byte(s.config.User)) == 1
			passMatch := subtle.ConstantTimeCompare([]byte(pass), []byte(s.config.Password)) == 1

			if !ok || !userMatch || !passMatch {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted Area"`)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("401 Unauthorized\n"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
