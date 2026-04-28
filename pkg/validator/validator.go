package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterValidation("username", validateUsername)
	validate.RegisterValidation("password", validatePassword)
}

// Validate 验证结构体
func Validate(s interface{}) error {
	return validate.Struct(s)
}

// ValidateVar 验证变量
func ValidateVar(field interface{}, tag string) error {
	return validate.Var(field, tag)
}

// 用户名验证：字母开头，允许字母数字下划线，长度4-20
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	matched, _ := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_]{3,19}$", username)
	return matched
}

// 密码验证：至少8位，包含字母和数字
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}
	hasLetter := regexp.MustCompile("[a-zA-Z]").MatchString(password)
	hasNumber := regexp.MustCompile("[0-9]").MatchString(password)
	return hasLetter && hasNumber
}

// 注册请求验证
type RegisterRequest struct {
	Username    string `json:"username" validate:"required,username"`
	Password    string `json:"password" validate:"required,password"`
	Email       string `json:"email" validate:"required,email"`
	DisplayName string `json:"display_name" validate:"required,min=2,max=50"`
	Language    string `json:"language" validate:"omitempty,oneof=zh-CN en-US"`
}

// 登录请求验证
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// 创建组织请求验证
type CreateOrganizationRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	NameEn      string `json:"name_en" validate:"omitempty,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

// 创建项目请求验证
type CreateProjectRequest struct {
	OrgID       string `json:"org_id" validate:"required"`
	Name        string `json:"name" validate:"required,min=2,max=100"`
	NameEn      string `json:"name_en" validate:"omitempty,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
	Language    string `json:"language" validate:"omitempty,oneof=zh-CN en-US"`
}

// 创建测试用例请求验证
type CreateTestCaseRequest struct {
	ProjectID     string   `json:"project_id" validate:"required"`
	ModuleID      string   `json:"module_id" validate:"omitempty"`
	Title         string   `json:"title" validate:"required,min=5,max=500"`
	TitleEn       string   `json:"title_en" validate:"omitempty,max=500"`
	Priority      string   `json:"priority" validate:"required,oneof=P0 P1 P2 P3"`
	Precondition  string   `json:"precondition" validate:"omitempty"`
	Steps         []Step   `json:"steps" validate:"required,min=1,dive"`
	Tags          []string `json:"tags" validate:"omitempty"`
	Language      string   `json:"language" validate:"omitempty,oneof=zh-CN en-US"`
	CustomFields  map[string]interface{} `json:"custom_fields" validate:"omitempty"`
}

// 测试步骤验证
type Step struct {
	StepNum    int    `json:"step_num" validate:"required,min=1"`
	Action     string `json:"action" validate:"required,min=1"`
	ActionEn   string `json:"action_en" validate:"omitempty"`
	Expected   string `json:"expected" validate:"required,min=1"`
	ExpectedEn string `json:"expected_en" validate:"omitempty"`
}

// 创建缺陷请求验证
type CreateDefectRequest struct {
	ProjectID       string   `json:"project_id" validate:"required"`
	ModuleID        string   `json:"module_id" validate:"omitempty"`
	Title           string   `json:"title" validate:"required,min=5,max=500"`
	TitleEn         string   `json:"title_en" validate:"omitempty,max=500"`
	Description     string   `json:"description" validate:"required"`
	ReproduceSteps  string   `json:"reproduce_steps" validate:"omitempty"`
	Severity        string   `json:"severity" validate:"required,oneof=critical major general minor trivial"`
	Priority        string   `json:"priority" validate:"required,oneof=P0 P1 P2 P3"`
	DefectType      string   `json:"defect_type" validate:"omitempty,oneof=functional performance ui compatibility security data"`
	RelatedCaseID   string   `json:"related_case_id" validate:"omitempty"`
	Language        string   `json:"language" validate:"omitempty,oneof=zh-CN en-US"`
	CustomFields    map[string]interface{} `json:"custom_fields" validate:"omitempty"`
}

// 获取验证错误信息（中英文）
func GetErrorMsg(lang string, err error) (zh, en string) {
	if err == nil {
		return "", ""
	}
	
	// 这里可以根据验证错误类型返回对应的中英文提示
	// 简化版本，实际项目中可以扩展
	zh = "参数验证失败"
	en = "Validation failed"
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrors {
			switch fieldErr.Tag() {
			case "required":
				zh = "字段不能为空"
				en = "Field is required"
			case "username":
				zh = "用户名格式不正确（字母开头，4-20位字母数字下划线）"
				en = "Invalid username format"
			case "password":
				zh = "密码格式不正确（至少8位，包含字母和数字）"
				en = "Invalid password format"
			case "email":
				zh = "邮箱格式不正确"
				en = "Invalid email format"
			case "min":
				zh = "长度不足"
				en = "Length is too short"
			case "max":
				zh = "长度超出限制"
				en = "Length exceeds limit"
			case "oneof":
				zh = "选项值不正确"
				en = "Invalid option value"
			}
			break
		}
	}
	
	return zh, en
}