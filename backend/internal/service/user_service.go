package service

import (
	"errors"
	"time"

	"github.com/Jarlene/tiaozao/backend/internal/config"
	"github.com/Jarlene/tiaozao/backend/internal/model"
	"github.com/Jarlene/tiaozao/backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户业务逻辑
type UserService struct {
	repo  *repository.UserRepository
	cfg   *config.JWTConfig
}

// NewUserService 创建用户服务实例
func NewUserService(repo *repository.UserRepository, cfg *config.Config) *UserService {
	return &UserService{repo: repo, cfg: &cfg.JWT}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string      `json:"token"`
	User  model.User  `json:"user"`
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Avatar string `json:"avatar"`
	Bio    string `json:"bio"`
}

// Register 用户注册
func (s *UserService) Register(req *RegisterRequest) (*model.User, error) {
	// 检查用户名是否已存在
	if _, err := s.repo.FindByUsername(req.Username); err == nil {
		return nil, errors.New("用户名已被占用")
	}

	// 检查邮箱是否已存在
	if req.Email != "" {
		if _, err := s.repo.FindByEmail(req.Email); err == nil {
			return nil, errors.New("邮箱已被占用")
		}
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	user := &model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   1,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, errors.New("创建用户失败")
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(req *LoginRequest) (*LoginResponse, error) {
	// 查找用户
	user, err := s.repo.FindByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, errors.New("登录失败")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("账户已被禁用")
	}

	// 生成JWT
	token, err := s.generateToken(user.ID, user.Username)
	if err != nil {
		return nil, errors.New("生成令牌失败")
	}

	return &LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

// GetProfile 获取用户资料
func (s *UserService) GetProfile(userID uint) (*model.User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, errors.New("获取用户信息失败")
	}
	return user, nil
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(userID uint, req *UpdateProfileRequest) (*model.User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}

	if err := s.repo.Update(user); err != nil {
		return nil, errors.New("更新用户信息失败")
	}

	return user, nil
}

// generateToken 生成JWT令牌
func (s *UserService) generateToken(userID uint, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Duration(s.cfg.ExpireHour) * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}
