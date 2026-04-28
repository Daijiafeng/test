package handler

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"testmind/internal/model"
	"testmind/internal/repository"
	"testmind/pkg/response"
)

type TestPlanHandler struct {
	db *repository.DB
}

func NewTestPlanHandler(db *repository.DB) *TestPlanHandler {
	return &TestPlanHandler{db: db}
}

// Create 创建测试计划
func (h *TestPlanHandler) Create(c *gin.Context) {
	projectID := c.Param("project_id")

	var req struct {
		Name        string `json:"name" binding:"required,min=2,max=200"`
		NameEn      string `json:"name_en" binding:"omitempty,max=200"`
		Version     string `json:"version" binding:"omitempty,max=50"`
		Description string `json:"description" binding:"omitempty"`
		StartDate   string `json:"start_date" binding:"omitempty"`
		EndDate     string `json:"end_date" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	userID := c.GetString("user_id")

	plan := &model.TestPlan{
		PlanID:      uuid.New().String()[:32],
		ProjectID:   projectID,
		Name:        req.Name,
		NameEn:      req.NameEn,
		Version:     req.Version,
		Description: req.Description,
		Status:      "draft",
		OwnerID:     userID,
		AISuggestion: model.JSONB{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			plan.StartDate = t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			plan.EndDate = t
		}
	}

	query := `
		INSERT INTO test_plans (plan_id, project_id, name, name_en, version, description, status, 
			start_date, end_date, owner_id, ai_suggestion, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := h.db.ExecContext(c.Request.Context(), query,
		plan.PlanID, plan.ProjectID, plan.Name, plan.NameEn, plan.Version, plan.Description,
		plan.Status, plan.StartDate, plan.EndDate, plan.OwnerID, plan.AISuggestion,
		plan.CreatedAt, plan.UpdatedAt)
	if err != nil {
		response.InternalError(c, "创建测试计划失败", "Failed to create test plan")
		return
	}

	response.Success(c, plan)
}

// Get 获取测试计划详情
func (h *TestPlanHandler) Get(c *gin.Context) {
	planID := c.Param("plan_id")

	var plan model.TestPlan
	query := `SELECT * FROM test_plans WHERE plan_id = $1`
	err := h.db.GetContext(c.Request.Context(), &plan, query, planID)
	if err == sql.ErrNoRows {
		response.NotFound(c, "测试计划不存在", "Test plan not found")
		return
	}
	if err != nil {
		response.InternalError(c, "查询测试计划失败", "Failed to get test plan")
		return
	}

	response.Success(c, plan)
}

// List 获取测试计划列表
func (h *TestPlanHandler) List(c *gin.Context) {
	projectID := c.Param("project_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	var plans []model.TestPlan
	query := `SELECT * FROM test_plans WHERE project_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	err := h.db.SelectContext(c.Request.Context(), &plans, query, projectID, pageSize, offset)
	if err != nil {
		response.InternalError(c, "查询测试计划列表失败", "Failed to list test plans")
		return
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM test_plans WHERE project_id = $1`
	h.db.GetContext(c.Request.Context(), &total, countQuery, projectID)

	response.SuccessWithPagination(c, plans, page, pageSize, total)
}

// Update 更新测试计划
func (h *TestPlanHandler) Update(c *gin.Context) {
	planID := c.Param("plan_id")

	var req struct {
		Name        string `json:"name" binding:"omitempty,min=2,max=200"`
		NameEn      string `json:"name_en" binding:"omitempty,max=200"`
		Version     string `json:"version" binding:"omitempty,max=50"`
		Description string `json:"description" binding:"omitempty"`
		Status      string `json:"status" binding:"omitempty,oneof=draft in_progress completed cancelled"`
		StartDate   string `json:"start_date" binding:"omitempty"`
		EndDate     string `json:"end_date" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	query := `
		UPDATE test_plans SET 
			name = COALESCE(NULLIF($1, ''), name),
			name_en = COALESCE(NULLIF($2, ''), name_en),
			version = COALESCE(NULLIF($3, ''), version),
			description = COALESCE(NULLIF($4, ''), description),
			status = COALESCE(NULLIF($5, ''), status),
			updated_at = NOW()
		WHERE plan_id = $6
	`
	result, err := h.db.ExecContext(c.Request.Context(), query,
		req.Name, req.NameEn, req.Version, req.Description, req.Status, planID)
	if err != nil {
		response.InternalError(c, "更新测试计划失败", "Failed to update test plan")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		response.NotFound(c, "测试计划不存在", "Test plan not found")
		return
	}

	response.Success(c, gin.H{"plan_id": planID, "updated": true})
}

// Delete 删除测试计划
func (h *TestPlanHandler) Delete(c *gin.Context) {
	planID := c.Param("plan_id")

	result, err := h.db.ExecContext(c.Request.Context(),
		`UPDATE test_plans SET status = 'cancelled', updated_at = NOW() WHERE plan_id = $1`, planID)
	if err != nil {
		response.InternalError(c, "删除测试计划失败", "Failed to delete test plan")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		response.NotFound(c, "测试计划不存在", "Test plan not found")
		return
	}

	response.Success(c, gin.H{"plan_id": planID, "status": "cancelled"})
}

// AddRequirement 添加需求关联
func (h *TestPlanHandler) AddRequirement(c *gin.Context) {
	planID := c.Param("plan_id")

	var req struct {
		RequirementType    string `json:"requirement_type" binding:"required,oneof=manual feishu_doc"`
		RequirementURL     string `json:"requirement_url" binding:"omitempty"`
		RequirementTitle   string `json:"requirement_title" binding:"required"`
		RequirementContent string `json:"requirement_content" binding:"omitempty"`
		FeishuDocID        string `json:"feishu_doc_id" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	query := `
		INSERT INTO plan_requirements (plan_id, requirement_type, requirement_url, requirement_title, 
			requirement_content, feishu_doc_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`
	_, err := h.db.ExecContext(c.Request.Context(), query,
		planID, req.RequirementType, req.RequirementURL, req.RequirementTitle,
		req.RequirementContent, req.FeishuDocID)
	if err != nil {
		response.InternalError(c, "添加需求关联失败", "Failed to add requirement")
		return
	}

	response.Success(c, gin.H{"plan_id": planID, "requirement_title": req.RequirementTitle})
}

// ListRequirements 获取计划关联的需求列表
func (h *TestPlanHandler) ListRequirements(c *gin.Context) {
	planID := c.Param("plan_id")

	var requirements []struct {
		ID               int       `json:"id" db:"id"`
		RequirementType  string    `json:"requirement_type" db:"requirement_type"`
		RequirementURL   string    `json:"requirement_url" db:"requirement_url"`
		RequirementTitle string    `json:"requirement_title" db:"requirement_title"`
		FeishuDocID      string    `json:"feishu_doc_id" db:"feishu_doc_id"`
		CreatedAt        time.Time `json:"created_at" db:"created_at"`
	}

	query := `SELECT id, requirement_type, requirement_url, requirement_title, feishu_doc_id, created_at 
		FROM plan_requirements WHERE plan_id = $1 ORDER BY created_at DESC`
	err := h.db.SelectContext(c.Request.Context(), &requirements, query, planID)
	if err != nil {
		response.InternalError(c, "查询需求列表失败", "Failed to list requirements")
		return
	}

	response.Success(c, requirements)
}

// AddCases 批量关联用例到计划
func (h *TestPlanHandler) AddCases(c *gin.Context) {
	planID := c.Param("plan_id")

	var req struct {
		CaseIDs []string `json:"case_ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	userID := c.GetString("user_id")

	// 批量插入
	query := `
		INSERT INTO plan_cases (plan_id, case_id, assignee_id, sort_order, created_at)
		VALUES ($1, $2, $3, 0, NOW())
		ON CONFLICT (plan_id, case_id) DO NOTHING
	`
	inserted := 0
	for i, caseID := range req.CaseIDs {
		_, err := h.db.ExecContext(c.Request.Context(), query, planID, caseID, userID, i)
		if err == nil {
			inserted++
		}
	}

	response.Success(c, gin.H{
		"plan_id":    planID,
		"total":      len(req.CaseIDs),
		"inserted":   inserted,
		"skipped":    len(req.CaseIDs) - inserted,
	})
}

// RemoveCases 批量取消关联用例
func (h *TestPlanHandler) RemoveCases(c *gin.Context) {
	planID := c.Param("plan_id")

	var req struct {
		CaseIDs []string `json:"case_ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	query := `DELETE FROM plan_cases WHERE plan_id = $1 AND case_id = ANY($2)`
	_, err := h.db.ExecContext(c.Request.Context(), query, planID, req.CaseIDs)
	if err != nil {
		response.InternalError(c, "取消关联失败", "Failed to remove cases")
		return
	}

	response.Success(c, gin.H{"plan_id": planID, "removed": len(req.CaseIDs)})
}

// GetProgress 获取计划执行进度
func (h *TestPlanHandler) GetProgress(c *gin.Context) {
	planID := c.Param("plan_id")

	// 用例总数
	var totalCases int
	h.db.GetContext(c.Request.Context(), &totalCases,
		`SELECT COUNT(*) FROM plan_cases WHERE plan_id = $1`, planID)

	// 按结果统计执行记录
	var stats []struct {
		Result string `json:"result" db:"result"`
		Count  int    `json:"count" db:"count"`
	}
	h.db.SelectContext(c.Request.Context(), &stats,
		`SELECT result, COUNT(*) as count FROM execution_records WHERE plan_id = $1 GROUP BY result`, planID)

	// 计算进度
	executedMap := make(map[string]int)
	totalExecuted := 0
	for _, s := range stats {
		executedMap[s.Result] = s.Count
		totalExecuted += s.Count
	}

	passRate := 0.0
	if totalExecuted > 0 {
		passRate = float64(executedMap["passed"]) / float64(totalExecuted) * 100
	}

	progress := gin.H{
		"plan_id":        planID,
		"total_cases":    totalCases,
		"total_executed": totalExecuted,
		"remaining":      totalCases - totalExecuted,
		"progress_rate":  0.0,
		"pass_rate":      passRate,
		"by_result":      executedMap,
	}

	if totalCases > 0 {
		progress["progress_rate"] = float64(totalExecuted) / float64(totalCases) * 100
	}

	response.Success(c, progress)
}