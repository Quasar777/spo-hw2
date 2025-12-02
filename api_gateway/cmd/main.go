package main

import (
	"api_gateway/internal/handler"
	"context"
	"fmt"
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

	usersHandler := handler.NewUserHandler(httpClient, usersServiceURL, gw.usersCB)
	ordersHandler := handler.NewOrdersHandler(httpClient, ordersServiceURL, gw.ordersCB)
	aggHandler   := handler.NewAggregationHandler(httpClient, gw.usersCB, gw.ordersCB, usersServiceURL, ordersServiceURL)
	healthHandler := handler.NewHealthHandler(gw.usersCB, gw.ordersCB)

	srv := &http.Server{
		Addr: fmt.Sprintf(":%v", port),
		Handler: initRouter(gw, usersHandler, ordersHandler, aggHandler, healthHandler),
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

func initRouter(gw *Gateway, users *handler.UsersHandler, orders *handler.OrdersHandler, agg *handler.AggregationHandler, health *handler.HealthHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/users/{userId}", users.GetUser)
	r.Post("/users", users.CreateUser)
	r.Get("/users", users.ListUsers)
	r.Put("/users", users.UpdateUser)
	r.Delete("/users/{userId}", users.DeleteUser)

	r.Get("/orders/{orderId}", orders.GetOrder)
	r.Post("/orders", orders.CreateOrder)
	r.Get("/orders", orders.ListOrders)
	r.Put("/orders", orders.UpdateOrder)
	r.Delete("/orders/{orderId}", orders.DeleteOrder)
	r.Get("/orders/status", orders.OrdersStatus)
	r.Get("/orders/health", orders.OrdersHealth)

	r.Get("/users/{userId}/details", agg.UserDetails)

	r.Get("/health", health.Health)
	r.Get("/status", health.Status)

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