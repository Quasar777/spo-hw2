package service

import (
	"regexp"
	"service_users/internal/model"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repository UserRepository
	jwtSecret  []byte
	tokenTTL   time.Duration
}

func NewUserService(r UserRepository) *UserService {
	return &UserService{
		repository: r,
		jwtSecret:  []byte("super-secret-key"), // TODO: брать из env
		tokenTTL:   24 * time.Hour,
	}
}

func (s *UserService) GetUser(id int) (*model.User, error) {
	return s.repository.GetByID(id)
}

func (s *UserService) GetAllUsers() ([]model.User, error) {
	return s.repository.GetAll()
}

func (s *UserService) CreateUser(req model.CreateUserRequest) (int, error) {
	if req.Name == "" || req.Email == "" {
		return 0, model.ErrMissingRequiredFields
	}
	if !isEmailValid(req.Email) {
		return 0, model.ErrInvalidEmail
	}

	return s.repository.Create(&req)
}

func (s *UserService) UpdateUser(req model.UpdateUserRequest) error {
	if req.Name == "" {
		return model.ErrMissingRequiredFields
	}

	_, err := s.repository.GetByID(req.ID)
	if err != nil {
		return err
	}

	return s.repository.Update(&req)
}

func (s *UserService) DeleteUser(id int) error {
	return s.repository.Delete(id)
}

func (s *UserService) Register(req model.RegisterRequest) (int, error) {
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return 0, model.ErrMissingRequiredFields
	}
	if !isEmailValid(req.Email) {
		return 0, model.ErrInvalidEmail
	}
	if len(req.Password) < 6 {
		return 0, model.ErrInvalidPassword
	}

	if _, err := s.repository.GetByEmail(req.Email); err == nil {
		return 0, model.ErrUniqueEmailConflict
	} else if err != model.ErrUserNotFound {
		return 0, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	createReq := model.CreateUserRequest{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hashed),
		Roles:        []string{"user"},
	}

	return s.repository.Create(&createReq)
}

type userClaims struct {
	UserID int      `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

func (s *UserService) Login(req model.LoginRequest) (string, error) {
	if req.Email == "" || req.Password == "" {
		return "", model.ErrMissingRequiredFields
	}

	user, err := s.repository.GetByEmail(req.Email)
	if err != nil {
		if err == model.ErrUserNotFound {
			return "", model.ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", model.ErrInvalidCredentials
	}

	now := time.Now()
	claims := userClaims{
		UserID: user.ID,
		Email:  user.Email,
		Roles:  user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(user.ID),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return signed, nil
}

// Helpers

func isEmailValid(e string) bool {
    emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
    return emailRegex.MatchString(e)
}