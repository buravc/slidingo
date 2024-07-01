package server

import (
	"net/http"
	"time"
)

type limiter struct {
	pool    chan struct{}
	timeout time.Duration

	handler http.Handler
}

func newLimiter(poolSize int, timeout time.Duration, handler http.Handler) *limiter {
	pool := make(chan struct{}, poolSize)
	return &limiter{
		pool,
		timeout,
		handler,
	}
}

func (l *limiter) SetTimeout(timeout time.Duration) {
	l.timeout = timeout
}

func (l *limiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	case l.pool <- struct{}{}:
		break
	case <-time.After(l.timeout):
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	l.handler.ServeHTTP(w, r)

	<-l.pool
}
