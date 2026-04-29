package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Organization 组织
type Organization struct {
	OrgID       string    `json:"org_id" db:"org_id"`
	Name        string    `json:"name" db:"name"`
	NameEn      string    `json:"name_en" db:"name_en"`
	Description string    `json:"description" db:"description"`
	OwnerID     string    `json:"owner_id" db:"owner_id"`
	Status      string    `json:"status" db:"status"`
	AIConfig    JSONB     `json:"ai_config" db:"ai_config"`
	Language    string    `json:"language" db:"language"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Project 项目
type Project struct {
	ProjectID   string    `json:"project_id" db:"project_id"`
	OrgID       string    `json:"org_id" db:"org_id"`
	Name        string    `json:"name" db:"name"`
	NameEn      string    `json:"name_en" db:"name_en"`
	Description string    `json:"description" db:"description"`
	Status      string    `json:"status" db:"status"`
	Settings    JSONB     `json:"settings" db:"settings"`
	Language    string    `json:"language" db:"language"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// User 用户
type User struct {
	UserID         string    `json:"user_id" db:"user_id"`
	OrgID          string    `json:"org_id" db:"org_id"`
	Username       string    `json:"username" db:"username"`
	Email          string    `json:"email" db:"email"`
	Phone          string    `json:"phone" db:"phone"`
	PasswordHash   string    `json:"-" db:"password_hash"`
	AvatarURL      string    `json:"avatar_url" db:"avatar_url"`
	DisplayName    string    `json:"display_name" db:"display_name"`
	DisplayNameEn  string    `json:"display_name_en" db:"display_name_en"`
	Language       string    `json:"language" db:"language"`
	Status         string    `json:"status" db:"status"`
	LastLoginAt    *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Role 角色
type Role struct {
	RoleID     string    `json:"role_id" db:"role_id"`
	OrgID      string    `json:"org_id" db:"org_id"`
	Name       string    `json:"name" db:"name"`
	NameEn     string    `json:"name_en" db:"name_en"`
	RoleType   string    `json:"role_type" db:"role_type"`
	Permissions JSONB    `json:"permissions" db:"permissions"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// ProjectMember 项目成员
type ProjectMember struct {
	ID        int       `json:"id" db:"id"`
	ProjectID string    `json:"project_id" db:"project_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	RoleID    string    `json:"role_id" db:"role_id"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
}

// TestCase 测试用例
type TestCase struct {
	CaseID         string    `json:"case_id" db:"case_id"`
	ProjectID      string    `json:"project_id" db:"project_id"`
	ModuleID       string    `json:"module_id" db:"module_id"`
	Title          string    `json:"title" db:"title"`
	TitleEn        string    `json:"title_en" db:"title_en"`
	Priority       string    `json:"priority" db:"priority"`
	Precondition   string    `json:"precondition" db:"precondition"`
	Steps          []Step    `json:"steps" db:"steps"`
	Tags           []string  `json:"tags" db:"tags"`
	Status         string    `json:"status" db:"status"`
	CreatedBy      string    `json:"created_by" db:"created_by"`
	CreatedMethod  string    `json:"created_method" db:"created_method"`
	Version        int       `json:"version" db:"version"`
	ParentCaseID   string    `json:"parent_case_id" db:"parent_case_id"`
	Language       string    `json:"language" db:"language"`
	FeishuRecordID string    `json:"feishu_record_id" db:"feishu_record_id"`
	CustomFields   JSONB     `json:"custom_fields" db:"custom_fields"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Step 测试步骤
type Step struct {
	StepNum   int    `json:"step_num"`
	Action    string `json:"action"`
	ActionEn  string `json:"action_en"`
	Expected  string `json:"expected"`
	ExpectedEn string `json:"expected_en"`
}

// TestPlan 测试计划
type TestPlan struct {
	PlanID       string    `json:"plan_id" db:"plan_id"`
	ProjectID    string    `json:"project_id" db:"project_id"`
	Name         string    `json:"name" db:"name"`
	NameEn       string    `json:"name_en" db:"name_en"`
	Version      string    `json:"version" db:"version"`
	Description  string    `json:"description" db:"description"`
	Status       string    `json:"status" db:"status"`
	StartDate    time.Time `json:"start_date" db:"start_date"`
	EndDate      time.Time `json:"end_date" db:"end_date"`
	OwnerID      string    `json:"owner_id" db:"owner_id"`
	AISuggestion JSONB     `json:"ai_suggestion" db:"ai_suggestion"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// ExecutionRecord 执行记录
type ExecutionRecord struct {
	ExecutionID      string    `json:"execution_id" db:"execution_id"`
	PlanID           string    `json:"plan_id" db:"plan_id"`
	CaseID           string    `json:"case_id" db:"case_id"`
	ProjectID        string    `json:"project_id" db:"project_id"`
	ExecutorID       string    `json:"executor_id" db:"executor_id"`
	ExecutionType    string    `json:"execution_type" db:"execution_type"`
	Result           string    `json:"result" db:"result"`
	ActualResult     string    `json:"actual_result" db:"actual_result"`
	Duration         int       `json:"duration" db:"duration"`
	Environment      JSONB     `json:"environment" db:"environment"`
	AIAnalysis       JSONB     `json:"ai_analysis" db:"ai_analysis"`
	RelatedDefectID  string    `json:"related_defect_id" db:"related_defect_id"`
	CustomFields     JSONB     `json:"custom_fields" db:"custom_fields"`
	ExecutedAt       time.Time `json:"executed_at" db:"executed_at"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// Defect 缺陷
type Defect struct {
	DefectID          string    `json:"defect_id" db:"defect_id"`
	ProjectID         string    `json:"project_id" db:"project_id"`
	ModuleID          string    `json:"module_id" db:"module_id"`
	Title             string    `json:"title" db:"title"`
	TitleEn           string    `json:"title_en" db:"title_en"`
	Description       string    `json:"description" db:"description"`
	DescriptionEn     string    `json:"description_en" db:"description_en"`
	ReproduceSteps    string    `json:"reproduce_steps" db:"reproduce_steps"`
	Severity          string    `json:"severity" db:"severity"`
	Priority          string    `json:"priority" db:"priority"`
	DefectType        string    `json:"defect_type" db:"defect_type"`
	Status            string    `json:"status" db:"status"`
	AssigneeID        string    `json:"assignee_id" db:"assignee_id"`
	ReporterID        string    `json:"reporter_id" db:"reporter_id"`
	DiscoveryStage    string    `json:"discovery_stage" db:"discovery_stage"`
	Environment       JSONB     `json:"environment" db:"environment"`
	RelatedCaseID     string    `json:"related_case_id" db:"related_case_id"`
	RelatedExecutionID string   `json:"related_execution_id" db:"related_execution_id"`
	AIClassification  string    `json:"ai_classification" db:"ai_classification"`
	AIRootCause       string    `json:"ai_root_cause" db:"ai_root_cause"`
	AIFixSuggestion   string    `json:"ai_fix_suggestion" db:"ai_fix_suggestion"`
	SimilarDefects    []string  `json:"similar_defects" db:"similar_defects"`
	CustomFields      JSONB     `json:"custom_fields" db:"custom_fields"`
	Language          string    `json:"language" db:"language"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// TestReport 测试报告
type TestReport struct {
	ReportID     string     `json:"report_id" db:"report_id"`
	ProjectID    string     `json:"project_id" db:"project_id"`
	PlanID       string     `json:"plan_id" db:"plan_id"`
	Title        string     `json:"title" db:"title"`
	TitleEn      string     `json:"title_en" db:"title_en"`
	ReportType   string     `json:"report_type" db:"report_type"`
	Status       string     `json:"status" db:"status"` // draft, generated, published, archived
	Language     string     `json:"language" db:"language"`
	Summary      string     `json:"summary" db:"summary"` // 文字总结
	Details      JSONB      `json:"details" db:"details"` // 详细数据
	AIEvaluation JSONB      `json:"ai_evaluation" db:"ai_evaluation"`
	Content      string     `json:"content" db:"content"`
	FeishuDocID  string     `json:"feishu_doc_id" db:"feishu_doc_id"`
	ShareURL     string     `json:"share_url" db:"share_url"`
	GeneratedBy  string     `json:"generated_by" db:"generated_by"`
	GeneratedAt  time.Time  `json:"generated_at" db:"generated_at"`
	StartDate    time.Time  `json:"start_date" db:"start_date"`
	EndDate      time.Time  `json:"end_date" db:"end_date"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at" db:"updated_at"`
}

// JSONB 用于 PostgreSQL JSONB 类型
type JSONB map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONB: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, j)
}

// Module 模块
type Module struct {
	ModuleID   string    `json:"module_id" db:"module_id"`
	ProjectID  string    `json:"project_id" db:"project_id"`
	ParentID   string    `json:"parent_id" db:"parent_id"`
	Name       string    `json:"name" db:"name"`
	NameEn     string    `json:"name_en" db:"name_en"`
	SortOrder  int       `json:"sort_order" db:"sort_order"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// CustomFieldDefinition 自定义字段定义
type CustomFieldDefinition struct {
	FieldID       string    `json:"field_id" db:"field_id"`
	ProjectID     string    `json:"project_id" db:"project_id"`
	EntityType    string    `json:"entity_type" db:"entity_type"`
	FieldName     string    `json:"field_name" db:"field_name"`
	FieldNameEn   string    `json:"field_name_en" db:"field_name_en"`
	FieldType     string    `json:"field_type" db:"field_type"`
	Options       JSONB     `json:"options" db:"options"`
	Required      bool      `json:"required" db:"required"`
	DefaultValue  string    `json:"default_value" db:"default_value"`
	SortOrder     int       `json:"sort_order" db:"sort_order"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// OAuthState OAuth状态
type OAuthState struct {
	StateID   string    `json:"state_id" db:"state_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	StateCode string    `json:"state_code" db:"state_code"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}

// FeishuCredential 飞书凭证
type FeishuCredential struct {
	CredentialID  string    `json:"credential_id" db:"credential_id"`
	UserID        string    `json:"user_id" db:"user_id"`
	AccessToken   string    `json:"access_token" db:"access_token"`
	RefreshToken  string    `json:"refresh_token" db:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	FeishuUserID  string    `json:"feishu_user_id" db:"feishu_user_id"`
	FeishuName    string    `json:"feishu_name" db:"feishu_name"`
	FeishuAvatar  string    `json:"feishu_avatar" db:"feishu_avatar"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// PlanRequirement 计划需求关联
type PlanRequirement struct {
	ID                int       `json:"id" db:"id"`
	PlanID            string    `json:"plan_id" db:"plan_id"`
	RequirementType   string    `json:"requirement_type" db:"requirement_type"`
	RequirementURL    string    `json:"requirement_url" db:"requirement_url"`
	RequirementTitle  string    `json:"requirement_title" db:"requirement_title"`
	RequirementContent string   `json:"requirement_content" db:"requirement_content"`
	FeishuDocID       string    `json:"feishu_doc_id" db:"feishu_doc_id"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// PlanCase 计划用例关联
type PlanCase struct {
	ID         int       `json:"id" db:"id"`
	PlanID     string    `json:"plan_id" db:"plan_id"`
	CaseID     string    `json:"case_id" db:"case_id"`
	AssigneeID string    `json:"assignee_id" db:"assignee_id"`
	SortOrder  int       `json:"sort_order" db:"sort_order"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// CaseVersion 用例版本历史
type CaseVersion struct {
	VersionID    int       `json:"version_id" db:"version_id"`
	CaseID       string    `json:"case_id" db:"case_id"`
	VersionNum   int       `json:"version_num" db:"version_num"`
	Title        string    `json:"title" db:"title"`
	Steps        JSONB     `json:"steps" db:"steps"`
	ChangedBy    string    `json:"changed_by" db:"changed_by"`
	ChangeSummary string   `json:"change_summary" db:"change_summary"`
	Snapshot     JSONB     `json:"snapshot" db:"snapshot"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// CaseReview 用例评审
type CaseReview struct {
	ReviewID  string    `json:"review_id" db:"review_id"`
	CaseID    string    `json:"case_id" db:"case_id"`
	ReviewerID string   `json:"reviewer_id" db:"reviewer_id"`
	Result    string    `json:"result" db:"result"`
	Comment   string    `json:"comment" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// DefectComment 缺陷评论
type DefectComment struct {
	CommentID  string    `json:"comment_id" db:"comment_id"`
	DefectID   string    `json:"defect_id" db:"defect_id"`
	AuthorID   string    `json:"author_id" db:"author_id"`
	Content    string    `json:"content" db:"content"`
	Visibility string    `json:"visibility" db:"visibility"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// AICaseGeneration AI用例生成任务
type AICaseGeneration struct {
	TaskID            string     `json:"task_id" db:"task_id"`
	ProjectID         string     `json:"project_id" db:"project_id"`
	SourceType        string     `json:"source_type" db:"source_type"`
	SourceContent     string     `json:"source_content" db:"source_content"`
	FeishuDocID       string     `json:"feishu_doc_id" db:"feishu_doc_id"`
	RequestedBy       string     `json:"requested_by" db:"requested_by"`
	Status            string     `json:"status" db:"status"`
	GeneratedCount    int        `json:"generated_count" db:"generated_count"`
	AppliedCount      int        `json:"applied_count" db:"applied_count"`
	AppliedAt         *time.Time `json:"applied_at" db:"applied_at"`
	ErrorMessage      string     `json:"error_message" db:"error_message"`
	PriorityStrategy  string     `json:"priority_strategy" db:"priority_strategy"`
	CaseStyle         string     `json:"case_style" db:"case_style"`
	Language          string     `json:"language" db:"language"`
	CustomPrompt      string     `json:"custom_prompt" db:"custom_prompt"`
	MaxCases          int        `json:"max_cases" db:"max_cases"`
	ModuleID          string     `json:"module_id" db:"module_id"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	CompletedAt       *time.Time `json:"completed_at" db:"completed_at"`
}

// KnowledgeDoc 知识库文档
type KnowledgeDoc struct {
	DocID        string    `json:"doc_id" db:"doc_id"`
	ProjectID    string    `json:"project_id" db:"project_id"`
	Title        string    `json:"title" db:"title"`
	DocType      string    `json:"doc_type" db:"doc_type"`
	SourceType   string    `json:"source_type" db:"source_type"`
	FeishuDocID  string    `json:"feishu_doc_id" db:"feishu_doc_id"`
	FeishuDocURL string    `json:"feishu_doc_url" db:"feishu_doc_url"`
	Content      string    `json:"content" db:"content"`
	Summary      string    `json:"summary" db:"summary"`
	Tags         []string  `json:"tags" db:"tags"`
	VectorID     string    `json:"vector_id" db:"vector_id"`
	Language     string    `json:"language" db:"language"`
	CreatedBy    string    `json:"created_by" db:"created_by"`
	UpdatedBy    string    `json:"updated_by" db:"updated_by"`
	Version      int       `json:"version" db:"version"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// AuditLog 操作日志
type AuditLog struct {
	LogID         int       `json:"log_id" db:"log_id"`
	EntityType    string    `json:"entity_type" db:"entity_type"`
	EntityID      string    `json:"entity_id" db:"entity_id"`
	OperationType string    `json:"operation_type" db:"operation_type"`
	OperatorID    string    `json:"operator_id" db:"operator_id"`
	BeforeValue   string    `json:"before_value" db:"before_value"`
	AfterValue    string    `json:"after_value" db:"after_value"`
	Details       string    `json:"details" db:"details"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}