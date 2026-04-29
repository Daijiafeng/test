package handler

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"testmind/internal/config"
	"testmind/internal/model"
	"testmind/internal/repository"
	"testmind/pkg/jwt"
	"testmind/pkg/response"
	"testmind/pkg/validator"
)

type UserHandler struct {
	jwtManager *jwt.JWTManager
	userRepo   *repository.UserRepository
}

func NewUserHandler(cfg *config.Config, db *repository.DB) *UserHandler {
	return &UserHandler{
		jwtManager: jwt.NewJWTManager(
			cfg.JWT.Secret,
			cfg.JWT.AccessTokenExp,
			cfg.JWT.RefreshTokenExp,
		),
		userRepo: repository.NewUserRepository(db),
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
	existingUser, _ := h.userRepo.FindByUsername(c.Request.Context(), req.Username)
	if existingUser != nil {
		response.BadRequest(c, "用户名已存在", "Username already exists")
		return
	}

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

	if user.DisplayName == "" {
		user.DisplayName = req.Username
	}
	if user.Language == "" {
		user.Language = "zh-CN"
	}

	log.Printf("Registering user: %+v", user)

	// 保存到数据库
	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		log.Printf("Failed to create user: %v", err)
		response.InternalError(c, "注册失败: "+err.Error(), "Failed to register user: "+err.Error())
		return
	}

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

	// 查询数据库验证用户名密码
	user, err := h.userRepo.FindByUsername(c.Request.Context(), req.Username)
	if err != nil {
		log.Printf("Failed to find user: %v", err)
		response.InternalError(c, "登录失败: "+err.Error(), "Failed to login: "+err.Error())
		return
	}
	if user == nil {
		log.Printf("User not found: %s", req.Username)
		response.Unauthorized(c)
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		log.Printf("Password mismatch for user %s", req.Username)
		response.Unauthorized(c)
		return
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		log.Printf("Failed to update last login time: %v", err)
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
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c)
		return
	}

	// 从数据库查询用户信息
	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "查询用户信息失败", "Failed to get user info")
		return
	}
	if user == nil {
		response.NotFound(c, "用户不存在", "User not found")
		return
	}

	response.Success(c, user)
}

// UpdateProfile 更新用户信息
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c)
		return
	}

	var req struct {
		DisplayName string `json:"display_name" binding:"omitempty,min=2,max=50"`
		AvatarURL   string `json:"avatar_url" binding:"omitempty,url"`
		Language    string `json:"language" binding:"omitempty,oneof=zh-CN en-US"`

		// 其他可更新字段
		Phone string `json:"phone" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	// 查询当前用户
	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		response.NotFound(c, "用户不存在", "User not found")
		return
	}

	// 更新字段
	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	if req.Language != "" {
		user.Language = req.Language
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	user.UpdatedAt = time.Now()

	// 保存到数据库
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		response.InternalError(c, "更新失败", "Failed to update profile")
		return
	}

	response.Success(c, user)
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