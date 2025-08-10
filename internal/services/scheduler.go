package services

import (
	"database/sql"
	"sync"
	"time"
	"webapi/internal/logger"
)

type SchedulerService struct {
	db       *sql.DB
	release  *ReleaseService
	stopChan chan bool
	wg       sync.WaitGroup
	running  bool
	mutex    sync.RWMutex
}

func NewSchedulerService(db *sql.DB, releaseService *ReleaseService) *SchedulerService {
	return &SchedulerService{
		db:       db,
		release:  releaseService,
		stopChan: make(chan bool),
	}
}

// Start 启动调度器
func (s *SchedulerService) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.running {
		return
	}
	
	s.running = true
	s.wg.Add(1)
	
	go s.run()
	logger.Info("Release scheduler started")
}

// Stop 停止调度器
func (s *SchedulerService) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if !s.running {
		return
	}
	
	s.running = false
	close(s.stopChan)
	s.wg.Wait()
	
	logger.Info("Release scheduler stopped")
}

// IsRunning 检查调度器是否运行中
func (s *SchedulerService) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.running
}

// run 调度器主循环
func (s *SchedulerService) run() {
	defer s.wg.Done()
	
	// 每分钟检查一次
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAndSyncReleases()
		}
	}
}

// checkAndSyncReleases 检查并同步需要更新的Releases
func (s *SchedulerService) checkAndSyncReleases() {
	// 查询需要同步的项目配置
	rows, err := s.db.Query(`
		SELECT project, sync_interval, last_sync_at 
		FROM release_configs 
		WHERE enabled = true AND auto_sync = true
	`)
	if err != nil {
		logger.Error("Failed to query release configs: " + err.Error())
		return
	}
	defer rows.Close()
	
	now := time.Now()
	
	for rows.Next() {
		var project string
		var syncInterval int
		var lastSyncAt sql.NullTime
		
		if err := rows.Scan(&project, &syncInterval, &lastSyncAt); err != nil {
			logger.Error("Failed to scan release config: " + err.Error())
			continue
		}
		
		// 检查是否需要同步
		shouldSync := false
		if !lastSyncAt.Valid {
			// 从未同步过
			shouldSync = true
		} else {
			// 检查是否超过同步间隔
			nextSyncTime := lastSyncAt.Time.Add(time.Duration(syncInterval) * time.Minute)
			shouldSync = now.After(nextSyncTime)
		}
		
		if shouldSync {
			logger.Info("Auto syncing releases for project: " + project)
			
			// 异步执行同步，避免阻塞其他项目
			go func(projectID string) {
				if err := s.release.SyncReleases(projectID); err != nil {
					logger.Error("Failed to auto sync releases for project " + projectID + ": " + err.Error())
				}
			}(project)
		}
	}
}

// GetSchedulerStatus 获取调度器状态
func (s *SchedulerService) GetSchedulerStatus() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// 获取启用自动同步的项目数量
	var enabledCount int
	s.db.QueryRow("SELECT COUNT(*) FROM release_configs WHERE enabled = true AND auto_sync = true").Scan(&enabledCount)
	
	return map[string]interface{}{
		"running":        s.running,
		"enabled_projects": enabledCount,
		"check_interval": "1 minute",
	}
}

// TriggerSync 手动触发所有项目的同步
func (s *SchedulerService) TriggerSync() error {
	rows, err := s.db.Query("SELECT project FROM release_configs WHERE enabled = true")
	if err != nil {
		return err
	}
	defer rows.Close()
	
	var projects []string
	for rows.Next() {
		var project string
		if err := rows.Scan(&project); err != nil {
			continue
		}
		projects = append(projects, project)
	}
	
	// 异步同步所有项目
	for _, project := range projects {
		go func(projectID string) {
			if err := s.release.SyncReleases(projectID); err != nil {
				logger.Error("Failed to sync releases for project " + projectID + ": " + err.Error())
			}
		}(project)
	}
	
	logger.Info("Triggered sync for all enabled projects")
	return nil
}