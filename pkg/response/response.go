package response

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	MessageI18n MessageI18n `json:"message_i18n,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type MessageI18n struct {
	ZhCN string `json:"zh-CN"`
	EnUS string `json:"en-US"`
}

type Pagination struct {
	Page     int   `json:"page"`
PageSize  int   `json:"page_size"`
	Total    int64 `json:"total"`
	HasMore  bool  `json:"has_more"`
}

// 成功响应
func Success(c *gin.Context, data interface{}) {
	lang := c.GetHeader("Accept-Language")
	msgZh := "成功"
	msgEn := "Success"

	resp := Response{
		Code:    0,
		Message: getMessage(lang, msgZh, msgEn),
		MessageI18n: MessageI18n{
			ZhCN: msgZh,
			EnUS: msgEn,
		},
		Data: data,
	}

	c.JSON(http.StatusOK, resp)
}

// 成功响应带分页
func SuccessWithPagination(c *gin.Context, data interface{}, page, pageSize int, total int64) {
	lang := c.GetHeader("Accept-Language")
	msgZh := "成功"
	msgEn := "Success"

	resp := Response{
		Code:    0,
		Message: getMessage(lang, msgZh, msgEn),
		MessageI18n: MessageI18n{
			ZhCN: msgZh,
			EnUS: msgEn,
		},
		Data: data,
		Pagination: &Pagination{
			Page:     page,
		PageSize: pageSize,
			Total:    total,
			HasMore:  int64(page * pageSize) < total,
		},
	}

	c.JSON(http.StatusOK, resp)
}

// 错误响应
func Error(c *gin.Context, code int, msgZh, msgEn string) {
	lang := c.GetHeader("Accept-Language")

	resp := Response{
		Code:    code,
		Message: getMessage(lang, msgZh, msgEn),
		MessageI18n: MessageI18n{
			ZhCN: msgZh,
			EnUS: msgEn,
		},
	}

	c.JSON(http.StatusOK, resp)
}

// 参数错误
func BadRequest(c *gin.Context, msgZh, msgEn string) {
	Error(c, 1001, msgZh, msgEn)
}

// 未授权
func Unauthorized(c *gin.Context) {
	Error(c, 1002, "未授权访问", "Unauthorized")
}

// 权限不足
func Forbidden(c *gin.Context) {
	Error(c, 1003, "权限不足", "Forbidden")
}

// 资源不存在
func NotFound(c *gin.Context, msgZh, msgEn string) {
	Error(c, 1004, msgZh, msgEn)
}

// 内部错误
func InternalError(c *gin.Context, msgZh, msgEn string) {
	Error(c, 1005, msgZh, msgEn)
}

func getMessage(lang, zh, en string) string {
	if lang == "en-US" || lang == "en" {
		return en
	}
	return zh
}

// 通用错误码
const (
	CodeSuccess           = 0
	CodeBadRequest        = 1001
	CodeUnauthorized      = 1002
	CodeForbidden         = 1003
	CodeNotFound          = 1004
	CodeInternalError     = 1005
	
	// 用户服务错误码 2000-2999
	CodeUserExists        = 2001
	CodeUserNotFound      = 2002
	CodePasswordInvalid   = 2003
	CodeUserDisabled      = 2004
	
	// 测试用例服务错误码 3000-3999
	CodeCaseNotFound      = 3001
	CodeCaseVersionError  = 3002
	CodeCaseReviewError   = 3003
	
	// 执行服务错误码 4000-4999
	CodeExecutionNotFound = 4001
	CodeExecutionError    = 4002
	
	// 缺陷服务错误码 5000-5999
	CodeDefectNotFound    = 5001
	CodeDefectStatusError = 5002
	
	// 报告服务错误码 6000-6999
	CodeReportNotFound    = 6001
	CodeReportGenError    = 6002
	
	// AI服务错误码 7000-7999
	CodeAIServiceError    = 7001
	CodeAITimeout         = 7002
	CodeAIQuotaExceeded   = 7003
	
	// 飞书服务错误码 8000-8999
	CodeFeishuAuthError   = 8001
	CodeFeishuAPIError    = 8002
	CodeFeishuDocNotFound = 8003
)

// MarshalJSON 自定义JSON序列化
func (r *Response) MarshalJSON() ([]byte, error) {
	// 如果 MessageI18n 为空，不返回
	if r.MessageI18n.ZhCN == "" && r.MessageI18n.EnUS == "" {
		type Alias Response
		return json.Marshal(&struct{ *Alias }{Alias: (*Alias)(r)})
	}
	return json.Marshal(r)
}