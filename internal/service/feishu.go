package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// FeishuService 飞书API服务
type FeishuService struct {
	appID       string
	appSecret   string
	httpClient  *http.Client
	baseURL     string
}

// NewFeishuService 创建飞书服务
func NewFeishuService(appID, appSecret string) *FeishuService {
	return &FeishuService{
		appID:     appID,
		appSecret: appSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://open.feishu.cn/open-apis",
	}
}

// ============================================
// 应用级 Token（tenant_access_token）
// ============================================

// GetTenantAccessToken 获取应用级Token
func (s *FeishuService) GetTenantAccessToken() (string, int, error) {
	url := s.baseURL + "/auth/v3/tenant_access_token/internal"

	body := map[string]string{
		"app_id":     s.appID,
		"app_secret": s.appSecret,
	}

	resp, err := s.post(url, body, "")
	if err != nil {
		return "", 0, err
	}

	var result struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return "", 0, err
	}

	if result.Code != 0 {
		return "", 0, fmt.Errorf("feishu API error: %d - %s", result.Code, result.Msg)
	}

	return result.TenantAccessToken, result.Expire, nil
}

// ============================================
// 用户级 Token（user_access_token）
// ============================================

// UserAccessTokenResponse 用户Token响应
type UserAccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshExpiresIn int `json:"refresh_expires_in"`
	OpenID       string `json:"open_id"`
	Name         string `json:"name"`
	AvatarURL    string `json:"avatar_url"`
	Email        string `json:"email"`
	UserID       string `json:"user_id"`
}

// GetUserAccessToken 使用code换取用户Token
func (s *FeishuService) GetUserAccessToken(code string) (*UserAccessTokenResponse, error) {
	// 先获取tenant_access_token
	tenantToken, _, err := s.GetTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant token: %w", err)
	}

	url := s.baseURL + "/authen/v1/oidc/access_token"

	body := map[string]string{
		"grant_type": "authorization_code",
		"code":       code,
	}

	resp, err := s.post(url, body, tenantToken)
	if err != nil {
		return nil, err
	}

	var result struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data UserAccessTokenResponse `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("feishu API error: %d - %s", result.Code, result.Msg)
	}

	return &result.Data, nil
}

// RefreshUserAccessToken 刷新用户Token
func (s *FeishuService) RefreshUserAccessToken(refreshToken string) (*UserAccessTokenResponse, error) {
	tenantToken, _, err := s.GetTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant token: %w", err)
	}

	url := s.baseURL + "/authen/v1/oidc/refresh_access_token"

	body := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}

	resp, err := s.post(url, body, tenantToken)
	if err != nil {
		return nil, err
	}

	var result struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data UserAccessTokenResponse `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("feishu API error: %d - %s", result.Code, result.Msg)
	}

	return &result.Data, nil
}

// ============================================
// 飞书文档 API
// ============================================

// DocumentContent 文档内容
type DocumentContent struct {
	Code    int                    `json:"code"`
	Msg     string                 `json:"msg"`
	Data    map[string]interface{} `json:"data"`
}

// GetDocumentContent 获取文档内容
func (s *FeishuService) GetDocumentContent(userToken, documentID string) (*DocumentContent, error) {
	url := fmt.Sprintf("%s/docx/v1/documents/%s/raw_content", s.baseURL, documentID)

	resp, err := s.get(url, userToken)
	if err != nil {
		return nil, err
	}

	var result DocumentContent
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("feishu API error: %d - %s", result.Code, result.Msg)
	}

	return &result, nil
}

