package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"testmind/internal/config"
	"testmind/internal/model"
	"testmind/pkg/jwt"
	"testmind/pkg/response"
	"testmind/pkg/validator"
)

type UserHandler struct {
	jwtManager *jwt.JWTManager
}

func NewUserHandler(cfg *config.Config) *UserHandler {
	return &UserHandler{
		jwtManager: jwt.NewJWTManager(
			cfg.JWT.Secret,
			cfg.JWT.AccessTokenExp,
			cfg.JWT.RefreshTokenExp,
		),
	}
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req validator.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	if err := validator.Validate(req); err != nil {
		zhMsg, enMsg := validator.GetErrorMsg(c.GetHeader("Accept-Language"), err)
		response.BadRequest(c, zhMsg, enMsg)
		return
	}

	// 检查用户名是否已存在
	// TODO: 查询数据库验证用户名唯一性

	// 密码加密
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.InternalError(c, "密码加密失败", "Password encryption failed")
		return
	}

	// 创建用户
	user := &model.User{
		UserID:        uuid.New().String()[:32],
		Username:      req.Username,
		PasswordHash:  string(passwordHash),
		Email:         req.Email,
		DisplayName:   req.DisplayName,
		Language:      req.Language,
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// TODO: 保存到数据库

	// 生成Token
	tokenPair, err := h.jwtManager.GenerateTokenPair(user.UserID, user.Username, "", "")
	if err != nil {
		response.InternalError(c, "生成令牌失败", "Token generation failed")
		return
	}

	response.Success(c, gin.H{
		"user":  user,
		"token": tokenPair,
	})
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req validator.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	if err := validator.Validate(req); err != nil {
		zhMsg, enMsg := validator.GetErrorMsg(c.GetHeader("Accept-Language"), err)
		response.BadRequest(c, zhMsg, enMsg)
		return
	}

	// TODO: 查询数据库验证用户名密码
	// 模拟用户数据
	user := &model.User{
		UserID:      "test_user_001",
		Username:    req.Username,
		Email:      "test@example.com",
		DisplayName: "测试用户",
		Language:    "zh-CN",
		Status:      "active",
		LastLoginAt: time.Now(),
	}

	// 生成Token
	tokenPair, err := h.jwtManager.GenerateTokenPair(user.UserID, user.Username, "", "")
	if err != nil {
		response.InternalError(c, "生成令牌失败", "Token generation failed")
		return
	}

	// TODO: 更新最后登录时间

	response.Success(c, gin.H{
		"user":  user,
		"token": tokenPair,
	})
}

// Refresh 刷新Token
func (h *UserHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "缺少刷新令牌", "Refresh token is required")
		return
	}

	tokenPair, err := h.jwtManager.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c)
		return
	}

	response.Success(c, tokenPair)
}

// GetProfile 获取用户信息
func (h *UserHandler) GetProfile(c *gin.Context) {
	// 从上下文获取用户ID（由中间件设置）
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c)
		return
	}

	// TODO: 从数据库查询用户信息
	// 模拟数据
	user := &model.User{
		UserID:      userID.(string),
		Username:    "test_user",
		Email:       "test@example.com",
		DisplayName: "测试用户",
		Language:    "zh-CN",
		Status:      "active",
		CreatedAt:   time.Now().AddDate(0, -1, 0),
	}

	response.Success(c, user)
}

// UpdateProfile 更新用户信息
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c)
		return
	}

	var req struct {
		DisplayName string `json:"display_name" binding:"omitempty,min=2,max=50"`
		AvatarURL   string `json:"avatar_url" binding:"omitempty,url"`
		Language    string `json:"language" binding:"omitempty,oneof=zh-CN en-US"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	// TODO: 更新数据库

	response.Success(c, gin.H{
		"user_id":      userID,
		"display_name": req.DisplayName,
		"avatar_url":   req.AvatarURL,
		"language":     req.Language,
		"updated_at":   time.Now(),
	})
}

// Logout 登出
func (h *UserHandler) Logout(c *gin.Context) {
	// TODO: 将Token加入黑名单（Redis）
	response.Success(c, gin.H{
		"message": "已登出",
		"message_i18n": gin.H{
			"zh-CN": "已登出",
			"en-US": "Logged out",
		},
	})
}