package server

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"time"

	"online_mermaid/internal/auth"
	"online_mermaid/internal/config"
	"online_mermaid/internal/db"
)

type Server struct {
	config Config
	http   *http.Server
}

type Config struct {
	AppConfig config.Config
	Auth      *auth.Handler
	Sessions  *auth.Store
	DB        *db.DB
	Assets    fs.FS
}

func New(cfg Config) *Server {
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/auth/login", cfg.Auth.Login)
	mux.HandleFunc("GET /api/auth/session", cfg.Auth.Session)

	// Protected routes
	mux.Handle("POST /api/auth/logout", cfg.Sessions.Require(http.HandlerFunc(cfg.Auth.Logout)))

	// Diagram management API
	mux.Handle("GET /api/diagrams", cfg.Sessions.Require(listDiagramsHandler(cfg.DB)))
	mux.Handle("GET /api/diagrams/next-name", cfg.Sessions.Require(getNextNameHandler(cfg.DB)))
	mux.Handle("POST /api/diagrams", cfg.Sessions.Require(createDiagramHandler(cfg.DB)))
	mux.Handle("GET /api/diagrams/{id}", cfg.Sessions.Require(getDiagramHandler(cfg.DB)))
	mux.Handle("PUT /api/diagrams/{id}", cfg.Sessions.Require(updateDiagramHandler(cfg.DB)))
	mux.Handle("DELETE /api/diagrams/{id}", cfg.Sessions.Require(deleteDiagramHandler(cfg.DB)))
	mux.Handle("POST /api/diagrams/{id}/duplicate", cfg.Sessions.Require(duplicateDiagramHandler(cfg.DB)))

	// Fallback for unmatched API routes
	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, "not_found", "API route not found")
	})

	// Frontend SPA handler
	mux.Handle("/", spaHandler(cfg.Assets))

	handler := securityHeaders(recoverPanic(requestLogger(mux)))
	if cfg.AppConfig.SecureCookie {
		handler = securityHeadersWithHSTS(recoverPanic(requestLogger(mux)))
	}

	return &Server{
		config: cfg,
		http: &http.Server{
			Addr:              cfg.AppConfig.Addr,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       90 * time.Second,
			MaxHeaderBytes:    1 << 20,
		},
	}
}

func (s *Server) HTTPServer() *http.Server { return s.http }
func (s *Server) Handler() http.Handler    { return s.http.Handler }

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"code": code, "message": message})
}
