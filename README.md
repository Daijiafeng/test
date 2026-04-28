# TestMind - AI驱动的测试管理平台

> 一站式测试管理解决方案，集成飞书生态，支持AI智能生成测试用例

## 🌟 核心特性

### 📝 测试用例管理
- 支持手工编写和AI自动生成
- 版本历史追踪，评审流程管理
- 模块化组织，标签分类
- 飞书多维表格同步

### 🧪 测试计划与执行
- 灵活的测试计划管理
- 关联需求文档（飞书云文档）
- 执行记录统计分析
- 快速创建缺陷

### 🐛 缺陷管理
- 完整的缺陷生命周期管理
- 状态流转规则
- 评论、历史追踪
- 批量操作、数据导出

### 📊 测试报告
- 自动生成统计报告
- Markdown格式导出
- 飞书云文档分享
- AI智能分析摘要

### 🔗 飞书集成
- OAuth 2.0 用户授权
- 需求文档导入
- 多维表格同步
- 测试报告分享

### 🤖 AI能力
- 智能生成测试用例
- 用例优化（增强/简化/翻译/格式化）
- 缺陷分析建议
- 报告智能摘要

---

## 📦 技术架构

### 后端技术栈
- **语言**: Go 1.22
- **Web框架**: Gin
- **数据库**: PostgreSQL 15
- **缓存**: Redis 7
- **ORM**: sqlx

### 部署方式
- Docker Compose（单机）
- Kubernetes（集群，预留）

### AI服务
- OpenAI API（云端）
- 本地模型（预留接口）

---

## 🚀 快速启动

### 方式一：Docker Compose（推荐）

```bash
# 1. 克隆仓库
git clone https://github.com/Daijiafeng/test.git
cd test

# 2. 配置环境变量
cp .env.example .env
# 编辑.env，填写飞书和AI配置

# 3. 启动所有服务
docker-compose up -d

# 4. 检查服务状态
docker-compose ps

# 5. 访问API
curl http://localhost:8080/health
```

### 方式二：本地开发

```bash
# 1. 启动数据库
docker-compose up -d postgres redis

# 2. 安装Go依赖
go mod tidy

# 3. 运行数据库迁移
psql -h localhost -U testmind -d testmind -f migrations/001_init.sql
psql -h localhost -U testmind -d testmind -f migrations/002_add_tables.sql

# 4. 启动API服务
go run cmd/user-svc/main.go

# API地址: http://localhost:8080/api/v1
```

---

## 📋 API文档

### API总览（103+接口）

| 模块 | 数量 | 路径前缀 |
|------|------|----------|
| 用户认证 | 6 | `/api/v1/auth` |
| 组织管理 | 5 | `/api/v1/organizations` |
| 项目管理 | 9 | `/api/v1/projects` |
| 测试计划 | 10 | `/api/v1/plans` |
| 测试用例 | 10 | `/api/v1/cases` |
| 模块管理 | 4 | `/api/v1/modules` |
| 自定义字段 | 4 | `/api/v1/custom-fields` |
| 测试执行 | 12 | `/api/v1/executions` |
| 缺陷管理 | 15 | `/api/v1/defects` |
| 测试报告 | 7 | `/api/v1/reports` |
| 飞书集成 | 7 | `/api/v1/feishu` |
| AI生成 | 5 | `/api/v1/ai` |
| 知识库 | 6 | `/api/v1/knowledge` |

### 认证流程

```bash
# 1. 注册用户
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"password123"}'

# 2. 登录获取Token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"password123"}'

# 3. 使用Token访问API
curl http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer <access_token>"
```

### 常用接口示例

#### 创建项目

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"测试项目","name_en":"Test Project"}'
```

#### 创建测试计划

```bash
curl -X POST http://localhost:8080/api/v1/projects/<project_id>/plans \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"v1.0测试计划","version":"1.0"}'
```

#### AI生成测试用例

```bash
curl -X POST http://localhost:8080/api/v1/projects/<project_id>/ai/generate \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "source_type": "text",
    "source_content": "用户登录功能：支持账号密码登录、手机验证码登录、第三方登录...",
    "max_cases": 10
  }'
