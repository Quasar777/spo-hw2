package handler

import (
	"bytes"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sony/gobreaker"
)

type UsersHandler struct {
	client  *http.Client
	cb      *gobreaker.CircuitBreaker
	baseURL string
}

func NewUserHandler(cl *http.Client, url string, cbr *gobreaker.CircuitBreaker) *UsersHandler {
	return &UsersHandler{
		client:  cl,
		baseURL: url,
		cb:      cbr,
	}
}

func (h *UsersHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	resp, err := h.doRequest(http.MethodGet, "/users/"+userID, nil, r)
	if err != nil {
		handleCBError(w, err, "Users")
		return
	}

	forwardResponse(w, resp)
}

func (h *UsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	resp, err := h.doRequest(http.MethodPost, "/users", body, r)
	if err != nil {
		handleCBError(w, err, "Users")
		return
	}
	forwardResponse(w, resp)
}

func (h *UsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	resp, err := h.doRequest(http.MethodGet, "/users", nil, r)
	if err != nil {
		handleCBError(w, err, "Users")
		return
	}
	forwardResponse(w, resp)
}

func (h *UsersHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	resp, err := h.doRequest(http.MethodPut, "/users", body, r)
	if err != nil {
		handleCBError(w, err, "Users")
		return
	}
	forwardResponse(w, resp)
}

func (h *UsersHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	resp, err := h.doRequest(http.MethodDelete, "/users/"+userID, nil, r)
	if err != nil {
		handleCBError(w, err, "Users")
		return
	}
	forwardResponse(w, resp)
}

// общий метод для запросов в users-service через circuit breaker
func (h *UsersHandler) doRequest(method, path string, body []byte, r *http.Request) (*http.Response, error) {
	url := h.baseURL + path

	result, err := h.cb.Execute(func() (interface{}, error) {
		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			return nil, err
		}

		// Прокидываем важные заголовки
		req.Header.Set("Content-Type", "application/json")
		if rid := r.Header.Get("X-Request-ID"); rid != "" {
			req.Header.Set("X-Request-ID", rid)
		}

		return h.client.Do(req)
	})

	if err != nil {
		return nil, err
	}

	return result.(*http.Response), nil
}
