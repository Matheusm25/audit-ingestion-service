package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/matheusm25/audit-ingestion-service/internal/http/handler"
	audit_http_handler "github.com/matheusm25/audit-ingestion-service/internal/http/handler/audits"
)

type Server struct {
	httpServer *http.Server
	router     *mux.Router
}

func apiKeyMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			providedKey := r.Header.Get("x-api-key")
			if providedKey == "" {
				http.Error(w, "Missing API key", http.StatusUnauthorized)
				return
			}
			if providedKey != apiKey {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func NewServer(port int, apiKey string, healthHandler *handler.HealthHandler, auditHandler *audit_http_handler.AuditHandler) *Server {
	router := mux.NewRouter()

	router.HandleFunc("/health", healthHandler.HealthCheckHandler).Methods("GET")
	router.HandleFunc("/health/services", healthHandler.HealthCheckServicesHandler).Methods("GET")

	auditsRoute := router.HandleFunc("/audits", auditHandler.ListAuditsHandler).Methods("GET")
	auditsRoute.Handler(apiKeyMiddleware(apiKey)(auditsRoute.GetHandler()))

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("0.0.0.0:%d", port),
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		router: router,
	}
}

func (s *Server) Start() error {
	log.Printf("Starting HTTP server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down HTTP server...")
	return s.httpServer.Shutdown(ctx)
}
