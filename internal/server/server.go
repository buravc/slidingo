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

	limiter *limiter
}

func New(address string, maxConcurrentRequestCount int, timeout time.Duration, counter request.Counter) *HTTPServer {
	handler := newHandler(counter)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.ServeHTTP)

	limiter := newLimiter(maxConcurrentRequestCount, timeout, mux)

	return &HTTPServer{
		Server: &http.Server{
			Addr: address,
			Handler: limiter,
		},

		limiter: limiter,
	}
}

func (s *HTTPServer) SetTimeout(timeout time.Duration) {
	s.limiter.SetTimeout(timeout)
}

func (s *HTTPServer) Start() error {
	log.Printf("listening on %s\n", s.Addr)

	if err := s.ListenAndServe(); err != nil {
		return fmt.Errorf("server shut down: %w", err)
	}

	return nil
}
