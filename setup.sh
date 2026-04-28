#!/bin/bash
# TestMind 一键初始化脚本
# 使用方法：chmod +x setup.sh && ./setup.sh

set -e

echo "🦞 TestMind 项目初始化..."
echo ""

# 创建目录结构
echo "📁 创建目录结构..."
mkdir -p cmd/user-svc
mkdir -p internal/{config,handler,middleware,model,repository,service}
mkdir -p pkg/{jwt,response,validator}
mkdir -p migrations
mkdir -p docs

# 初始化 Go 模块
echo "📦 初始化 Go 模块..."
go mod init testmind

# 下载依赖
echo "⬇️  下载依赖..."
go get github.com/gin-gonic/gin@v1.9.1
go get github.com/go-playground/validator/v10@v10.16.0
go get github.com/golang-jwt/jwt/v5@v5.2.0
go get github.com/google/uuid@v1.5.0
go get github.com/jmoiron/sqlx@v1.3.5
go get github.com/lib/pq@v1.10.9
golang.org/x/crypto@v0.18.0
go mod tidy

echo ""
echo "✅ 项目初始化完成！"
echo ""
echo "📋 下一步："
echo "  1. 创建 PostgreSQL 数据库: createdb testmind"
echo "  2. 执行迁移脚本: psql -d testmind -f migrations/001_init.sql"
echo "  3. 配置环境变量: cp .env.example .env && 编辑 .env"
echo "  4. 启动服务: go run cmd/user-svc/main.go"
echo ""
echo "🦞 祝你开发愉快！"