# TestMind - AI测试管理平台

## 本地开发指南

### 前置要求

- Go 1.22+
- PostgreSQL 15+（或使用 Docker）
- Docker + Docker Compose（推荐）

---

## 方式一：Docker Compose（推荐）

### 1. 启动 PostgreSQL

```bash
docker-compose up -d postgres redis
```

数据库会自动执行 `migrations/001_init.sql`，创建所有表和预置角色。

### 2. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env`，确认数据库配置：
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=testmind
DB_PASSWORD=testmind123
DB_NAME=testmind
```

### 3. 安装依赖

```bash
go mod tidy
```

### 4. 启动服务

```bash
go run cmd/user-svc/main.go
```

### 5. 验证

```bash
curl http://localhost:8080/health
# 返回：{"status":"ok","service":"user-svc","version":"1.0.0"}
```

---

## 方式二：手动安装 PostgreSQL

### 1. 安装 PostgreSQL

```bash
# macOS
brew install postgresql@15
brew services start postgresql@15

# Ubuntu
sudo apt install postgresql-15
sudo systemctl start postgresql
```

### 2. 创建数据库

```bash
psql -U postgres
CREATE DATABASE testmind;
CREATE USER testmind WITH PASSWORD 'testmind123';
GRANT ALL PRIVILEGES ON DATABASE testmind TO testmind;
\q
```

### 3. 执行迁移脚本

```bash
psql -U testmind -d testmind -f migrations/001_init.sql
```

### 4-5. 同方式一

---

## API 测试

### 用户注册

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "Admin1234",
    "email": "admin@testmind.io",
    "display_name": "管理员",
    "language": "zh-CN"
  }'
```

### 用户登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "Admin1234"
  }'
```

返回：
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "user": {...},
    "token": {
      "access_token": "...",
      "refresh_token": "...",
      "expires_in": 7200
    }
  }
}
```

### 创建组织（需要Token）

```bash
curl -X POST http://localhost:8080/api/v1/organizations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "name": "测试团队",
    "name_en": "Test Team",
    "description": "这是一个测试团队"
  }'
```

### 创建项目

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "org_id": "<org_id>",
    "name": "TestMind 项目",
    "name_en": "TestMind Project",
    "description": "AI测试管理平台开发"
  }'
```

---

## 项目结构

```
testmind/
├── cmd/
│   └── user-svc/main.go       # 服务入口
├── internal/
│   ├── config/config.go       # 配置管理
│   ├── handler/
│   │   ├── user.go            # 用户API
│   │   ├── organization.go    # 组织API
│   │   └── project.go         # 项目API
│   ├── middleware/auth.go     # JWT认证中间件
│   ├── model/model.go         # 数据模型
│   └── repository/repository.go # 数据库访问层
├── pkg/
│   ├── jwt/jwt.go             # JWT工具
│   ├── response/response.go   # 统一响应格式
│   └── validator/validator.go # 参数校验
├── migrations/
│   └── 001_init.sql           # 数据库初始化
├── docker-compose.yml         # Docker Compose 配置
├── Dockerfile                 # Docker 构建文件
├── go.mod                     # Go 依赖管理
├── .env.example               # 环境变量模板
└── README.md                  # 本文档
```

---

## 已实现功能

| 模块 | API数量 | 功能 |
|------|---------|------|
| 用户认证 | 6 | 注册、登录、刷新Token、获取/更新用户信息、登出 |
| 组织管理 | 4 | 创建、查询、更新、列表 |
| 项目管理 | 7 | 创建、查询、更新、删除、成员管理（添加/移除/角色） |
| **合计** | **17** | — |

---

## 下一步开发

- [ ] 测试计划模块
- [ ] 测试用例模块
- [ ] AI用例生成模块
- [ ] 飞书OAuth集成
- [ ] 测试执行模块
- [ ] 缺陷管理模块

---

## 常见问题

### 数据库连接失败

检查 `.env` 配置，确认 `DB_HOST`、`DB_PORT`、`DB_USER`、`DB_PASSWORD` 正确。

### JWT Token 无效

确认 `JWT_SECRET` 设置正确（至少32字符）。

### 端口被占用

修改 `.env` 中的 `SERVER_PORT`，默认 8080。

---

🦞 Happy Coding!