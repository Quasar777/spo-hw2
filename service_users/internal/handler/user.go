package handler

import (
	"encoding/json"
	"net/http"
	"service_users/internal/model"
	"service_users/internal/service"
	"strconv"
	"strings"
	"time"

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

func (c *UserController) GetMany(w http.ResponseWriter, r *http.Request) {
	users, err := c.service.GetAllUsers()
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	c.writeJSON(w, http.StatusOK, users)
}

func (c *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var reqUser model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&reqUser); err != nil {
		http.Error(w, `{"error": "Invalid JSON}`, http.StatusBadRequest)
		return
	}

	id, err := c.service.CreateUser(reqUser)
	if err != nil {
		switch err {
		case model.ErrMissingRequiredFields:
			http.Error(w, `{"error": "Missing required fields"}`, http.StatusBadRequest)
		case model.ErrInvalidEmail:
			http.Error(w, `{"error": "Invalid email"}`, http.StatusBadRequest)
		case model.ErrUniqueEmailConflict:
			http.Error(w, `{"error": "User with this email is already exists"}`, http.StatusConflict)
		default:
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		}
		return
	}

	response := map[string]interface{}{
		"id": id,
		"message": "User created succesfully",
	}

	c.writeJSON(w, http.StatusCreated, response)
}

func (c *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var reqUser model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&reqUser); err != nil {
		http.Error(w, `{"error": "Invalid JSON}`, http.StatusBadRequest)
		return
	}
	if reqUser.ID == 0 {
		http.Error(w, `{"error": "Id is required"}`, http.StatusBadRequest)
		return
	}

	err := c.service.UpdateUser(reqUser)

	if err != nil {
		switch err {
		case model.ErrUserNotFound:
			http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		case model.ErrMissingRequiredFields:
			http.Error(w, `{"error": "Missing required fields"}`, http.StatusBadRequest)
		case model.ErrInvalidEmail:
			http.Error(w, `{"error": "Invalid email"}`, http.StatusBadRequest)
		case model.ErrUniqueEmailConflict:
			http.Error(w, `{"error": "User with this email is already exists"}`, http.StatusConflict)
		default:
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		}
		return
	}

	response := map[string]string {
		"message": "User updated successfully",
	}

	c.writeJSON(w, http.StatusOK, response)
}

func (c *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, `{"error": "Invalid id"}`, http.StatusBadRequest)
		return
	}

	err = c.service.DeleteUser(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]string {
		"message": "User deleted successfully",
	}

	c.writeJSON(w, http.StatusOK, response)
}

// Health Check 
func (c *UserController) Health(w http.ResponseWriter, r *http.Request) {
	response := map[string]any {
		"status":   "OK",
		"service":  "Users Service",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	c.writeJSON(w, http.StatusOK, response)
}

// Service status
func (c *UserController) Status(w http.ResponseWriter, r *http.Request) {
	response := map[string]string {
		"status": "Users service is running",
	}

	c.writeJSON(w, http.StatusOK, response)
}

// Helpers
func (c *UserController) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
	}
}