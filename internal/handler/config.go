package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"testmind/internal/model"
	"testmind/internal/repository"
	"testmind/pkg/response"
)

// ModuleHandler 模块管理
type ModuleHandler struct {
	db *repository.DB
}

func NewModuleHandler(db *repository.DB) *ModuleHandler {
	return &ModuleHandler{db: db}
}

// Create 创建模块
func (h *ModuleHandler) Create(c *gin.Context) {
	projectID := c.Param("project_id")

	var req struct {
		ParentID  string `json:"parent_id" binding:"omitempty"`
		Name      string `json:"name" binding:"required,min=1,max=200"`
		NameEn    string `json:"name_en" binding:"omitempty,max=200"`
		SortOrder int    `json:"sort_order" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	module := &model.Module{
		ModuleID:  uuid.New().String()[:32],
		ProjectID: projectID,
		ParentID:  req.ParentID,
		Name:      req.Name,
		NameEn:    req.NameEn,
		SortOrder: req.SortOrder,
		CreatedAt: time.Now(),
	}

	query := `
		INSERT INTO modules (module_id, project_id, parent_id, name, name_en, sort_order, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := h.db.ExecContext(c.Request.Context(), query,
		module.ModuleID, module.ProjectID, module.ParentID, module.Name,
		module.NameEn, module.SortOrder, module.CreatedAt)
	if err != nil {
		response.InternalError(c, "创建模块失败", "Failed to create module")
		return
	}

	response.Success(c, module)
}

// List 获取模块树
func (h *ModuleHandler) List(c *gin.Context) {
	projectID := c.Param("project_id")

	var modules []model.Module
	query := `SELECT * FROM modules WHERE project_id = $1 ORDER BY sort_order, created_at`
	err := h.db.SelectContext(c.Request.Context(), &modules, query, projectID)
	if err != nil {
		response.InternalError(c, "查询模块失败", "Failed to list modules")
		return
	}

	// 构建树形结构
	tree := buildModuleTree(modules, "")
	response.Success(c, tree)
}

// Update 更新模块
func (h *ModuleHandler) Update(c *gin.Context) {
	moduleID := c.Param("module_id")

	var req struct {
		Name      string `json:"name" binding:"omitempty,min=1,max=200"`
		NameEn    string `json:"name_en" binding:"omitempty,max=200"`
		ParentID  string `json:"parent_id" binding:"omitempty"`
		SortOrder int    `json:"sort_order" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	query := `
		UPDATE modules SET 
			name = COALESCE(NULLIF($1, ''), name),
			name_en = COALESCE(NULLIF($2, ''), name_en),
			parent_id = COALESCE(NULLIF($3, ''), parent_id),
			sort_order = $4
		WHERE module_id = $5
	`
	result, err := h.db.ExecContext(c.Request.Context(), query,
		req.Name, req.NameEn, req.ParentID, req.SortOrder, moduleID)
	if err != nil {
		response.InternalError(c, "更新模块失败", "Failed to update module")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		response.NotFound(c, "模块不存在", "Module not found")
		return
	}

	response.Success(c, gin.H{"module_id": moduleID, "updated": true})
}

// Delete 删除模块
func (h *ModuleHandler) Delete(c *gin.Context) {
	moduleID := c.Param("module_id")

	// 检查模块下是否有用例
	var caseCount int
	h.db.GetContext(c.Request.Context(), &caseCount,
		`SELECT COUNT(*) FROM test_cases WHERE module_id = $1`, moduleID)
	if caseCount > 0 {
		response.Error(c, 3001, "该模块下存在测试用例，无法删除", "Module has test cases, cannot delete")
		return
	}

	result, err := h.db.ExecContext(c.Request.Context(),
		`DELETE FROM modules WHERE module_id = $1`, moduleID)
	if err != nil {
		response.InternalError(c, "删除模块失败", "Failed to delete module")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		response.NotFound(c, "模块不存在", "Module not found")
		return
	}

	response.Success(c, gin.H{"module_id": moduleID, "deleted": true})
}

// ModuleTreeNode 模块树节点
type ModuleTreeNode struct {
	ModuleID  string            `json:"module_id"`
	Name      string            `json:"name"`
	NameEn    string            `json:"name_en"`
	ParentID  string            `json:"parent_id"`
	SortOrder int               `json:"sort_order"`
	Children  []*ModuleTreeNode `json:"children"`
}

func buildModuleTree(modules []model.Module, parentID string) []*ModuleTreeNode {
	var tree []*ModuleTreeNode
	for _, m := range modules {
		if (parentID == "" && m.ParentID == "") || m.ParentID == parentID {
			node := &ModuleTreeNode{
				ModuleID:  m.ModuleID,
				Name:      m.Name,
				NameEn:    m.NameEn,
				ParentID:  m.ParentID,
				SortOrder: m.SortOrder,
				Children:  buildModuleTree(modules, m.ModuleID),
			}
			tree = append(tree, node)
		}
	}
	return tree
}

// CustomFieldHandler 自定义字段管理
type CustomFieldHandler struct {
	db *repository.DB
}

func NewCustomFieldHandler(db *repository.DB) *CustomFieldHandler {
	return &CustomFieldHandler{db: db}
}

// Create 创建自定义字段
func (h *CustomFieldHandler) Create(c *gin.Context) {
	projectID := c.Param("project_id")

	var req struct {
		EntityType   string                 `json:"entity_type" binding:"required,oneof=test_case defect execution"`
		FieldName    string                 `json:"field_name" binding:"required,min=1,max=100"`
		FieldNameEn  string                 `json:"field_name_en" binding:"omitempty,max=100"`
		FieldType    string                 `json:"field_type" binding:"required,oneof=text number single_select multi_select date user multi_user attachment url"`
		Options      map[string]interface{} `json:"options" binding:"omitempty"`
		Required     bool                   `json:"required"`
		DefaultValue string                 `json:"default_value" binding:"omitempty"`
		SortOrder    int                    `json:"sort_order" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	field := &model.CustomFieldDefinition{
		FieldID:      uuid.New().String()[:32],
		ProjectID:    projectID,
		EntityType:   req.EntityType,
		FieldName:    req.FieldName,
		FieldNameEn:  req.FieldNameEn,
		FieldType:    req.FieldType,
		Options:      model.JSONB(req.Options),
		Required:     req.Required,
		DefaultValue: req.DefaultValue,
		SortOrder:    req.SortOrder,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO custom_field_definitions (field_id, project_id, entity_type, field_name, field_name_en, 
			field_type, options, required, default_value, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := h.db.ExecContext(c.Request.Context(), query,
		field.FieldID, field.ProjectID, field.EntityType, field.FieldName, field.FieldNameEn,
		field.FieldType, field.Options, field.Required, field.DefaultValue, field.SortOrder,
		field.CreatedAt, field.UpdatedAt)
	if err != nil {
		response.InternalError(c, "创建自定义字段失败", "Failed to create custom field")
		return
	}

	response.Success(c, field)
}

// List 获取自定义字段列表
func (h *CustomFieldHandler) List(c *gin.Context) {
	projectID := c.Param("project_id")
	entityType := c.Query("entity_type")

	query := `SELECT * FROM custom_field_definitions WHERE project_id = $1`
	args := []interface{}{projectID}

	if entityType != "" {
		query += ` AND entity_type = $2`
		args = append(args, entityType)
	}

	query += ` ORDER BY sort_order, created_at`

	var fields []model.CustomFieldDefinition
	err := h.db.SelectContext(c.Request.Context(), &fields, query, args...)
	if err != nil {
		response.InternalError(c, "查询自定义字段失败", "Failed to list custom fields")
		return
	}

	response.Success(c, fields)
}

// Update 更新自定义字段
func (h *CustomFieldHandler) Update(c *gin.Context) {
	fieldID := c.Param("field_id")

	var req struct {
		FieldName    string                 `json:"field_name" binding:"omitempty,min=1,max=100"`
		FieldNameEn  string                 `json:"field_name_en" binding:"omitempty,max=100"`
		Options      map[string]interface{} `json:"options" binding:"omitempty"`
		Required     *bool                  `json:"required"`
		DefaultValue string                 `json:"default_value" binding:"omitempty"`
		SortOrder    int                    `json:"sort_order" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	query := `
		UPDATE custom_field_definitions SET 
			field_name = COALESCE(NULLIF($1, ''), field_name),
			field_name_en = COALESCE(NULLIF($2, ''), field_name_en),
			default_value = COALESCE(NULLIF($3, ''), default_value),
			sort_order = $4,
			updated_at = NOW()
		WHERE field_id = $5
	`
	args := []interface{}{req.FieldName, req.FieldNameEn, req.DefaultValue, req.SortOrder, fieldID}

	if req.Options != nil {
		query = `
			UPDATE custom_field_definitions SET 
				field_name = COALESCE(NULLIF($1, ''), field_name),
				field_name_en = COALESCE(NULLIF($2, ''), field_name_en),
				options = $3,
				default_value = COALESCE(NULLIF($4, ''), default_value),
				sort_order = $5,
				updated_at = NOW()
			WHERE field_id = $6
		`
		args = []interface{}{req.FieldName, req.FieldNameEn, model.JSONB(req.Options), req.DefaultValue, req.SortOrder, fieldID}
	}

	result, err := h.db.ExecContext(c.Request.Context(), query, args...)
	if err != nil {
		response.InternalError(c, "更新自定义字段失败", "Failed to update custom field")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		response.NotFound(c, "自定义字段不存在", "Custom field not found")
		return
	}

	response.Success(c, gin.H{"field_id": fieldID, "updated": true})
}

// Delete 删除自定义字段
func (h *CustomFieldHandler) Delete(c *gin.Context) {
	fieldID := c.Param("field_id")

	result, err := h.db.ExecContext(c.Request.Context(),
		`DELETE FROM custom_field_definitions WHERE field_id = $1`, fieldID)
	if err != nil {
		response.InternalError(c, "删除自定义字段失败", "Failed to delete custom field")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		response.NotFound(c, "自定义字段不存在", "Custom field not found")
		return
	}

	response.Success(c, gin.H{"field_id": fieldID, "deleted": true})
}

// 确保 strconv 被引用
var _ = strconv.Itoa