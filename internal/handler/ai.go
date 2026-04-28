package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"testmind/internal/config"
	"testmind/internal/model"
	"testmind/internal/repository"
	"testmind/pkg/response"
)

type AIHandler struct {
	cfg *config.Config
	db  *repository.DB
}

func NewAIHandler(cfg *config.Config, db *repository.DB) *AIHandler {
	return &AIHandler{cfg: cfg, db: db}
}

// GenerateRequest AI生成请求
type GenerateRequest struct {
	SourceType       string                 `json:"source_type" binding:"required,oneof=manual feishu_doc text"`
	SourceContent    string                 `json:"source_content" binding:"omitempty"`
	FeishuDocID      string                 `json:"feishu_doc_id" binding:"omitempty"`
	FeishuDocURL     string                 `json:"feishu_doc_url" binding:"omitempty"`
	ProjectID        string                 `json:"project_id" binding:"required"`
	ModuleID         string                 `json:"module_id" binding:"omitempty"`
	PriorityStrategy string                 `json:"priority_strategy" binding:"omitempty,oneof=auto high medium low"`
	CaseStyle        string                 `json:"case_style" binding:"omitempty,oneof=standard concise detailed"`
	Language         string                 `json:"language" binding:"omitempty,oneof=zh-CN en-US"`
	CustomPrompt     string                 `json:"custom_prompt" binding:"omitempty"`
	MaxCases         int                    `json:"max_cases" binding:"omitempty,min=1,max=200"`
	CustomFields     map[string]interface{} `json:"custom_fields" binding:"omitempty"`
}

