package handler

import (
	"encoding/json"
	"net/http"
	"service_users/internal/model"
	"service_users/internal/service"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UserController struct {
	service service.UserService
}

func NewUserController(s service.UserService) *UserController {
	return &UserController{service: s}
}

func (c *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, `{"error": "Invalid id"}`, http.StatusBadRequest)
		return
	}
	
	var user *model.User
	user, err = c.service.GetUser(id)

	if err != nil {
		switch err {
		case model.ErrUserNotFound:
			http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		default:
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		}
		return
	}
	
	c.writeJSON(w, http.StatusOK, user)
}


// Helpers
func (c *UserController) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
	}
}