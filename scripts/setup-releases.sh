#!/bin/bash

# MenthaMC WebAPI Release功能设置脚本

set -e

echo "🚀 开始设置Release自动拉取功能..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到Go环境，请先安装Go 1.21+"
    exit 1
fi

# 检查PostgreSQL连接
if ! command -v psql &> /dev/null; then
    echo "⚠️  警告: 未找到psql命令，请确保PostgreSQL已安装"
fi

# 执行数据库迁移
echo "📊 执行数据库迁移..."
if [ -n "$DB_URL" ]; then
    psql "$DB_URL" -f sql/releases_migration.sql
    echo "✅ 数据库迁移完成"
else
    echo "⚠️  请设置DB_URL环境变量或手动执行: psql -d webapi -f sql/releases_migration.sql"
fi

# 构建命令行工具
echo "🔨 构建Release管理工具..."
go build -o bin/release-manager cmd/release-manager/main.go
echo "✅ 构建完成: bin/release-manager"

# 创建示例配置
echo "📝 创建示例配置..."

# Paper项目配置
echo "配置Paper项目..."
./bin/release-manager \
    -action=config \
    -project=paper \
    -owner=PaperMC \
    -repo=Paper \
    -interval=60 \
    -auto \
    -enabled

echo "✅ Paper项目配置完成"

# 测试同步
echo "🔄 测试同步功能..."
./bin/release-manager -action=sync -project=paper

echo "📋 查看配置列表..."
./bin/release-manager -action=list

echo ""
echo "🎉 Release功能设置完成！"
echo ""
echo "📚 使用说明:"
echo "  1. 启动WebAPI服务: go run main.go"
echo "  2. 查看Release: curl http://localhost:32767/v2/projects/paper/releases"
echo "  3. 管理配置: ./bin/release-manager -action=list"
echo ""
echo "📖 详细文档请查看: docs/RELEASES.md"