package service

import (
	"errors"
	"fmt"
	"time"

	"agnos-assessment/internal/config"
	"agnos-assessment/internal/model"
	"agnos-assessment/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrStaffAlreadyExists = errors.New("username already exists in this hospital")
	ErrInvalidCredentials = errors.New("invalid username, password, or hospital")
)

type JWTClaims struct {
	StaffID  string `json:"staff_id"`
	Username string `json:"username"`
	Hospital string `json:"hospital"`
	jwt.RegisteredClaims
}

type AuthService interface {
	CreateStaff(req *model.StaffCreateRequest) (*model.Staff, error)
	Login(req *model.StaffLoginRequest) (string, *model.Staff, error)
	ValidateToken(tokenStr string) (*JWTClaims, error)
}

type authService struct {
	staffRepo repository.StaffRepository
	cfg       *config.Config
}

func NewAuthService(staffRepo repository.StaffRepository, cfg *config.Config) AuthService {
	return &authService{
		staffRepo: staffRepo,
		cfg:       cfg,
	}
}

func (s *authService) CreateStaff(req *model.StaffCreateRequest) (*model.Staff, error) {
	// Check if username already exists in this hospital
	existing, err := s.staffRepo.FindByUsernameAndHospital(req.Username, req.Hospital)
	if err != nil {
		return nil, fmt.Errorf("failed to query staff: %w", err)
	}
	if existing != nil {
		return nil, ErrStaffAlreadyExists
	}

	// Hash password with bcrypt cost 12
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newStaff := &model.Staff{
		ID:           uuid.New(),
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		Hospital:     req.Hospital,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.staffRepo.Create(newStaff); err != nil {
		return nil, fmt.Errorf("failed to persist staff: %w", err)
	}

	return newStaff, nil
}

func (s *authService) Login(req *model.StaffLoginRequest) (string, *model.Staff, error) {
	staff, err := s.staffRepo.FindByUsernameAndHospital(req.Username, req.Hospital)
	if err != nil {
		return "", nil, fmt.Errorf("database error: %w", err)
	}
	if staff == nil {
		return "", nil, ErrInvalidCredentials
	}

	// Compare bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(req.Password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// Generate JWT
	expireTime := time.Now().Add(time.Duration(s.cfg.JWTExpireHrs) * time.Hour)
	claims := &JWTClaims{
		StaffID:  staff.ID.String(),
		Username: staff.Username,
		Hospital: staff.Hospital,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   staff.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(s.cfg.JWTSecret))

	return tokenStr, staff, nil
}

func (s *authService) ValidateToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, _ := token.Claims.(*JWTClaims)
	return claims, nil
}
