# 开发环境启动脚本 (PowerShell)

Write-Host "🚀 Starting LeavesMC WebAPI v2 Development Environment" -ForegroundColor Green

# 检查是否存在 .env 文件
if (-not (Test-Path ".env")) {
    Write-Host "⚠️  .env file not found. Copying from .env.example..." -ForegroundColor Yellow
    Copy-Item ".env.example" ".env"
    Write-Host "📝 Please edit .env file with your configuration before running again." -ForegroundColor Yellow
    exit 1
}

# 检查 Go 版本
Write-Host "🔍 Checking Go version..." -ForegroundColor Cyan
go version

# 下载依赖
Write-Host "📦 Downloading dependencies..." -ForegroundColor Cyan
go mod download
go mod tidy

# 检查数据库连接（可选）
Write-Host "🗄️  Database connection will be tested on startup..." -ForegroundColor Cyan

# 启动应用
Write-Host "🎯 Starting application..." -ForegroundColor Cyan
if (Get-Command air -ErrorAction SilentlyContinue) {
    Write-Host "🔥 Using Air for hot reload..." -ForegroundColor Green
    air
} else {
    Write-Host "📌 Air not found, running with go run..." -ForegroundColor Yellow
    Write-Host "💡 Install Air for hot reload: go install github.com/cosmtrek/air@latest" -ForegroundColor Blue
    go run main.go
}