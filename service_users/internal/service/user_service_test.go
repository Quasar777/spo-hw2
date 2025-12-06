package service_test

import (
	"service_users/internal/model"
	"service_users/internal/repository"
	"service_users/internal/service"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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

func TestUserService_Register_Success(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	req := model.RegisterRequest{
		Email:    "reguser@example.com",
		Name:     "Reg User",
		Password: "secret123",
	}

	id, err := svc.Register(req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if id == 0 {
		t.Fatalf("expected non-zero id, got 0")
	}

	u, err := svc.GetUser(id)
	if err != nil {
		t.Fatalf("expected user after register, got error: %v", err)
	}

	if u.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, u.Email)
	}
	if u.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, u.Name)
	}
	if len(u.Roles) == 0 || u.Roles[0] != "user" {
		t.Errorf("expected roles [user], got %+v", u.Roles)
	}
	if u.PasswordHash == "" {
		t.Fatalf("expected password hash to be set")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		t.Fatalf("stored password hash does not match original password: %v", err)
	}
}

func TestUserService_Register_MissingFields(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	tests := []struct {
		name string
		req  model.RegisterRequest
	}{
		{
			name: "empty name",
			req: model.RegisterRequest{
				Email:    "test@example.com",
				Password: "secret123",
			},
		},
		{
			name: "empty email",
			req: model.RegisterRequest{
				Name:     "Test",
				Password: "secret123",
			},
		},
		{
			name: "empty password",
			req: model.RegisterRequest{
				Name:  "Test",
				Email: "test@example.com",
			},
		},
		{
			name: "all empty",
			req:  model.RegisterRequest{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id, err := svc.Register(tc.req)
			if err != model.ErrMissingRequiredFields {
				t.Fatalf("expected ErrMissingRequiredFields, got: %v", err)
			}
			if id != 0 {
				t.Fatalf("expected id = 0 on error, got %d", id)
			}
		})
	}
}

func TestUserService_Register_InvalidEmail(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	req := model.RegisterRequest{
		Name:     "Test",
		Email:    "invalid-email",
		Password: "secret123",
	}

	id, err := svc.Register(req)
	if err != model.ErrInvalidEmail {
		t.Fatalf("expected ErrInvalidEmail, got: %v", err)
	}
	if id != 0 {
		t.Fatalf("expected id = 0 on error, got %d", id)
	}
}

func TestUserService_Register_ShortPassword(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	req := model.RegisterRequest{
		Name:     "Test",
		Email:    "test@example.com",
		Password: "123", // меньше 6 символов
	}

	id, err := svc.Register(req)
	if err != model.ErrInvalidPassword {
		t.Fatalf("expected ErrInvalidPassword, got: %v", err)
	}
	if id != 0 {
		t.Fatalf("expected id = 0 on error, got %d", id)
	}
}

func TestUserService_Register_EmailConflict(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	// сначала регистрируем нового пользователя
	first := model.RegisterRequest{
		Name:     "User1",
		Email:    "same@example.com",
		Password: "secret123",
	}
	_, err := svc.Register(first)
	if err != nil {
		t.Fatalf("unexpected error on first register: %v", err)
	}

	// потом пытаемся зарегистрировать с тем же email
	second := model.RegisterRequest{
		Name:     "User2",
		Email:    "same@example.com",
		Password: "anotherpass",
	}
	id, err := svc.Register(second)
	if err != model.ErrUniqueEmailConflict {
		t.Fatalf("expected ErrUniqueEmailConflict, got: %v", err)
	}
	if id != 0 {
		t.Fatalf("expected id = 0 on conflict, got %d", id)
	}
}

