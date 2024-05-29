package server

import (
	"fmt"
	"log"
	"net/http"
)

type HTTPServer struct {
	*http.Server
}

func New(address string) *HTTPServer {
	return &HTTPServer{
		Server: &http.Server{
			Addr: address,
		},
	}
}

func (s *HTTPServer) SetHandler(handler http.Handler) {
	s.Handler = handler
}

func (s *HTTPServer) Start() error {
	log.Printf("listening on %s\n", s.Addr)

	if err := s.ListenAndServe(); err != nil {
		return fmt.Errorf("server shut down: %w", err)
	}

	return nil
}
