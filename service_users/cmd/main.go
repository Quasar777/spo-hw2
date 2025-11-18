package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"service_users/internal/handler"
	"service_users/internal/repository"
	"service_users/internal/service"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	port     = "8000"
	shutdownTimeout = 5 * time.Second
)

func main() {
	// Dependency injection
	userRepository := repository.NewUserRepository()
	userService := service.NewUserService(userRepository)
	user := handler.NewUserController(*userService)

	srv := &http.Server{
		Addr: fmt.Sprintf(":%v", port),
		Handler: initRouter(user),
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	
	go func() {
		log.Println("starting user-service on port", port)
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

func initRouter(user *handler.UserController) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Routes
	r.Get("/users", user.GetMany)
	r.Get("/users/{id}", user.GetUser)
	r.Post("/users", user.CreateUser)
	r.Put("/users", user.UpdateUser)
	r.Delete("/users/{id}",user.DeleteUser)
	r.Get("/users/health", user.Health)
	r.Get("/users/status", user.Status)	
	
	return r
}