```

---

## 🔧 配置说明

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SERVER_PORT` | API端口 | 8080 |
| `DB_HOST` | 数据库地址 | localhost |
| `DB_PORT` | 数据库端口 | 5432 |
| `DB_USER` | 数据库用户 | testmind |
| `DB_PASSWORD` | 数据库密码 | testmind123 |
| `DB_NAME` | 数据库名 | testmind |
| `JWT_SECRET` | JWT密钥 | 必须设置 |
| `FEISHU_APP_ID` | 飞书App ID | 必须设置 |
| `FEISHU_APP_SECRET` | 飞书App Secret | 必须设置 |
| `AI_CLOUD_API_KEY` | AI服务API Key | 必须设置 |

---

## 📊 数据库设计

### 核心表（26张）

| 分类 | 表名 |
|------|------|
| 用户 | users, roles, organizations |
| 项目 | projects, project_members |
| 用例 | test_cases, modules, case_versions, case_reviews |
| 计划 | test_plans, plan_requirements, plan_cases |
| 执行 | execution_records |
| 缺陷 | defects, defect_comments |
| 报告 | test_reports |
| 飞书 | oauth_states, feishu_credentials |
| AI | ai_case_generations |
| 知识库 | knowledge_docs |
| 配置 | custom_field_definitions, audit_logs |

---

## 🔗 集成配置

### 飞书应用配置

1. 创建飞书自建应用：https://open.feishu.cn/app
2. 配置权限：
   - `contact:user.base:read` - 获取用户基本信息
   - `docx:document` - 文档操作
   - `bitable:record` - 多维表格操作
3. 配置回调地址：`https://your-domain/api/v1/feishu/oauth/callback`
4. 获取App ID和App Secret，填入`.env`

### AI服务配置

1. 获取OpenAI API Key：https://platform.openai.com/api-keys
2. 填入`.env`的`AI_CLOUD_API_KEY`
3. 可选：指定模型`AI_CLOUD_MODEL=gpt-4`

---

## 📝 开发指南

### 项目结构

```
testmind/
├── cmd/user-svc/main.go       # 主入口
├── internal/
│   ├── config/                # 配置
│   ├── handler/               # API处理器（13个）
│   ├── middleware/            # 中间件
│   ├── model/                 # 数据模型
│   ├── repository/            # 数据库操作
│   └── service/               # 业务服务
├── migrations/                # SQL迁移
├── pkg/                       # 公共工具
├── docker-compose.yml         # Docker编排
├── Dockerfile                 # 容器镜像
└── README.md                  # 使用说明
```

### 本地开发步骤

```bash
# 1. 启动数据库
docker-compose up -d postgres

# 2. 运行迁移
psql -h localhost -U testmind -d testmind -f migrations/*.sql

# 3. 热重载开发（可选安装air）
go install github.com/cosmtrek/air@latest
air

# 4. 测试API
go test ./...
```

---

## 📈 进度与路线图

### 已完成 ✅
- [x] 需求文档（PRD v1.1）
- [x] 技术方案设计
- [x] 任务拆分（87项）
- [x] 前端原型设计（11页面）
- [x] 后端API框架（103+接口）
- [x] 数据库设计（26张表）
- [x] Docker部署配置
- [x] 飞书OAuth集成
- [x] AI服务接口

### 进行中 🚧
- [ ] 飞书API实际对接测试
- [ ] AI用例生成优化
- [ ] 前端Vue骨架开发

### 计划中 📋
- [ ] 自动化执行对接（Selenium/Appium）
- [ ] 性能测试模块
- [ ] 移动端小程序/App
- [ ] 国际化完整支持

---

## 📚 文档资源

| 文档 | 链接 |
|------|------|
| 测试报告 | https://www.feishu.cn/docx/RMpudXBqQoV6oTxYurqccQLsnjf |
| 需求文档 | https://www.feishu.cn/docx/IjFodERgqo51hQxb74Qc8cYanbf |
| 技术方案 | https://www.feishu.cn/docx/HMcZdEaDXozERfxkEicc2I3ynNY |
| 任务拆分 | https://www.feishu.cn/docx/UIQPdX7DmoAygPxGm3xckqCQndf |
| 前端原型 | https://kbcineyoxn.aiforce.cloud/app/app_4k15s96snqqh8 |

---

## 🤝 贡献指南

```bash
# Fork仓库
git clone https://github.com/Daijiafeng/test.git

# 创建分支
git checkout -b feature/your-feature

# 提交代码
git commit -am "Add your feature"

# 推送分支
git push origin feature/your-feature

# 创建Pull Request
```

---

## 📄 License

MIT License

---

## 👥 联系方式

- 作者：戴佳峰
- 部门：技术组
- 飞书OpenID：ou_a05a67f1576635de6c0b12e19d4111ff

---

🦞 **TestMind** - 让测试更智能