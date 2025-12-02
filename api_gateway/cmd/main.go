package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sony/gobreaker"
)

const (
	usersServiceURL  = "http://service_users:8000"
	ordersServiceURL = "http://service_orders:8000"
	port     = "8000"
	shutdownTimeout = 5 * time.Second
)

var httpClient = &http.Client{
	Timeout: 3 * time.Second,
}

type Gateway struct {
	usersCB  *gobreaker.CircuitBreaker
	ordersCB *gobreaker.CircuitBreaker
}

func main() {
	gw := &Gateway{
		usersCB:  newCircuitBreaker("users-service"),
		ordersCB: newCircuitBreaker("orders-service"),
	}

	srv := &http.Server{
		Addr: fmt.Sprintf(":%v", port),
		Handler: initRouter(gw),
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	
	go func() {
		log.Println("starting api-gateway on port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server starting failed: %v", err)
		}
	}() 
	
	<-ctx.Done()

	shutDownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	
	log.Printf("shutting down server gracefully")
	if err := srv.Shutdown(shutDownCtx); err != nil {
		log.Println("error when shutting down:", err)
	} else {
		log.Println("server stopped")
	}
}

func initRouter(gw *Gateway) *chi.Mux {
	r := chi.NewRouter()

	// Общие мидлвары
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS примерно как в express().use(cors())
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// USERS
	r.Get("/users/{userId}", gw.getUser)
	r.Post("/users", gw.createUser)
	r.Get("/users", gw.listUsers)
	r.Put("/users/{userId}", gw.updateUser)
	r.Delete("/users/{userId}", gw.deleteUser)

	// ORDERS
	r.Get("/orders/{orderId}", gw.getOrder)
	r.Post("/orders", gw.createOrder)
	r.Get("/orders", gw.listOrders)
	r.Put("/orders/{orderId}", gw.updateOrder)
	r.Delete("/orders/{orderId}", gw.deleteOrder)
	r.Get("/orders/status", gw.ordersStatus)
	r.Get("/orders/health", gw.ordersHealth)

	// Агрегация
	r.Get("/users/{userId}/details", gw.userDetails)

	// Health / Status самого шлюза
	r.Get("/health", gw.health)
	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status": "API Gateway is running",
		})
	})

	return r
}

func newCircuitBreaker(name string) *gobreaker.CircuitBreaker {
	settings := gobreaker.Settings{
		Name: name,
		// Через сколько ошибок и при каком проценте фейлов открывать "пробку"
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < 5 {
				return false
			}
			errorRate := float64(counts.TotalFailures) / float64(counts.Requests)
			return errorRate >= 0.5
		},
		Timeout: 3 * time.Second, // сколько ждать перед попыткой "полечить" сервис
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Printf("circuit %s changed from %s to %s", name, from.String(), to.String())
		},
	}
	return gobreaker.NewCircuitBreaker(settings)
}

// Общий helper для JSON-ответов
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Обёртка, которая делает HTTP-запрос через circuit breaker
func (g *Gateway) doRequest(cb *gobreaker.CircuitBreaker, method, url string, body []byte, r *http.Request) (*http.Response, error) {
	result, err := cb.Execute(func() (interface{}, error) {
		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			return nil, err
		}

		// Прокидываем заголовки, которые нам важны
		req.Header.Set("Content-Type", "application/json")
		if rid := r.Header.Get("X-Request-ID"); rid != "" {
			req.Header.Set("X-Request-ID", rid)
		}

		return httpClient.Do(req)
	})

	if err != nil {
		return nil, err
	}

	return result.(*http.Response), nil
}

// Helper: прочитать ответ сервиса и отдать клиенту
func forwardResponse(w http.ResponseWriter, resp *http.Response) {
	defer resp.Body.Close()

	for k, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// --- USERS handlers ---

func (g *Gateway) getUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	url := usersServiceURL + "/users/" + userID

	resp, err := g.doRequest(g.usersCB, http.MethodGet, url, nil, r)
	if err != nil {
		g.handleCBError(w, err, "Users")
		return
	}
	forwardResponse(w, resp)
}

func (g *Gateway) createUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	url := usersServiceURL + "/users"
	resp, err := g.doRequest(g.usersCB, http.MethodPost, url, body, r)
	if err != nil {
		g.handleCBError(w, err, "Users")
		return
	}
	forwardResponse(w, resp)
}

func (g *Gateway) listUsers(w http.ResponseWriter, r *http.Request) {
	url := usersServiceURL + "/users"
	resp, err := g.doRequest(g.usersCB, http.MethodGet, url, nil, r)
	if err != nil {
		g.handleCBError(w, err, "Users")
		return
	}
	forwardResponse(w, resp)
}

func (g *Gateway) updateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	url := usersServiceURL + "/users/" + userID
	resp, err := g.doRequest(g.usersCB, http.MethodPut, url, body, r)
	if err != nil {
		g.handleCBError(w, err, "Users")
		return
	}
	forwardResponse(w, resp)
}

func (g *Gateway) deleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	url := usersServiceURL + "/users/" + userID
	resp, err := g.doRequest(g.usersCB, http.MethodDelete, url, nil, r)
	if err != nil {
		g.handleCBError(w, err, "Users")
		return
	}
	forwardResponse(w, resp)
}

// --- ORDERS handlers ---

