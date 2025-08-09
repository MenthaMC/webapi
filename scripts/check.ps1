# 项目结构检查脚本

Write-Host "🔍 Checking LeavesMC WebAPI v2 Go Project Structure" -ForegroundColor Green

$errors = @()
$warnings = @()

# 检查必需文件
$requiredFiles = @(
    "main.go",
    "go.mod",
    ".env.example",
    "README.md",
    "Dockerfile",
    "docker-compose.yml",
    "Makefile",
    "sql/init.sql",
    "public/docs.html",
    "public/api-v2.json"
)

Write-Host "`n📁 Checking required files..." -ForegroundColor Cyan
foreach ($file in $requiredFiles) {
    if (Test-Path $file) {
        Write-Host "  ✅ $file" -ForegroundColor Green
    } else {
        Write-Host "  ❌ $file" -ForegroundColor Red
        $errors += "Missing file: $file"
    }
}

# 检查目录结构
$requiredDirs = @(
    "internal/app",
    "internal/config",
    "internal/database",
    "internal/handlers",
    "internal/middleware",
    "internal/models",
    "internal/services",
    "internal/utils",
    "public",
    "sql",
    "scripts"
)

Write-Host "`n📂 Checking directory structure..." -ForegroundColor Cyan
foreach ($dir in $requiredDirs) {
    if (Test-Path $dir -PathType Container) {
        Write-Host "  ✅ $dir/" -ForegroundColor Green
    } else {
        Write-Host "  ❌ $dir/" -ForegroundColor Red
        $errors += "Missing directory: $dir"
    }
}

# 检查 Go 文件
$goFiles = Get-ChildItem -Path "internal" -Filter "*.go" -Recurse
Write-Host "`n🔧 Found $($goFiles.Count) Go files in internal/" -ForegroundColor Cyan

# 检查关键的 Go 文件
$keyGoFiles = @(
    "internal/app/app.go",
    "internal/config/config.go",
    "internal/database/database.go",
    "internal/handlers/handlers.go",
    "internal/services/services.go"
)

foreach ($file in $keyGoFiles) {
    if (Test-Path $file) {
        Write-Host "  ✅ $file" -ForegroundColor Green
    } else {
        Write-Host "  ❌ $file" -ForegroundColor Red
        $errors += "Missing Go file: $file"
    }
}

# 检查 .env 文件
Write-Host "`n⚙️  Checking configuration..." -ForegroundColor Cyan
if (Test-Path ".env") {
    Write-Host "  ✅ .env file exists" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  .env file not found (will be created from .env.example)" -ForegroundColor Yellow
    $warnings += ".env file not found"
}

# 检查 Go 模块
Write-Host "`n📦 Checking Go module..." -ForegroundColor Cyan
if (Test-Path "go.mod") {
    $goMod = Get-Content "go.mod" -Raw
    if ($goMod -match "module webapi-v2-neo") {
        Write-Host "  ✅ Go module configured correctly" -ForegroundColor Green
    } else {
        Write-Host "  ⚠️  Go module name might be incorrect" -ForegroundColor Yellow
        $warnings += "Go module name verification needed"
    }
}

# 总结
Write-Host "`n📊 Summary:" -ForegroundColor Magenta
if ($errors.Count -eq 0) {
    Write-Host "  ✅ All required files and directories are present!" -ForegroundColor Green
    Write-Host "  🚀 Project structure is ready for development" -ForegroundColor Green
} else {
    Write-Host "  ❌ Found $($errors.Count) errors:" -ForegroundColor Red
    foreach ($error in $errors) {
        Write-Host "    - $error" -ForegroundColor Red
    }
}

if ($warnings.Count -gt 0) {
    Write-Host "  ⚠️  Found $($warnings.Count) warnings:" -ForegroundColor Yellow
    foreach ($warning in $warnings) {
        Write-Host "    - $warning" -ForegroundColor Yellow
    }
}

Write-Host "`n🎯 Next steps:" -ForegroundColor Blue
Write-Host "  1. Copy .env.example to .env and configure" -ForegroundColor White
Write-Host "  2. Run 'go mod download' to install dependencies" -ForegroundColor White
Write-Host "  3. Set up PostgreSQL database" -ForegroundColor White
Write-Host "  4. Run 'go run main.go' or 'make dev' to start" -ForegroundColor White