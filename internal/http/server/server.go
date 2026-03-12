package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/matheusm25/audit-ingestion-service/internal/http/handler"
)

type Server struct {
	httpServer *http.Server
	router     *mux.Router
}

func NewServer(port int) *Server {
	router := mux.NewRouter()

	// Register health check endpoint
	router.HandleFunc("/health", handler.HealthCheckHandler).Methods("GET")

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("localhost:%d", port),
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
