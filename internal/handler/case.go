package handler

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"testmind/internal/model"
	"testmind/internal/repository"
	"testmind/pkg/response"
)

type TestCaseHandler struct {
	db *repository.DB
}

func NewTestCaseHandler(db *repository.DB) *TestCaseHandler {
	return &TestCaseHandler{db: db}
}

// Create 创建测试用例
func (h *TestCaseHandler) Create(c *gin.Context) {
	projectID := c.Param("project_id")

	var req struct {
		ModuleID      string                   `json:"module_id" binding:"omitempty"`
		Title         string                   `json:"title" binding:"required,min=5,max=500"`
		TitleEn       string                   `json:"title_en" binding:"omitempty,max=500"`
		Priority      string                   `json:"priority" binding:"required,oneof=P0 P1 P2 P3"`
		Precondition  string                   `json:"precondition" binding:"omitempty"`
		Steps         []model.Step             `json:"steps" binding:"required,min=1"`
		Tags          []string                 `json:"tags" binding:"omitempty"`
		Language      string                   `json:"language" binding:"omitempty,oneof=zh-CN en-US"`
		CustomFields  map[string]interface{}   `json:"custom_fields" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	userID := c.GetString("user_id")

	// 序列化步骤
	stepsJSON, _ := json.Marshal(req.Steps)

	tc := &model.TestCase{
		CaseID:        uuid.New().String()[:32],
		ProjectID:     projectID,
		ModuleID:      req.ModuleID,
		Title:         req.Title,
		TitleEn:       req.TitleEn,
		Priority:      req.Priority,
		Precondition:  req.Precondition,
		Steps:         req.Steps,
		Tags:          req.Tags,
		Status:        "draft",
		CreatedBy:     userID,
		CreatedMethod: "manual",
		Version:       1,
		Language:      stringOrDefault(req.Language, "zh-CN"),
		CustomFields:  model.JSONB(req.CustomFields),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	query := `
		INSERT INTO test_cases (case_id, project_id, module_id, title, title_en, priority, 
			precondition, steps, tags, status, created_by, created_method, version, 
			language, custom_fields, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`
	_, err := h.db.ExecContext(c.Request.Context(), query,
		tc.CaseID, tc.ProjectID, tc.ModuleID, tc.Title, tc.TitleEn, tc.Priority,
		tc.Precondition, stepsJSON, tc.Tags, tc.Status, tc.CreatedBy, tc.CreatedMethod,
		tc.Version, tc.Language, tc.CustomFields, tc.CreatedAt, tc.UpdatedAt)
	if err != nil {
		response.InternalError(c, "创建测试用例失败", "Failed to create test case")
		return
	}

	response.Success(c, tc)
}

// Get 获取测试用例详情
func (h *TestCaseHandler) Get(c *gin.Context) {
	caseID := c.Param("case_id")

	var tc model.TestCase
	query := `SELECT * FROM test_cases WHERE case_id = $1`
	err := h.db.GetContext(c.Request.Context(), &tc, query, caseID)
	if err == sql.ErrNoRows {
		response.NotFound(c, "测试用例不存在", "Test case not found")
		return
	}
	if err != nil {
		response.InternalError(c, "查询测试用例失败", "Failed to get test case")
		return
	}

	// Steps 字段已在数据库层面正确反序列化，无需额外处理
	response.Success(c, tc)
}

// List 获取测试用例列表（支持筛选）
func (h *TestCaseHandler) List(c *gin.Context) {
	projectID := c.Param("project_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	moduleID := c.Query("module_id")
	priority := c.Query("priority")
	status := c.Query("status")
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	// 构建查询
	baseQuery := `SELECT * FROM test_cases WHERE project_id = $1`
	countQuery := `SELECT COUNT(*) FROM test_cases WHERE project_id = $1`
	args := []interface{}{projectID}
	argIdx := 2

	if moduleID != "" {
		baseQuery += ` AND module_id = $` + strconv.Itoa(argIdx)
		countQuery += ` AND module_id = $` + strconv.Itoa(argIdx)
		args = append(args, moduleID)
		argIdx++
	}
	if priority != "" {
		baseQuery += ` AND priority = $` + strconv.Itoa(argIdx)
		countQuery += ` AND priority = $` + strconv.Itoa(argIdx)
		args = append(args, priority)
		argIdx++
	}
	if status != "" {
		baseQuery += ` AND status = $` + strconv.Itoa(argIdx)
		countQuery += ` AND status = $` + strconv.Itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	if keyword != "" {
		baseQuery += ` AND (title ILIKE $` + strconv.Itoa(argIdx) + ` OR precondition ILIKE $` + strconv.Itoa(argIdx) + `)`
		countQuery += ` AND (title ILIKE $` + strconv.Itoa(argIdx) + ` OR precondition ILIKE $` + strconv.Itoa(argIdx) + `)`
		args = append(args, "%"+keyword+"%")
		argIdx++
	}

	baseQuery += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, pageSize, offset)

	var cases []model.TestCase
	err := h.db.SelectContext(c.Request.Context(), &cases, baseQuery, args...)
	if err != nil {
		response.InternalError(c, "查询测试用例列表失败", "Failed to list test cases")
		return
	}

	var total int64
	h.db.GetContext(c.Request.Context(), &total, countQuery, args[:argIdx-2]...)

	response.SuccessWithPagination(c, cases, page, pageSize, total)
}

// Update 更新测试用例
func (h *TestCaseHandler) Update(c *gin.Context) {
	caseID := c.Param("case_id")

	var req struct {
		ModuleID     string                 `json:"module_id" binding:"omitempty"`
		Title        string                 `json:"title" binding:"omitempty,min=5,max=500"`
		TitleEn      string                 `json:"title_en" binding:"omitempty,max=500"`
		Priority     string                 `json:"priority" binding:"omitempty,oneof=P0 P1 P2 P3"`
		Precondition string                 `json:"precondition" binding:"omitempty"`
		Steps        []model.Step           `json:"steps" binding:"omitempty,min=1"`
		Tags         []string               `json:"tags" binding:"omitempty"`
		Status       string                 `json:"status" binding:"omitempty,oneof=draft reviewing approved deprecated"`
		CustomFields map[string]interface{} `json:"custom_fields" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	// 获取当前用例（用于版本快照）
	var current model.TestCase
	err := h.db.GetContext(c.Request.Context(), &current,
		`SELECT * FROM test_cases WHERE case_id = $1`, caseID)
	if err == sql.ErrNoRows {
		response.NotFound(c, "测试用例不存在", "Test case not found")
		return
	}

	// 保存版本快照
	snapshot, _ := json.Marshal(current)
	h.db.ExecContext(c.Request.Context(), `
		INSERT INTO case_versions (case_id, version_num, title, steps, changed_by, change_summary, snapshot, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`, caseID, current.Version, current.Title, current.Steps, c.GetString("user_id"), "更新用例", snapshot)

	// 构建更新
	updates := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Title != "" {
		updates = append(updates, `title = $`+strconv.Itoa(argIdx))
		args = append(args, req.Title)
		argIdx++
	}
	if req.TitleEn != "" {
		updates = append(updates, `title_en = $`+strconv.Itoa(argIdx))
		args = append(args, req.TitleEn)
		argIdx++
	}
	if req.Priority != "" {
		updates = append(updates, `priority = $`+strconv.Itoa(argIdx))
		args = append(args, req.Priority)
		argIdx++
	}
	if req.Precondition != "" {
		updates = append(updates, `precondition = $`+strconv.Itoa(argIdx))
		args = append(args, req.Precondition)
		argIdx++
	}
	if req.Steps != nil {
		stepsJSON, _ := json.Marshal(req.Steps)
		updates = append(updates, `steps = $`+strconv.Itoa(argIdx))
		args = append(args, stepsJSON)
		argIdx++
	}
	if req.Tags != nil {
		updates = append(updates, `tags = $`+strconv.Itoa(argIdx))
		args = append(args, req.Tags)
		argIdx++
	}
	if req.Status != "" {
		updates = append(updates, `status = $`+strconv.Itoa(argIdx))
		args = append(args, req.Status)
		argIdx++
	}
	if req.CustomFields != nil {
		updates = append(updates, `custom_fields = $`+strconv.Itoa(argIdx))
		args = append(args, model.JSONB(req.CustomFields))
		argIdx++
	}

	updates = append(updates, `version = version + 1`)
	updates = append(updates, `updated_at = NOW()`)

	// WHERE case_id
	args = append(args, caseID)

	query := `UPDATE test_cases SET ` + joinStrings(updates, ", ") + ` WHERE case_id = $` + strconv.Itoa(argIdx)
	_, err = h.db.ExecContext(c.Request.Context(), query, args...)
	if err != nil {
		response.InternalError(c, "更新测试用例失败", "Failed to update test case")
		return
	}

	response.Success(c, gin.H{"case_id": caseID, "updated": true})
}

// Delete 删除测试用例
func (h *TestCaseHandler) Delete(c *gin.Context) {
	caseID := c.Param("case_id")

	result, err := h.db.ExecContext(c.Request.Context(),
		`UPDATE test_cases SET status = 'deprecated', updated_at = NOW() WHERE case_id = $1`, caseID)
	if err != nil {
		response.InternalError(c, "删除测试用例失败", "Failed to delete test case")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		response.NotFound(c, "测试用例不存在", "Test case not found")
		return
	}

	response.Success(c, gin.H{"case_id": caseID, "status": "deprecated"})
}

// GetVersions 获取用例版本历史
func (h *TestCaseHandler) GetVersions(c *gin.Context) {
	caseID := c.Param("case_id")

	var versions []struct {
		VersionID     int       `json:"version_id" db:"version_id"`
		VersionNum    int       `json:"version_num" db:"version_num"`
		Title         string    `json:"title" db:"title"`
		ChangedBy     string    `json:"changed_by" db:"changed_by"`
		ChangeSummary string    `json:"change_summary" db:"change_summary"`
		CreatedAt     time.Time `json:"created_at" db:"created_at"`
	}

	query := `SELECT version_id, version_num, title, changed_by, change_summary, created_at 
		FROM case_versions WHERE case_id = $1 ORDER BY version_num DESC`
	err := h.db.SelectContext(c.Request.Context(), &versions, query, caseID)
	if err != nil {
		response.InternalError(c, "查询版本历史失败", "Failed to get versions")
		return
	}

	response.Success(c, versions)
}

// Review 提交/处理用例评审
func (h *TestCaseHandler) Review(c *gin.Context) {
	caseID := c.Param("case_id")

	var req struct {
		Result  string `json:"result" binding:"required,oneof=approved rejected"`
		Comment string `json:"comment" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	userID := c.GetString("user_id")
	reviewID := uuid.New().String()[:32]

	query := `
		INSERT INTO case_reviews (review_id, case_id, reviewer_id, result, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	_, err := h.db.ExecContext(c.Request.Context(), query, reviewID, caseID, userID, req.Result, req.Comment)
	if err != nil {
		response.InternalError(c, "提交评审失败", "Failed to submit review")
		return
	}

	// 更新用例状态
	status := "approved"
	if req.Result == "rejected" {
		status = "draft"
	}
	h.db.ExecContext(c.Request.Context(),
		`UPDATE test_cases SET status = $1, updated_at = NOW() WHERE case_id = $2`, status, caseID)

	response.Success(c, gin.H{
		"review_id": reviewID,
		"case_id":   caseID,
		"result":    req.Result,
		"status":    status,
	})
}

// BatchCreate 批量创建用例（AI生成后使用）
func (h *TestCaseHandler) BatchCreate(c *gin.Context) {
	projectID := c.Param("project_id")

	var req struct {
		Cases []struct {
			ModuleID     string                 `json:"module_id" binding:"omitempty"`
			Title        string                 `json:"title" binding:"required"`
			TitleEn      string                 `json:"title_en" binding:"omitempty"`
			Priority     string                 `json:"priority" binding:"required,oneof=P0 P1 P2 P3"`
			Precondition string                 `json:"precondition" binding:"omitempty"`
			Steps        []model.Step           `json:"steps" binding:"required,min=1"`
			Tags         []string               `json:"tags" binding:"omitempty"`
			CustomFields map[string]interface{} `json:"custom_fields" binding:"omitempty"`
		} `json:"cases" binding:"required,min=1"`
		CreatedMethod string `json:"created_method" binding:"omitempty,oneof=manual ai_generated ai_optimized"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	userID := c.GetString("user_id")
	createdMethod := stringOrDefault(req.CreatedMethod, "ai_generated")
	inserted := 0
	var caseIDs []string

	for _, tc := range req.Cases {
		caseID := uuid.New().String()[:32]
		stepsJSON, _ := json.Marshal(tc.Steps)
		customFields := model.JSONB{}
		if tc.CustomFields != nil {
			customFields = model.JSONB(tc.CustomFields)
		}

		query := `
			INSERT INTO test_cases (case_id, project_id, module_id, title, title_en, priority,
				precondition, steps, tags, status, created_by, created_method, version,
				language, custom_fields, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'draft', $10, $11, 1, 'zh-CN', $12, NOW(), NOW())
		`
		_, err := h.db.ExecContext(c.Request.Context(), query,
			caseID, projectID, tc.ModuleID, tc.Title, tc.TitleEn, tc.Priority,
			tc.Precondition, stepsJSON, tc.Tags, userID, createdMethod, customFields)
		if err == nil {
			inserted++
			caseIDs = append(caseIDs, caseID)
		}
	}

	response.Success(c, gin.H{
		"project_id": projectID,
		"total":      len(req.Cases),
		"inserted":   inserted,
		"case_ids":   caseIDs,
	})
}

// Search 全文搜索用例
func (h *TestCaseHandler) Search(c *gin.Context) {
	projectID := c.Param("project_id")
	keyword := c.Query("q")

	if keyword == "" {
		response.BadRequest(c, "请输入搜索关键词", "Search keyword is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	offset := (page - 1) * pageSize

	var cases []model.TestCase
	query := `
		SELECT * FROM test_cases 
		WHERE project_id = $1 AND (title ILIKE $2 OR precondition ILIKE $2)
		ORDER BY updated_at DESC LIMIT $3 OFFSET $4
	`
	err := h.db.SelectContext(c.Request.Context(), &cases, query, projectID, "%"+keyword+"%", pageSize, offset)
	if err != nil {
		response.InternalError(c, "搜索失败", "Search failed")
		return
	}

	var total int64
	h.db.GetContext(c.Request.Context(), &total,
		`SELECT COUNT(*) FROM test_cases WHERE project_id = $1 AND (title ILIKE $2 OR precondition ILIKE $2)`,
		projectID, "%"+keyword+"%")

	response.SuccessWithPagination(c, cases, page, pageSize, total)
}

func joinStrings(ss []string, sep string) string {
	if len(ss) == 0 {
		return ""
	}
	result := ss[0]
	for i := 1; i < len(ss); i++ {
		result += sep + ss[i]
	}
	return result
}