# Makefile for TestMind

.PHONY: all build run test clean docker-up docker-down migrate help

# 默认目标
all: build

# 构建
build:
	@echo "Building TestMind API..."
	@go build -o bin/testmind-api ./cmd/user-svc
	@echo "Build complete: bin/testmind-api"

# 运行
run:
	@echo "Running TestMind API..."
	@go run ./cmd/user-svc/main.go

# 测试
test:
	@echo "Running tests..."
	@go test -v ./...

# 代码检查
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# 格式化代码
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# 清理
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f testmind-api
	@echo "Clean complete"

# Docker启动
docker-up:
	@echo "Starting Docker services..."
	@docker-compose up -d
	@echo "Services started"
	@docker-compose ps

# Docker停止
docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down
	@echo "Services stopped"

# Docker日志
docker-logs:
	@docker-compose logs -f testmind-api

# 数据库迁移
migrate:
	@echo "Running database migrations..."
	@psql -h localhost -U testmind -d testmind -f migrations/001_init.sql
	@psql -h localhost -U testmind -d testmind -f migrations/002_add_tables.sql
	@echo "Migrations complete"

# 数据库重置
db-reset:
	@echo "Resetting database..."
	@docker-compose down -v
	@docker-compose up -d postgres
	@sleep 5
	@$(MAKE) migrate

# 依赖更新
deps:
	@echo "Updating dependencies..."
	@go mod tidy
	@go mod download
	@echo "Dependencies updated"

# 安装工具
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed"

# 生成Swagger文档（需要swag工具）
swagger:
	@echo "Generating Swagger docs..."
	@swag init -g cmd/user-svc/main.go -o docs/swagger
	@echo "Swagger docs generated"

# 帮助
help:
	@echo "TestMind Makefile Commands:"
	@echo ""
	@echo "  make build        - 编译项目"
	@echo "  make run          - 运行项目"
	@echo "  make test         - 运行测试"
	@echo "  make lint         - 代码检查"
	@echo "  make fmt          - 格式化代码"
	@echo "  make clean        - 清理构建文件"
	@echo ""
	@echo "  make docker-up    - 启动Docker服务"
	@echo "  make docker-down  - 停止Docker服务"
	@echo "  make docker-logs  - 查看服务日志"
	@echo ""
	@echo "  make migrate      - 执行数据库迁移"
	@echo "  make db-reset     - 重置数据库"
	@echo ""
	@echo "  make deps         - 更新依赖"
	@echo "  make swagger      - 生成API文档"
	@echo ""
	@echo "  make help         - 显示帮助信息"

# 快速启动（开发模式）
dev: docker-up migrate run

# 生产构建
prod:
	@echo "Building for production..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o bin/testmind-api ./cmd/user-svc
	@echo "Production build complete"