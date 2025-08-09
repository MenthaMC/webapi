package app

import (
	"database/sql"
	"webapi/internal/config"
	"webapi/internal/handlers"
	"webapi/internal/middleware"

	"github.com/gin-gonic/gin"
)

type App struct {
	config *config.Config
	db     *sql.DB
	router *gin.Engine
}

func New(cfg *config.Config, database *sql.DB) *App {
	// 设置 Gin 模式
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	
	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	app := &App{
		config: cfg,
		db:     database,
		router: router,
	}

	// 设置路由
	app.setupRoutes()

	return app
}

func (a *App) Run(addr string) error {
	return a.router.Run(addr)
}

func (a *App) setupRoutes() {
	h := handlers.New(a.config, a.db)

	// 静态文件和文档
	a.router.GET("/", h.RedirectToAPI)
	a.router.GET("/favicon.ico", h.ServeFavicon)
	a.router.GET("/v2/docs", h.ServeDocs)
	a.router.GET("/v2/api", h.ServeAPISpec)

	// API 路由组
	v2 := a.router.Group("/v2")
	{
		// 项目相关
		v2.GET("/projects", h.GetProjects)
		v2.GET("/projects/:project", h.GetProject)
		v2.GET("/projects/:project/versions/:version", h.GetVersion)
		v2.GET("/projects/:project/versions/:version/builds", h.GetVersionBuilds)
		v2.GET("/projects/:project/versions/:version/builds/:build", h.GetBuild)
		v2.GET("/projects/:project/versions/:version/latestGroupBuildId", h.GetLatestGroupBuildId)
		v2.GET("/projects/:project/versions/:version/differ/:verRef", h.GetVersionDiffer)
		v2.GET("/projects/:project/version_group/:family", h.GetVersionGroup)
		v2.GET("/projects/:project/version_group/:family/builds", h.GetVersionGroupBuilds)
		v2.GET("/projects/:project/versions/:version/builds/:build/downloads/:download", h.DownloadBuild)

		// 需要认证的路由
		authenticated := v2.Group("/")
		authenticated.Use(middleware.Authentication(a.config.JWT))
		{
			// 提交
			authenticated.POST("/commit/build", h.CommitBuild)
			authenticated.POST("/commit/build/download_source", h.CommitDownloadSource)
			
			// 删除
			authenticated.POST("/delete/build/download_source", h.DeleteDownloadSource)
		}
	}

	// 404 处理
	a.router.GET("/404/", h.Handle404)
	a.router.NoRoute(h.Handle404)
}