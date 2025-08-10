package handlers

import (
	"time"
	"webapi/internal/utils"

	"github.com/gin-gonic/gin"
)

// TriggerReleaseSync 手动触发releases同步
func (h *Handlers) TriggerReleaseSync(c *gin.Context) {
	if h.scheduler == nil {
		utils.BadRequestResponse(c, "Release sync scheduler is not enabled")
		return
	}

	err := h.scheduler.TriggerSync()
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"message": "Release sync triggered successfully",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// SyncProjectReleases 同步指定项目的releases
func (h *Handlers) SyncProjectReleases(c *gin.Context) {
	projectID := c.Param("project")

	if h.scheduler == nil {
		utils.BadRequestResponse(c, "Release sync scheduler is not enabled")
		return
	}

	err := h.scheduler.SyncProject(projectID)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"message":    "Project release sync completed successfully",
		"project_id": projectID,
		"time":       time.Now().Format(time.RFC3339),
	})
}

// GetSyncStatus 获取同步状态
func (h *Handlers) GetSyncStatus(c *gin.Context) {
	if h.scheduler == nil {
		utils.SuccessResponse(c, map[string]interface{}{
			"enabled": false,
			"message": "Release sync scheduler is not configured",
		})
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"enabled": true,
		"running": h.scheduler.IsRunning(),
		"time":    time.Now().Format(time.RFC3339),
	})
}