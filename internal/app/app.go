package app

import (
	"database/sql"
	"time"
	"webapi/internal/config"
	"webapi/internal/handlers"
	"webapi/internal/middleware"
	"webapi/internal/services"

	"github.com/gin-gonic/gin"
)

type App struct {
	config    *config.Config
	db        *sql.DB
	router    *gin.Engine
	scheduler *services.SchedulerService
}

func New(cfg *config.Config, database *sql.DB) *App {
	// 创建服务
	svc := services.New(database)
	
	// 创建调度器（如果配置了GitHub token）
	var scheduler *services.SchedulerService
	if cfg.GitHub.Token != "" {
		interval, err := time.ParseDuration(cfg.GitHub.SyncInterval)
		if err != nil {
			interval = time.Hour // 默认1小时
		}
		scheduler = services.NewSchedulerService(svc, cfg.GitHub.Token, interval)
		scheduler.Start() // 启动调度器
	}

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
		config:    cfg,
		db:        database,
		router:    router,
		scheduler: scheduler,
	}

	// 设置路由
	app.setupRoutes()

	return app
}

func (a *App) Run(addr string) error {
	return a.router.Run(addr)
}

func (a *App) GetRouter() *gin.Engine {
	return a.router
}

func (a *App) Stop() {
	if a.scheduler != nil {
		a.scheduler.Stop()
	}
}

func (a *App) setupRoutes() {
	h := handlers.New(a.config, a.db, a.scheduler)

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

		// Release同步相关路由
		v2.GET("/sync/status", h.GetSyncStatus)
		
		// 需要认证的路由
		authenticated := v2.Group("/")
		authenticated.Use(middleware.Authentication(a.config.JWT))
		{
			// 提交
			authenticated.POST("/commit/build", h.CommitBuild)
			authenticated.POST("/commit/build/download_source", h.CommitDownloadSource)
			
			// 删除
			authenticated.POST("/delete/build/download_source", h.DeleteDownloadSource)
			
			// Release同步
			authenticated.POST("/sync/trigger", h.TriggerReleaseSync)
			authenticated.POST("/projects/:project/sync", h.SyncProjectReleases)
		}
	}

	// 404 处理
	a.router.GET("/404/", h.Handle404)
	a.router.NoRoute(h.Handle404)
}