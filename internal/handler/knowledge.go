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

type KnowledgeHandler struct {
	db *repository.DB
}

func NewKnowledgeHandler(db *repository.DB) *KnowledgeHandler {
	return &KnowledgeHandler{db: db}
}

// Create 创建知识库文档
func (h *KnowledgeHandler) Create(c *gin.Context) {
	projectID := c.Param("project_id")

	var req struct {
		Title        string   `json:"title" binding:"required,min=2,max=500"`
		DocType      string   `json:"doc_type" binding:"required,oneof=requirement design api other"`
		SourceType   string   `json:"source_type" binding:"required,oneof=manual feishu_doc upload"`
		FeishuDocID  string   `json:"feishu_doc_id" binding:"omitempty"`
		FeishuDocURL string   `json:"feishu_doc_url" binding:"omitempty"`
		Content      string   `json:"content" binding:"omitempty"`
		Summary      string   `json:"summary" binding:"omitempty"`
		Tags         []string `json:"tags" binding:"omitempty"`
		Language     string   `json:"language" binding:"omitempty,oneof=zh-CN en-US"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	userID := c.GetString("user_id")
	docID := uuid.New().String()[:32]

	doc := &model.KnowledgeDoc{
		DocID:        docID,
		ProjectID:    projectID,
		Title:        req.Title,
		DocType:      req.DocType,
		SourceType:   req.SourceType,
		FeishuDocID:  req.FeishuDocID,
		FeishuDocURL: req.FeishuDocURL,
		Content:      req.Content,
		Summary:      req.Summary,
		Tags:         req.Tags,
		Language:     stringOrDefault(req.Language, "zh-CN"),
		CreatedBy:    userID,
		Version:      1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO knowledge_docs (doc_id, project_id, title, doc_type, source_type, 
			feishu_doc_id, feishu_doc_url, content, summary, tags, language, 
			created_by, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := h.db.ExecContext(c.Request.Context(), query,
		doc.DocID, doc.ProjectID, doc.Title, doc.DocType, doc.SourceType,
		doc.FeishuDocID, doc.FeishuDocURL, doc.Content, doc.Summary, doc.Tags,
		doc.Language, doc.CreatedBy, doc.Version, doc.CreatedAt, doc.UpdatedAt)
	if err != nil {
		response.InternalError(c, "创建文档失败", "Failed to create document")
		return
	}

	response.Success(c, doc)
}

// Get 获取文档详情
func (h *KnowledgeHandler) Get(c *gin.Context) {
	docID := c.Param("doc_id")

	var doc model.KnowledgeDoc
	query := `SELECT * FROM knowledge_docs WHERE doc_id = $1`
	err := h.db.GetContext(c.Request.Context(), &doc, query, docID)
	if err == sql.ErrNoRows {
		response.NotFound(c, "文档不存在", "Document not found")
		return
	}
	if err != nil {
		response.InternalError(c, "查询文档失败", "Failed to get document")
		return
	}

	response.Success(c, doc)
}

// List 获取文档列表
func (h *KnowledgeHandler) List(c *gin.Context) {
	projectID := c.Param("project_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	docType := c.Query("doc_type")
	sourceType := c.Query("source_type")
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	baseQuery := `SELECT * FROM knowledge_docs WHERE project_id = $1`
	countQuery := `SELECT COUNT(*) FROM knowledge_docs WHERE project_id = $1`
	args := []interface{}{projectID}
	argIdx := 2

	if docType != "" {
		baseQuery += ` AND doc_type = $` + strconv.Itoa(argIdx)
		countQuery += ` AND doc_type = $` + strconv.Itoa(argIdx)
		args = append(args, docType)
		argIdx++
	}
	if sourceType != "" {
		baseQuery += ` AND source_type = $` + strconv.Itoa(argIdx)
		countQuery += ` AND source_type = $` + strconv.Itoa(argIdx)
		args = append(args, sourceType)
		argIdx++
	}
	if keyword != "" {
		baseQuery += ` AND (title ILIKE $` + strconv.Itoa(argIdx) + ` OR content ILIKE $` + strconv.Itoa(argIdx) + `)`
		countQuery += ` AND (title ILIKE $` + strconv.Itoa(argIdx) + ` OR content ILIKE $` + strconv.Itoa(argIdx) + `)`
		args = append(args, "%"+keyword+"%")
		argIdx++
	}

	baseQuery += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, pageSize, offset)

	var docs []model.KnowledgeDoc
	err := h.db.SelectContext(c.Request.Context(), &docs, baseQuery, args...)
	if err != nil {
		response.InternalError(c, "查询文档列表失败", "Failed to list documents")
		return
	}

	var total int64
	h.db.GetContext(c.Request.Context(), &total, countQuery, args[:argIdx-2]...)

	response.SuccessWithPagination(c, docs, page, pageSize, total)
}

// Update 更新文档
func (h *KnowledgeHandler) Update(c *gin.Context) {
	docID := c.Param("doc_id")

	var req struct {
		Title    string   `json:"title" binding:"omitempty,min=2,max=500"`
		Content  string   `json:"content" binding:"omitempty"`
		Summary  string   `json:"summary" binding:"omitempty"`
		Tags     []string `json:"tags" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	userID := c.GetString("user_id")

	query := `
		UPDATE knowledge_docs SET 
			title = COALESCE(NULLIF($1, ''), title),
			content = COALESCE(NULLIF($2, ''), content),
			summary = COALESCE(NULLIF($3, ''), summary),
			tags = COALESCE($4, tags),
			updated_by = $5,
			version = version + 1,
			updated_at = NOW()
		WHERE doc_id = $6
	`
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	result, err := h.db.ExecContext(c.Request.Context(), query,
		req.Title, req.Content, req.Summary, tags, userID, docID)
	if err != nil {
		response.InternalError(c, "更新文档失败", "Failed to update document")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		response.NotFound(c, "文档不存在", "Document not found")
		return
	}

	response.Success(c, gin.H{"doc_id": docID, "updated": true})
}

// Delete 删除文档
func (h *KnowledgeHandler) Delete(c *gin.Context) {
	docID := c.Param("doc_id")

	result, err := h.db.ExecContext(c.Request.Context(),
		`DELETE FROM knowledge_docs WHERE doc_id = $1`, docID)
	if err != nil {
		response.InternalError(c, "删除文档失败", "Failed to delete document")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		response.NotFound(c, "文档不存在", "Document not found")
		return
	}

	response.Success(c, gin.H{"doc_id": docID, "deleted": true})
}

// SyncFeishu 从飞书同步文档内容
func (h *KnowledgeHandler) SyncFeishu(c *gin.Context) {
	docID := c.Param("doc_id")

	// 获取文档信息
	var doc model.KnowledgeDoc
	err := h.db.GetContext(c.Request.Context(), &doc,
		`SELECT * FROM knowledge_docs WHERE doc_id = $1`, docID)
	if err == sql.ErrNoRows {
		response.NotFound(c, "文档不存在", "Document not found")
		return
	}

	if doc.FeishuDocID == "" {
		response.BadRequest(c, "该文档不是飞书文档", "Document is not from Feishu")
		return
	}

	userID := c.GetString("user_id")

	// TODO: 实际调用飞书API获取文档内容
	// 这里模拟同步结果
	content := "飞书文档内容（同步更新）"

	// 更新文档
	h.db.ExecContext(c.Request.Context(),
		`UPDATE knowledge_docs SET content = $1, version = version + 1, updated_by = $2, updated_at = NOW() WHERE doc_id = $3`,
		content, userID, docID)

	response.Success(c, gin.H{
		"doc_id":  docID,
		"synced":  true,
		"version": doc.Version + 1,
	})
}

func stringOrDefault(s, defaultVal string) string {
	if s == "" {
		return defaultVal
	}
	return s
}

var _ = strconv.Itoa