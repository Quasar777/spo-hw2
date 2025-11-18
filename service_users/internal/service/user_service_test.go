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