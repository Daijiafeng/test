# 多阶段构建
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制go.mod和go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o testmind-api ./cmd/user-svc

# 运行阶段
FROM alpine:3.18

RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/testmind-api .
COPY --from=builder /app/migrations ./migrations

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# 启动
CMD ["./testmind-api"]