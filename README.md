# TestMind - AI测试管理平台

## 快速开始

### 前置要求

- Go 1.22+
- PostgreSQL 15+
- Redis 7+

### 1. 初始化数据库

```bash
psql -U testmind -d testmind -f migrations/001_init.sql
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 填入实际配置
```

### 3. 安装依赖

```bash
go mod tidy
```

### 4. 启动用户服务

```bash
go run cmd/user-svc/main.go
```

### 5. 验证

```bash
# 健康检查
curl http://localhost:8080/health

# 用户注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testadmin",
    "password": "Test1234",
    "email": "admin@testmind.io",
    "display_name": "管理员",
    "language": "zh-CN"
  }'

# 用户登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testadmin",
    "password": "Test1234"
  }'
```

## 项目结构

```
testmind/
├── cmd/
│   └── user-svc/         # 用户服务入口
├── internal/
│   ├── config/           # 配置管理
│   ├── handler/          # HTTP处理器
│   ├── middleware/        # 中间件（认证、跨域等）
│   ├── model/            # 数据模型
│   ├── repository/       # 数据访问层
│   └── service/          # 业务逻辑层
├── pkg/
│   ├── jwt/              # JWT工具
│   ├── response/         # 响应格式
│   └── validator/        # 参数校验
├── migrations/           # 数据库迁移
└── docs/                 # 文档
```

## 开发进度

### ✅ 已完成
- [x] 数据库DDL设计（20张表）
- [x] 配置管理模块
- [x] 数据模型定义
- [x] JWT认证工具
- [x] 统一响应格式
- [x] 参数校验器
- [x] 用户注册/登录/刷新Token API
- [x] 认证中间件
- [x] API路由框架

### 🚧 开发中
- [ ] 数据库连接层（repository）
- [ ] 组织/项目管理
- [ ] 测试计划模块
- [ ] 测试用例模块

### 📋 待开发
- [ ] AI用例生成引擎
- [ ] 飞书OAuth集成
- [ ] 测试执行模块
- [ ] 缺陷管理模块
- [ ] 测试报告模块
- [ ] 知识库模块