// GetDocumentBlocks 获取文档块内容
func (s *FeishuService) GetDocumentBlocks(userToken, documentID string, pageSize int, pageToken string) (*DocumentContent, error) {
	url := fmt.Sprintf("%s/docx/v1/documents/%s/blocks?page_size=%d", s.baseURL, documentID, pageSize)
	if pageToken != "" {
		url += "&page_token=" + pageToken
	}

	resp, err := s.get(url, userToken)
	if err != nil {
		return nil, err
	}

	var result DocumentContent
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ============================================
// 飞书多维表格 API
// ============================================

// BitableRecord 多维表格记录
type BitableRecord struct {
	RecordID string                 `json:"record_id"`
	Fields   map[string]interface{} `json:"fields"`
}

// CreateBitableRecord 创建多维表格记录
func (s *FeishuService) CreateBitableRecord(userToken, appToken, tableID string, fields map[string]interface{}) (*BitableRecord, error) {
	url := fmt.Sprintf("%s/bitable/v1/apps/%s/tables/%s/records", s.baseURL, appToken, tableID)

	body := map[string]interface{}{
		"fields": fields,
	}

	resp, err := s.post(url, body, userToken)
	if err != nil {
		return nil, err
	}

	var result struct {
		Code int          `json:"code"`
		Msg  string       `json:"msg"`
		Data struct {
			Record BitableRecord `json:"record"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("feishu API error: %d - %s", result.Code, result.Msg)
	}

	return &result.Data.Record, nil
}

// BatchCreateBitableRecords 批量创建多维表格记录
func (s *FeishuService) BatchCreateBitableRecords(userToken, appToken, tableID string, records []map[string]interface{}) (int, error) {
	url := fmt.Sprintf("%s/bitable/v1/apps/%s/tables/%s/records/batch_create", s.baseURL, appToken, tableID)

	var recordsData []map[string]interface{}
	for _, fields := range records {
		recordsData = append(recordsData, map[string]interface{}{"fields": fields})
	}

	body := map[string]interface{}{
		"records": recordsData,
	}

	resp, err := s.post(url, body, userToken)
	if err != nil {
		return 0, err
	}

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Records []BitableRecord `json:"records"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, err
	}

	if result.Code != 0 {
		return 0, fmt.Errorf("feishu API error: %d - %s", result.Code, result.Msg)
	}

	return len(result.Data.Records), nil
}

// ============================================
// 飞书云文档 API（创建文档）
// ============================================

// CreateDocument 创建飞书云文档
func (s *FeishuService) CreateDocument(userToken, title, content string) (string, string, error) {
	url := s.baseURL + "/docx/v1/documents"

	body := map[string]interface{}{
		"title": title,
	}

	resp, err := s.post(url, body, userToken)
	if err != nil {
		return "", "", err
	}

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Document struct {
				DocumentID string `json:"document_id"`
				Title      string `json:"title"`
			} `json:"document"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return "", "", err
	}

	if result.Code != 0 {
		return "", "", fmt.Errorf("feishu API error: %d - %s", result.Code, result.Msg)
	}

	docID := result.Data.Document.DocumentID

	// 写入内容
	if content != "" {
		blockURL := fmt.Sprintf("%s/docx/v1/documents/%s/blocks/%s/children", s.baseURL, docID, docID)
		blocks := s.textToBlocks(content)
		blockBody := map[string]interface{}{
			"children": blocks,
		}
		s.post(blockURL, blockBody, userToken)
	}

	docURL := fmt.Sprintf("https://www.feishu.cn/docx/%s", docID)
	return docID, docURL, nil
}

// ============================================
// 用户信息 API
// ============================================

// GetUserInfo 获取用户信息
func (s *FeishuService) GetUserInfo(userToken string) (map[string]interface{}, error) {
	url := s.baseURL + "/authen/v1/user_info"

	resp, err := s.get(url, userToken)
	if err != nil {
		return nil, err
	}

	var result struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("feishu API error: %d - %s", result.Code, result.Msg)
	}

	return result.Data, nil
}

// ============================================
// 辅助方法
// ============================================

func (s *FeishuService) post(url string, body interface{}, token string) ([]byte, error) {
	bodyBytes, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (s *FeishuService) get(url string, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// textToBlocks 将文本转换为飞书文档块
func (s *FeishuService) textToBlocks(content string) []map[string]interface{} {
	var blocks []map[string]interface{}

	// 简单实现：将内容作为一个文本块
	blocks = append(blocks, map[string]interface{}{
		"block_type": 2, // text
		"text": map[string]interface{}{
			"elements": []map[string]interface{}{
				{
					"text_run": map[string]interface{}{
						"content": content,
					},
				},
			},
		},
	})

	return blocks
}