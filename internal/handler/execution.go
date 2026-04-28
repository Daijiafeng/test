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

type ExecutionHandler struct {
	db *repository.DB
}

func NewExecutionHandler(db *repository.DB) *ExecutionHandler {
	return &ExecutionHandler{db: db}
}

// Create 创建执行记录
func (h *ExecutionHandler) Create(c *gin.Context) {
	planID := c.Param("plan_id")

	var req struct {
		CaseID         string                 `json:"case_id" binding:"required"`
		ExecutionType  string                 `json:"execution_type" binding:"required,oneof=manual auto_web auto_api auto_mobile auto_performance"`
		Result         string                 `json:"result" binding:"required,oneof=passed failed blocked skipped"`
		ActualResult   string                 `json:"actual_result" binding:"omitempty"`
		Duration       int                    `json:"duration" binding:"omitempty"` // 秒
		Environment    map[string]interface{} `json:"environment" binding:"omitempty"`
		RelatedDefectID string                 `json:"related_defect_id" binding:"omitempty"`
		CustomFields   map[string]interface{} `json:"custom_fields" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	userID := c.GetString("user_id")

	// 获取用例所属项目
	var projectID string
	h.db.GetContext(c.Request.Context(), &projectID,
		`SELECT project_id FROM test_cases WHERE case_id = $1`, req.CaseID)

	exec := &model.ExecutionRecord{
		ExecutionID:      uuid.New().String()[:32],
		PlanID:           planID,
		CaseID:           req.CaseID,
		ProjectID:        projectID,
		ExecutorID:       userID,
		ExecutionType:    req.ExecutionType,
		Result:           req.Result,
		ActualResult:     req.ActualResult,
		Duration:         req.Duration,
		Environment:      model.JSONB(req.Environment),
		RelatedDefectID:  req.RelatedDefectID,
		CustomFields:     model.JSONB(req.CustomFields),
		ExecutedAt:       time.Now(),
		CreatedAt:        time.Now(),
	}

	query := `
		INSERT INTO execution_records (execution_id, plan_id, case_id, project_id, executor_id, 
			execution_type, result, actual_result, duration, environment, related_defect_id, 
			custom_fields, executed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := h.db.ExecContext(c.Request.Context(), query,
		exec.ExecutionID, exec.PlanID, exec.CaseID, exec.ProjectID, exec.ExecutorID,
		exec.ExecutionType, exec.Result, exec.ActualResult, exec.Duration, exec.Environment,
		exec.RelatedDefectID, exec.CustomFields, exec.ExecutedAt, exec.CreatedAt)
	if err != nil {
		response.InternalError(c, "创建执行记录失败", "Failed to create execution record")
		return
	}

	response.Success(c, exec)
}

// Get 获取执行记录详情
func (h *ExecutionHandler) Get(c *gin.Context) {
	executionID := c.Param("execution_id")

	var exec model.ExecutionRecord
	query := `SELECT * FROM execution_records WHERE execution_id = $1`
	err := h.db.GetContext(c.Request.Context(), &exec, query, executionID)
	if err == sql.ErrNoRows {
		response.NotFound(c, "执行记录不存在", "Execution record not found")
		return
	}
	if err != nil {
		response.InternalError(c, "查询执行记录失败", "Failed to get execution record")
		return
	}

	response.Success(c, exec)
}

// List 获取执行记录列表
func (h *ExecutionHandler) List(c *gin.Context) {
	planID := c.Param("plan_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	result := c.Query("result")
	executorID := c.Query("executor_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	baseQuery := `SELECT * FROM execution_records WHERE plan_id = $1`
	countQuery := `SELECT COUNT(*) FROM execution_records WHERE plan_id = $1`
	args := []interface{}{planID}
	argIdx := 2

	if result != "" {
		baseQuery += ` AND result = $` + strconv.Itoa(argIdx)
		countQuery += ` AND result = $` + strconv.Itoa(argIdx)
		args = append(args, result)
		argIdx++
	}
	if executorID != "" {
		baseQuery += ` AND executor_id = $` + strconv.Itoa(argIdx)
		countQuery += ` AND executor_id = $` + strconv.Itoa(argIdx)
		args = append(args, executorID)
		argIdx++
	}

	baseQuery += ` ORDER BY executed_at DESC LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, pageSize, offset)

	var executions []model.ExecutionRecord
	err := h.db.SelectContext(c.Request.Context(), &executions, baseQuery, args...)
	if err != nil {
		response.InternalError(c, "查询执行记录列表失败", "Failed to list execution records")
		return
	}

	var total int64
	h.db.GetContext(c.Request.Context(), &total, countQuery, args[:argIdx-2]...)

	response.SuccessWithPagination(c, executions, page, pageSize, total)
}

// Update 更新执行记录
func (h *ExecutionHandler) Update(c *gin.Context) {
	executionID := c.Param("execution_id")

	var req struct {
		Result          string                 `json:"result" binding:"omitempty,oneof=passed failed blocked skipped"`
		ActualResult    string                 `json:"actual_result" binding:"omitempty"`
		Duration        int                    `json:"duration" binding:"omitempty"`
		Environment     map[string]interface{} `json:"environment" binding:"omitempty"`
		RelatedDefectID string                 `json:"related_defect_id" binding:"omitempty"`
		CustomFields    map[string]interface{} `json:"custom_fields" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	query := `
		UPDATE execution_records SET 
			result = COALESCE(NULLIF($1, ''), result),
			actual_result = COALESCE(NULLIF($2, ''), actual_result),
			duration = CASE WHEN $3 = 0 THEN duration ELSE $3 END,
			related_defect_id = COALESCE(NULLIF($4, ''), related_defect_id)
		WHERE execution_id = $5
	`
	args := []interface{}{req.Result, req.ActualResult, req.Duration, req.RelatedDefectID, executionID}

	if req.Environment != nil {
		query = `
			UPDATE execution_records SET 
				result = COALESCE(NULLIF($1, ''), result),
				actual_result = COALESCE(NULLIF($2, ''), actual_result),
				duration = CASE WHEN $3 = 0 THEN duration ELSE $3 END,
				environment = $4,
				related_defect_id = COALESCE(NULLIF($5, ''), related_defect_id)
			WHERE execution_id = $6
		`
		envJSON, _ := json.Marshal(req.Environment)
		args = []interface{}{req.Result, req.ActualResult, req.Duration, envJSON, req.RelatedDefectID, executionID}
	}

	result, err := h.db.ExecContext(c.Request.Context(), query, args...)
	if err != nil {
		response.InternalError(c, "更新执行记录失败", "Failed to update execution record")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		response.NotFound(c, "执行记录不存在", "Execution record not found")
		return
	}

	response.Success(c, gin.H{"execution_id": executionID, "updated": true})
}

// BatchCreate 批量创建执行记录
func (h *ExecutionHandler) BatchCreate(c *gin.Context) {
	planID := c.Param("plan_id")

	var req struct {
		Executions []struct {
			CaseID         string                 `json:"case_id" binding:"required"`
			ExecutionType  string                 `json:"execution_type" binding:"required,oneof=manual auto_web auto_api auto_mobile auto_performance"`
			Result         string                 `json:"result" binding:"required,oneof=passed failed blocked skipped"`
			ActualResult   string                 `json:"actual_result" binding:"omitempty"`
			Duration       int                    `json:"duration" binding:"omitempty"`
			Environment    map[string]interface{} `json:"environment" binding:"omitempty"`
			RelatedDefectID string                 `json:"related_defect_id" binding:"omitempty"`
		} `json:"executions" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	userID := c.GetString("user_id")
	inserted := 0
	var executionIDs []string

	for _, e := range req.Executions {
		// 获取用例所属项目
		var projectID string
		h.db.GetContext(c.Request.Context(), &projectID,
			`SELECT project_id FROM test_cases WHERE case_id = $1`, e.CaseID)

		executionID := uuid.New().String()[:32]
		envJSON, _ := json.Marshal(e.Environment)

		query := `
			INSERT INTO execution_records (execution_id, plan_id, case_id, project_id, executor_id,
				execution_type, result, actual_result, duration, environment, related_defect_id, 
				executed_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		`
		_, err := h.db.ExecContext(c.Request.Context(), query,
			executionID, planID, e.CaseID, projectID, userID,
			e.ExecutionType, e.Result, e.ActualResult, e.Duration, envJSON, e.RelatedDefectID)
		if err == nil {
			inserted++
			executionIDs = append(executionIDs, executionID)
		}
	}

	response.Success(c, gin.H{
		"plan_id":       planID,
		"total":         len(req.Executions),
		"inserted":      inserted,
		"execution_ids": executionIDs,
	})
}

// CreateDefect 从执行记录快速创建缺陷
func (h *ExecutionHandler) CreateDefect(c *gin.Context) {
	executionID := c.Param("execution_id")

	// 获取执行记录
	var exec model.ExecutionRecord
	err := h.db.GetContext(c.Request.Context(), &exec,
		`SELECT * FROM execution_records WHERE execution_id = $1`, executionID)
	if err == sql.ErrNoRows {
		response.NotFound(c, "执行记录不存在", "Execution record not found")
		return
	}
	if err != nil {
		response.InternalError(c, "查询执行记录失败", "Failed to get execution record")
		return
	}

	// 获取用例信息
	var tc model.TestCase
	h.db.GetContext(c.Request.Context(), &tc,
		`SELECT * FROM test_cases WHERE case_id = $1`, exec.CaseID)

	userID := c.GetString("user_id")
	defectID := uuid.New().String()[:32]

	// 创建缺陷
	defect := &model.Defect{
		DefectID:           defectID,
		ProjectID:          exec.ProjectID,
		Title:              "[执行失败] " + tc.Title,
		Description:        exec.ActualResult,
		Severity:           "general",
		Priority:           "P2",
		Status:             "open",
		ReporterID:         userID,
		RelatedCaseID:      exec.CaseID,
		RelatedExecutionID: executionID,
		Language:           "zh-CN",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	query := `
		INSERT INTO defects (defect_id, project_id, title, description, severity, priority,
			status, reporter_id, related_case_id, related_execution_id, language, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err = h.db.ExecContext(c.Request.Context(), query,
		defect.DefectID, defect.ProjectID, defect.Title, defect.Description,
		defect.Severity, defect.Priority, defect.Status, defect.ReporterID,
		defect.RelatedCaseID, defect.RelatedExecutionID, defect.Language,
		defect.CreatedAt, defect.UpdatedAt)
	if err != nil {
		response.InternalError(c, "创建缺陷失败", "Failed to create defect")
		return
	}

	// 更新执行记录关联的缺陷
	h.db.ExecContext(c.Request.Context(),
		`UPDATE execution_records SET related_defect_id = $1 WHERE execution_id = $2`,
		defectID, executionID)

	response.Success(c, gin.H{
		"execution_id": executionID,
		"defect_id":    defectID,
		"defect":       defect,
	})
}

// GetStatistics 获取执行统计
func (h *ExecutionHandler) GetStatistics(c *gin.Context) {
	planID := c.Param("plan_id")

	// 按结果统计
	var resultStats []struct {
		Result string `json:"result" db:"result"`
		Count  int    `json:"count" db:"count"`
	}
	h.db.SelectContext(c.Request.Context(), &resultStats,
		`SELECT result, COUNT(*) as count FROM execution_records WHERE plan_id = $1 GROUP BY result`,
		planID)

	// 按执行类型统计
	var typeStats []struct {
		ExecutionType string `json:"execution_type" db:"execution_type"`
		Count         int    `json:"count" db:"count"`
	}
	h.db.SelectContext(c.Request.Context(), &typeStats,
		`SELECT execution_type, COUNT(*) as count FROM execution_records WHERE plan_id = $1 GROUP BY execution_type`,
		planID)

	// 按执行人统计
	var executorStats []struct {
		ExecutorID string `json:"executor_id" db:"executor_id"`
		Count      int    `json:"count" db:"count"`
	}
	h.db.SelectContext(c.Request.Context(), &executorStats,
		`SELECT executor_id, COUNT(*) as count FROM execution_records WHERE plan_id = $1 GROUP BY executor_id`,
		planID)

	// 计算总数和通过率
	var totalExecuted int
	var passedCount int
	resultMap := make(map[string]int)
	for _, s := range resultStats {
		resultMap[s.Result] = s.Count
		totalExecuted += s.Count
		if s.Result == "passed" {
			passedCount = s.Count
		}
	}

	passRate := 0.0
	if totalExecuted > 0 {
		passRate = float64(passedCount) / float64(totalExecuted) * 100
	}

	response.Success(c, gin.H{
		"plan_id":         planID,
		"total_executed":  totalExecuted,
		"pass_rate":       passRate,
		"by_result":       resultMap,
		"by_type":         typeStats,
		"by_executor":     executorStats,
	})
}

// ListByCase 获取用例的执行历史
func (h *ExecutionHandler) ListByCase(c *gin.Context) {
	caseID := c.Param("case_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	var executions []model.ExecutionRecord
	query := `SELECT * FROM execution_records WHERE case_id = $1 ORDER BY executed_at DESC LIMIT $2 OFFSET $3`
	err := h.db.SelectContext(c.Request.Context(), &executions, query, caseID, pageSize, offset)
	if err != nil {
		response.InternalError(c, "查询执行历史失败", "Failed to get execution history")
		return
	}

	var total int64
	h.db.GetContext(c.Request.Context(), &total,
		`SELECT COUNT(*) FROM execution_records WHERE case_id = $1`, caseID)

	response.SuccessWithPagination(c, executions, page, pageSize, total)
}

// 外部自动化接口预留

// AutoRun 触发自动化执行（预留接口）
func (h *ExecutionHandler) AutoRun(c *gin.Context) {
	planID := c.Param("plan_id")

	var req struct {
		ExecutionType string `json:"execution_type" binding:"required,oneof=auto_web auto_api auto_mobile auto_performance"`
		Environment   string `json:"environment" binding:"omitempty"`
		Config        string `json:"config" binding:"omitempty"` // JSON配置
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	// TODO: 实现自动化执行逻辑
	// 1. 查询计划下的用例
	// 2. 调用对应的自动化框架
	// 3. 创建异步任务
	// 4. 返回任务ID

	taskID := uuid.New().String()[:32]

	response.Success(c, gin.H{
		"plan_id":        planID,
		"task_id":        taskID,
		"status":         "pending",
		"execution_type": req.ExecutionType,
		"message":        "自动化执行任务已创建",
	})
}

// AutoStatus 查询自动化执行状态（预留接口）
func (h *ExecutionHandler) AutoStatus(c *gin.Context) {
	taskID := c.Param("task_id")

	// TODO: 实现状态查询逻辑
	// 从任务队列或数据库查询任务状态

	response.Success(c, gin.H{
		"task_id":    taskID,
		"status":     "completed",
		"progress":   100,
		"message":    "自动化执行已完成",
		"created_at": time.Now().Add(-time.Hour),
		"updated_at": time.Now(),
	})
}

// ExternalResult 外部框架回传执行结果（预留接口）
func (h *ExecutionHandler) ExternalResult(c *gin.Context) {
	executionID := c.Param("execution_id")

	var req struct {
		Result       string `json:"result" binding:"required,oneof=passed failed blocked skipped"`
		ActualResult string `json:"actual_result" binding:"omitempty"`
		Duration     int    `json:"duration" binding:"omitempty"`
		LogURL       string `json:"log_url" binding:"omitempty"`
		ScreenshotURL string `json:"screenshot_url" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	// 更新执行记录
	query := `
		UPDATE execution_records SET 
			result = $1,
			actual_result = COALESCE($2, actual_result),
			duration = CASE WHEN $3 = 0 THEN duration ELSE $3 END
		WHERE execution_id = $4
	`
	_, err := h.db.ExecContext(c.Request.Context(), query,
		req.Result, req.ActualResult, req.Duration, executionID)
	if err != nil {
		response.InternalError(c, "更新执行结果失败", "Failed to update execution result")
		return
	}

	response.Success(c, gin.H{
		"execution_id": executionID,
		"result":       req.Result,
		"updated":      true,
	})
}

var _ = strconv.Itoa