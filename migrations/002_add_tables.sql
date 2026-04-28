-- TestMind 数据库初始化脚本 v2.0
-- 补充缺失的表结构

-- =====================================================
-- OAuth 状态表（飞书授权流程）
-- =====================================================
CREATE TABLE IF NOT EXISTS oauth_states (
    state_id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL,
    state_code VARCHAR(64) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_oauth_states_code ON oauth_states(state_code);
CREATE INDEX idx_oauth_states_user ON oauth_states(user_id);

-- =====================================================
-- 飞书用户凭证表
-- =====================================================
CREATE TABLE IF NOT EXISTS feishu_credentials (
    credential_id VARCHAR(32) PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    expires_at TIMESTAMP NOT NULL,
    feishu_user_id VARCHAR(64),
    feishu_name VARCHAR(200),
    feishu_avatar TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_feishu_cred_user ON feishu_credentials(user_id);

-- =====================================================
-- 测试计划-需求关联表
-- =====================================================
CREATE TABLE IF NOT EXISTS plan_requirements (
    id SERIAL PRIMARY KEY,
    plan_id VARCHAR(32) NOT NULL REFERENCES test_plans(plan_id) ON DELETE CASCADE,
    requirement_type VARCHAR(20) NOT NULL DEFAULT 'manual',  -- manual, feishu_doc
    requirement_url TEXT,
    requirement_title VARCHAR(500) NOT NULL,
    requirement_content TEXT,
    feishu_doc_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_plan_req_plan ON plan_requirements(plan_id);

-- =====================================================
-- 测试计划-用例关联表
-- =====================================================
CREATE TABLE IF NOT EXISTS plan_cases (
    id SERIAL PRIMARY KEY,
    plan_id VARCHAR(32) NOT NULL REFERENCES test_plans(plan_id) ON DELETE CASCADE,
    case_id VARCHAR(32) NOT NULL REFERENCES test_cases(case_id) ON DELETE CASCADE,
    assignee_id VARCHAR(32),
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(plan_id, case_id)
);

CREATE INDEX idx_plan_cases_plan ON plan_cases(plan_id);
CREATE INDEX idx_plan_cases_case ON plan_cases(case_id);

-- =====================================================
-- 用例版本历史表
-- =====================================================
CREATE TABLE IF NOT EXISTS case_versions (
    version_id SERIAL PRIMARY KEY,
    case_id VARCHAR(32) NOT NULL REFERENCES test_cases(case_id) ON DELETE CASCADE,
    version_num INT NOT NULL,
    title VARCHAR(500),
    steps JSONB,
    changed_by VARCHAR(32) NOT NULL,
    change_summary VARCHAR(500),
    snapshot JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_case_versions_case ON case_versions(case_id);

-- =====================================================
-- 用例评审表
-- =====================================================
CREATE TABLE IF NOT EXISTS case_reviews (
    review_id VARCHAR(32) PRIMARY KEY,
    case_id VARCHAR(32) NOT NULL REFERENCES test_cases(case_id) ON DELETE CASCADE,
    reviewer_id VARCHAR(32) NOT NULL,
    result VARCHAR(20) NOT NULL,  -- approved, rejected
    comment TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_case_reviews_case ON case_reviews(case_id);

-- =====================================================
-- 缺陷评论表
-- =====================================================
CREATE TABLE IF NOT EXISTS defect_comments (
    comment_id VARCHAR(32) PRIMARY KEY,
    defect_id VARCHAR(32) NOT NULL REFERENCES defects(defect_id) ON DELETE CASCADE,
    author_id VARCHAR(32) NOT NULL,
    content TEXT NOT NULL,
    visibility VARCHAR(20) DEFAULT 'public',  -- public, internal
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_defect_comments_defect ON defect_comments(defect_id);

-- =====================================================
-- AI用例生成任务表
-- =====================================================
CREATE TABLE IF NOT EXISTS ai_case_generations (
    task_id VARCHAR(32) PRIMARY KEY,
    project_id VARCHAR(32) NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    source_type VARCHAR(20) NOT NULL,  -- manual, feishu_doc, text
    source_content TEXT,
    feishu_doc_id VARCHAR(100),
    requested_by VARCHAR(32) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending, processing, completed, failed
    generated_count INT DEFAULT 0,
    applied_count INT DEFAULT 0,
    applied_at TIMESTAMP,
    error_message TEXT,
    priority_strategy VARCHAR(20) DEFAULT 'auto',  -- auto, high, medium, low
    case_style VARCHAR(20) DEFAULT 'standard',  -- standard, concise, detailed
    language VARCHAR(10) DEFAULT 'zh-CN',
    custom_prompt TEXT,
    max_cases INT DEFAULT 50,
    module_id VARCHAR(32),
    created_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE INDEX idx_ai_gen_project ON ai_case_generations(project_id);
CREATE INDEX idx_ai_gen_status ON ai_case_generations(status);

-- =====================================================
-- 知识库文档表
-- =====================================================
CREATE TABLE IF NOT EXISTS knowledge_docs (
    doc_id VARCHAR(32) PRIMARY KEY,
    project_id VARCHAR(32) NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    doc_type VARCHAR(20) NOT NULL DEFAULT 'requirement',  -- requirement, design, api, other
    source_type VARCHAR(20) NOT NULL DEFAULT 'manual',  -- manual, feishu_doc, upload
    feishu_doc_id VARCHAR(100),
    feishu_doc_url TEXT,
    content TEXT,
    summary TEXT,
    tags TEXT[],
    vector_id VARCHAR(100),  -- Milvus向量ID
    language VARCHAR(10) DEFAULT 'zh-CN',
    created_by VARCHAR(32) NOT NULL,
    updated_by VARCHAR(32),
    version INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_knowledge_project ON knowledge_docs(project_id);
CREATE INDEX idx_knowledge_type ON knowledge_docs(doc_type);

-- =====================================================
-- 数据库迁移版本表
-- =====================================================
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(20) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT NOW()
);

-- 记录当前版本
INSERT INTO schema_migrations (version) VALUES ('2.0') ON CONFLICT DO NOTHING;
