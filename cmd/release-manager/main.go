package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"webapi/internal/config"
	"webapi/internal/database"
	"webapi/internal/models"
	"webapi/internal/services"
)

func main() {
	var (
		action    = flag.String("action", "", "操作类型: config, sync, list")
		project   = flag.String("project", "", "项目ID")
		owner     = flag.String("owner", "", "仓库所有者")
		repo      = flag.String("repo", "", "仓库名称")
		token     = flag.String("token", "", "GitHub访问令牌")
		interval  = flag.Int("interval", 60, "同步间隔(分钟)")
		autoSync  = flag.Bool("auto", false, "启用自动同步")
		enabled   = flag.Bool("enabled", true, "启用配置")
	)
	flag.Parse()

	if *action == "" {
		fmt.Println("使用方法:")
		fmt.Println("  配置项目: -action=config -project=<项目ID> -owner=<所有者> -repo=<仓库名> [-token=<令牌>] [-interval=<间隔>] [-auto] [-enabled=false]")
		fmt.Println("  同步项目: -action=sync -project=<项目ID>")
		fmt.Println("  列出配置: -action=list")
		os.Exit(1)
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	db, err := database.Init(cfg.Database.URL)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.Close()

	// 创建服务
	releaseService := services.NewReleaseService(db)

	switch *action {
	case "config":
		if *project == "" || *owner == "" || *repo == "" {
			log.Fatal("配置操作需要指定 project, owner, repo 参数")
		}
		
		config := &models.ReleaseConfig{
			Project:      *project,
			RepoOwner:    *owner,
			RepoName:     *repo,
			AccessToken:  *token,
			AutoSync:     *autoSync,
			SyncInterval: *interval,
			Enabled:      *enabled,
		}
		
		if err := releaseService.SaveReleaseConfig(config); err != nil {
			log.Fatalf("保存配置失败: %v", err)
		}
		
		fmt.Printf("项目 %s 的Release配置已保存\n", *project)
		
	case "sync":
		if *project == "" {
			log.Fatal("同步操作需要指定 project 参数")
		}
		
		fmt.Printf("开始同步项目 %s 的Releases...\n", *project)
		if err := releaseService.SyncReleases(*project); err != nil {
			log.Fatalf("同步失败: %v", err)
		}
		fmt.Printf("项目 %s 同步完成\n", *project)
		
	case "list":
		rows, err := db.Query(`
			SELECT project, repo_owner, repo_name, auto_sync, sync_interval, 
			       last_sync_at, enabled
			FROM release_configs
			ORDER BY project
		`)
		if err != nil {
			log.Fatalf("查询配置失败: %v", err)
		}
		defer rows.Close()
		
		fmt.Println("Release配置列表:")
		fmt.Println("================")
		
		for rows.Next() {
			var config models.ReleaseConfig
			var lastSyncAt *string
			
			err := rows.Scan(
				&config.Project, &config.RepoOwner, &config.RepoName,
				&config.AutoSync, &config.SyncInterval, &lastSyncAt, &config.Enabled,
			)
			if err != nil {
				log.Printf("扫描行失败: %v", err)
				continue
			}
			
			fmt.Printf("项目: %s\n", config.Project)
			fmt.Printf("  仓库: %s/%s\n", config.RepoOwner, config.RepoName)
			fmt.Printf("  自动同步: %t\n", config.AutoSync)
			fmt.Printf("  同步间隔: %d分钟\n", config.SyncInterval)
			fmt.Printf("  启用状态: %t\n", config.Enabled)
			if lastSyncAt != nil {
				fmt.Printf("  最后同步: %s\n", *lastSyncAt)
			} else {
				fmt.Printf("  最后同步: 从未同步\n")
			}
			fmt.Println()
		}
		
	default:
		log.Fatalf("未知操作: %s", *action)
	}
}