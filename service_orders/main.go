package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// В отличие от users, здесь в Node у тебя свободная структура заказа:
// newOrder = { id, ...orderData }
// Чтобы максимально повторить поведение, будем хранить map[string]any.
var (
	mu         sync.RWMutex
	fakeOrders = map[int]map[string]any{}
	currentID  = 1
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/orders/status", statusHandler)
	r.Get("/orders/health", healthHandler)
	r.Get("/orders/{orderId}", getOrderHandler)
	r.Get("/orders", listOrdersHandler)
	r.Post("/orders", createOrderHandler)
	r.Put("/orders/{orderId}", updateOrderHandler)
	r.Delete("/orders/{orderId}", deleteOrderHandler)

	port := ":8000"
	log.Printf("Orders service running on port %s\n", port)
	if err := http.ListenAndServe("0.0.0.0"+port, r); err != nil {
		log.Fatal(err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "Orders service is running",
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "OK",
		"service":   "Orders Service",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func getOrderHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "orderId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order id"})
		return
	}

	mu.RLock()
	order, ok := fakeOrders[id]
	mu.RUnlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Order not found"})
		return
	}

	writeJSON(w, http.StatusOK, order)
}

func listOrdersHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	orders := make([]map[string]any, 0, len(fakeOrders))
	for _, o := range fakeOrders {
		orders = append(orders, o)
	}
	mu.RUnlock()

	// В Node-версии у тебя есть фильтр по userId (parseInt + filter)
	// Попробуем повторить.
	userIdParam := r.URL.Query().Get("userId")
	if userIdParam != "" {
		userID, err := strconv.Atoi(userIdParam)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid userId"})
			return
		}

		filtered := make([]map[string]any, 0)
		for _, o := range orders {
			// JSON числа декодируются как float64
			if val, ok := o["userId"]; ok {
				switch v := val.(type) {
				case float64:
					if int(v) == userID {
						filtered = append(filtered, o)
					}
				case int:
					if v == userID {
						filtered = append(filtered, o)
					}
				}
			}
		}
		orders = filtered
	}

	writeJSON(w, http.StatusOK, orders)
}

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	mu.Lock()
	id := currentID
	currentID++

	payload["id"] = id
	fakeOrders[id] = payload
	mu.Unlock()

	writeJSON(w, http.StatusCreated, payload)
}

func updateOrderHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "orderId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order id"})
		return
	}

	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, ok := fakeOrders[id]; !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Order not found"})
		return
	}

	// Поведение как в Node: newOrder = { id, ...orderData }
	newOrder := map[string]any{
		"id": id,
	}
	for k, v := range updates {
		newOrder[k] = v
	}
	fakeOrders[id] = newOrder

	writeJSON(w, http.StatusOK, newOrder)
}

func deleteOrderHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "orderId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order id"})
		return
	}

	mu.Lock()
	order, ok := fakeOrders[id]
	if !ok {
		mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Order not found"})
		return
	}
	delete(fakeOrders, id)
	mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"message":     "Order deleted",
		"deletedOrder": order,
	})
}