// GenerateCase AI生成测试用例
func (h *AIHandler) GenerateCase(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	userID := c.GetString("user_id")
	taskID := uuid.New().String()[:32]

	// 创建AI生成任务
	task := &model.AICaseGeneration{
		TaskID:          taskID,
		ProjectID:       req.ProjectID,
		SourceType:      req.SourceType,
		SourceContent:   req.SourceContent,
		FeishuDocID:     req.FeishuDocID,
		RequestedBy:     userID,
		Status:          "pending",
		PriorityStrategy: req.PriorityStrategy,
		CaseStyle:       req.CaseStyle,
		Language:        req.Language,
		CustomPrompt:    req.CustomPrompt,
		MaxCases:        req.MaxCases,
		CreatedAt:       time.Now(),
	}

	query := `
		INSERT INTO ai_case_generations (task_id, project_id, source_type, source_content, 
			feishu_doc_id, requested_by, status, priority_strategy, case_style, language, 
			custom_prompt, max_cases, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := h.db.ExecContext(c.Request.Context(), query,
		task.TaskID, task.ProjectID, task.SourceType, task.SourceContent,
		task.FeishuDocID, task.RequestedBy, task.Status, task.PriorityStrategy,
		task.CaseStyle, task.Language, task.CustomPrompt, task.MaxCases, task.CreatedAt)
	if err != nil {
		response.InternalError(c, "创建AI任务失败", "Failed to create AI task")
		return
	}

	// 如果是飞书文档，获取文档内容
	if req.SourceType == "feishu_doc" && req.FeishuDocID != "" {
		content, err := h.fetchFeishuDocument(c, userID, req.FeishuDocID)
		if err != nil {
			// 更新任务状态为失败
			h.db.ExecContext(c.Request.Context(),
				`UPDATE ai_case_generations SET status = 'failed', error_message = $1 WHERE task_id = $2`,
				err.Error(), taskID)
			response.InternalError(c, "获取飞书文档失败", "Failed to fetch Feishu document")
			return
		}
		task.SourceContent = content
	}

	// TODO: 实际调用AI服务生成用例
	// 这里模拟生成结果
	generatedCases := h.mockGenerateCases(task)

	// 更新任务状态
	h.db.ExecContext(c.Request.Context(),
		`UPDATE ai_case_generations SET status = 'completed', generated_count = $1, completed_at = NOW() WHERE task_id = $2`,
		len(generatedCases), taskID)

	response.Success(c, gin.H{
		"task_id":         taskID,
		"status":          "completed",
		"generated_count": len(generatedCases),
		"cases":           generatedCases,
		"message":         "AI用例生成完成",
	})
}

// GetTaskStatus 获取AI生成任务状态
func (h *AIHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("task_id")

	var task struct {
		TaskID         string     `json:"task_id" db:"task_id"`
		ProjectID      string     `json:"project_id" db:"project_id"`
		Status         string     `json:"status" db:"status"`
		GeneratedCount int        `json:"generated_count" db:"generated_count"`
		ErrorMessage   string     `json:"error_message" db:"error_message"`
		CreatedAt      time.Time  `json:"created_at" db:"created_at"`
		CompletedAt    *time.Time `json:"completed_at" db:"completed_at"`
	}

	err := h.db.GetContext(c.Request.Context(), &task,
		`SELECT task_id, project_id, status, generated_count, error_message, created_at, completed_at 
		FROM ai_case_generations WHERE task_id = $1`, taskID)
	if err == sql.ErrNoRows {
		response.NotFound(c, "任务不存在", "Task not found")
		return
	}
	if err != nil {
		response.InternalError(c, "查询任务失败", "Failed to get task")
		return
	}

	response.Success(c, task)
}

// ListTasks 获取AI生成任务列表
func (h *AIHandler) ListTasks(c *gin.Context) {
	projectID := c.Param("project_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	baseQuery := `SELECT * FROM ai_case_generations WHERE project_id = $1`
	countQuery := `SELECT COUNT(*) FROM ai_case_generations WHERE project_id = $1`
	args := []interface{}{projectID}
	argIdx := 2

	if status != "" {
		baseQuery += ` AND status = $` + strconv.Itoa(argIdx)
		countQuery += ` AND status = $` + strconv.Itoa(argIdx)
		args = append(args, status)
		argIdx++
	}

	baseQuery += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, pageSize, offset)

	var tasks []model.AICaseGeneration
	err := h.db.SelectContext(c.Request.Context(), &tasks, baseQuery, args...)
	if err != nil {
		response.InternalError(c, "查询任务列表失败", "Failed to list tasks")
		return
	}

	var total int64
	h.db.GetContext(c.Request.Context(), &total, countQuery, args[:argIdx-2]...)

	response.SuccessWithPagination(c, tasks, page, pageSize, total)
}

// ApplyCases 应用AI生成的用例（保存到数据库）
func (h *AIHandler) ApplyCases(c *gin.Context) {
	taskID := c.Param("task_id")

	var req struct {
		CaseIDs    []string `json:"case_ids" binding:"omitempty"`
		ApplyAll   bool     `json:"apply_all"`
		ModuleID   string   `json:"module_id" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	// 获取任务生成的用例
	var task model.AICaseGeneration
	err := h.db.GetContext(c.Request.Context(), &task,
		`SELECT * FROM ai_case_generations WHERE task_id = $1`, taskID)
	if err == sql.ErrNoRows {
		response.NotFound(c, "任务不存在", "Task not found")
		return
	}

	// TODO: 从AI结果中获取用例数据
	// 这里使用模拟数据
	generatedCases := h.mockGenerateCases(&task)

	userID := c.GetString("user_id")
	inserted := 0
	var appliedCaseIDs []string

	for i, tc := range generatedCases {
		// 如果指定了特定用例，只应用那些
		if len(req.CaseIDs) > 0 && !contains(req.CaseIDs, tc["case_id"].(string)) {
			continue
		}

		caseID := uuid.New().String()[:32]
		stepsJSON, _ := json.Marshal(tc["steps"])

		query := `
			INSERT INTO test_cases (case_id, project_id, module_id, title, title_en, priority,
				precondition, steps, status, created_by, created_method, version, language, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'draft', $9, 'ai_generated', 1, $10, NOW(), NOW())
		`
		moduleID := req.ModuleID
		if moduleID == "" {
			moduleID = task.ModuleID
		}

		_, err := h.db.ExecContext(c.Request.Context(), query,
			caseID, task.ProjectID, moduleID, tc["title"], tc["title_en"], tc["priority"],
			tc["precondition"], stepsJSON, userID, task.Language)
		if err == nil {
			inserted++
			appliedCaseIDs = append(appliedCaseIDs, caseID)
		}
	}

	// 更新任务的应用状态
	h.db.ExecContext(c.Request.Context(),
		`UPDATE ai_case_generations SET applied_count = $1, applied_at = NOW() WHERE task_id = $2`,
		inserted, taskID)

	response.Success(c, gin.H{
		"task_id":       taskID,
		"total":         len(generatedCases),
		"applied":       inserted,
		"applied_ids":   appliedCaseIDs,
	})
}

