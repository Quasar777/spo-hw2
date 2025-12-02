package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/sony/gobreaker"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func forwardResponse(w http.ResponseWriter, resp *http.Response) {
	defer resp.Body.Close()

	for k, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

// общий helper для ошибок circuit breaker’а
func handleCBError(w http.ResponseWriter, err error, serviceName string) {
	if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": serviceName + " service temporarily unavailable",
		})
		return
	}

	writeJSON(w, http.StatusInternalServerError, map[string]string{
		"error": "Internal server error",
	})
}