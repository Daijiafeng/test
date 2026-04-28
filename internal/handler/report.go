package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"testmind/internal/model"
	"testmind/internal/repository"
	"testmind/pkg/response"
)

type ReportHandler struct {
	db *repository.DB
}

func NewReportHandler(db *repository.DB) *ReportHandler {
	return &ReportHandler{db: db}
}

// Generate 生成测试报告
func (h *ReportHandler) Generate(c *gin.Context) {
	projectID := c.Param("project_id")

	var req struct {
		PlanID         string   `json:"plan_id" binding:"required"`
		Title          string   `json:"title" binding:"omitempty,max=200"`
		ReportType     string   `json:"report_type" binding:"required,oneof=summary detailed execution defect custom"`
		IncludeCases   bool     `json:"include_cases"`
		IncludeDefects bool     `json:"include_defects"`
		DateRange      []string `json:"date_range" binding:"omitempty,len=2"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	userID := c.GetString("user_id")
	reportID := uuid.New().String()[:32]

	// 生成报告标题
	title := req.Title
	if title == "" {
		title = "测试报告_" + time.Now().Format("20060102")
	}

	// 计算开始和结束时间
	var startDate, endDate time.Time
	if len(req.DateRange) == 2 {
		startDate, _ = time.Parse("2006-01-02", req.DateRange[0])
		endDate, _ = time.Parse("2006-01-02", req.DateRange[1])
	} else {
		// 默认最近一周
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -7)
	}

	// 收集统计数据
	stats := h.collectStatistics(c.Request.Context(), req.PlanID, startDate, endDate)

	// 构建报告内容
	content := h.buildReportContent(stats, req.ReportType, req.IncludeCases, req.IncludeDefects)

	report := &model.TestReport{
		ReportID:     reportID,
		ProjectID:    projectID,
		PlanID:       req.PlanID,
		Title:        title,
		ReportType:   req.ReportType,
		Status:       "generated",
		GeneratedBy:  userID,
		GeneratedAt:  time.Now(),
		StartDate:    startDate,
		EndDate:      endDate,
		Summary:      stats["summary"],
		Details:      model.JSONB(stats),
		Language:     "zh-CN",
		CreatedAt:    time.Now(),
	}

	query := `
		INSERT INTO test_reports (report_id, project_id, plan_id, title, report_type, status,
			generated_by, generated_at, start_date, end_date, summary, details, language, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := h.db.ExecContext(c.Request.Context(), query,
		report.ReportID, report.ProjectID, report.PlanID, report.Title, report.ReportType,
		report.Status, report.GeneratedBy, report.GeneratedAt, report.StartDate, report.EndDate,
		report.Summary, report.Details, report.Language, report.CreatedAt)
	if err != nil {
		response.InternalError(c, "生成报告失败", "Failed to generate report")
		return
	}

	response.Success(c, gin.H{
		"report_id":     reportID,
		"title":         title,
		"report_type":   req.ReportType,
		"generated_at":  report.GeneratedAt,
		"statistics":    stats,
	})
}

// Get 获取报告详情
func (h *ReportHandler) Get(c *gin.Context) {
	reportID := c.Param("report_id")

	var report model.TestReport
	query := `SELECT * FROM test_reports WHERE report_id = $1`
	err := h.db.GetContext(c.Request.Context(), &report, query, reportID)
	if err == sql.ErrNoRows {
		response.NotFound(c, "报告不存在", "Report not found")
		return
	}
	if err != nil {
		response.InternalError(c, "查询报告失败", "Failed to get report")
		return
	}

	response.Success(c, report)
}

// List 获取报告列表
func (h *ReportHandler) List(c *gin.Context) {
	projectID := c.Param("project_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	reportType := c.Query("report_type")
	planID := c.Query("plan_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	baseQuery := `SELECT * FROM test_reports WHERE project_id = $1`
	countQuery := `SELECT COUNT(*) FROM test_reports WHERE project_id = $1`
	args := []interface{}{projectID}
	argIdx := 2

	if reportType != "" {
		baseQuery += ` AND report_type = $` + strconv.Itoa(argIdx)
		countQuery += ` AND report_type = $` + strconv.Itoa(argIdx)
		args = append(args, reportType)
		argIdx++
	}
	if planID != "" {
		baseQuery += ` AND plan_id = $` + strconv.Itoa(argIdx)
		countQuery += ` AND plan_id = $` + strconv.Itoa(argIdx)
		args = append(args, planID)
		argIdx++
	}

	baseQuery += ` ORDER BY generated_at DESC LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, pageSize, offset)

	var reports []model.TestReport
	err := h.db.SelectContext(c.Request.Context(), &reports, baseQuery, args...)
	if err != nil {
		response.InternalError(c, "查询报告列表失败", "Failed to list reports")
		return
	}

	var total int64
	h.db.GetContext(c.Request.Context(), &total, countQuery, args[:argIdx-2]...)

	response.SuccessWithPagination(c, reports, page, pageSize, total)
}

// Update 更新报告
func (h *ReportHandler) Update(c *gin.Context) {
	reportID := c.Param("report_id")

	var req struct {
		Title    string                 `json:"title" binding:"omitempty,min=2,max=200"`
		Summary  string                 `json:"summary" binding:"omitempty"`
		Details  map[string]interface{} `json:"details" binding:"omitempty"`
		Status   string                 `json:"status" binding:"omitempty,oneof=draft generated published archived"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	query := `
		UPDATE test_reports SET 
			title = COALESCE(NULLIF($1, ''), title),
			summary = COALESCE(NULLIF($2, ''), summary),
			details = COALESCE($3, details),
			status = COALESCE(NULLIF($4, ''), status),
			updated_at = NOW()
		WHERE report_id = $5
	`
	detailsJSON := model.JSONB{}
	if req.Details != nil {
		detailsJSON = model.JSONB(req.Details)
	}

	result, err := h.db.ExecContext(c.Request.Context(), query,
		req.Title, req.Summary, detailsJSON, req.Status, reportID)
	if err != nil {
		response.InternalError(c, "更新报告失败", "Failed to update report")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		response.NotFound(c, "报告不存在", "Report not found")
		return
	}

	response.Success(c, gin.H{"report_id": reportID, "updated": true})
}

// Delete 删除报告
func (h *ReportHandler) Delete(c *gin.Context) {
	reportID := c.Param("report_id")

	result, err := h.db.ExecContext(c.Request.Context(),
		`DELETE FROM test_reports WHERE report_id = $1`, reportID)
	if err != nil {
		response.InternalError(c, "删除报告失败", "Failed to delete report")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		response.NotFound(c, "报告不存在", "Report not found")
		return
	}

	response.Success(c, gin.H{"report_id": reportID, "deleted": true})
}

// Export 导出报告
func (h *ReportHandler) Export(c *gin.Context) {
	reportID := c.Param("report_id")
	format := c.DefaultQuery("format", "json")

	var report model.TestReport
	err := h.db.GetContext(c.Request.Context(), &report,
		`SELECT * FROM test_reports WHERE report_id = $1`, reportID)
	if err == sql.ErrNoRows {
		response.NotFound(c, "报告不存在", "Report not found")
		return
	}

	switch format {
	case "json":
		response.Success(c, gin.H{
			"report_id": reportID,
			"format":    "json",
			"data":      report,
		})
	case "markdown":
		md := h.convertToMarkdown(report)
		response.Success(c, gin.H{
			"report_id": reportID,
			"format":    "markdown",
			"content":   md,
		})
	default:
		response.BadRequest(c, "不支持的导出格式", "Unsupported export format")
	}
}

// Share 分享报告（飞书文档）
func (h *ReportHandler) Share(c *gin.Context) {
	reportID := c.Param("report_id")

	var req struct {
		Target string `json:"target" binding:"required,oneof=feishu_doc feishu_bitable email"`
		Title  string `json:"title" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	var report model.TestReport
	err := h.db.GetContext(c.Request.Context(), &report,
		`SELECT * FROM test_reports WHERE report_id = $1`, reportID)
	if err == sql.ErrNoRows {
		response.NotFound(c, "报告不存在", "Report not found")
		return
	}

	// TODO: 实际调用飞书API创建文档
	// 这里返回模拟结果
	shareURL := ""
	switch req.Target {
	case "feishu_doc":
		shareURL = fmt.Sprintf("https://www.feishu.cn/docx/%s", uuid.New().String()[:16])
	case "feishu_bitable":
		shareURL = fmt.Sprintf("https://www.feishu.cn/base/%s", uuid.New().String()[:16])
	}

	// 更新报告状态为已发布
	h.db.ExecContext(c.Request.Context(),
		`UPDATE test_reports SET status = 'published', share_url = $1, updated_at = NOW() WHERE report_id = $2`,
		shareURL, reportID)

	response.Success(c, gin.H{
		"report_id": reportID,
		"target":    req.Target,
		"share_url": shareURL,
		"status":    "published",
	})
}

// collectStatistics 收集统计数据
func (h *ReportHandler) collectStatistics(ctx interface{}, planID string, startDate, endDate time.Time) map[string]interface{} {
	stats := make(map[string]interface{})

	// 用例统计
	var caseStats struct {
		Total    int
		Passed   int
		Failed   int
		Blocked  int
		Skipped  int
	}

	// 执行统计
	query := `
		SELECT COUNT(*) as total,
			SUM(CASE WHEN result = 'passed' THEN 1 ELSE 0 END) as passed,
			SUM(CASE WHEN result = 'failed' THEN 1 ELSE 0 END) as failed,
			SUM(CASE WHEN result = 'blocked' THEN 1 ELSE 0 END) as blocked,
			SUM(CASE WHEN result = 'skipped' THEN 1 ELSE 0 END) as skipped
		FROM execution_records WHERE plan_id = $1
	`
	h.db.GetContext(ctx, &caseStats, query, planID)

	passRate := 0.0
	if caseStats.Total > 0 {
		passRate = float64(caseStats.Passed) / float64(caseStats.Total) * 100
	}

	stats["execution"] = gin.H{
		"total":    caseStats.Total,
		"passed":   caseStats.Passed,
		"failed":   caseStats.Failed,
		"blocked":  caseStats.Blocked,
		"skipped":  caseStats.Skipped,
		"pass_rate": fmt.Sprintf("%.2f%%", passRate),
	}

	// 缺陷统计
	var defectStats struct {
		Total    int
		Fatal    int
		Critical int
		General  int
		Open     int
		Closed   int
	}

	defectQuery := `
		SELECT COUNT(*) as total,
			SUM(CASE WHEN severity = 'fatal' THEN 1 ELSE 0 END) as fatal,
			SUM(CASE WHEN severity = 'critical' THEN 1 ELSE 0 END) as critical,
			SUM(CASE WHEN severity = 'general' THEN 1 ELSE 0 END) as general,
			SUM(CASE WHEN status NOT IN ('closed', 'rejected') THEN 1 ELSE 0 END) as open,
			SUM(CASE WHEN status = 'closed' THEN 1 ELSE 0 END) as closed
		FROM defects WHERE project_id = (SELECT project_id FROM test_plans WHERE plan_id = $1)
	`
	h.db.GetContext(ctx, &defectStats, defectQuery, planID)

	stats["defects"] = gin.H{
		"total":    defectStats.Total,
		"fatal":    defectStats.Fatal,
		"critical": defectStats.Critical,
		"general":  defectStats.General,
		"open":     defectStats.Open,
		"closed":   defectStats.Closed,
	}

	// 汇总文字
	summary := fmt.Sprintf("本次测试共执行%d条用例，通过%d条，失败%d条，通过率%s。发现%d个缺陷，其中致命%d个，严重%d个，一般%d个，未关闭%d个。",
		caseStats.Total, caseStats.Passed, caseStats.Failed, fmt.Sprintf("%.2f%%", passRate),
		defectStats.Total, defectStats.Fatal, defectStats.Critical, defectStats.General, defectStats.Open)

	stats["summary"] = summary

	return stats
}

// buildReportContent 构建报告内容
func (h *ReportHandler) buildReportContent(stats map[string]interface{}, reportType string, includeCases, includeDefects bool) map[string]interface{} {
	content := make(map[string]interface{})

	content["statistics"] = stats
	content["report_type"] = reportType

	// TODO: 根据类型添加详细的用例列表和缺陷列表

	return content
}

// convertToMarkdown 转换为Markdown格式
func (h *ReportHandler) convertToMarkdown(report model.TestReport) string {
	md := fmt.Sprintf("# %s\n\n", report.Title)
	md += fmt.Sprintf("**生成时间：** %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))
	md += fmt.Sprintf("**报告类型：** %s\n\n", report.ReportType)
	md += fmt.Sprintf("**测试周期：** %s ~ %s\n\n", report.StartDate.Format("2006-01-02"), report.EndDate.Format("2006-01-02"))

	md += "## 测试总结\n\n"
	md += report.Summary + "\n\n"

	// 添加统计数据
	if report.Details != nil {
		detailsJSON, _ := json.MarshalIndent(report.Details, "", "  ")
		md += "## 详细数据\n\n```json\n" + string(detailsJSON) + "\n```\n"
	}

	return md
}

var _ = strconv.Itoa