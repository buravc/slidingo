package server

import (
	"encoding/json"
	"net/http"
	"slidingo/internal/request"
	"time"
)

type handler struct {
	requestCounter request.Counter
}

func newHandler(requestCounter request.Counter) http.Handler {
	return &handler{
		requestCounter,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2*time.Second)
	count := h.requestCounter.Count()
	resp, err := json.Marshal(count)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
