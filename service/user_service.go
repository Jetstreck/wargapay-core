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
	"gorm.io/gorm"
)

type UserService interface {
	Register(req dto.RegisterRequest) (*dto.UserDetail, error)
	Login(req dto.LoginRequest) (string, *dto.UserDetail, error)
	GetProfile(userID uint) (*dto.UserDetail, error)
	ListUsers() ([]dto.UserDetail, error)
	UpdateUserRole(userID uint, role string) (*dto.UserDetail, error)
}

type userService struct {
	userRepo     repository.UserRepository
	emailService EmailService
	notifSvc     NotificationService
}

func NewUserService(userRepo repository.UserRepository, emailService EmailService, notifSvc ...NotificationService) UserService {
	svc := &userService{
		userRepo:     userRepo,
		emailService: emailService,
	}
	if len(notifSvc) > 0 {
		svc.notifSvc = notifSvc[0]
	}
	return svc
}

func (s *userService) Register(req dto.RegisterRequest) (*dto.UserDetail, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Semua pendaftaran baru selalu berstatus warga ("user").
	// Hak akses menjadi admin hanya dapat diatur oleh pengurus RT (admin terdaftar) via dashboard.
	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "user",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	if s.notifSvc != nil {
		s.notifSvc.NotifyWelcome(user)
	} else {
		go s.emailService.SendNotificationEmail(user.Email, "Selamat Datang di WargaPay", "Akun anda berhasil dibuat.")
	}

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("email atau password salah")
		}
		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", nil, errors.New("email atau password salah")
	}

	// Generate JWT Token
	claims := jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", nil, err
	}

	return tokenString, &dto.UserDetail{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (s *userService) GetProfile(userID uint) (*dto.UserDetail, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	return &dto.UserDetail{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (s *userService) ListUsers() ([]dto.UserDetail, error) {
	users, err := s.userRepo.FindAll()
	if err != nil {
		return nil, err
	}

	var dtos []dto.UserDetail
	for _, u := range users {
		dtos = append(dtos, dto.UserDetail{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
			Role:  u.Role,
		})
	}
	return dtos, nil
}

func (s *userService) UpdateUserRole(userID uint, role string) (*dto.UserDetail, error) {
	if role != "admin" && role != "user" {
		return nil, errors.New("role harus bernilai 'admin' atau 'user'")
	}
	if err := s.userRepo.UpdateRole(userID, role); err != nil {
		return nil, err
	}
	return s.GetProfile(userID)
}
