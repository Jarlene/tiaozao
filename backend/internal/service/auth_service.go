package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jarlene/tiaozao/internal/config"
	"github.com/jarlene/tiaozao/internal/model"
	"github.com/jarlene/tiaozao/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid verification code")
	ErrCodeExpired        = errors.New("verification code expired")
	ErrCodeRecentlySent   = errors.New("verification code sent recently, please wait")
)

type AuthService struct {
	userRepo *repository.UserRepository
	smsRepo  *repository.SmsRepository
	cfg      *config.SMSConfig
}

func NewAuthService(userRepo *repository.UserRepository, smsRepo *repository.SmsRepository, cfg *config.SMSConfig) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		smsRepo:  smsRepo,
		cfg:      cfg,
	}
}

// SendCode generates and "sends" a verification code to the given phone
func (s *AuthService) SendCode(phone string) error {
	// Generate 6-digit code
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	// If mock mode is on, use the configured code
	if s.cfg.MockMode {
		code = s.cfg.MockCode
	}

	sms := &model.SmsCode{
		Phone:    phone,
		Code:     code,
		ExpireAt: time.Now().Add(time.Duration(s.cfg.ExpireMinutes) * time.Minute),
	}

	if err := s.smsRepo.Save(sms); err != nil {
		return fmt.Errorf("failed to save SMS code: %w", err)
	}

	// In production, send via SMS provider (Aliyun, Twilio, etc.)
	// For now, log the code
	fmt.Printf("[SMS] Verification code for %s: %s (expires in %d min)\n",
		phone, code, s.cfg.ExpireMinutes)

	return nil
}

// LoginWithCode verifies the code and returns the authenticated user
func (s *AuthService) LoginWithCode(phone, code string) (*model.User, error) {
	// Find the latest valid code
	sms, err := s.smsRepo.FindLatestValid(phone)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check expiration
	if time.Now().After(sms.ExpireAt) {
		return nil, ErrCodeExpired
	}

	// Verify code
	if sms.Code != code {
		return nil, ErrInvalidCredentials
	}

	// Mark code as used
	if err := s.smsRepo.MarkUsed(sms.ID); err != nil {
		return nil, fmt.Errorf("failed to mark code as used: %w", err)
	}

	// Find or create user
	user, err := s.userRepo.FindByPhone(phone)
	if err != nil {
		// Create new user if not found
		user = &model.User{
			Phone:    phone,
			Nickname: fmt.Sprintf("用户%s", phone[len(phone)-4:]),
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	return user, nil
}
