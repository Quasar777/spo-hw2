package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"service_users/internal/handler"
	"service_users/internal/model"
	"service_users/internal/repository"
	"service_users/internal/service"
)

func newTestController() (*handler.UserController, *service.UserService, *repository.UserRepository) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)
	ctrl := handler.NewUserController(*svc)
	return ctrl, svc, repo
}

func TestRegisterHandler_Success(t *testing.T) {
	ctrl, _, _ := newTestController()

	r := chi.NewRouter()
	r.Post("/auth/register", ctrl.Register)

	body := []byte(`{"email":"test@example.com","name":"Test User","password":"secret123"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var resp struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success = true, got false")
	}
	if resp.Data == nil {
		t.Fatalf("expected data field, got nil")
	}
	if _, ok := resp.Data["id"]; !ok {
		t.Fatalf("expected 'id' in data, got: %+v", resp.Data)
	}
}

func TestRegisterHandler_InvalidJSON(t *testing.T) {
	ctrl, _, _ := newTestController()

	r := chi.NewRouter()
	r.Post("/auth/register", ctrl.Register)

	body := []byte(`{invalid json`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !contains(rr.Body.String(), "invalid JSON") {
		t.Fatalf("expected error message about invalid JSON, got: %s", rr.Body.String())
	}
}

func TestRegisterHandler_MissingFields(t *testing.T) {
	ctrl, _, _ := newTestController()

	r := chi.NewRouter()
	r.Post("/auth/register", ctrl.Register)

	body := []byte(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !contains(rr.Body.String(), "email, name and password are required") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestRegisterHandler_EmailConflict(t *testing.T) {
	ctrl, svc, _ := newTestController()

	r := chi.NewRouter()
	r.Post("/auth/register", ctrl.Register)

	_, err := svc.Register(model.RegisterRequest{
		Email:    "same@example.com",
		Name:     "User1",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("unexpected error on first register: %v", err)
	}

	body := []byte(`{"email":"same@example.com","name":"User2","password":"anotherpass"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusConflict, rr.Code, rr.Body.String())
	}
	if !contains(rr.Body.String(), "user with this email already exists") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestLoginHandler_Success(t *testing.T) {
	ctrl, svc, _ := newTestController()

	r := chi.NewRouter()
	r.Post("/auth/login", ctrl.Login)
	r.Post("/auth/register", ctrl.Register)

	_, err := svc.Register(model.RegisterRequest{
		Email:    "login@example.com",
		Name:     "Login User",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("unexpected error on register: %v", err)
	}

	body := []byte(`{"email":"login@example.com","password":"secret123"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success = true, got false")
	}
	token, ok := resp.Data["token"].(string)
	if !ok || token == "" {
		t.Fatalf("expected non-empty token in data, got: %+v", resp.Data)
	}
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	ctrl, svc, _ := newTestController()

	r := chi.NewRouter()
	r.Post("/auth/login", ctrl.Login)

	_, err := svc.Register(model.RegisterRequest{
		Email:    "login2@example.com",
		Name:     "User",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("unexpected error on register: %v", err)
	}

	body := []byte(`{"email":"login2@example.com","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
	if !contains(rr.Body.String(), "invalid email or password") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestLoginHandler_MissingFields(t *testing.T) {
	ctrl, _, _ := newTestController()

	r := chi.NewRouter()
	r.Post("/auth/login", ctrl.Login)

	body := []byte(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !contains(rr.Body.String(), "email and password are required") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestGetMeHandler_Success(t *testing.T) {
	ctrl, svc, _ := newTestController()

	r := chi.NewRouter()

	r.With(ctrl.AuthMiddleware).Get("/users/me", ctrl.GetMe)

	id, err := svc.Register(model.RegisterRequest{
		Email:    "me@example.com",
		Name:     "Me User",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("unexpected error on register: %v", err)
	}
	if id == 0 {
		t.Fatalf("expected non-zero id")
	}

	token, err := svc.Login(model.LoginRequest{
		Email:    "me@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("unexpected error on login: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var user model.User
	if err := json.NewDecoder(rr.Body).Decode(&user); err != nil {
		t.Fatalf("failed to decode user: %v", err)
	}

	if user.ID != id {
		t.Errorf("expected user ID %d, got %d", id, user.ID)
	}
	if user.Email != "me@example.com" {
		t.Errorf("expected email me@example.com, got %s", user.Email)
	}
}

func TestGetMeHandler_MissingAuthorizationHeader(t *testing.T) {
	ctrl, _, _ := newTestController()

	r := chi.NewRouter()
	r.With(ctrl.AuthMiddleware).Get("/users/me", ctrl.GetMe)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if !contains(rr.Body.String(), "missing Authorization header") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestGetMeHandler_InvalidToken(t *testing.T) {
	ctrl, _, _ := newTestController()

	r := chi.NewRouter()
	r.With(ctrl.AuthMiddleware).Get("/users/me", ctrl.GetMe)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if !contains(rr.Body.String(), "invalid or expired token") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

// helpers

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}