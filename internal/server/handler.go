package server

import (
	"encoding/json"
	"net/http"
	"slidingo/internal/request"
)

type handler struct {
	requestCounter request.Counter
}

func NewHandler(requestCounter request.Counter) http.Handler {
	return &handler{
		requestCounter,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	count := h.requestCounter.Count()
	resp, err := json.Marshal(count)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
