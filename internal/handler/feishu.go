package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"testmind/internal/config"
	"testmind/internal/repository"
	"testmind/pkg/response"
)

type FeishuHandler struct {
	cfg *config.Config
	db  *repository.DB
}

func NewFeishuHandler(cfg *config.Config, db *repository.DB) *FeishuHandler {
	return &FeishuHandler{cfg: cfg, db: db}
}

// OAuthConfig 飞书OAuth配置
type OAuthConfig struct {
	AppID       string
	AppSecret   string
	RedirectURI string
}

// GetOAuthConfig 获取OAuth配置
func (h *FeishuHandler) GetOAuthConfig(c *gin.Context) {
	config := gin.H{
		"app_id":       h.cfg.Feishu.AppID,
		"redirect_uri": h.cfg.Feishu.RedirectURI,
		"auth_url":     fmt.Sprintf("https://open.feishu.cn/open-apis/authen/v1/authorize?app_id=%s&redirect_uri=%s&state=%s",
			h.cfg.Feishu.AppID, url.QueryEscape(h.cfg.Feishu.RedirectURI), generateState()),
	}

	response.Success(c, config)
}

// OAuthAuthorize 发起OAuth授权
func (h *FeishuHandler) OAuthAuthorize(c *gin.Context) {
	state := generateState()
	
	// 存储state用于验证
	userID := c.GetString("user_id")
	stateKey := uuid.New().String()[:32]
	
	query := `
		INSERT INTO oauth_states (state_id, user_id, state_code, created_at, expires_at)
		VALUES ($1, $2, $3, NOW(), NOW() + INTERVAL '10 minutes')
	`
	h.db.ExecContext(c.Request.Context(), query, stateKey, userID, state)

	authURL := fmt.Sprintf(
		"https://open.feishu.cn/open-apis/authen/v1/authorize?app_id=%s&redirect_uri=%s&state=%s",
		h.cfg.Feishu.AppID,
		url.QueryEscape(h.cfg.Feishu.RedirectURI),
		state,
	)

	response.Success(c, gin.H{
		"auth_url": authURL,
		"state":    state,
		"expires":  "10 minutes",
	})
}

// OAuthCallback 处理OAuth回调
func (h *FeishuHandler) OAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		response.BadRequest(c, "缺少授权码或状态", "Missing code or state")
		return
	}

	// 验证state
	var stateRecord struct {
		UserID   string
		StateID  string
		ExpiresAt time.Time
	}
	
	err := h.db.GetContext(c.Request.Context(), &stateRecord,
		`SELECT user_id, state_id, expires_at FROM oauth_states WHERE state_code = $1`, state)
	if err != nil || stateRecord.ExpiresAt.Before(time.Now()) {
		response.BadRequest(c, "授权状态无效或已过期", "Invalid or expired state")
		return
	}

	// 使用code换取access_token
	tokenResp, err := h.exchangeCodeForToken(code)
	if err != nil {
		response.InternalError(c, "获取飞书Token失败", "Failed to get Feishu token")
		return
	}

	// 存储用户飞书凭证
	userID := stateRecord.UserID
	credentialID := uuid.New().String()[:32]
	
	query := `
		INSERT INTO feishu_credentials (credential_id, user_id, access_token, refresh_token, 
			expires_at, feishu_user_id, feishu_name, feishu_avatar, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`
	h.db.ExecContext(c.Request.Context(), query,
		credentialID, userID, tokenResp.AccessToken, tokenResp.RefreshToken,
		time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		tokenResp.FeishuUserID, tokenResp.Name, tokenResp.AvatarURL)

	// 删除已使用的state
	h.db.ExecContext(c.Request.Context(),
		`DELETE FROM oauth_states WHERE state_code = $1`, state)

	response.Success(c, gin.H{
		"status":           "authorized",
		"feishu_user_id":   tokenResp.FeishuUserID,
		"feishu_name":      tokenResp.Name,
		"expires_at":       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	})
}

// RefreshToken 刷新飞书Token
func (h *FeishuHandler) RefreshToken(c *gin.Context) {
	userID := c.GetString("user_id")

	// 获取当前凭证
	var cred struct {
		RefreshToken string
		CredentialID string
	}
	
	err := h.db.GetContext(c.Request.Context(), &cred,
		`SELECT credential_id, refresh_token FROM feishu_credentials WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`,
		userID)
	if err != nil {
		response.NotFound(c, "未找到飞书授权信息", "Feishu credential not found")
		return
	}

	// 刷新token
	tokenResp, err := h.refreshFeishuToken(cred.RefreshToken)
	if err != nil {
		response.InternalError(c, "刷新飞书Token失败", "Failed to refresh Feishu token")
		return
	}

	// 更新凭证
	h.db.ExecContext(c.Request.Context(),
		`UPDATE feishu_credentials SET access_token = $1, expires_at = $2, updated_at = NOW() WHERE credential_id = $3`,
		tokenResp.AccessToken, time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second), cred.CredentialID)

	response.Success(c, gin.H{
		"status":     "refreshed",
		"expires_at": time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	})
}

