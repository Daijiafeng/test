package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"testmind/internal/model"
	"testmind/internal/repository"
	"testmind/pkg/response"
	"testmind/pkg/validator"
)

type ProjectHandler struct {
	projectRepo *repository.ProjectRepository
	memberRepo  *repository.ProjectMemberRepository
	roleRepo    *repository.RoleRepository
}

func NewProjectHandler(db *repository.DB) *ProjectHandler {
	return &ProjectHandler{
		projectRepo: repository.NewProjectRepository(db),
		memberRepo:  repository.NewProjectMemberRepository(db),
		roleRepo:    repository.NewRoleRepository(db),
	}
}

// Create 创建项目
func (h *ProjectHandler) Create(c *gin.Context) {
	var req validator.CreateProjectRequest
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

	// 创建项目
	project := &model.Project{
		ProjectID:   uuid.New().String()[:32],
		OrgID:       req.OrgID,
		Name:        req.Name,
		NameEn:      req.NameEn,
		Description: req.Description,
		Status:      "active",
		Settings:    model.JSONB{},
		Language:    stringOrDefault(req.Language, "zh-CN"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.projectRepo.Create(c.Request.Context(), project); err != nil {
		response.InternalError(c, "创建项目失败", "Failed to create project")
		return
	}

	// 创建者自动成为项目管理员
	member := &model.ProjectMember{
		ProjectID: project.ProjectID,
		UserID:    userID,
		RoleID:    "role_project_admin",
		JoinedAt:  time.Now(),
	}

	if err := h.memberRepo.Add(c.Request.Context(), member); err != nil {
		response.InternalError(c, "添加项目成员失败", "Failed to add project member")
		return
	}

	response.Success(c, project)
}

// Get 获取项目详情
func (h *ProjectHandler) Get(c *gin.Context) {
	projectID := c.Param("project_id")

	project, err := h.projectRepo.FindByID(c.Request.Context(), projectID)
	if err != nil {
		response.InternalError(c, "查询项目失败", "Failed to get project")
		return
	}

	if project == nil {
		response.NotFound(c, "项目不存在", "Project not found")
		return
	}

	if project.Status == "deleted" {
		response.NotFound(c, "项目不存在", "Project not found")
		return
	}

	response.Success(c, project)
}

// Update 更新项目
func (h *ProjectHandler) Update(c *gin.Context) {
	projectID := c.Param("project_id")

	var req struct {
		Name        string `json:"name" binding:"omitempty,min=2,max=100"`
		NameEn      string `json:"name_en" binding:"omitempty,max=100"`
		Description string `json:"description" binding:"omitempty,max=500"`
		Language    string `json:"language" binding:"omitempty,oneof=zh-CN en-US"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	project, err := h.projectRepo.FindByID(c.Request.Context(), projectID)
	if err != nil || project == nil {
		response.NotFound(c, "项目不存在", "Project not found")
		return
	}

	if req.Name != "" {
		project.Name = req.Name
	}
	if req.NameEn != "" {
		project.NameEn = req.NameEn
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.Language != "" {
		project.Language = req.Language
	}
	project.UpdatedAt = time.Now()

	if err := h.projectRepo.Update(c.Request.Context(), project); err != nil {
		response.InternalError(c, "更新项目失败", "Failed to update project")
		return
	}

	response.Success(c, project)
}

// Delete 删除项目（软删除）
func (h *ProjectHandler) Delete(c *gin.Context) {
	projectID := c.Param("project_id")

	project, err := h.projectRepo.FindByID(c.Request.Context(), projectID)
	if err != nil || project == nil {
		response.NotFound(c, "项目不存在", "Project not found")
		return
	}

	if err := h.projectRepo.Delete(c.Request.Context(), projectID); err != nil {
		response.InternalError(c, "删除项目失败", "Failed to delete project")
		return
	}

	response.Success(c, gin.H{
		"project_id": projectID,
		"status":     "deleted",
	})
}

// ListMembers 获取项目成员列表
func (h *ProjectHandler) ListMembers(c *gin.Context) {
	projectID := c.Param("project_id")

	members, err := h.memberRepo.ListByProject(c.Request.Context(), projectID)
	if err != nil {
		response.InternalError(c, "查询项目成员失败", "Failed to list members")
		return
	}

	response.Success(c, members)
}

// AddMember 添加项目成员
func (h *ProjectHandler) AddMember(c *gin.Context) {
	projectID := c.Param("project_id")

	var req struct {
		UserID string `json:"user_id" binding:"required"`
		RoleID string `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	// 检查是否已是成员
	isMember, err := h.memberRepo.IsMember(c.Request.Context(), projectID, req.UserID)
	if err != nil {
		response.InternalError(c, "检查成员失败", "Failed to check membership")
		return
	}

	if isMember {
		response.Error(c, 2001, "该用户已是项目成员", "User is already a project member")
		return
	}

	// 验证角色是否存在
	role, err := h.roleRepo.FindByID(c.Request.Context(), req.RoleID)
	if err != nil || role == nil {
		response.NotFound(c, "角色不存在", "Role not found")
		return
	}

	member := &model.ProjectMember{
		ProjectID: projectID,
		UserID:    req.UserID,
		RoleID:    req.RoleID,
		JoinedAt:  time.Now(),
	}

	if err := h.memberRepo.Add(c.Request.Context(), member); err != nil {
		response.InternalError(c, "添加项目成员失败", "Failed to add project member")
		return
	}

	response.Success(c, member)
}

// RemoveMember 移除项目成员
func (h *ProjectHandler) RemoveMember(c *gin.Context) {
	projectID := c.Param("project_id")
	userID := c.Param("user_id")

	if err := h.memberRepo.Remove(c.Request.Context(), projectID, userID); err != nil {
		response.InternalError(c, "移除项目成员失败", "Failed to remove project member")
		return
	}

	response.Success(c, gin.H{
		"project_id": projectID,
		"user_id":    userID,
		"removed":    true,
	})
}

// UpdateMemberRole 更新项目成员角色
func (h *ProjectHandler) UpdateMemberRole(c *gin.Context) {
	projectID := c.Param("project_id")
	userID := c.Param("user_id")

	var req struct {
		RoleID string `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误", "Invalid request format")
		return
	}

	if err := h.memberRepo.UpdateRole(c.Request.Context(), projectID, userID, req.RoleID); err != nil {
		response.InternalError(c, "更新成员角色失败", "Failed to update member role")
		return
	}

	response.Success(c, gin.H{
		"project_id": projectID,
		"user_id":    userID,
		"role_id":    req.RoleID,
		"updated":    true,
	})
}