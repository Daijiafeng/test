package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"testmind/internal/config"
	"testmind/internal/model"
	"testmind/internal/repository"
	"testmind/pkg/response"
	"testmind/pkg/validator"
)

type OrganizationHandler struct {
	orgRepo *repository.OrganizationRepository
	userRepo *repository.UserRepository
}

func NewOrganizationHandler(db *repository.DB) *OrganizationHandler {
	return &OrganizationHandler{
		orgRepo: repository.NewOrganizationRepository(db),
		userRepo: repository.NewUserRepository(db),
	}
}

// Create 创建组织
func (h *OrganizationHandler) Create(c *gin.Context) {
	var req validator.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	if err := validator.Validate(req); err != nil {
		zhMsg, enMsg := validator.GetErrorMsg(c.GetHeader("Accept-Language"), err)
		response.BadRequest(c, zhMsg, enMsg)
		return
	}

	userID := c.GetString("user_id")

	// 创建组织
	org := &model.Organization{
		OrgID:       uuid.New().String()[:32],
		Name:        req.Name,
		NameEn:      req.NameEn,
		Description: req.Description,
		OwnerID:     userID,
		Status:      "active",
		AIConfig:    model.JSONB{},
		Language:    "zh-CN",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.orgRepo.Create(c.Request.Context(), org); err != nil {
		response.InternalError(c, "创建组织失败", "Failed to create organization")
		return
	}

	response.Success(c, org)
}

// Get 获取组织详情
func (h *OrganizationHandler) Get(c *gin.Context) {
	orgID := c.Param("org_id")

	org, err := h.orgRepo.FindByID(c.Request.Context(), orgID)
	if err != nil {
		response.InternalError(c, "查询组织失败", "Failed to get organization")
		return
	}

	if org == nil {
		response.NotFound(c, "组织不存在", "Organization not found")
		return
	}

	response.Success(c, org)
}

// Update 更新组织
func (h *OrganizationHandler) Update(c *gin.Context) {
	orgID := c.Param("org_id")

	var req struct {
		Name        string `json:"name" binding:"omitempty,min=2,max=100"`
		NameEn      string `json:"name_en" binding:"omitempty,max=100"`
		Description string `json:"description" binding:"omitempty,max=500"`
		AIConfig    map[string]interface{} `json:"ai_config" binding:"omitempty"`
		Language    string `json:"language" binding:"omitempty,oneof=zh-CN en-US"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	org, err := h.orgRepo.FindByID(c.Request.Context(), orgID)
	if err != nil || org == nil {
		response.NotFound(c, "组织不存在", "Organization not found")
		return
	}

	// 更新字段
	if req.Name != "" {
		org.Name = req.Name
	}
	if req.NameEn != "" {
		org.NameEn = req.NameEn
	}
	if req.Description != "" {
		org.Description = req.Description
	}
	if req.AIConfig != nil {
		org.AIConfig = model.JSONB(req.AIConfig)
	}
	if req.Language != "" {
		org.Language = req.Language
	}
	org.UpdatedAt = time.Now()

	if err := h.orgRepo.Update(c.Request.Context(), org); err != nil {
		response.InternalError(c, "更新组织失败", "Failed to update organization")
		return
	}

	response.Success(c, org)
}

// ListProjects 获取组织下的项目列表
func (h *OrganizationHandler) ListProjects(c *gin.Context) {
	orgID := c.Param("org_id")

	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "50")

	pageInt := 1
	pageSizeInt := 50
	if p, err := validator.ValidateVar(page, "min=1"); err == nil {
		pageInt = parseInt(page)
	}
	if ps, err := validator.ValidateVar(pageSize, "min=1,max=100"); err == nil {
		pageSizeInt = parseInt(pageSize)
	}

	offset := (pageInt - 1) * pageSizeInt

	projectRepo := repository.NewProjectRepository(h.orgRepo.db)
	projects, err := projectRepo.ListByOrg(c.Request.Context(), orgID, pageSizeInt, offset)
	if err != nil {
		response.InternalError(c, "查询项目列表失败", "Failed to list projects")
		return
	}

	total, err := projectRepo.CountByOrg(c.Request.Context(), orgID)
	if err != nil {
		response.InternalError(c, "统计项目数量失败", "Failed to count projects")
		return
	}

	response.SuccessWithPagination(c, projects, pageInt, pageSizeInt, total)
}

// List 获取用户拥有的组织列表
func (h *OrganizationHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")

	orgs, err := h.orgRepo.ListByOwner(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "查询组织列表失败", "Failed to list organizations")
		return
	}

	response.Success(c, orgs)
}

func parseInt(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	return result
}