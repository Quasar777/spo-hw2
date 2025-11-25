package service_test

import (
	"service_users/internal/model"
	"service_users/internal/repository"
	"service_users/internal/service"
	"testing"
)

func TestUserService_GetExistingUser_Success(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	user, err := svc.GetUser(1)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user == nil {
		t.Fatalf("expected user, got nil")
	}
	if user.ID != 1 {
		t.Errorf("expected ID = 1, got %d", user.ID)
	}
	if user.Name != "Alice" {
		t.Errorf("expected Name = Alice, got %s", user.Name)
	}
}

func TestUserService_GetUserWithNotExistingId_NotFound(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	user, err := svc.GetUser(999)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != model.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil user, got: %+v", user)
	}
}

func TestUserService_GetUserWithWrongId_ErrInvalidId(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	user, err := svc.GetUser(999)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != model.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil user, got: %+v", user)
	}
}

func TestUserService_GetAllUsers_Success(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	_, err := svc.GetAllUsers()
	
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}	
}

func TestUserService_CreateUser_Success(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	req := model.CreateUserRequest{
		Name:  "Bob",
		Email: "bob@example.com",
	}

	id, err := svc.CreateUser(req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if id == 0 {
		t.Fatalf("expected non-zero id, got 0")
	}

	// проверим, что юзер реально появился
	user, err := svc.GetUser(id)
	if err != nil {
		t.Fatalf("expected user, got error: %v", err)
	}
	if user.Name != req.Name || user.Email != req.Email {
		t.Errorf("unexpected user data: %+v", user)
	}
}

func TestUserService_CreateUser_MissingFields(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	tt := []struct {
		name string
		req  model.CreateUserRequest
	}{
		{
			name: "empty name",
			req: model.CreateUserRequest{
				Name:  "",
				Email: "test@example.com",
			},
		},
		{
			name: "empty email",
			req: model.CreateUserRequest{
				Name:  "Test",
				Email: "",
			},
		},
		{
			name: "both empty",
			req:  model.CreateUserRequest{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			id, err := svc.CreateUser(tc.req)
			if err != model.ErrMissingRequiredFields {
				t.Fatalf("expected ErrMissingRequiredFields, got: %v", err)
			}
			if id != 0 {
				t.Fatalf("expected id = 0 on error, got %d", id)
			}
		})
	}
}

func TestUserService_CreateUser_InvalidEmail(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	req := model.CreateUserRequest{
		Name:  "Test",
		Email: "invalid-email",
	}

	id, err := svc.CreateUser(req)
	if err != model.ErrInvalidEmail {
		t.Fatalf("expected ErrInvalidEmail, got: %v", err)
	}
	if id != 0 {
		t.Fatalf("expected id = 0 on error, got %d", id)
	}
}

func TestUserService_CreateUser_EmailConflict(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	// в репозитории уже есть alice@example.com
	req := model.CreateUserRequest{
		Name:  "Another Alice",
		Email: "alice@example.com",
	}

	id, err := svc.CreateUser(req)
	if err != model.ErrUniqueEmailConflict {
		t.Fatalf("expected ErrUniqueEmailConflict, got: %v", err)
	}
	if id != 0 {
		t.Fatalf("expected id = 0 on conflict, got %d", id)
	}
}

func TestUserService_UpdateUser_Success(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	req := model.UpdateUserRequest{
		ID:   1,
		Name: "New Alice",
	}

	if err := svc.UpdateUser(req); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	user, err := svc.GetUser(1)
	if err != nil {
		t.Fatalf("GetUser error: %v", err)
	}
	if user.Name != "New Alice" {
		t.Errorf("expected name = New Alice, got %s", user.Name)
	}
}

func TestUserService_UpdateUser_MissingName(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	req := model.UpdateUserRequest{
		ID:   1,
		Name: "",
	}

	err := svc.UpdateUser(req)
	if err != model.ErrMissingRequiredFields {
		t.Fatalf("expected ErrMissingRequiredFields, got: %v", err)
	}
}

func TestUserService_UpdateUser_NotFound(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	req := model.UpdateUserRequest{
		ID:   999,
		Name: "Ghost",
	}

	err := svc.UpdateUser(req)
	if err != model.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	// сначала убеждаемся, что юзер есть
	if _, err := svc.GetUser(1); err != nil {
		t.Fatalf("expected user 1 to exist, got error: %v", err)
	}

	if err := svc.DeleteUser(1); err != nil {
		t.Fatalf("expected no error on delete, got: %v", err)
	}

	// теперь должен быть not found
	if _, err := svc.GetUser(1); err != model.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound after delete, got: %v", err)
	}
}