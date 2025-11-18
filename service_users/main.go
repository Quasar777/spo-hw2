package main

import (
	"encoding/json"
	"log"
	"net/http"
	"service_users/internal/model"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)



var (
	mu        sync.RWMutex
	fakeUsers = map[int]*model.User{}
	currentID = 1
)

func main() {
	r := chi.NewRouter()

	// Простые полезные мидлвары
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// JSON маршруты
	r.Get("/users", getUsersHandler)
	r.Post("/users", createUserHandler)
	r.Get("/users/health", healthHandler)
	r.Get("/users/status", statusHandler)
	r.Get("/users/{userId}", getUserHandler)
	r.Put("/users/{userId}", updateUserHandler)
	r.Delete("/users/{userId}", deleteUserHandler)

	port := ":8000" // как и в Node
	log.Printf("Users service running on port %s\n", port)
	if err := http.ListenAndServe("0.0.0.0"+port, r); err != nil {
		log.Fatal(err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "OK",
		"service":  "Users Service",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "Users service is running",
	})
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	users := make([]*model.User, 0, len(fakeUsers))
	for _, u := range fakeUsers {
		users = append(users, u)
	}

	writeJSON(w, http.StatusOK, users)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	now := time.Now()

	mu.Lock()
	id := currentID
	currentID++

	user := &model.User{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Копируем все поля из payload (минимально, чтобы повторить поведение Node)
	if email, ok := payload["email"].(string); ok {
		user.Email = email
	}
	if name, ok := payload["name"].(string); ok {
		user.Name = name
	}

	fakeUsers[id] = user
	mu.Unlock()

	// Возвращаем как в Node: статус 201 и созданный юзер
	writeJSON(w, http.StatusCreated, user)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	mu.RLock()
	user, ok := fakeUsers[id]
	mu.RUnlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	mu.Lock()
	user, ok := fakeUsers[id]
	if !ok {
		mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	// Обновляем только те поля, которые точно знаем
	if email, ok := updates["email"].(string); ok {
		user.Email = email
	}
	if name, ok := updates["name"].(string); ok {
		user.Name = name
	}
	user.UpdatedAt = time.Now()

	fakeUsers[id] = user
	mu.Unlock()

	writeJSON(w, http.StatusOK, user)
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	mu.Lock()
	user, ok := fakeUsers[id]
	if !ok {
		mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}
	delete(fakeUsers, id)
	mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"message":     "User deleted",
		"deletedUser": user,
	})
}