// OptimizeCase AI优化已有用例
func (h *AIHandler) OptimizeCase(c *gin.Context) {
	caseID := c.Param("case_id")

	var req struct {
		OptimizeType string `json:"optimize_type" binding:"required,oneof=enhance simplify translate format"`
		TargetLang   string `json:"target_lang" binding:"omitempty,oneof=zh-CN en-US"`
		CustomPrompt string `json:"custom_prompt" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	// 获取原用例
	var tc model.TestCase
	err := h.db.GetContext(c.Request.Context(), &tc,
		`SELECT * FROM test_cases WHERE case_id = $1`, caseID)
	if err == sql.ErrNoRows {
		response.NotFound(c, "用例不存在", "Test case not found")
		return
	}

	// TODO: 实际调用AI服务优化用例
	// 这里模拟优化结果
	optimizedCase := h.mockOptimizeCase(tc, req.OptimizeType, req.TargetLang)

	// 创建版本快照
	snapshot, _ := json.Marshal(tc)
	h.db.ExecContext(c.Request.Context(), `
		INSERT INTO case_versions (case_id, version_num, title, steps, changed_by, change_summary, snapshot, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`, caseID, tc.Version, tc.Title, tc.Steps, c.GetString("user_id"), "AI优化", snapshot)

	// 更新用例
	stepsJSON, _ := json.Marshal(optimizedCase["steps"])
	h.db.ExecContext(c.Request.Context(),
		`UPDATE test_cases SET title = $1, title_en = $2, precondition = $3, steps = $4, 
			language = $5, version = version + 1, updated_at = NOW() WHERE case_id = $6`,
		optimizedCase["title"], optimizedCase["title_en"], optimizedCase["precondition"],
		stepsJSON, optimizedCase["language"], caseID)

	response.Success(c, gin.H{
		"case_id":        caseID,
		"optimize_type":  req.OptimizeType,
		"optimized_case": optimizedCase,
		"version":        tc.Version + 1,
	})
}

// fetchFeishuDocument 获取飞书文档内容
func (h *AIHandler) fetchFeishuDocument(c *gin.Context, userID, docID string) (string, error) {
	// 获取用户飞书凭证
	var accessToken string
	err := h.db.GetContext(c.Request.Context(), &accessToken,
		`SELECT access_token FROM feishu_credentials WHERE user_id = $1 AND expires_at > NOW()`,
		userID)
	if err != nil {
		return "", fmt.Errorf("请先授权飞书")
	}

	// TODO: 实际调用飞书API获取文档
	// 这里返回模拟内容
	return fmt.Sprintf("飞书文档 %s 的内容（待实现）", docID), nil
}

// mockGenerateCases 模拟生成测试用例
func (h *AIHandler) mockGenerateCases(task *model.AICaseGeneration) []map[string]interface{} {
	cases := []map[string]interface{}{}
	
	// 根据源内容生成模拟用例
	content := task.SourceContent
	if content == "" {
		content = "默认需求内容"
	}

	// 简单模拟：生成3个用例
	for i := 1; i <= 3; i++ {
		caseData := map[string]interface{}{
			"case_id":       fmt.Sprintf("ai_gen_%d", i),
			"title":         fmt.Sprintf("AI生成用例%d：验证%s功能", i, extractKeyword(content)),
			"title_en":      fmt.Sprintf("AI Generated Case %d: Verify %s functionality", i, extractKeyword(content)),
			"priority":      determinePriority(task.PriorityStrategy),
			"precondition":  "系统正常运行，用户已登录",
			"steps": []model.Step{
				{
					StepNum:   1,
					Action:    "进入功能页面",
					Expected:  "页面正常加载",
					StepType:  "action",
				},
				{
					StepNum:   2,
					Action:    "执行核心操作",
					Expected:  "操作成功执行",
					StepType:  "action",
				},
				{
					StepNum:   3,
					Action:    "验证结果",
					Expected:  "结果符合预期",
					StepType:  "verify",
				},
			},
			"language": task.Language,
		}
		cases = append(cases, caseData)
	}

	return cases
}

// mockOptimizeCase 模拟优化用例
func (h *AIHandler) mockOptimizeCase(tc model.TestCase, optimizeType, targetLang string) map[string]interface{} {
	optimized := map[string]interface{}{
		"title":        tc.Title,
		"title_en":     tc.TitleEn,
		"precondition": tc.Precondition,
		"steps":        tc.Steps,
		"language":     tc.Language,
	}

	switch optimizeType {
	case "enhance":
		optimized["title"] = "[增强] " + tc.Title
		optimized["precondition"] = tc.Precondition + "（已补充前置条件）"
	case "simplify":
		optimized["title"] = strings.Replace(tc.Title, "验证", "测试", 1)
	case "translate":
		if targetLang == "en-US" {
			optimized["title_en"] = "[Translated] " + tc.Title
			optimized["language"] = "en-US"
		} else {
			optimized["language"] = "zh-CN"
		}
	case "format":
		optimized["title"] = "[格式化] " + tc.Title
	}

	return optimized
}

// extractKeyword 从内容中提取关键词
func extractKeyword(content string) string {
	// 简单模拟：取前10个字符作为关键词
	if len(content) > 10 {
		return content[:10]
	}
	return content
}

// determinePriority 根据策略确定优先级
func determinePriority(strategy string) string {
	switch strategy {
	case "high":
		return "P0"
	case "medium":
		return "P1"
	case "low":
		return "P2"
	default:
		return "P1"
	}
}

// contains 检查字符串是否在列表中
func contains(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}

var _ = strconv.Itoa