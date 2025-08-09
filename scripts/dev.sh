#!/bin/bash

# 开发环境启动脚本

set -e

echo "🚀 Starting LeavesMC WebAPI v2 Development Environment"

# 检查是否存在 .env 文件
if [ ! -f .env ]; then
    echo "⚠️  .env file not found. Copying from .env.example..."
    cp .env.example .env
    echo "📝 Please edit .env file with your configuration before running again."
    exit 1
fi

# 检查 Go 版本
echo "🔍 Checking Go version..."
go version

# 下载依赖
echo "📦 Downloading dependencies..."
go mod download
go mod tidy

# 检查数据库连接（可选）
echo "🗄️  Database connection will be tested on startup..."

# 启动应用
echo "🎯 Starting application..."
if command -v air &> /dev/null; then
    echo "🔥 Using Air for hot reload..."
    air
else
    echo "📌 Air not found, running with go run..."
    echo "💡 Install Air for hot reload: go install github.com/cosmtrek/air@latest"
    go run main.go
fi