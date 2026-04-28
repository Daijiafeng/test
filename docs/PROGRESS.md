# TestMind API 开发进度

## 📊 当前进度

**最后更新：** 2026-04-28

### ✅ 已完成的模块

| 模块 | API数量 | 状态 |
|------|---------|------|
| 用户认证 | 6 | ✅ 完成 |
| 组织管理 | 4 | ✅ 完成 |
| 项目管理 | 7 | ✅ 完成 |
| 测试计划 | 10 | ✅ 完成 |
| 测试用例 | 10 | ✅ 完成 |
| 模块管理 | 4 | ✅ 完成 |
| 自定义字段 | 4 | ✅ 完成 |
| **合计** | **45** | — |

---

## 🔌 API 列表

### 用户认证

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | /api/v1/auth/register | 用户注册 |
| POST | /api/v1/auth/login | 用户登录 |
| POST | /api/v1/auth/refresh | 刷新Token |
| GET | /api/v1/auth/profile | 获取用户信息 |
| PUT | /api/v1/auth/profile | 更新用户信息 |
| POST | /api/v1/auth/logout | 登出 |

### 组织管理

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | /api/v1/organizations | 创建组织 |
| GET | /api/v1/organizations | 获取组织列表 |
| GET | /api/v1/organizations/:org_id | 获取组织详情 |
| PUT | /api/v1/organizations/:org_id | 更新组织 |
| GET | /api/v1/organizations/:org_id/projects | 获取组织下的项目 |

### 项目管理

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | /api/v1/projects | 创建项目 |
| GET | /api/v1/projects/:project_id | 获取项目详情 |
| PUT | /api/v1/projects/:project_id | 更新项目 |
| DELETE | /api/v1/projects/:project_id | 删除项目 |
| GET | /api/v1/projects/:project_id/members | 获取项目成员 |
| POST | /api/v1/projects/:project_id/members | 添加成员 |
| DELETE | /api/v1/projects/:project_id/members/:user_id | 移除成员 |
| PUT | /api/v1/projects/:project_id/members/:user_id | 更新成员角色 |

### 测试计划

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | /api/v1/projects/:project_id/plans | 创建测试计划 |
| GET | /api/v1/projects/:project_id/plans | 获取计划列表 |
| GET | /api/v1/plans/:plan_id | 获取计划详情 |
| PUT | /api/v1/plans/:plan_id | 更新计划 |
| DELETE | /api/v1/plans/:plan_id | 删除计划 |
| POST | /api/v1/plans/:plan_id/requirements | 添加需求关联 |
| GET | /api/v1/plans/:plan_id/requirements | 获取需求列表 |
| POST | /api/v1/plans/:plan_id/cases | 批量关联用例 |
| DELETE | /api/v1/plans/:plan_id/cases | 批量取消关联 |
| GET | /api/v1/plans/:plan_id/progress | 获取执行进度 |

### 测试用例

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | /api/v1/projects/:project_id/cases | 创建用例 |
| GET | /api/v1/projects/:project_id/cases | 获取用例列表（支持筛选） |
| GET | /api/v1/projects/:project_id/cases/search | 全文搜索用例 |
| POST | /api/v1/projects/:project_id/cases/batch | 批量创建用例 |
| GET | /api/v1/cases/:case_id | 获取用例详情 |
| PUT | /api/v1/cases/:case_id | 更新用例 |
| DELETE | /api/v1/cases/:case_id | 删除用例 |
| GET | /api/v1/cases/:case_id/versions | 获取版本历史 |
| POST | /api/v1/cases/:case_id/review | 提交评审 |

### 模块管理

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | /api/v1/projects/:project_id/modules | 创建模块 |
| GET | /api/v1/projects/:project_id/modules | 获取模块树 |
| PUT | /api/v1/modules/:module_id | 更新模块 |
| DELETE | /api/v1/modules/:module_id | 删除模块 |

### 自定义字段

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | /api/v1/projects/:project_id/custom-fields | 创建自定义字段 |
| GET | /api/v1/projects/:project_id/custom-fields | 获取字段列表 |
| PUT | /api/v1/custom-fields/:field_id | 更新字段 |
| DELETE | /api/v1/custom-fields/:field_id | 删除字段 |

---

## 📋 待开发模块

| 模块 | 预计API数量 | 优先级 |
|------|-------------|--------|
| 测试执行 | 8 | 高 |
| 缺陷管理 | 12 | 高 |
| 测试报告 | 6 | 中 |
| AI用例生成 | 4 | 高 |
| 飞书OAuth | 5 | 高 |
| 知识库 | 4 | 低 |

---

## 🚀 快速启动

```bash
# 拉取最新代码
git pull origin master

# 启动数据库
docker-compose up -d postgres

# 安装依赖
go mod tidy

# 启动服务
go run cmd/user-svc/main.go
```

---

**Git 提交历史：**
```
c63ab17 feat: add test plan, test case, module and custom field APIs
179af36 feat: add Docker support and improve README
98e3e2e docs: add setup script for local initialization
f2c67d2 feat: add database repository and project management APIs
47f6c01 Initial commit: TestMind backend framework
```