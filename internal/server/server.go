package server

import (
	"fmt"
	"log"
	"net/http"
	"slidingo/internal/request"
	"time"
)

type HTTPServer struct {
	*http.Server
}

func New(address string, maxConcurrentRequestCount int, counter request.Counter) *HTTPServer {
	handler := newHandler(counter)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.ServeHTTP)

	limiter := newLimiter(maxConcurrentRequestCount, time.Millisecond*300, mux)

	return &HTTPServer{
		Server: &http.Server{
			Addr: address,
			Handler: limiter,
		},
	}
}

func (s *HTTPServer) Start() error {
	log.Printf("listening on %s\n", s.Addr)

	if err := s.ListenAndServe(); err != nil {
		return fmt.Errorf("server shut down: %w", err)
	}

	return nil
}