func TestUserService_Login_Success(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	regReq := model.RegisterRequest{
		Email:    "login@example.com",
		Name:     "Login User",
		Password: "secret123",
	}

	id, err := svc.Register(regReq)
	if err != nil {
		t.Fatalf("unexpected error on register: %v", err)
	}
	if id == 0 {
		t.Fatalf("expected non-zero id")
	}

	// логинимся
	loginReq := model.LoginRequest{
		Email:    "login@example.com",
		Password: "secret123",
	}

	token, err := svc.Login(loginReq)
	if err != nil {
		t.Fatalf("expected no error on login, got: %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}

	// проверяем, что токен нормально парсится
	info, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("expected token to be parsed, got error: %v", err)
	}
	if info.UserID != id {
		t.Errorf("expected UserID %d, got %d", id, info.UserID)
	}
	if info.Email != regReq.Email {
		t.Errorf("expected Email %s, got %s", regReq.Email, info.Email)
	}
	if len(info.Roles) == 0 || info.Roles[0] != "user" {
		t.Errorf("expected roles [user], got %+v", info.Roles)
	}
}

func TestUserService_Login_InvalidPassword(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	regReq := model.RegisterRequest{
		Email:    "login2@example.com",
		Name:     "Login User 2",
		Password: "secret123",
	}

	_, err := svc.Register(regReq)
	if err != nil {
		t.Fatalf("unexpected error on register: %v", err)
	}

	badLogin := model.LoginRequest{
		Email:    "login2@example.com",
		Password: "wrong-password",
	}

	token, err := svc.Login(badLogin)
	if err != model.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
	if token != "" {
		t.Fatalf("expected empty token on error, got: %s", token)
	}
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	req := model.LoginRequest{
		Email:    "no_such_user@example.com",
		Password: "secret123",
	}

	token, err := svc.Login(req)
	if err != model.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
	if token != "" {
		t.Fatalf("expected empty token on error, got: %s", token)
	}
}

func TestUserService_Login_MissingFields(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	tests := []struct {
		name string
		req  model.LoginRequest
	}{
		{
			name: "empty email",
			req: model.LoginRequest{
				Email:    "",
				Password: "secret123",
			},
		},
		{
			name: "empty password",
			req: model.LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
		},
		{
			name: "both empty",
			req:  model.LoginRequest{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token, err := svc.Login(tc.req)
			if err != model.ErrMissingRequiredFields {
				t.Fatalf("expected ErrMissingRequiredFields, got: %v", err)
			}
			if token != "" {
				t.Fatalf("expected empty token on error, got: %s", token)
			}
		})
	}
}

func TestUserService_ParseToken_InvalidSignature(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	// создаём токен с другим секретом
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 1,
		"email":   "test@example.com",
		"roles":   []string{"user"},
	})

	// подпишем другим ключом
	otherSecret := []byte("another-secret")
	signed, err := token.SignedString(otherSecret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	if _, err := svc.ParseToken(signed); err == nil {
		t.Fatalf("expected error on token with invalid signature, got nil")
	}
}

func TestUserService_GetCurrentUser_Success(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	u, err := svc.GetCurrentUser(1)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if u == nil {
		t.Fatalf("expected user, got nil")
	}
	if u.ID != 1 {
		t.Errorf("expected ID = 1, got %d", u.ID)
	}
}

func TestUserService_UpdateProfile_Success(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	if _, err := svc.GetUser(1); err != nil {
		t.Fatalf("expected user 1 to exist, got error: %v", err)
	}

	req := model.UpdateProfileRequest{
		Name: "Updated Name",
	}

	u, err := svc.UpdateProfile(1, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if u.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, u.Name)
	}
}

func TestUserService_UpdateProfile_MissingName(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	req := model.UpdateProfileRequest{
		Name: "",
	}

	u, err := svc.UpdateProfile(1, req)
	if err != model.ErrMissingRequiredFields {
		t.Fatalf("expected ErrMissingRequiredFields, got: %v", err)
	}
	if u != nil {
		t.Fatalf("expected nil user on error, got: %+v", u)
	}
}

func TestUserService_UpdateProfile_UserNotFound(t *testing.T) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)

	req := model.UpdateProfileRequest{
		Name: "Ghost",
	}

	u, err := svc.UpdateProfile(999, req)
	if err != model.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}
	if u != nil {
		t.Fatalf("expected nil user on error, got: %+v", u)
	}
}