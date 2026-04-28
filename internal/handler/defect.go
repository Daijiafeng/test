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

type DefectHandler struct {
	db *repository.DB
}

func NewDefectHandler(db *repository.DB) *DefectHandler {
	return &DefectHandler{db: db}
}

// Create 创建缺陷
func (h *DefectHandler) Create(c *gin.Context) {
	projectID := c.Param("project_id")

	var req struct {
		Title              string                 `json:"title" binding:"required,min=5,max=500"`
		TitleEn            string                 `json:"title_en" binding:"omitempty,max=500"`
		Description        string                 `json:"description" binding:"omitempty"`
		Severity           string                 `json:"severity" binding:"required,oneof=fatal critical general minor suggestion"`
		Priority           string                 `json:"priority" binding:"required,oneof=P0 P1 P2 P3"`
		ModuleID           string                 `json:"module_id" binding:"omitempty"`
		RelatedCaseID      string                 `json:"related_case_id" binding:"omitempty"`
		RelatedExecutionID string                 `json:"related_execution_id" binding:"omitempty"`
		Tags               []string               `json:"tags" binding:"omitempty"`
		CustomFields       map[string]interface{} `json:"custom_fields" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	userID := c.GetString("user_id")

	defect := &model.Defect{
		DefectID:           uuid.New().String()[:32],
		ProjectID:          projectID,
		Title:              req.Title,
		TitleEn:            req.TitleEn,
		Description:        req.Description,
		Severity:           req.Severity,
		Priority:           req.Priority,
		ModuleID:           req.ModuleID,
		Status:             "open",
		ReporterID:         userID,
		RelatedCaseID:      req.RelatedCaseID,
		RelatedExecutionID: req.RelatedExecutionID,
		Tags:               req.Tags,
		Language:           "zh-CN",
		CustomFields:       model.JSONB(req.CustomFields),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	query := `
		INSERT INTO defects (defect_id, project_id, title, title_en, description, severity, priority,
			module_id, status, reporter_id, related_case_id, related_execution_id, tags, language,
			custom_fields, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`
	_, err := h.db.ExecContext(c.Request.Context(), query,
		defect.DefectID, defect.ProjectID, defect.Title, defect.TitleEn, defect.Description,
		defect.Severity, defect.Priority, defect.ModuleID, defect.Status, defect.ReporterID,
		defect.RelatedCaseID, defect.RelatedExecutionID, defect.Tags, defect.Language,
		defect.CustomFields, defect.CreatedAt, defect.UpdatedAt)
	if err != nil {
		response.InternalError(c, "创建缺陷失败", "Failed to create defect")
		return
	}

	response.Success(c, defect)
}

// Get 获取缺陷详情
func (h *DefectHandler) Get(c *gin.Context) {
	defectID := c.Param("defect_id")

	var defect model.Defect
	query := `SELECT * FROM defects WHERE defect_id = $1`
	err := h.db.GetContext(c.Request.Context(), &defect, query, defectID)
	if err == sql.ErrNoRows {
		response.NotFound(c, "缺陷不存在", "Defect not found")
		return
	}
	if err != nil {
		response.InternalError(c, "查询缺陷失败", "Failed to get defect")
		return
	}

	response.Success(c, defect)
}

// List 获取缺陷列表（支持筛选）
func (h *DefectHandler) List(c *gin.Context) {
	projectID := c.Param("project_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	status := c.Query("status")
	severity := c.Query("severity")
	priority := c.Query("priority")
	moduleID := c.Query("module_id")
	reporterID := c.Query("reporter_id")
	assigneeID := c.Query("assignee_id")
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	baseQuery := `SELECT * FROM defects WHERE project_id = $1`
	countQuery := `SELECT COUNT(*) FROM defects WHERE project_id = $1`
	args := []interface{}{projectID}
	argIdx := 2

	if status != "" {
		baseQuery += ` AND status = $` + strconv.Itoa(argIdx)
		countQuery += ` AND status = $` + strconv.Itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	if severity != "" {
		baseQuery += ` AND severity = $` + strconv.Itoa(argIdx)
		countQuery += ` AND severity = $` + strconv.Itoa(argIdx)
		args = append(args, severity)
		argIdx++
	}
	if priority != "" {
		baseQuery += ` AND priority = $` + strconv.Itoa(argIdx)
		countQuery += ` AND priority = $` + strconv.Itoa(argIdx)
		args = append(args, priority)
		argIdx++
	}
	if moduleID != "" {
		baseQuery += ` AND module_id = $` + strconv.Itoa(argIdx)
		countQuery += ` AND module_id = $` + strconv.Itoa(argIdx)
		args = append(args, moduleID)
		argIdx++
	}
	if reporterID != "" {
		baseQuery += ` AND reporter_id = $` + strconv.Itoa(argIdx)
		countQuery += ` AND reporter_id = $` + strconv.Itoa(argIdx)
		args = append(args, reporterID)
		argIdx++
	}
	if assigneeID != "" {
		baseQuery += ` AND assignee_id = $` + strconv.Itoa(argIdx)
		countQuery += ` AND assignee_id = $` + strconv.Itoa(argIdx)
		args = append(args, assigneeID)
		argIdx++
	}
	if keyword != "" {
		baseQuery += ` AND (title ILIKE $` + strconv.Itoa(argIdx) + ` OR description ILIKE $` + strconv.Itoa(argIdx) + `)`
		countQuery += ` AND (title ILIKE $` + strconv.Itoa(argIdx) + ` OR description ILIKE $` + strconv.Itoa(argIdx) + `)`
		args = append(args, "%"+keyword+"%")
		argIdx++
	}

	baseQuery += ` ORDER BY priority, severity, created_at DESC LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, pageSize, offset)

	var defects []model.Defect
	err := h.db.SelectContext(c.Request.Context(), &defects, baseQuery, args...)
	if err != nil {
		response.InternalError(c, "查询缺陷列表失败", "Failed to list defects")
		return
	}

	var total int64
	h.db.GetContext(c.Request.Context(), &total, countQuery, args[:argIdx-2]...)

	response.SuccessWithPagination(c, defects, page, pageSize, total)
}

// Update 更新缺陷
func (h *DefectHandler) Update(c *gin.Context) {
	defectID := c.Param("defect_id")

	var req struct {
		Title        string                 `json:"title" binding:"omitempty,min=5,max=500"`
		TitleEn      string                 `json:"title_en" binding:"omitempty,max=500"`
		Description  string                 `json:"description" binding:"omitempty"`
		Severity     string                 `json:"severity" binding:"omitempty,oneof=fatal critical general minor suggestion"`
		Priority     string                 `json:"priority" binding:"omitempty,oneof=P0 P1 P2 P3"`
		ModuleID     string                 `json:"module_id" binding:"omitempty"`
		Tags         []string               `json:"tags" binding:"omitempty"`
		CustomFields map[string]interface{} `json:"custom_fields" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	query := `
		UPDATE defects SET 
			title = COALESCE(NULLIF($1, ''), title),
			title_en = COALESCE(NULLIF($2, ''), title_en),
			description = COALESCE(NULLIF($3, ''), description),
			severity = COALESCE(NULLIF($4, ''), severity),
			priority = COALESCE(NULLIF($5, ''), priority),
			module_id = COALESCE(NULLIF($6, ''), module_id),
			tags = COALESCE($7, tags),
			custom_fields = COALESCE($8, custom_fields),
			updated_at = NOW()
		WHERE defect_id = $9
	`
	tagsJSON := req.Tags
	if tagsJSON == nil {
		tagsJSON = []string{}
	}
	customFields := model.JSONB{}
	if req.CustomFields != nil {
		customFields = model.JSONB(req.CustomFields)
	}

	result, err := h.db.ExecContext(c.Request.Context(), query,
		req.Title, req.TitleEn, req.Description, req.Severity, req.Priority,
		req.ModuleID, tagsJSON, customFields, defectID)
	if err != nil {
		response.InternalError(c, "更新缺陷失败", "Failed to update defect")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		response.NotFound(c, "缺陷不存在", "Defect not found")
		return
	}

	response.Success(c, gin.H{"defect_id": defectID, "updated": true})
}

// TransitionStatus 状态流转
func (h *DefectHandler) TransitionStatus(c *gin.Context) {
	defectID := c.Param("defect_id")

	var req struct {
		TargetStatus string `json:"target_status" binding:"required,oneof=open in_progress fixed verified closed reopened rejected"`
		Comment      string `json:"comment" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	userID := c.GetString("user_id")

	// 获取当前缺陷
	var current model.Defect
	err := h.db.GetContext(c.Request.Context(), &current,
		`SELECT * FROM defects WHERE defect_id = $1`, defectID)
	if err == sql.ErrNoRows {
		response.NotFound(c, "缺陷不存在", "Defect not found")
		return
	}

	// 验证状态流转规则
	validTransitions := map[string][]string{
		"open":         {"in_progress", "rejected"},
		"in_progress":  {"fixed", "rejected"},
		"fixed":        {"verified", "in_progress"},
		"verified":     {"closed", "in_progress"},
		"closed":       {"reopened"},
		"reopened":     {"in_progress"},
		"rejected":     {"reopened"},
	}

	allowed := false
	for _, s := range validTransitions[current.Status] {
		if s == req.TargetStatus {
			allowed = true
			break
		}
	}

	if !allowed {
		response.Error(c, 5002, "不允许的状态流转", "Invalid status transition")
		return
	}

	// 更新状态
	now := time.Now()
	query := `UPDATE defects SET status = $1, updated_at = $2 WHERE defect_id = $3`
	_, err = h.db.ExecContext(c.Request.Context(), query, req.TargetStatus, now, defectID)
	if err != nil {
		response.InternalError(c, "状态流转失败", "Failed to transition status")
		return
	}

	// 记录操作历史
	auditID := uuid.New().String()[:32]
	auditQuery := `
		INSERT INTO audit_logs (log_id, entity_type, entity_id, operation_type, operator_id, 
			before_value, after_value, details, created_at)
		VALUES ($1, 'defect', $2, 'status_transition', $3, $4, $5, $6, NOW())
	`
	h.db.ExecContext(c.Request.Context(), auditQuery,
		auditID, defectID, userID, current.Status, req.TargetStatus, req.Comment)

	response.Success(c, gin.H{
		"defect_id":     defectID,
		"from_status":   current.Status,
		"to_status":     req.TargetStatus,
		"updated_at":    now,
	})
}

// Assign 分配缺陷
func (h *DefectHandler) Assign(c *gin.Context) {
	defectID := c.Param("defect_id")

	var req struct {
		AssigneeID string `json:"assignee_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	query := `UPDATE defects SET assignee_id = $1, updated_at = NOW() WHERE defect_id = $2`
	result, err := h.db.ExecContext(c.Request.Context(), query, req.AssigneeID, defectID)
	if err != nil {
		response.InternalError(c, "分配缺陷失败", "Failed to assign defect")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		response.NotFound(c, "缺陷不存在", "Defect not found")
		return
	}

	response.Success(c, gin.H{"defect_id": defectID, "assignee_id": req.AssigneeID})
}

// AddComment 添加评论
func (h *DefectHandler) AddComment(c *gin.Context) {
	defectID := c.Param("defect_id")

	var req struct {
		Content    string `json:"content" binding:"required,min=1,max=5000"`
		Visibility string `json:"visibility" binding:"omitempty,oneof=public internal"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	userID := c.GetString("user_id")
	commentID := uuid.New().String()[:32]

	query := `
		INSERT INTO defect_comments (comment_id, defect_id, author_id, content, visibility, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	_, err := h.db.ExecContext(c.Request.Context(), query,
		commentID, defectID, userID, req.Content, req.Visibility)
	if err != nil {
		response.InternalError(c, "添加评论失败", "Failed to add comment")
		return
	}

	response.Success(c, gin.H{
		"comment_id": commentID,
		"defect_id":  defectID,
		"author_id":  userID,
		"content":    req.Content,
	})
}

// ListComments 获取评论列表
func (h *DefectHandler) ListComments(c *gin.Context) {
	defectID := c.Param("defect_id")

	var comments []struct {
		CommentID  string    `json:"comment_id" db:"comment_id"`
		AuthorID   string    `json:"author_id" db:"author_id"`
		Content    string    `json:"content" db:"content"`
		Visibility string    `json:"visibility" db:"visibility"`
		CreatedAt  time.Time `json:"created_at" db:"created_at"`
	}

	query := `SELECT comment_id, author_id, content, visibility, created_at 
		FROM defect_comments WHERE defect_id = $1 ORDER BY created_at ASC`
	err := h.db.SelectContext(c.Request.Context(), &comments, query, defectID)
	if err != nil {
		response.InternalError(c, "查询评论失败", "Failed to list comments")
		return
	}

	response.Success(c, comments)
}

// GetHistory 获取缺陷历史
func (h *DefectHandler) GetHistory(c *gin.Context) {
	defectID := c.Param("defect_id")

	var history []struct {
		LogID           int       `json:"log_id" db:"log_id"`
		OperationType   string    `json:"operation_type" db:"operation_type"`
		OperatorID      string    `json:"operator_id" db:"operator_id"`
		BeforeValue     string    `json:"before_value" db:"before_value"`
		AfterValue      string    `json:"after_value" db:"after_value"`
		Details         string    `json:"details" db:"details"`
		CreatedAt       time.Time `json:"created_at" db:"created_at"`
	}

	query := `SELECT log_id, operation_type, operator_id, before_value, after_value, details, created_at 
		FROM audit_logs WHERE entity_type = 'defect' AND entity_id = $1 ORDER BY created_at DESC`
	err := h.db.SelectContext(c.Request.Context(), &history, query, defectID)
	if err != nil {
		response.InternalError(c, "查询历史失败", "Failed to get history")
		return
	}

	response.Success(c, history)
}

// Statistics 缺陷统计
func (h *DefectHandler) Statistics(c *gin.Context) {
	projectID := c.Param("project_id")

	// 按状态统计
	var statusStats []struct {
		Status string `json:"status" db:"status"`
		Count  int    `json:"count" db:"count"`
	}
	h.db.SelectContext(c.Request.Context(), &statusStats,
		`SELECT status, COUNT(*) as count FROM defects WHERE project_id = $1 GROUP BY status`,
		projectID)

	// 按严重程度统计
	var severityStats []struct {
		Severity string `json:"severity" db:"severity"`
		Count    int    `json:"count" db:"count"`
	}
	h.db.SelectContext(c.Request.Context(), &severityStats,
		`SELECT severity, COUNT(*) as count FROM defects WHERE project_id = $1 GROUP BY severity`,
		projectID)

	// 按优先级统计
	var priorityStats []struct {
		Priority string `json:"priority" db:"priority"`
		Count    int    `json:"count" db:"count"`
	}
	h.db.SelectContext(c.Request.Context(), &priorityStats,
		`SELECT priority, COUNT(*) as count FROM defects WHERE project_id = $1 GROUP BY priority`,
		projectID)

	// 总数
	var total int64
	h.db.GetContext(c.Request.Context(), &total,
		`SELECT COUNT(*) FROM defects WHERE project_id = $1`, projectID)

	// 未关闭数
	var openCount int64
	h.db.GetContext(c.Request.Context(), &openCount,
		`SELECT COUNT(*) FROM defects WHERE project_id = $1 AND status NOT IN ('closed', 'rejected')`,
		projectID)

	statusMap := make(map[string]int)
	for _, s := range statusStats {
		statusMap[s.Status] = s.Count
	}

	severityMap := make(map[string]int)
	for _, s := range severityStats {
		severityMap[s.Severity] = s.Count
	}

	priorityMap := make(map[string]int)
	for _, s := range priorityStats {
		priorityMap[s.Priority] = s.Count
	}

	response.Success(c, gin.H{
		"project_id":   projectID,
		"total":        total,
		"open_count":   openCount,
		"closed_count": total - openCount,
		"by_status":    statusMap,
		"by_severity":  severityMap,
		"by_priority":  priorityMap,
	})
}

// Search 全文搜索缺陷
func (h *DefectHandler) Search(c *gin.Context) {
	projectID := c.Param("project_id")
	keyword := c.Query("q")

	if keyword == "" {
		response.BadRequest(c, "请输入搜索关键词", "Search keyword is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	offset := (page - 1) * pageSize

	var defects []model.Defect
	query := `
		SELECT * FROM defects 
		WHERE project_id = $1 AND (title ILIKE $2 OR description ILIKE $2)
		ORDER BY updated_at DESC LIMIT $3 OFFSET $4
	`
	err := h.db.SelectContext(c.Request.Context(), &defects, query, projectID, "%"+keyword+"%", pageSize, offset)
	if err != nil {
		response.InternalError(c, "搜索失败", "Search failed")
		return
	}

	var total int64
	h.db.GetContext(c.Request.Context(), &total,
		`SELECT COUNT(*) FROM defects WHERE project_id = $1 AND (title ILIKE $2 OR description ILIKE $2)`,
		projectID, "%"+keyword+"%")

	response.SuccessWithPagination(c, defects, page, pageSize, total)
}

// BatchUpdate 批量更新缺陷
func (h *DefectHandler) BatchUpdate(c *gin.Context) {
	projectID := c.Param("project_id")

	var req struct {
		DefectIDs  []string `json:"defect_ids" binding:"required,min=1"`
		Status     string   `json:"status" binding:"omitempty,oneof=open in_progress fixed verified closed reopened rejected"`
		Priority   string   `json:"priority" binding:"omitempty,oneof=P0 P1 P2 P3"`
		AssigneeID string   `json:"assignee_id" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	updates := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Status != "" {
		updates = append(updates, `status = $`+strconv.Itoa(argIdx))
		args = append(args, req.Status)
		argIdx++
	}
	if req.Priority != "" {
		updates = append(updates, `priority = $`+strconv.Itoa(argIdx))
		args = append(args, req.Priority)
		argIdx++
	}
	if req.AssigneeID != "" {
		updates = append(updates, `assignee_id = $`+strconv.Itoa(argIdx))
		args = append(args, req.AssigneeID)
		argIdx++
	}

	if len(updates) == 0 {
		response.BadRequest(c, "请提供更新内容", "No update fields provided")
		return
	}

	updates = append(updates, `updated_at = NOW()`)
	args = append(args, req.DefectIDs, projectID)

	query := `UPDATE defects SET ` + joinStrings(updates, ", ") + 
		` WHERE defect_id = ANY($` + strconv.Itoa(argIdx) + `) AND project_id = $` + strconv.Itoa(argIdx+1)

	result, err := h.db.ExecContext(c.Request.Context(), query, args...)
	if err != nil {
		response.InternalError(c, "批量更新失败", "Failed to batch update")
		return
	}

	rows, _ := result.RowsAffected()

	response.Success(c, gin.H{
		"project_id": projectID,
		"total":      len(req.DefectIDs),
		"updated":    rows,
	})
}

// Export 导出缺陷
func (h *DefectHandler) Export(c *gin.Context) {
	projectID := c.Param("project_id")

	// 获取所有缺陷
	var defects []model.Defect
	query := `SELECT * FROM defects WHERE project_id = $1 ORDER BY priority, severity, created_at DESC`
	err := h.db.SelectContext(c.Request.Context(), &defects, query, projectID)
	if err != nil {
		response.InternalError(c, "导出失败", "Export failed")
		return
	}

	// 返回JSON格式（前端可转换为Excel/CSV）
	response.Success(c, gin.H{
		"project_id": projectID,
		"total":      len(defects),
		"defects":    defects,
		"format":     "json",
	})
}

// 确保 strconv 被引用
var _ = strconv.Itoa