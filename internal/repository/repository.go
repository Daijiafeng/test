package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"testmind/internal/config"
	"testmind/internal/model"
)

type DB struct {
	*sqlx.DB
}

func NewDB(cfg *config.DatabaseConfig) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DB{db}, nil
}

// ============================================================
// 用户相关
// ============================================================

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (user_id, org_id, username, email, phone, password_hash, 
			avatar_url, display_name, display_name_en, language, status, created_at, updated_at)
		VALUES (:user_id, :org_id, :username, :email, :phone, :password_hash,
			:avatar_url, :display_name, :display_name_en, :language, :status, :created_at, :updated_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE username = $1`
	err := r.db.GetContext(ctx, &user, query, username)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) FindByID(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE user_id = $1`
	err := r.db.GetContext(ctx, &user, query, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE email = $1`
	err := r.db.GetContext(ctx, &user, query, email)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users SET 
			display_name = :display_name,
			display_name_en = :display_name_en,
			avatar_url = :avatar_url,
			language = :language,
			last_login_at = :last_login_at,
			updated_at = :updated_at
		WHERE user_id = :user_id
	`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *UserRepository) List(ctx context.Context, orgID string, limit, offset int) ([]model.User, error) {
	var users []model.User
	query := `SELECT * FROM users WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &users, query, orgID, limit, offset)
	return users, err
}

// ============================================================
// 组织相关
// ============================================================

type OrganizationRepository struct {
	db *DB
}

func NewOrganizationRepository(db *DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(ctx context.Context, org *model.Organization) error {
	query := `
		INSERT INTO organizations (org_id, name, name_en, description, owner_id, status, ai_config, language, created_at, updated_at)
		VALUES (:org_id, :name, :name_en, :description, :owner_id, :status, :ai_config, :language, :created_at, :updated_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, org)
	return err
}

func (r *OrganizationRepository) FindByID(ctx context.Context, orgID string) (*model.Organization, error) {
	var org model.Organization
	query := `SELECT * FROM organizations WHERE org_id = $1`
	err := r.db.GetContext(ctx, &org, query, orgID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &org, err
}

func (r *OrganizationRepository) Update(ctx context.Context, org *model.Organization) error {
	query := `
		UPDATE organizations SET 
			name = :name,
			name_en = :name_en,
			description = :description,
			ai_config = :ai_config,
			language = :language,
			updated_at = :updated_at
		WHERE org_id = :org_id
	`
	_, err := r.db.NamedExecContext(ctx, query, org)
	return err
}

func (r *OrganizationRepository) ListByOwner(ctx context.Context, ownerID string) ([]model.Organization, error) {
	var orgs []model.Organization
	query := `SELECT * FROM organizations WHERE owner_id = $1 ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &orgs, query, ownerID)
	return orgs, err
}

// ============================================================
// 项目相关
// ============================================================

type ProjectRepository struct {
	db *DB
}

func NewProjectRepository(db *DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, project *model.Project) error {
	query := `
		INSERT INTO projects (project_id, org_id, name, name_en, description, status, settings, language, created_at, updated_at)
		VALUES (:project_id, :org_id, :name, :name_en, :description, :status, :settings, :language, :created_at, :updated_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, project)
	return err
}

func (r *ProjectRepository) FindByID(ctx context.Context, projectID string) (*model.Project, error) {
	var project model.Project
	query := `SELECT * FROM projects WHERE project_id = $1`
	err := r.db.GetContext(ctx, &project, query, projectID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &project, err
}

func (r *ProjectRepository) Update(ctx context.Context, project *model.Project) error {
	query := `
		UPDATE projects SET 
			name = :name,
			name_en = :name_en,
			description = :description,
			settings = :settings,
			language = :language,
			updated_at = :updated_at
		WHERE project_id = :project_id
	`
	_, err := r.db.NamedExecContext(ctx, query, project)
	return err
}

func (r *ProjectRepository) Delete(ctx context.Context, projectID string) error {
	query := `UPDATE projects SET status = 'deleted', updated_at = NOW() WHERE project_id = $1`
	_, err := r.db.ExecContext(ctx, query, projectID)
	return err
}

func (r *ProjectRepository) ListByOrg(ctx context.Context, orgID string, limit, offset int) ([]model.Project, error) {
	var projects []model.Project
	query := `SELECT * FROM projects WHERE org_id = $1 AND status != 'deleted' ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &projects, query, orgID, limit, offset)
	return projects, err
}

func (r *ProjectRepository) CountByOrg(ctx context.Context, orgID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM projects WHERE org_id = $1 AND status != 'deleted'`
	err := r.db.GetContext(ctx, &count, query, orgID)
	return count, err
}

// ============================================================
// 项目成员相关
// ============================================================

type ProjectMemberRepository struct {
	db *DB
}

func NewProjectMemberRepository(db *DB) *ProjectMemberRepository {
	return &ProjectMemberRepository{db: db}
}

func (r *ProjectMemberRepository) Add(ctx context.Context, member *model.ProjectMember) error {
	query := `
		INSERT INTO project_members (project_id, user_id, role_id, joined_at)
		VALUES (:project_id, :user_id, :role_id, :joined_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, member)
	return err
}

func (r *ProjectMemberRepository) Remove(ctx context.Context, projectID, userID string) error {
	query := `DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, projectID, userID)
	return err
}

func (r *ProjectMemberRepository) UpdateRole(ctx context.Context, projectID, userID, roleID string) error {
	query := `UPDATE project_members SET role_id = $1 WHERE project_id = $2 AND user_id = $3`
	_, err := r.db.ExecContext(ctx, query, roleID, projectID, userID)
	return err
}

func (r *ProjectMemberRepository) ListByProject(ctx context.Context, projectID string) ([]model.ProjectMember, error) {
	var members []model.ProjectMember
	query := `
		SELECT pm.*, u.username, u.display_name, r.name as role_name
		FROM project_members pm
		LEFT JOIN users u ON pm.user_id = u.user_id
		LEFT JOIN roles r ON pm.role_id = r.role_id
		WHERE pm.project_id = $1
		ORDER BY pm.joined_at DESC
	`
	err := r.db.SelectContext(ctx, &members, query, projectID)
	return members, err
}

func (r *ProjectMemberRepository) IsMember(ctx context.Context, projectID, userID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM project_members WHERE project_id = $1 AND user_id = $2)`
	err := r.db.GetContext(ctx, &exists, query, projectID, userID)
	return exists, err
}

// ============================================================
// 角色相关
// ============================================================

type RoleRepository struct {
	db *DB
}

func NewRoleRepository(db *DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) FindByID(ctx context.Context, roleID string) (*model.Role, error) {
	var role model.Role
	query := `SELECT * FROM roles WHERE role_id = $1`
	err := r.db.GetContext(ctx, &role, query, roleID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &role, err
}

func (r *RoleRepository) ListSystemRoles(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	query := `SELECT * FROM roles WHERE role_type = 'system'`
	err := r.db.SelectContext(ctx, &roles, query)
	return roles, err
}

func (r *RoleRepository) ListByOrg(ctx context.Context, orgID string) ([]model.Role, error) {
	var roles []model.Role
	query := `SELECT * FROM roles WHERE org_id = $1 OR org_id IS NULL`
	err := r.db.SelectContext(ctx, &roles, query, orgID)
	return roles, err
}