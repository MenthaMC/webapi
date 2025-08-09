.PHONY: build run test clean dev docker-build docker-run

# 变量定义
BINARY_NAME=webapi-v2-neo
DOCKER_IMAGE=webapi-v2-neo
VERSION?=latest

# 构建
build:
	go build -o $(BINARY_NAME) main.go

# 运行
run:
	go run main.go

# 开发模式（使用 air 热重载，需要先安装 air）
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Installing..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

# 测试
test:
	go test -v ./...

# 测试覆盖率
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 清理
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Please install it first."; \
		echo "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 依赖管理
deps:
	go mod download
	go mod tidy

# 生产构建
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BINARY_NAME) main.go

# Docker 构建
docker-build:
	docker build -t $(DOCKER_IMAGE):$(VERSION) .

# Docker 运行
docker-run:
	docker run -p 32767:32767 --env-file .env $(DOCKER_IMAGE):$(VERSION)

# 数据库迁移（如果有迁移工具）
migrate-up:
	@echo "Running database migrations..."
	# 这里可以添加数据库迁移命令

migrate-down:
	@echo "Rolling back database migrations..."
	# 这里可以添加数据库回滚命令

# 生成 API 文档（如果使用 swag）
docs:
	@if command -v swag > /dev/null; then \
		swag init; \
	else \
		echo "swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# 安装开发工具
install-tools:
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest

# 帮助
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  dev           - Run in development mode with hot reload"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  build-prod    - Build for production"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  docs          - Generate API documentation"
	@echo "  install-tools - Install development tools"
	@echo "  help          - Show this help"