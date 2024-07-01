package server

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/sync/semaphore"
)

type limiter struct {
	semaphore *semaphore.Weighted
	timeout   time.Duration

	handler http.Handler
}

func newLimiter(poolSize int, timeout time.Duration, handler http.Handler) *limiter {
	sph := semaphore.NewWeighted(int64(poolSize))
	return &limiter{
		sph,
		timeout,
		handler,
	}
}

func (l *limiter) SetTimeout(timeout time.Duration) {
	l.timeout = timeout
}

func (l *limiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancelFunc := context.WithTimeout(r.Context(), l.timeout)
	defer cancelFunc()
	if err := l.semaphore.Acquire(ctx, 1); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	l.handler.ServeHTTP(w, r)

	l.semaphore.Release(1)
}
