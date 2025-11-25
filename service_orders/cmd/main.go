package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"service_orders/internal/handler"
	"service_orders/internal/repository"
	"service_orders/internal/service"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	port            = "8000"
	shutdownTimeout = 5 * time.Second
)

func main() {
	// DI
	orderRepo := repository.NewInMemoryOrderRepository()
	orderService := service.NewOrderService(orderRepo)
	orderController := handler.NewOrderController(*orderService)

	r := initRouter(orderController)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: r,
	}

	// graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Println("starting orders-service on port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server starting failed: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	log.Println("shutting down server gracefully")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Println("error when shutting down:", err)
	} else {
		log.Println("server stopped")
	}
}

func initRouter(order *handler.OrderController) *chi.Mux {
	r := chi.NewRouter()

	// middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// routes
	r.Get("/orders/status", order.Status)
	r.Get("/orders/health", order.Health)
	r.Get("/orders/{id}", order.GetOrder)
	r.Get("/orders", order.ListOrders)
	r.Post("/orders", order.CreateOrder)
	r.Put("/orders/{id}", order.UpdateOrder)
	r.Delete("/orders/{id}", order.DeleteOrder)

	return r
}
