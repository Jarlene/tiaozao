package service

import (
	stdErrors "errors"
	"regexp"
	"strings"
	"time"

	"flea-market/internal/config"
	"flea-market/internal/model"
	"flea-market/internal/repository"
	"flea-market/internal/storage"
	"flea-market/pkg/errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const DefaultInitialBalance int64 = 1000 // 新用户默认赠送 10.00 元（单位：分，开发环境默认值，上线前调整为更小值或 0）
var (
	reUpper  = regexp.MustCompile(`[A-Z]`)
	reLower  = regexp.MustCompile(`[a-z]`)
	reDigit  = regexp.MustCompile(`[0-9]`)
	reEmail  = regexp.MustCompile(`^([^@]+)`)
)

type AuthService struct {
	userRepo repository.UserRepository
	minio    *storage.MinIOClient // TODO: 预留，后续用于头像上传功能
	cfg      *config.JWTConfig
}

func NewAuthService(userRepo repository.UserRepository, minio *storage.MinIOClient, cfg *config.JWTConfig) *AuthService {
	return &AuthService{userRepo: userRepo, minio: minio, cfg: cfg}
}

type RegisterReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type UserProfile struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Role      int    `json:"role"` // 1=user, 2=admin
	CreatedAt string `json:"created_at"`
}

type UpdateProfileReq struct {
	Nickname string `json:"nickname" binding:"max=100"`
}

func (s *AuthService) Register(req *RegisterReq) (*TokenPair, *UserProfile, int, error) {
	// 验证密码强度
	if !isValidPassword(req.Password) {
		return nil, nil, errors.ErrWeakPassword, nil
	}

	// 哈希密码
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, errors.ErrInternal, err
	}

	user := &model.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		Nickname:     extractNickname(req.Email),
		Status:       1,
		Balance:      DefaultInitialBalance,
	}

	if err := s.userRepo.Create(user); err != nil {
		// 检测唯一约束冲突（并行注册时的竞态处理）
		if isDuplicateKeyError(err) {
			return nil, nil, errors.ErrEmailExists, nil
		}
		return nil, nil, errors.ErrInternal, err
	}

	tokens, err := s.generateTokenPair(user)
	if err != nil {
		return nil, nil, errors.ErrInternal, err
	}

	profile := &UserProfile{
		ID:            user.ID,
		Email:         user.Email,
		Nickname:      user.Nickname,
		AvatarURL:     user.AvatarURL,
		Role:          user.Role,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
	}

	return tokens, profile, errors.Success, nil
}

func (s *AuthService) Login(req *LoginReq) (*TokenPair, *UserProfile, int, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.ErrInvalidCredentials, nil
		}
		return nil, nil, errors.ErrInternal, err
	}

	if user.Status == 0 {
		return nil, nil, errors.ErrInvalidCredentials, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, nil, errors.ErrInvalidCredentials, nil
	}

	tokens, err := s.generateTokenPair(user)
	if err != nil {
		return nil, nil, errors.ErrInternal, err
	}

	profile := &UserProfile{
		ID:            user.ID,
		Email:         user.Email,
		Nickname:      user.Nickname,
		AvatarURL:     user.AvatarURL,
		Role:          user.Role,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
	}

	return tokens, profile, errors.Success, nil
}

func (s *AuthService) Refresh(refreshToken string) (*TokenPair, int, error) {
	claims := &RefreshClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.ErrInvalidRefresh, nil
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, errors.ErrInvalidRefresh, nil
	}

	tokens, err := s.generateTokenPair(user)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	return tokens, errors.Success, nil
}

func (s *AuthService) GetProfile(userID uint) (*UserProfile, int, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	return &UserProfile{
		ID:            user.ID,
		Email:         user.Email,
		Nickname:      user.Nickname,
		AvatarURL:     user.AvatarURL,
		Role:          user.Role,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
	}, errors.Success, nil
}

func (s *AuthService) UpdateProfile(userID uint, req *UpdateProfileReq) (*UserProfile, int, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.ErrInternal, err
	}

	return &UserProfile{
		ID:            user.ID,
		Email:         user.Email,
		Nickname:      user.Nickname,
		AvatarURL:     user.AvatarURL,
		Role:          user.Role,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
	}, errors.Success, nil
}

type AccessClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID  uint   `json:"user_id"`
	TokenID string `json:"token_id"`
	jwt.RegisteredClaims
}

func (s *AuthService) generateTokenPair(user *model.User) (*TokenPair, error) {
	now := time.Now()

	// Access Token
	accessClaims := &AccessClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   "access",
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return nil, err
	}

	// Refresh Token
	refreshClaims := &RefreshClaims{
		UserID:  user.ID,
		TokenID: uuid.New().String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   "refresh",
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.AccessTokenTTL.Seconds()),
	}, nil
}

// 密码策略：最小8位，至少包含1个大写字母、1个小写字母、1个数字
func isValidPassword(password string) bool {
	if len(password) < 8 || len(password) > 72 {
		return false
	}
	return reUpper.MatchString(password) && reLower.MatchString(password) && reDigit.MatchString(password)
}

// 从邮箱提取昵称（@前面部分）
func extractNickname(email string) string {
	matches := reEmail.FindStringSubmatch(email)
	if len(matches) > 1 {
		return matches[1]
	}
	return "用户"
}

// isDuplicateKeyError 检测 PostgreSQL 唯一约束冲突错误
// 在并行注册时，FindByEmail 检查可能通过但 Create 时仍会触发唯一约束
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// GORM v1.25.10+ 提供了 ErrDuplicatedKey 哨兵错误
	if stdErrors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	// 兼容旧版本和 pgx 驱动的原生错误
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint") || strings.Contains(msg, "23505")
}
