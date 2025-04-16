package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

// Server handles configuration, startup, and graceful shutdown of an HTTP server.
type Server struct {
	srv    http.Server
	logger *slog.Logger
}

// New accepts the dependencies required to create a [Server], and returns one.
func New(host, port string, h http.Handler, logger *slog.Logger) *Server {
	return &Server{
		srv: http.Server{
			Addr:    net.JoinHostPort(host, port),
			Handler: h,
		},
		logger: logger,
	}
}

// ListenAndServe starts an [http.Server] that listens on the configured host and port. When ctx is cancelled, the server shuts down within 10 seconds and the function returns.
func (s *Server) ListenAndServe(ctx context.Context) {
	go func() {
		s.logger.Info(fmt.Sprintf("Serving on %v", s.srv.Addr))
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.Error("Error while shutting down server", "err", err)
		}
	}()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Error while shutting down server", "err", err)
		}
	}()
	// Wait for the shutdown goroutine to terminate before returning.
	wg.Wait()
}
