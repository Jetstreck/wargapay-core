package service

import (
	"errors"
	"os"
	"time"
	"wargapay-core/dto"
	"wargapay-core/entity"
	"wargapay-core/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(req dto.RegisterRequest) (*dto.UserDetail, error)
	Login(req dto.LoginRequest) (string, *dto.UserDetail, error)
	ListUsers() ([]dto.UserDetail, error)
	UpdateUserRole(id uint, role string) (*dto.UserDetail, error)
}

type userService struct {
	userRepo     repository.UserRepository
	emailService EmailService
}

func NewUserService(userRepo repository.UserRepository, emailService EmailService) UserService {
	return &userService{
		userRepo:     userRepo,
		emailService: emailService,
	}
}

func (s *userService) Register(req dto.RegisterRequest) (*dto.UserDetail, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	role := "user"
	if req.Role != "" {
		role = req.Role
	}

	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	go s.emailService.SendNotificationEmail(user.Email, "Selamat Datang di WargaPay", "Akun anda berhasil dibuat.")

	return &dto.UserDetail{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (s *userService) Login(req dto.LoginRequest) (string, *dto.UserDetail, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return "", nil, errors.New("email atau password salah")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", nil, errors.New("email atau password salah")
	}

	claims := jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", nil, err
	}

	userDetail := &dto.UserDetail{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}

	return tokenString, userDetail, nil
}

func (s *userService) ListUsers() ([]dto.UserDetail, error) {
	users, err := s.userRepo.FindAll()
	if err != nil {
		return nil, err
	}

	var results []dto.UserDetail
	for _, u := range users {
		results = append(results, dto.UserDetail{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
			Role:  u.Role,
		})
	}
	return results, nil
}

func (s *userService) UpdateUserRole(id uint, role string) (*dto.UserDetail, error) {
	if role != "admin" && role != "user" {
		return nil, errors.New("role harus berupa 'admin' atau 'user'")
	}
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	if err := s.userRepo.UpdateRole(id, role); err != nil {
		return nil, err
	}
	user.Role = role
	return &dto.UserDetail{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}
