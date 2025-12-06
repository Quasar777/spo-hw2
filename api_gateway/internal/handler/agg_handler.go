package handler

import (
	"api_gateway/internal/model"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sony/gobreaker"
)

type AggregationHandler struct {
	client        *http.Client
	usersCB       *gobreaker.CircuitBreaker
	ordersCB      *gobreaker.CircuitBreaker
	usersBaseURL  string
	ordersBaseURL string
}

func NewAggregationHandler(
	client *http.Client,
	usersCB *gobreaker.CircuitBreaker,
	ordersCB *gobreaker.CircuitBreaker,
	usersBaseURL string,
	ordersBaseURL string,
) *AggregationHandler {
	return &AggregationHandler{
		client:        client,
		usersCB:       usersCB,
		ordersCB:      ordersCB,
		usersBaseURL:  usersBaseURL,
		ordersBaseURL: ordersBaseURL,
	}
}

type result struct {
	resp *http.Response
	err  error
}

func (h *AggregationHandler) doUsersRequest(method, path string, body []byte, r *http.Request) (*http.Response, error) {
	url := h.usersBaseURL + path

	res, err := h.usersCB.Execute(func() (interface{}, error) {
		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		if rid := r.Header.Get("X-Request-ID"); rid != "" {
			req.Header.Set("X-Request-ID", rid)
		}

		return h.client.Do(req)
	})
	if err != nil {
		return nil, err
	}

	return res.(*http.Response), nil
}

func (h *AggregationHandler) doOrdersRequest(method, path string, body []byte, r *http.Request) (*http.Response, error) {
	url := h.ordersBaseURL + path

	res, err := h.ordersCB.Execute(func() (interface{}, error) {
		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		if rid := r.Header.Get("X-Request-ID"); rid != "" {
			req.Header.Set("X-Request-ID", rid)
		}

		return h.client.Do(req)
	})
	if err != nil {
		return nil, err
	}

	return res.(*http.Response), nil
}

func (h *AggregationHandler) UserDetails(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	if userIDStr == "" {
		http.Error(w, `{"error": "userId is required"}`, http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid userId"}`, http.StatusBadRequest)
		return
	}

	userPath := "/users/" + userIDStr
	ordersPath := "/orders"

	userCh := make(chan result, 1)
	ordersCh := make(chan result, 1)

	go func() {
		resp, err := h.doUsersRequest(http.MethodGet, userPath, nil, r)
		userCh <- result{resp: resp, err: err}
	}()

	go func() {
		resp, err := h.doOrdersRequest(http.MethodGet, ordersPath, nil, r)
		ordersCh <- result{resp: resp, err: err}
	}()

	userRes := <-userCh
	ordersRes := <-ordersCh

	if userRes.err != nil {
		handleCBError(w, userRes.err, "Users")
		return
	}
	if ordersRes.err != nil {
		handleCBError(w, ordersRes.err, "Orders")
		return
	}

	defer userRes.resp.Body.Close()
	defer ordersRes.resp.Body.Close()

	if userRes.resp.StatusCode == http.StatusNotFound {
		body, _ := io.ReadAll(userRes.resp.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(body)
		return
	}

	if userRes.resp.StatusCode >= 400 {
		body, _ := io.ReadAll(userRes.resp.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		if len(body) > 0 {
			_, _ = w.Write(body)
		} else {
			_, _ = w.Write([]byte(`{"error":"failed to fetch user"}`))
		}
		return
	}

	if ordersRes.resp.StatusCode >= 400 {
		body, _ := io.ReadAll(ordersRes.resp.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		if len(body) > 0 {
			_, _ = w.Write(body)
		} else {
			_, _ = w.Write([]byte(`{"error":"failed to fetch orders"}`))
		}
		return
	}

	var user model.User
	if err := json.NewDecoder(userRes.resp.Body).Decode(&user); err != nil {
		http.Error(w, `{"error": "failed to parse user"}`, http.StatusInternalServerError)
		return
	}

	var orders []model.Order
	if err := json.NewDecoder(ordersRes.resp.Body).Decode(&orders); err != nil {
		http.Error(w, `{"error": "failed to parse orders"}`, http.StatusInternalServerError)
		return
	}

	filtered := make([]model.Order, 0)
	for _, o := range orders {
		if o.UserId == userID {
			filtered = append(filtered, o)
		}
	}

	response := model.UserDetailsResponse{
		User:   user,
		Orders: filtered,
	}

	writeJSON(w, http.StatusOK, response)
}