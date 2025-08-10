package services

import (
	"context"
	"fmt"
	"sync"
	"time"
	"webapi/internal/logger"
)

type SchedulerService struct {
	services      *Services
	githubService *GitHubService
	interval      time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	running       bool
	mu            sync.RWMutex
}

func NewSchedulerService(services *Services, githubToken string, interval time.Duration) *SchedulerService {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &SchedulerService{
		services:      services,
		githubService: NewGitHubService(githubToken),
		interval:      interval,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start 启动定时任务
func (s *SchedulerService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.running {
		logger.Warn("Scheduler is already running")
		return
	}
	
	s.running = true
	s.wg.Add(1)
	
	go s.run()
	logger.Info("Release sync scheduler started")
}

// Stop 停止定时任务
func (s *SchedulerService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.running {
		return
	}
	
	s.running = false
	s.cancel()
	s.wg.Wait()
	
	logger.Info("Release sync scheduler stopped")
}

// IsRunning 检查调度器是否正在运行
func (s *SchedulerService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// run 执行定时任务
func (s *SchedulerService) run() {
	defer s.wg.Done()
	
	// 立即执行一次同步
	s.syncAllProjects()
	
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.syncAllProjects()
		}
	}
}

// syncAllProjects 同步所有项目的releases
func (s *SchedulerService) syncAllProjects() {
	logger.Info("Starting scheduled release sync for all projects")
	
	projects, err := s.services.Project.GetAll()
	if err != nil {
		logger.Errorf("Failed to get projects for sync: %v", err)
		return
	}
	
	successCount := 0
	errorCount := 0
	
	for _, project := range projects {
		// 只同步GitHub仓库
		if !s.isGitHubRepo(project.Repo) {
			continue
		}
		
		err := s.githubService.SyncProjectReleases(&project, s.services)
		if err != nil {
			logger.Errorf("Failed to sync releases for project %s: %v", project.ID, err)
			errorCount++
		} else {
			successCount++
		}
		
		// 检查是否需要停止
		select {
		case <-s.ctx.Done():
			return
		default:
		}
		
		// 添加延迟以避免API限制
		time.Sleep(1 * time.Second)
	}
	
	logger.Infof("Release sync completed. Success: %d, Errors: %d", successCount, errorCount)
}

// isGitHubRepo 检查是否为GitHub仓库
func (s *SchedulerService) isGitHubRepo(repoURL string) bool {
	return repoURL != "" && (
		len(repoURL) > 19 && repoURL[:19] == "https://github.com/" ||
		len(repoURL) > 15 && repoURL[:15] == "git@github.com:")
}

// TriggerSync 手动触发同步
func (s *SchedulerService) TriggerSync() error {
	go s.syncAllProjects()
	return nil
}

// SyncProject 同步指定项目
func (s *SchedulerService) SyncProject(projectID string) error {
	project, err := s.services.Project.GetByID(projectID)
	if err != nil {
		return err
	}
	
	if project == nil {
		return fmt.Errorf("project %s not found", projectID)
	}
	
	if !s.isGitHubRepo(project.Repo) {
		return fmt.Errorf("project %s is not a GitHub repository", projectID)
	}
	
	return s.githubService.SyncProjectReleases(project, s.services)
}