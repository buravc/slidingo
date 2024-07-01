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
	timeoutChan := time.After(l.timeout)
	for {
		select {
		case <-timeoutChan:
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		default:
			currentCounter := l.counter.Load()
			if currentCounter > 0 && l.counter.CompareAndSwap(currentCounter, currentCounter-1) {
				l.handler.ServeHTTP(w, r)
				l.counter.Add(+1)
				return
			}
		}
	}

}
