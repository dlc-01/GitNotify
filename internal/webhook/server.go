package webhook

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/dlc-01/GitNotify/internal/config"
)

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func NewServerConfig(cfg *config.WebhookConfig) ServerConfig {
	port := cfg.Port
	if port == 0 {
		port = 8080
	}
	host := cfg.Host
	return ServerConfig{
		Host:         host,
		Port:         port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

type Server struct {
	srv     *http.Server
	handler *Handler
	log     *slog.Logger
}

func NewServer(cfg ServerConfig, handler *Handler, log *slog.Logger) *Server {
	mux := http.NewServeMux()
	mux.Handle("/webhook", handler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return &Server{
		srv: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler:      mux,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		handler: handler,
		log:     log,
	}
}

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	s.log.Info("webhook server started",
		slog.String("addr", s.srv.Addr),
	)

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	case <-ctx.Done():
		s.log.Info("shutting down webhook server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}

		s.log.Info("webhook server stopped")
		return nil
	}
}
