package handlers

import (
	"database/sql"
	"webapi/internal/config"
	"webapi/internal/services"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	config   *config.Config
	db       *sql.DB
	services *services.Services
}

func New(cfg *config.Config, database *sql.DB) *Handlers {
	return &Handlers{
		config:   cfg,
		db:       database,
		services: services.New(database),
	}
}

// StartScheduler 启动调度器
func (h *Handlers) StartScheduler() {
	h.services.Scheduler.Start()
}

// StopScheduler 停止调度器
func (h *Handlers) StopScheduler() {
	h.services.Scheduler.Stop()
}

// GetSchedulerStatus 获取调度器状态
func (h *Handlers) GetSchedulerStatus(c *gin.Context) {
	status := h.services.Scheduler.GetSchedulerStatus()
	c.JSON(200, status)
}

// TriggerAllSync 触发所有项目同步
func (h *Handlers) TriggerAllSync(c *gin.Context) {
	if err := h.services.Scheduler.TriggerSync(); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Sync triggered for all projects"})
}
