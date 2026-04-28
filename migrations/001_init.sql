-- TestMind 数据库初始化脚本
-- 版本: v1.0
-- 日期: 2026-04-28
-- 数据库: PostgreSQL 15+

-- ============================================================
-- 1. 用户与权限
-- ============================================================

-- 组织（租户）
CREATE TABLE organizations (
    org_id          VARCHAR(32) PRIMARY KEY,
    name            VARCHAR(200) NOT NULL,
    name_en         VARCHAR(200),
    description     TEXT,
    owner_id        VARCHAR(32) NOT NULL,
    status          VARCHAR(20) DEFAULT 'active',
    ai_config       JSONB DEFAULT '{}',
    language        VARCHAR(10) DEFAULT 'zh-CN',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 项目
CREATE TABLE projects (
    project_id      VARCHAR(32) PRIMARY KEY,
    org_id          VARCHAR(32) NOT NULL REFERENCES organizations(org_id) ON DELETE CASCADE,
    name            VARCHAR(200) NOT NULL,
    name_en         VARCHAR(200),
    description     TEXT,
    status          VARCHAR(20) DEFAULT 'active',
    settings        JSONB DEFAULT '{}',
    language        VARCHAR(10) DEFAULT 'zh-CN',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_projects_org_id ON projects(org_id);

-- 用户
CREATE TABLE users (
    user_id         VARCHAR(32) PRIMARY KEY,
    org_id          VARCHAR(32) REFERENCES organizations(org_id) ON DELETE SET NULL,
    username        VARCHAR(100) NOT NULL UNIQUE,
    email           VARCHAR(200),
    phone           VARCHAR(20),
    password_hash   VARCHAR(200) NOT NULL,
    avatar_url      VARCHAR(500),
    display_name    VARCHAR(100),
    display_name_en VARCHAR(100),
    language        VARCHAR(10) DEFAULT 'zh-CN',
    status          VARCHAR(20) DEFAULT 'active',
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 飞书用户授权信息
CREATE TABLE user_feishu_auth (
    id              SERIAL PRIMARY KEY,
    user_id         VARCHAR(32) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    feishu_open_id  VARCHAR(64) UNIQUE,
    access_token    TEXT NOT NULL,
    refresh_token   TEXT NOT NULL,
    token_expires_at TIMESTAMPTZ,
    scopes          TEXT[],
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id)
);

-- 角色
CREATE TABLE roles (
    role_id         VARCHAR(32) PRIMARY KEY,
    org_id          VARCHAR(32) REFERENCES organizations(org_id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    name_en         VARCHAR(100),
    role_type       VARCHAR(20) DEFAULT 'custom',
    permissions     JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 项目成员
CREATE TABLE project_members (
    id              SERIAL PRIMARY KEY,
    project_id      VARCHAR(32) NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    user_id         VARCHAR(32) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    role_id         VARCHAR(32) REFERENCES roles(role_id) ON DELETE SET NULL,
    joined_at       TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, user_id)
);
CREATE INDEX idx_pm_project ON project_members(project_id);
CREATE INDEX idx_pm_user ON project_members(user_id);

-- ============================================================
-- 2. 系统配置
-- ============================================================

-- 自定义字段定义
CREATE TABLE custom_field_definitions (
    field_id        VARCHAR(32) PRIMARY KEY,
    project_id      VARCHAR(32) NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    entity_type     VARCHAR(20) NOT NULL,
    field_name      VARCHAR(100) NOT NULL,
    field_name_en   VARCHAR(100),
    field_type      VARCHAR(20) NOT NULL,
    options         JSONB,
    required        BOOLEAN DEFAULT FALSE,
    default_value   TEXT,
    sort_order      INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_cfd_project_entity ON custom_field_definitions(project_id, entity_type);

-- 模块（树形结构）
CREATE TABLE modules (
    module_id       VARCHAR(32) PRIMARY KEY,
    project_id      VARCHAR(32) NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    parent_id       VARCHAR(32),
    name            VARCHAR(200) NOT NULL,
    name_en         VARCHAR(200),
    sort_order      INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_modules_project ON modules(project_id);

-- 环境管理
CREATE TABLE environments (
    env_id          VARCHAR(32) PRIMARY KEY,
    project_id      VARCHAR(32) NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    name_en         VARCHAR(100),
    base_url        VARCHAR(500),
    description     TEXT,
    config          JSONB DEFAULT '{}',
    sort_order      INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- 3. 测试计划
-- ============================================================

CREATE TABLE test_plans (
    plan_id         VARCHAR(32) PRIMARY KEY,
    project_id      VARCHAR(32) NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    name            VARCHAR(200) NOT NULL,
    name_en         VARCHAR(200),
    version         VARCHAR(50),
    description     TEXT,
    status          VARCHAR(20) DEFAULT 'draft',
    start_date      DATE,
    end_date        DATE,
    owner_id        VARCHAR(32) REFERENCES users(user_id) ON DELETE SET NULL,
    ai_suggestion   JSONB,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_plans_project ON test_plans(project_id);

-- 计划关联需求
CREATE TABLE plan_requirements (
    id              SERIAL PRIMARY KEY,
    plan_id         VARCHAR(32) NOT NULL REFERENCES test_plans(plan_id) ON DELETE CASCADE,
    requirement_type VARCHAR(20) DEFAULT 'manual',
    requirement_url VARCHAR(500),
    requirement_title VARCHAR(200),
    requirement_content TEXT,
    feishu_doc_id   VARCHAR(100),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 计划关联用例
CREATE TABLE plan_cases (
    id              SERIAL PRIMARY KEY,
    plan_id         VARCHAR(32) NOT NULL REFERENCES test_plans(plan_id) ON DELETE CASCADE,
    case_id         VARCHAR(32) NOT NULL REFERENCES test_cases(case_id) ON DELETE CASCADE,
    assignee_id     VARCHAR(32) REFERENCES users(user_id) ON DELETE SET NULL,
    sort_order      INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(plan_id, case_id)
);

-- ============================================================
-- 4. 测试用例
-- ============================================================

CREATE TABLE test_cases (
    case_id         VARCHAR(32) PRIMARY KEY,
    project_id      VARCHAR(32) NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    module_id       VARCHAR(32) REFERENCES modules(module_id) ON DELETE SET NULL,
    title           VARCHAR(500) NOT NULL,
    title_en        VARCHAR(500),
    priority        VARCHAR(5) DEFAULT 'P2',
    precondition    TEXT,
    steps           JSONB NOT NULL DEFAULT '[]',
    tags            TEXT[],
    status          VARCHAR(20) DEFAULT 'draft',
    created_by      VARCHAR(32) NOT NULL REFERENCES users(user_id),
    created_method  VARCHAR(20) DEFAULT 'manual',
    version         INT DEFAULT 1,
    parent_case_id  VARCHAR(32),
    language        VARCHAR(10) DEFAULT 'zh-CN',
    feishu_record_id VARCHAR(100),
    custom_fields   JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_cases_project ON test_cases(project_id);
CREATE INDEX idx_cases_module ON test_cases(project_id, module_id);
CREATE INDEX idx_cases_status ON test_cases(project_id, status);

-- 用例版本历史
CREATE TABLE case_versions (
    version_id      SERIAL PRIMARY KEY,
    case_id         VARCHAR(32) NOT NULL REFERENCES test_cases(case_id) ON DELETE CASCADE,
    version_num     INT NOT NULL,
    title           VARCHAR(500),
    steps           JSONB,
    changed_by      VARCHAR(32) REFERENCES users(user_id) ON DELETE SET NULL,
    change_summary  TEXT,
    snapshot        JSONB,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_case_versions ON case_versions(case_id, version_num);

-- 用例评审
CREATE TABLE case_reviews (
    review_id       VARCHAR(32) PRIMARY KEY,
    case_id         VARCHAR(32) NOT NULL REFERENCES test_cases(case_id) ON DELETE CASCADE,
    reviewer_id     VARCHAR(32) NOT NULL REFERENCES users(user_id),
    result          VARCHAR(20),
    comment         TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- 5. 测试执行
-- ============================================================

CREATE TABLE execution_records (
    execution_id    VARCHAR(32) PRIMARY KEY,
    plan_id         VARCHAR(32) REFERENCES test_plans(plan_id) ON DELETE SET NULL,
    case_id         VARCHAR(32) NOT NULL REFERENCES test_cases(case_id) ON DELETE CASCADE,
    project_id      VARCHAR(32) NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    executor_id     VARCHAR(32) NOT NULL REFERENCES users(user_id),
    execution_type  VARCHAR(20) DEFAULT 'manual',
    result          VARCHAR(20) NOT NULL,
    actual_result   TEXT,
    duration        INT,
    environment     JSONB,
    ai_analysis     JSONB,
    related_defect_id VARCHAR(32),
    custom_fields   JSONB DEFAULT '{}',
    executed_at     TIMESTAMPTZ DEFAULT NOW(),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_exec_plan ON execution_records(plan_id);
CREATE INDEX idx_exec_case ON execution_records(case_id);
CREATE INDEX idx_exec_project ON execution_records(project_id);
CREATE INDEX idx_exec_result ON execution_records(project_id, result);

-- ============================================================
-- 6. 缺陷管理
-- ============================================================

CREATE TABLE defects (
    defect_id       VARCHAR(32) PRIMARY KEY,
    project_id      VARCHAR(32) NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    module_id       VARCHAR(32) REFERENCES modules(module_id) ON DELETE SET NULL,
    title           VARCHAR(500) NOT NULL,
    title_en        VARCHAR(500),
    description     TEXT,
    description_en  TEXT,
    reproduce_steps TEXT,
    severity        VARCHAR(20) DEFAULT 'general',
    priority        VARCHAR(5) DEFAULT 'P2',
    defect_type     VARCHAR(20),
    status          VARCHAR(20) DEFAULT 'open',
    assignee_id     VARCHAR(32) REFERENCES users(user_id) ON DELETE SET NULL,
    reporter_id     VARCHAR(32) NOT NULL REFERENCES users(user_id),
    discovery_stage VARCHAR(20),
    environment     JSONB,
    related_case_id VARCHAR(32),
    related_execution_id VARCHAR(32),
    ai_classification VARCHAR(20),
    ai_root_cause   TEXT,
    ai_fix_suggestion TEXT,
    similar_defects TEXT[],
    custom_fields   JSONB DEFAULT '{}',
    language        VARCHAR(10) DEFAULT 'zh-CN',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_defects_project ON defects(project_id);
CREATE INDEX idx_defects_status ON defects(project_id, status);
CREATE INDEX idx_defects_assignee ON defects(assignee_id);

-- 缺陷评论
CREATE TABLE defect_comments (
    comment_id      VARCHAR(32) PRIMARY KEY,
    defect_id       VARCHAR(32) NOT NULL REFERENCES defects(defect_id) ON DELETE CASCADE,
    author_id       VARCHAR(32) NOT NULL REFERENCES users(user_id),
    content         TEXT NOT NULL,
    mention_ids     TEXT[],
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_comments_defect ON defect_comments(defect_id);

-- 缺陷状态变更历史
CREATE TABLE defect_status_history (
    id              SERIAL PRIMARY KEY,
    defect_id       VARCHAR(32) NOT NULL REFERENCES defects(defect_id) ON DELETE CASCADE,
    from_status     VARCHAR(20),
    to_status       VARCHAR(20) NOT NULL,
    operator_id     VARCHAR(32) NOT NULL REFERENCES users(user_id),
    comment         TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- 7. 测试报告
-- ============================================================

CREATE TABLE test_reports (
    report_id       VARCHAR(32) PRIMARY KEY,
    project_id      VARCHAR(32) NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    plan_id         VARCHAR(32) REFERENCES test_plans(plan_id) ON DELETE SET NULL,
    title           VARCHAR(200) NOT NULL,
    title_en        VARCHAR(200),
    report_type     VARCHAR(20) DEFAULT 'plan',
    language        VARCHAR(10) DEFAULT 'zh-CN',
    summary         JSONB,
    ai_evaluation   JSONB,
    content         TEXT,
    feishu_doc_id   VARCHAR(100),
    created_by      VARCHAR(32) NOT NULL REFERENCES users(user_id),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_reports_project ON test_reports(project_id);

-- ============================================================
-- 8. 知识库
-- ============================================================

CREATE TABLE knowledge_docs (
    doc_id          VARCHAR(32) PRIMARY KEY,
    project_id      VARCHAR(32) NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    title           VARCHAR(200) NOT NULL,
    title_en        VARCHAR(200),
    category        VARCHAR(50),
    content         TEXT,
    content_en      TEXT,
    tags            TEXT[],
    language        VARCHAR(10) DEFAULT 'zh-CN',
    created_by      VARCHAR(32) NOT NULL REFERENCES users(user_id),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- 9. 审计日志
-- ============================================================

CREATE TABLE audit_logs (
    id              BIGSERIAL PRIMARY KEY,
    org_id          VARCHAR(32),
    project_id      VARCHAR(32),
    user_id         VARCHAR(32),
    action          VARCHAR(50) NOT NULL,
    entity_type     VARCHAR(50) NOT NULL,
    entity_id       VARCHAR(32),
    old_value       JSONB,
    new_value       JSONB,
    ip_address      VARCHAR(50),
    user_agent      TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_audit_org ON audit_logs(org_id);
CREATE INDEX idx_audit_project ON audit_logs(project_id);
CREATE INDEX idx_audit_time ON audit_logs(created_at);

-- ============================================================
-- 10. 初始数据
-- ============================================================

-- 系统预置角色
INSERT INTO roles (role_id, org_id, name, name_en, role_type, permissions) VALUES
('role_super_admin', NULL, '超级管理员', 'Super Admin', 'system', '{"all": true}'),
('role_project_admin', NULL, '项目管理员', 'Project Admin', 'system', '{"project": {"all": true}, "plans": {"all": true}, "cases": {"all": true}, "executions": {"all": true}, "defects": {"all": true}, "reports": {"all": true}, "ai": {"all": true}}'),
('role_test_manager', NULL, '测试经理', 'Test Manager', 'system', '{"plans": {"all": true}, "cases": {"all": true}, "executions": {"all": true}, "defects": {"all": true}, "reports": {"all": true}, "ai": {"all": true}}'),
('role_test_engineer', NULL, '测试工程师', 'Test Engineer', 'system', '{"plans": {"read": true, "execute": true}, "cases": {"create": true, "read": true, "execute": true}, "executions": {"create": true, "read": true}, "defects": {"create": true, "read": true, "update": true}, "reports": {"create": true, "read": true}, "ai": {"case_generation": true, "execution": true}}'),
('role_readonly', NULL, '只读成员', 'Read Only', 'system', '{"plans": {"read": true}, "cases": {"read": true}, "executions": {"read": true}, "defects": {"read": true}, "reports": {"read": true}}');

-- ============================================================
-- 完成提示
-- ============================================================
-- 执行完毕后，所有表已创建，预置角色已插入
-- 下一步：启动 user-svc 服务并测试 /auth/register 和 /auth/login 接口
