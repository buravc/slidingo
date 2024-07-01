package server

import (
	"net/http"
	"sync/atomic"
	"time"
)

type limiter struct {
	counter *atomic.Int32
	timeout time.Duration

	handler http.Handler
}

func newLimiter(poolSize int, timeout time.Duration, handler http.Handler) *limiter {
	var counter atomic.Int32
	counter.Store(int32(poolSize))
	return &limiter{
		&counter,
		timeout,
		handler,
	}
}

func (l *limiter) SetTimeout(timeout time.Duration) {
	l.timeout = timeout
}

func (l *limiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	currentCounter := l.counter.Load()
	if currentCounter > 0 && l.counter.CompareAndSwap(currentCounter, currentCounter-1) {
		l.handler.ServeHTTP(w, r)
		l.counter.Add(+1)
		return
	}

	w.WriteHeader(http.StatusServiceUnavailable)
}
