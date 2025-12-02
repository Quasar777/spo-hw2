package handler

import (
	"net/http"

	"github.com/sony/gobreaker"
)

type HealthHandler struct {
	usersCB  *gobreaker.CircuitBreaker
	ordersCB *gobreaker.CircuitBreaker
}

func NewHealthHandler(
	usersCB *gobreaker.CircuitBreaker,
	ordersCB *gobreaker.CircuitBreaker,
) *HealthHandler {
	return &HealthHandler{
		usersCB:  usersCB,
		ordersCB: ordersCB,
	}
}

// GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "API Gateway is running",
		"circuits": map[string]any{
			"users": map[string]any{
				"state": h.usersCB.State().String(),
				"stats": h.usersCB.Counts(),
			},
			"orders": map[string]any{
				"state": h.ordersCB.State().String(),
				"stats": h.ordersCB.Counts(),
			},
		},
	})
}

// GET /status
func (h *HealthHandler) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "API Gateway is running",
	})
}