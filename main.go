package main

import (
	"fmt"
	"log"
	"webapi/internal/app"
	"webapi/internal/config"
	"webapi/internal/database"
	"webapi/internal/logger"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger.Init(cfg.LogLevel)

	// 初始化数据库
	db, err := database.Init(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 创建应用
	application := app.New(cfg, db)

	// 启动服务器
	logger.Info("MenthaMC WebAPI serve (Powered by Gin)")
	logger.Info(fmt.Sprintf("> Ready! Available at http://localhost:%d", cfg.Port))
	
	if err := application.Run(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}