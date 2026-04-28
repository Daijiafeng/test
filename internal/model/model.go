package model

import (
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
	LastLoginAt    time.Time `json:"last_login_at" db:"last_login_at"`
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
	ReportID     string    `json:"report_id" db:"report_id"`
	ProjectID    string    `json:"project_id" db:"project_id"`
	PlanID       string    `json:"plan_id" db:"plan_id"`
	Title        string    `json:"title" db:"title"`
	TitleEn      string    `json:"title_en" db:"title_en"`
	ReportType   string    `json:"report_type" db:"report_type"`
	Language     string    `json:"language" db:"language"`
	Summary      JSONB     `json:"summary" db:"summary"`
	AIEvaluation JSONB     `json:"ai_evaluation" db:"ai_evaluation"`
	Content      string    `json:"content" db:"content"`
	FeishuDocID  string    `json:"feishu_doc_id" db:"feishu_doc_id"`
	CreatedBy    string    `json:"created_by" db:"created_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// JSONB 用于 PostgreSQL JSONB 类型
type JSONB map[string]interface{}

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