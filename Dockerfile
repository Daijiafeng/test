# 构建阶段
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -o user-svc ./cmd/user-svc

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 复制构建产物
COPY --from=builder /app/user-svc .
COPY --from=builder /app/migrations ./migrations

# 暴露端口
EXPOSE 8080

# 运行
CMD ["./user-svc"]