func (g *Gateway) getOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderId")
	url := ordersServiceURL + "/orders/" + orderID
	resp, err := g.doRequest(g.ordersCB, http.MethodGet, url, nil, r)
	if err != nil {
		g.handleCBError(w, err, "Orders")
		return
	}
	forwardResponse(w, resp)
}

func (g *Gateway) createOrder(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	url := ordersServiceURL + "/orders"
	resp, err := g.doRequest(g.ordersCB, http.MethodPost, url, body, r)
	if err != nil {
		g.handleCBError(w, err, "Orders")
		return
	}
	forwardResponse(w, resp)
}

func (g *Gateway) listOrders(w http.ResponseWriter, r *http.Request) {
	// Прокинем query-параметры (например, ?userId=)
	query := ""
	if r.URL.RawQuery != "" {
		query = "?" + r.URL.RawQuery
	}
	url := ordersServiceURL + "/orders" + query

	resp, err := g.doRequest(g.ordersCB, http.MethodGet, url, nil, r)
	if err != nil {
		g.handleCBError(w, err, "Orders")
		return
	}
	forwardResponse(w, resp)
}

func (g *Gateway) updateOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderId")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	url := ordersServiceURL + "/orders/" + orderID
	resp, err := g.doRequest(g.ordersCB, http.MethodPut, url, body, r)
	if err != nil {
		g.handleCBError(w, err, "Orders")
		return
	}
	forwardResponse(w, resp)
}

func (g *Gateway) deleteOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderId")
	url := ordersServiceURL + "/orders/" + orderID
	resp, err := g.doRequest(g.ordersCB, http.MethodDelete, url, nil, r)
	if err != nil {
		g.handleCBError(w, err, "Orders")
		return
	}
	forwardResponse(w, resp)
}

func (g *Gateway) ordersStatus(w http.ResponseWriter, r *http.Request) {
	url := ordersServiceURL + "/orders/status"
	resp, err := g.doRequest(g.ordersCB, http.MethodGet, url, nil, r)
	if err != nil {
		g.handleCBError(w, err, "Orders")
		return
	}
	forwardResponse(w, resp)
}

func (g *Gateway) ordersHealth(w http.ResponseWriter, r *http.Request) {
	url := ordersServiceURL + "/orders/health"
	resp, err := g.doRequest(g.ordersCB, http.MethodGet, url, nil, r)
	if err != nil {
		g.handleCBError(w, err, "Orders")
		return
	}
	forwardResponse(w, resp)
}

// --- Aggregation: /users/{userId}/details ---

func (g *Gateway) userDetails(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	// 1) достаём юзера
	userURL := usersServiceURL + "/users/" + userID

	// 2) достаём все заказы
	ordersURL := ordersServiceURL + "/orders"

	type result struct {
		resp *http.Response
		err  error
	}

	userCh := make(chan result, 1)
	ordersCh := make(chan result, 1)

	go func() {
		resp, err := g.doRequest(g.usersCB, http.MethodGet, userURL, nil, r)
		userCh <- result{resp, err}
	}()

	go func() {
		resp, err := g.doRequest(g.ordersCB, http.MethodGet, ordersURL, nil, r)
		ordersCh <- result{resp, err}
	}()

	userRes := <-userCh
	ordersRes := <-ordersCh

	if userRes.err != nil {
		g.handleCBError(w, userRes.err, "Users")
		return
	}
	if ordersRes.err != nil {
		g.handleCBError(w, ordersRes.err, "Orders")
		return
	}

	defer userRes.resp.Body.Close()
	defer ordersRes.resp.Body.Close()

	if userRes.resp.StatusCode == http.StatusNotFound {
		// Просто прокинем как 404
		body, _ := io.ReadAll(userRes.resp.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(body)
		return
	}

	// Распарсим user
	var user any
	if err := json.NewDecoder(userRes.resp.Body).Decode(&user); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to parse user"})
		return
	}

	// Распарсим orders (ожидаем массив)
	var orders []map[string]any
	if err := json.NewDecoder(ordersRes.resp.Body).Decode(&orders); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to parse orders"})
		return
	}

	// Фильтруем заказы по userId (как в Node-коде)
	filtered := make([]map[string]any, 0)
	for _, o := range orders {
		if val, ok := o["userId"]; ok {
			switch v := val.(type) {
			case float64:
				// userId в заказе как число (float64 из JSON)
				if fmt.Sprintf("%d", int(v)) == userID {
					filtered = append(filtered, o)
				}
			case int:
				if fmt.Sprintf("%d", v) == userID {
					filtered = append(filtered, o)
				}
			case string:
				if v == userID {
					filtered = append(filtered, o)
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":   user,
		"orders": filtered,
	})
}

// --- Health шлюза ---

func (g *Gateway) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "API Gateway is running",
		"circuits": map[string]any{
			"users": map[string]any{
				"state": g.usersCB.State().String(),
				"stats": g.usersCB.Counts(),
			},
			"orders": map[string]any{
				"state": g.ordersCB.State().String(),
				"stats": g.ordersCB.Counts(),
			},
		},
	})
}

// --- Ошибки circuit breaker ---

func (g *Gateway) handleCBError(w http.ResponseWriter, err error, service string) {
	log.Printf("%s service error: %v", service, err)

	if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": service + " service temporarily unavailable",
		})
		return
	}

	writeJSON(w, http.StatusInternalServerError, map[string]string{
		"error": "Internal server error",
	})
}
