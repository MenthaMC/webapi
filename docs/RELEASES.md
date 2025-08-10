# 自动拉取Releases功能

本功能允许WebAPI自动从GitHub仓库拉取Release信息，并提供API接口查询和管理。

## 功能特性

- 🔄 自动定时同步GitHub Releases
- 📋 提供完整的Release信息查询API
- ⚙️ 灵活的配置管理
- 🔐 支持私有仓库（通过访问令牌）
- 📊 调度器状态监控
- 🛠️ 命令行管理工具

## 数据库表结构

### release_configs
存储项目的Release同步配置
- `project`: 项目ID
- `repo_owner`: GitHub仓库所有者
- `repo_name`: GitHub仓库名称
- `access_token`: GitHub访问令牌（私有仓库需要）
- `auto_sync`: 是否启用自动同步
- `sync_interval`: 同步间隔（分钟）
- `enabled`: 配置是否启用

### releases
存储Release信息
- `project`: 项目ID
- `tag_name`: 标签名称
- `name`: Release名称
- `body`: Release描述
- `draft`: 是否为草稿
- `prerelease`: 是否为预发布版本
- `published_at`: 发布时间

### release_assets
存储Release附件信息
- `release_id`: 关联的Release ID
- `name`: 文件名
- `browser_download_url`: 下载链接
- `size`: 文件大小

## API接口

### 查询接口

#### 获取项目Releases列表
```
GET /v2/projects/{project}/releases?limit=20&offset=0
```

响应示例：
```json
{
  "releases": [
    {
      "id": 1,
      "project": "mint",
      "tag_name": "v1.21.4",
      "name": "Mint 1.21.4",
      "body": "Release notes...",
      "draft": false,
      "prerelease": false,
      "created_at": "2024-01-01T00:00:00Z",
      "published_at": "2024-01-01T00:00:00Z",
      "html_url": "https://github.com/MenthaMC/Mint/releases/tag/v1.20.4",
      "assets": [
        {
          "name": "mint-1.20.4.jar",
          "size": 12345678,
          "browser_download_url": "https://github.com/MenthaMC/Mint/releases/tag/v1.20.4/paper-1.20.4.jar"
        }
      ]
    }
  ],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "count": 1
  }
}
```

#### 获取最新Release
```
GET /v2/projects/{project}/releases/latest
```

#### 获取Release配置
```
GET /v2/projects/{project}/releases/config
```

### 管理接口（需要认证）

#### 保存Release配置
```
POST /v2/projects/{project}/releases/config
Content-Type: application/json

{
  "repo_owner": "PaperMC",
  "repo_name": "Paper",
  "access_token": "ghp_xxxxxxxxxxxx",
  "auto_sync": true,
  "sync_interval": 30,
  "enabled": true
}
```

#### 手动同步Releases
```
POST /v2/projects/{project}/releases/sync
```

#### 获取调度器状态
```
GET /v2/admin/scheduler/status
```

响应示例：
```json
{
  "running": true,
  "enabled_projects": 3,
  "check_interval": "1 minute"
}
```

#### 触发所有项目同步
```
POST /v2/admin/scheduler/trigger
```

## 命令行工具

项目提供了命令行工具来管理Release配置：

### 配置项目
```bash
go run cmd/release-manager/main.go \
  -action=config \
  -project=paper \
  -owner=PaperMC \
  -repo=Paper \
  -token=ghp_xxxxxxxxxxxx \
  -interval=30 \
  -auto \
  -enabled
```

### 手动同步项目
```bash
go run cmd/release-manager/main.go \
  -action=sync \
  -project=paper
```

### 列出所有配置
```bash
go run cmd/release-manager/main.go -action=list
```

## 部署步骤

### 1. 数据库迁移
执行SQL脚本创建相关表：
```bash
psql -d webapi -f sql/releases_migration.sql
```

### 2. 配置项目
使用命令行工具或API配置需要同步的项目：
```bash
go run cmd/release-manager/main.go \
  -action=config \
  -project=paper \
  -owner=PaperMC \
  -repo=Paper \
  -auto \
  -interval=60
```

### 3. 启动服务
启动WebAPI服务，调度器会自动开始工作：
```bash
go run main.go
```

### 4. 验证功能
检查调度器状态：
```bash
curl http://localhost:32767/v2/admin/scheduler/status \
  -H "Authentication: YOUR_JWT_TOKEN"
```

## 配置说明

### GitHub访问令牌
- 公开仓库：不需要访问令牌
- 私有仓库：需要具有`repo`权限的Personal Access Token

### 同步间隔
- 最小间隔：1分钟
- 推荐间隔：30-60分钟
- 调度器每分钟检查一次是否需要同步

### 自动同步
- 启用后会根据设定的间隔自动同步
- 可以随时通过API手动触发同步
- 调度器在应用启动时自动启动

## 监控和日志

系统会记录以下日志：
- 调度器启动/停止
- 自动同步触发
- 同步成功/失败
- GitHub API调用错误

可以通过日志监控同步状态和排查问题。

## 故障排除

### 常见问题

1. **GitHub API限制**
   - 未认证请求：60次/小时
   - 认证请求：5000次/小时
   - 建议配置访问令牌

2. **同步失败**
   - 检查网络连接
   - 验证仓库名称和所有者
   - 确认访问令牌权限

3. **调度器未运行**
   - 检查应用启动日志
   - 验证数据库连接
   - 确认配置正确

### 调试命令
```bash
# 检查配置
go run cmd/release-manager/main.go -action=list

# 手动同步测试
go run cmd/release-manager/main.go -action=sync -project=paper

# 查看API响应
curl http://localhost:32767/v2/projects/paper/releases/latest