// GetCredential 获取飞书授权状态
func (h *FeishuHandler) GetCredential(c *gin.Context) {
	userID := c.GetString("user_id")

	var cred struct {
		CredentialID   string    `json:"credential_id"`
		FeishuUserID   string    `json:"feishu_user_id"`
		FeishuName     string    `json:"feishu_name"`
		FeishuAvatar   string    `json:"feishu_avatar"`
		ExpiresAt      time.Time `json:"expires_at"`
		CreatedAt      time.Time `json:"created_at"`
	}

	err := h.db.GetContext(c.Request.Context(), &cred,
		`SELECT credential_id, feishu_user_id, feishu_name, feishu_avatar, expires_at, created_at 
		FROM feishu_credentials WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`,
		userID)
	if err != nil {
		response.Success(c, gin.H{
			"authorized": false,
			"message":    "未授权飞书",
		})
		return
	}

	isExpired := cred.ExpiresAt.Before(time.Now())

	response.Success(c, gin.H{
		"authorized":     !isExpired,
		"credential_id":  cred.CredentialID,
		"feishu_user_id": cred.FeishuUserID,
		"feishu_name":    cred.FeishuName,
		"feishu_avatar":  cred.FeishuAvatar,
		"expires_at":     cred.ExpiresAt,
		"is_expired":     isExpired,
	})
}

// RevokeAuth 撤销飞书授权
func (h *FeishuHandler) RevokeAuth(c *gin.Context) {
	userID := c.GetString("user_id")

	result, err := h.db.ExecContext(c.Request.Context(),
		`DELETE FROM feishu_credentials WHERE user_id = $1`, userID)
	if err != nil {
		response.InternalError(c, "撤销授权失败", "Failed to revoke authorization")
		return
	}

	rows, _ := result.RowsAffected()

	response.Success(c, gin.H{
		"revoked":    true,
		"deleted":    rows,
		"user_id":    userID,
	})
}

// FetchDocument 获取飞书文档内容（用户授权后）
func (h *FeishuHandler) FetchDocument(c *gin.Context) {
	userID := c.GetString("user_id")
	docID := c.Param("doc_id")

	// 获取用户凭证
	var accessToken string
	err := h.db.GetContext(c.Request.Context(), &accessToken,
		`SELECT access_token FROM feishu_credentials WHERE user_id = $1 AND expires_at > NOW() ORDER BY created_at DESC LIMIT 1`,
		userID)
	if err != nil {
		response.BadRequest(c, "请先授权飞书", "Please authorize Feishu first")
		return
	}

	// 调用飞书API获取文档内容
	content, err := h.fetchFeishuDocument(accessToken, docID)
	if err != nil {
		response.InternalError(c, "获取飞书文档失败", "Failed to fetch Feishu document")
		return
	}

	response.Success(c, gin.H{
		"doc_id":  docID,
		"content": content,
	})
}

// exchangeCodeForToken 使用code换取access_token
func (h *FeishuHandler) exchangeCodeForToken(code string) (*TokenResponse, error) {
	apiURL := "https://open.feishu.cn/open-apis/authen/v1/access_token"

	reqBody := gin.H{
		"app_id":     h.cfg.Feishu.AppID,
		"app_secret": h.cfg.Feishu.AppSecret,
		"grant_type": "authorization_code",
		"code":       code,
	}

	resp, err := h.callFeishuAPI(apiURL, reqBody)
	if err != nil {
		return nil, err
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(resp, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// refreshFeishuToken 刷新token
func (h *FeishuHandler) refreshFeishuToken(refreshToken string) (*TokenResponse, error) {
	apiURL := "https://open.feishu.cn/open-apis/authen/v1/refresh_access_token"

	reqBody := gin.H{
		"app_id":        h.cfg.Feishu.AppID,
		"app_secret":    h.cfg.Feishu.AppSecret,
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}

	resp, err := h.callFeishuAPI(apiURL, reqBody)
	if err != nil {
		return nil, err
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(resp, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// fetchFeishuDocument 获取飞书文档内容
func (h *FeishuHandler) fetchFeishuDocument(accessToken, docID string) (string, error) {
	// TODO: 实际调用飞书API
	// 这里返回模拟内容
	return "飞书文档内容（待实现）", nil
}

// callFeishuAPI 调用飞书API
func (h *FeishuHandler) callFeishuAPI(apiURL string, reqBody gin.H) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// TokenResponse 飞书Token响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	FeishuUserID string `json:"open_id"`
	Name         string `json:"name"`
	AvatarURL    string `json:"avatar_url"`
}

// generateState 生成随机state
func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}