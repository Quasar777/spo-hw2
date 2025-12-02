package handler

import (
	"bytes"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sony/gobreaker"
)

type OrdersHandler struct {
	client  *http.Client
	cb      *gobreaker.CircuitBreaker
	baseURL string
}

func NewOrdersHandler(cl *http.Client, url string, cbr *gobreaker.CircuitBreaker) *OrdersHandler {
	return &OrdersHandler{
		client:  cl,
		cb:      cbr,
		baseURL: url,
	}
}

func (h *OrdersHandler) doRequest(method, path string, body []byte, r *http.Request) (*http.Response, error) {
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

func (h *OrdersHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderId")

	resp, err := h.doRequest(http.MethodGet, "/orders/"+orderID, nil, r)
	if err != nil {
		handleCBError(w, err, "Orders")
		return
	}
	forwardResponse(w, resp)
}

func (h *OrdersHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	resp, err := h.doRequest(http.MethodPost, "/orders", body, r)
	if err != nil {
		handleCBError(w, err, "Orders")
		return
	}
	forwardResponse(w, resp)
}

func (h *OrdersHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	path := "/orders"
	if r.URL.RawQuery != "" {
		path = path + "?" + r.URL.RawQuery
	}

	resp, err := h.doRequest(http.MethodGet, path, nil, r)
	if err != nil {
		handleCBError(w, err, "Orders")
		return
	}
	forwardResponse(w, resp)
}

func (h *OrdersHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderId")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	resp, err := h.doRequest(http.MethodPut, "/orders/"+orderID, body, r)
	if err != nil {
		handleCBError(w, err, "Orders")
		return
	}
	forwardResponse(w, resp)
}

func (h *OrdersHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderId")

	resp, err := h.doRequest(http.MethodDelete, "/orders/"+orderID, nil, r)
	if err != nil {
		handleCBError(w, err, "Orders")
		return
	}
	forwardResponse(w, resp)
}

func (h *OrdersHandler) OrdersStatus(w http.ResponseWriter, r *http.Request) {
	resp, err := h.doRequest(http.MethodGet, "/orders/status", nil, r)
	if err != nil {
		handleCBError(w, err, "Orders")
		return
	}
	forwardResponse(w, resp)
}

func (h *OrdersHandler) OrdersHealth(w http.ResponseWriter, r *http.Request) {
	resp, err := h.doRequest(http.MethodGet, "/orders/health", nil, r)
	if err != nil {
		handleCBError(w, err, "Orders")
		return
	}
	forwardResponse(w, resp)
}