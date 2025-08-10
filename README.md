# MenthaMC WebAPI
这是 MenthaMC WebAPI 使用 Gin 框架构建的 RESTful API 服务。

## 功能特性

- 项目管理和版本控制
- 构建信息查询和管理
- 文件下载服务
- **GitHub Releases 自动同步** ✨
- JWT 认证
- PostgreSQL 数据库支持
- 自动数据库迁移
- Swagger API 文档

## 技术栈

- **语言**: Go 1.21+
- **Web框架**: Gin
- **数据库**: PostgreSQL
- **认证**: JWT (ES256)
- **日志**: Logrus
- **配置**: 环境变量 + .env 文件

## 快速开始(Linux)

### 1. 环境准备

确保已安装：
- Go 1.21+
- PostgreSQL 12+

### 2. 克隆项目

```bash
git clone https://github.com/MenthaMC/webapi
cd webapi
```

### 3. 安装依赖

```bash
go mod download
```

### 4. 配置环境变量

复制环境变量模板：
```bash
cp .env.example .env
```

编辑 `.env` 文件，配置必要的环境变量：
- `DB_URL`: PostgreSQL 连接字符串
- `API_PUBLIC_KEY`: JWT 公钥
- `API_PRIVATE_KEY`: JWT 私钥

### 5. 初始化数据库
```bash
docker-compose up -d
```
使用docker创建数据库，应用启动时会自动执行数据库初始化脚本。

### 6. 启动服务

```bash
go run main.go
```

服务将在 `http://localhost:32767` 启动。

## API 文档

启动服务后，访问 `http://localhost:32767/v2/docs` 查看 Swagger API 文档。

## 项目结构

```
.
├── main.go                 # 应用入口
├── go.mod                  # Go 模块文件
├── .env.example           # 环境变量模板
├── README.md              # 项目文档
├── internal/              # 内部包
│   ├── app/              # 应用配置
│   ├── config/           # 配置管理
│   ├── database/         # 数据库连接
│   ├── handlers/         # HTTP 处理器
│   ├── logger/           # 日志配置
│   ├── middleware/       # 中间件
│   ├── models/           # 数据模型
│   ├── services/         # 业务逻辑
│   └── utils/            # 工具函数
├── public/               # 静态文件
│   ├── docs.html        # API 文档页面
│   ├── api-v2.json      # OpenAPI 规范
│   └── favicon.ico      # 网站图标
└── sql/                 # 数据库脚本
    └── init.sql         # 初始化脚本
```

## API 端点

### 查询接口

- `GET /v2/projects` - 获取项目列表
- `GET /v2/projects/{project}` - 获取项目详情
- `GET /v2/projects/{project}/versions/{version}` - 获取版本信息
- `GET /v2/projects/{project}/versions/{version}/builds` - 获取构建列表
- `GET /v2/projects/{project}/versions/{version}/builds/{build}` - 获取构建详情
- `GET /v2/projects/{project}/versions/{version}/latestGroupBuildId` - 获取最新构建ID
- `GET /v2/projects/{project}/versions/{version}/differ/{verRef}` - 获取版本差异
- `GET /v2/projects/{project}/version_group/{family}` - 获取版本组信息
- `GET /v2/projects/{project}/version_group/{family}/builds` - 获取版本组构建列表

### 下载接口

- `GET /v2/projects/{project}/versions/{version}/builds/{build}/downloads/{download}` - 下载构建文件

### 管理接口（需要认证）

- `POST /v2/commit/build` - 提交新构建
- `POST /v2/commit/build/download_source` - 添加下载源
- `POST /v2/delete/build/download_source` - 删除下载源

## 认证

管理接口需要 JWT 认证。在请求头中添加：
```
Authentication: <JWT_TOKEN>
```

## 开发

### 运行测试

```bash
go test ./...
```

### 构建

```bash
go build -o webapi main.go
```

### 生产部署

1. 构建二进制文件
2. 配置环境变量
3. 确保数据库可访问
4. 运行二进制文件

## 环境变量

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| PORT | 否 | 32767 | 服务端口 |
| DB_URL | 是 | - | PostgreSQL 连接字符串 |
| LOG_LEVEL | 否 | info | 日志级别 |
| API_PUBLIC_KEY | 是 | - | JWT 公钥 |
| API_PRIVATE_KEY | 是 | - | JWT 私钥 |
| API_ISSUER | 否 | MenthaMC | JWT 发行者 |
| API_SUBJECT | 否 | leaves-ci | JWT 主题 |
| API_ALGO | 否 | ES256 | JWT 算法 |
| COMMIT_BUILD_WEBHOOK_URL | 否 | - | 构建提交 Webhook URL |
| GITHUB_TOKEN | 否 | - | GitHub Personal Access Token |
| GITHUB_SYNC_INTERVAL | 否 | 1h | GitHub Releases 同步间隔 |

## GitHub Releases 自动同步

新增的 GitHub Releases 自动同步功能可以自动从 GitHub 仓库拉取 Releases 信息并同步到数据库中。

详细配置和使用说明请参考：[RELEASES_SYNC.md](./RELEASES_SYNC.md)

### 快速配置

1. 获取 GitHub Personal Access Token
2. 在 `.env` 文件中配置：
   ```env
   GITHUB_TOKEN=your_github_token_here
   GITHUB_SYNC_INTERVAL=1h
   ```
3. 重启服务即可自动开始同步

### 新增 API 端点

- `GET /v2/sync/status` - 获取同步状态
- `POST /v2/sync/trigger` - 手动触发全量同步（需要认证）
- `POST /v2/projects/{project}/sync` - 同步指定项目（需要认证）

## 许可证

[MIT License](./LICENSE)
