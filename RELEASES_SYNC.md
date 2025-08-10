# GitHub Releases 自动同步功能

本功能允许 MenthaMC WebAPI 自动从 GitHub 仓库拉取 Releases 信息并同步到数据库中。

## 功能特性

- 自动定期同步所有项目的 GitHub Releases
- 支持手动触发同步
- 支持单个项目的同步
- 自动创建版本和版本组
- 解析 Release 资产（JAR 文件）
- 创建下载记录

## 配置

在 `.env` 文件中添加以下配置：

```env
# GitHub 配置
GITHUB_TOKEN=your_github_token_here
GITHUB_SYNC_INTERVAL=1h
```

### 配置说明

- `GITHUB_TOKEN`: GitHub Personal Access Token，用于访问 GitHub API
- `GITHUB_SYNC_INTERVAL`: 同步间隔，支持格式如 `1h`、`30m`、`2h30m` 等

## GitHub Token 获取

1. 登录 GitHub
2. 进入 Settings > Developer settings > Personal access tokens
3. 点击 "Generate new token"
4. 选择适当的权限（至少需要 `public_repo` 权限）
5. 复制生成的 token 到配置文件中

## API 端点

### 获取同步状态
```
GET /v2/sync/status
```

返回示例：
```json
{
  "success": true,
  "data": {
    "enabled": true,
    "running": true,
    "time": "2024-01-01T12:00:00Z"
  }
}
```

### 手动触发全量同步（需要认证）
```
POST /v2/sync/trigger
```

### 同步指定项目（需要认证）
```
POST /v2/projects/{project}/sync
```

## 工作原理

1. **定期扫描**: 根据配置的间隔时间，自动扫描所有项目
2. **GitHub API 调用**: 使用 GitHub API 获取仓库的 Releases 信息
3. **数据解析**: 解析 Release 信息，提取版本号、构建信息等
4. **版本管理**: 自动创建版本组和版本记录
5. **构建记录**: 创建对应的构建记录和下载链接

## 版本组规则

版本组根据版本号自动创建，规则如下：
- `1.20.4` -> 版本组 `1.20`
- `1.19.2` -> 版本组 `1.19`
- `2.0.1` -> 版本组 `2.0`

## 支持的仓库格式

- `https://github.com/owner/repo`
- `https://github.com/owner/repo.git`
- `git@github.com:owner/repo.git`

## 注意事项

1. 只同步包含 JAR 文件的 Releases
2. 跳过草稿状态的 Releases
3. 已存在的版本不会重复创建
4. API 调用有频率限制，建议设置合理的同步间隔
5. 需要确保 GitHub Token 有足够的权限访问目标仓库

## 日志

同步过程中的日志会记录在应用日志中，包括：
- 同步开始和结束时间
- 成功同步的项目数量
- 错误信息和失败原因
- 跳过的版本信息

## 故障排除

### 常见问题

1. **Token 权限不足**
   - 确保 Token 有 `public_repo` 权限
   - 对于私有仓库，需要 `repo` 权限

2. **API 频率限制**
   - 增加同步间隔时间
   - 使用认证的 Token 可获得更高的频率限制

3. **仓库 URL 格式错误**
   - 确保仓库 URL 格式正确
   - 支持的格式见上述说明

4. **网络连接问题**
   - 检查服务器网络连接
   - 确保可以访问 GitHub API

### 调试模式

设置日志级别为 `debug` 可获得更详细的同步信息：

```env
LOG_LEVEL=debug