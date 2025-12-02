package handler

import (
	"encoding/json"
	"net/http"
	"service_orders/internal/model"
	"service_orders/internal/service"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type OrderController struct {
	service service.OrderService
}

func NewOrderController(s service.OrderService) *OrderController {
	return &OrderController{service: s}
}

// GET /orders/status
func (c *OrderController) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "Orders service is running",
	})
}

// GET /orders/health
func (c *OrderController) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "OK",
		"service":   "Orders Service",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// GET /orders/{id}
func (c *OrderController) GetOrder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid id"}`, http.StatusBadRequest)
		return
	}

	order, err := c.service.GetOrder(id)
	if err != nil {
		switch err {
		case model.ErrOrderNotFound:
			http.Error(w, `{"error": "Order not found"}`, http.StatusNotFound)
		default:
			http.Error(w, `{"error": "Server error"}`, http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, order)
}

// GET /orders?userId=...
func (c *OrderController) ListOrders(w http.ResponseWriter, r *http.Request) {
	userIdParam := r.URL.Query().Get("userId")

	var userID *int
	if userIdParam != "" {
		parsed, err := strconv.Atoi(userIdParam)
		if err != nil {
			http.Error(w, `{"error": "invalid userId"}`, http.StatusBadRequest)
			return
		}
		userID = &parsed
	}

	orders, err := c.service.ListOrders(userID)
	if err != nil {
		http.Error(w, `{"error": "Server error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, orders)
}

// POST /orders
func (c *OrderController) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req model.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
		return
	}

	id, err := c.service.CreateOrder(req)
	if err != nil {
		switch err {
		case model.ErrMissingRequiredFields:
			http.Error(w, `{"error": "Missing required fields"}`, http.StatusBadRequest)
		case model.ErrInvalidPrice:
			http.Error(w, `{"error": "Invalid price"}`, http.StatusBadRequest)
		default:
			http.Error(w, `{"error": "Server error"}`, http.StatusInternalServerError)
		}
		return
	}

	response := map[string]interface{}{
		"id": id,
		"message": "Order created succesfully",
	}

	writeJSON(w, http.StatusCreated, response)
}

// PUT /orders/{id}
func (c *OrderController) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if req.ID == 0 {
		http.Error(w, `{"error": "invalid ID"}`, http.StatusBadRequest)
		return
	}

	err := c.service.UpdateOrder(req)
	if err != nil {
		switch err {
		case model.ErrMissingRequiredFields:
			http.Error(w, `{"error": "Missing required fields"}`, http.StatusBadRequest)
		case model.ErrInvalidPrice:
			http.Error(w, `{"error": "Invalid price"}`, http.StatusBadRequest)
		case model.ErrOrderNotFound:
			http.Error(w, `{"error": "Order not found"}`, http.StatusNotFound)
		default:
			http.Error(w, `{"error": "Server error"}`, http.StatusInternalServerError)
		}
		return
	}

	response := map[string]string{
		"message": "Order updated successfully",
	}

	writeJSON(w, http.StatusOK, response)
}

// DELETE /orders/{id}
func (c *OrderController) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "invalid id"}`, http.StatusBadRequest)
		return
	}

	err = c.service.DeleteOrder(id)
	if err != nil {
		switch err {
		case model.ErrOrderNotFound:
			http.Error(w, `{"error": "Order not found"}`, http.StatusNotFound)
		default:
			http.Error(w, `{"error": "Server error"}`, http.StatusInternalServerError)
		}
		return
	}

	response := map[string]string{
		"message": "Order deleted successfully",
	}

	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}