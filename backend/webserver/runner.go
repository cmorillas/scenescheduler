// backend/webserver/runner.go
//
// This file defines the lifecycle orchestration for the WebServer module.
// It is responsible for starting, supervising, and gracefully stopping the server.
//
// Contents:
// - Public Lifecycle Methods (Run, Stop)
// - Internal Lifecycle Helpers

package webserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"scenescheduler/backend/eventbus"
)

// =============================================================================
// Public Lifecycle Methods
// =============================================================================

// Run starts the web server and its sub-components. It blocks until the
// server stops due to an error or context cancellation.
// The context for this module already exists from the constructor.
func (s *WebServer) Run() {
	defer s.cleanup()

	s.logger.InfoGui("WebServer starting", "port", s.config.Port, "tls", s.config.EnableTLS)

	// WHEP handler is purely reactive (HTTP-based), no need to Run() it
	// It will be shut down in cleanup()

	// Publish the started event so other components know the server is active.
	eventbus.Publish(s.bus, eventbus.WebServerStarted{
		Port:      s.config.Port,
		UseTLS:    s.config.EnableTLS,
		IPs:       getLocalIPs(),
		Timestamp: time.Now(),
	})

	// Start a goroutine to listen for context cancellation and trigger shutdown.
	go func() {
		<-s.ctx.Done()
		s.logger.InfoGui("Context cancelled, initiating web server shutdown.")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("HTTP server graceful shutdown failed", "error", err)
		}
	}()

	// Start the main HTTP server loop. This is a blocking call.
	var err error
	if s.config.EnableTLS {
		err = s.httpServer.ListenAndServeTLS(s.config.CertFilePath, s.config.KeyFilePath)
	} else {
		err = s.httpServer.ListenAndServe()
	}

	// If ListenAndServe exits and it's not because the server was closed, it's an error.
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("WebServer stopped with an error", "error", err)
	} else {
		s.logger.InfoGui("WebServer stopped gracefully.")
	}
}

// Stop requests the web server to shut down gracefully.
// This method is idempotent and safe to call multiple times.
func (s *WebServer) Stop() {
	s.stopOnce.Do(func() {
		s.logger.InfoGui("Stop requested for WebServer.")
		if s.cancel != nil {
			s.cancel()
		}
	})
}

// =============================================================================
// Internal Lifecycle Helpers
// =============================================================================

// cleanup orchestrates the graceful shutdown of all module resources.
// It is idempotent thanks to sync.Once.
func (s *WebServer) cleanup() {
	s.cleanupOnce.Do(func() {
		s.logger.Debug("Cleaning up WebServer resources.")

		// Close all active WebSocket connections.
		if s.wsHandler != nil {
			s.wsHandler.CloseAll()
		}

		// Shutdown WHEP handler and close all WebRTC sessions.
		if s.whepHandler != nil {
			s.whepHandler.Shutdown()
		}

		// Shutdown preview manager and kill all active preview processes.
		if s.previewManager != nil {
			s.logger.Debug("Shutting down preview manager")
			if err := s.previewManager.Shutdown(); err != nil {
				s.logger.Error("Error shutting down preview manager", "error", err)
			}
		}

		// Unsubscribe from all event bus topics to prevent memory leaks.
		s.unsubscribeAllEvents()

		// Publish the final stopped event.
		eventbus.Publish(s.bus, eventbus.WebServerStopped{
			Reason:    eventbus.ReasonGracefulShutdown,
			Timestamp: time.Now(),
		})
	})
}