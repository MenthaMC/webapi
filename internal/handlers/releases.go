package handlers

import (
	"net/http"
	"strconv"
	"webapi/internal/logger"
	"webapi/internal/models"

	"github.com/gin-gonic/gin"
)

// GetReleases 获取项目的Releases列表
func (h *Handlers) GetReleases(c *gin.Context) {
	projectID := c.Param("project")
	
	// 解析分页参数
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}
	
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}
	
	// 检查项目是否存在
	project, err := h.services.Project.GetByID(projectID)
	if err != nil {
		logger.Error("Failed to get project: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	
	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// 获取Releases
	releases, err := h.services.Release.GetReleases(projectID, limit, offset)
	if err != nil {
		logger.Error("Failed to get releases: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"releases": releases,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(releases),
		},
	})
}

// GetLatestRelease 获取项目的最新Release
func (h *Handlers) GetLatestRelease(c *gin.Context) {
	projectID := c.Param("project")
	
	// 检查项目是否存在
	project, err := h.services.Project.GetByID(projectID)
	if err != nil {
		logger.Error("Failed to get project: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	
	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// 获取最新Release
	release, err := h.services.Release.GetLatestRelease(projectID)
	if err != nil {
		logger.Error("Failed to get latest release: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	
	if release == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No releases found"})
		return
	}
	
	c.JSON(http.StatusOK, release)
}

// GetReleaseConfig 获取项目的Release配置
func (h *Handlers) GetReleaseConfig(c *gin.Context) {
	projectID := c.Param("project")
	
	// 检查项目是否存在
	project, err := h.services.Project.GetByID(projectID)
	if err != nil {
		logger.Error("Failed to get project: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	
	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// 获取Release配置
	config, err := h.services.Release.GetReleaseConfig(projectID)
	if err != nil {
		logger.Error("Failed to get release config: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	
	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Release config not found"})
		return
	}
	
	// 隐藏敏感信息
	config.AccessToken = ""
	
	c.JSON(http.StatusOK, config)
}

// SaveReleaseConfig 保存Release配置
func (h *Handlers) SaveReleaseConfig(c *gin.Context) {
	projectID := c.Param("project")
	
	// 检查项目是否存在
	project, err := h.services.Project.GetByID(projectID)
	if err != nil {
		logger.Error("Failed to get project: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	
	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	var config models.ReleaseConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	
	// 设置项目ID
	config.Project = projectID
	
	// 验证必填字段
	if config.RepoOwner == "" || config.RepoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "repo_owner and repo_name are required"})
		return
	}
	
	// 设置默认值
	if config.SyncInterval <= 0 {
		config.SyncInterval = 60 // 默认60分钟
	}
	
	// 保存配置
	if err := h.services.Release.SaveReleaseConfig(&config); err != nil {
		logger.Error("Failed to save release config: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	
	logger.Info("Release config saved for project: " + projectID)
	c.JSON(http.StatusOK, gin.H{"message": "Release config saved successfully"})
}

// SyncReleases 手动同步Releases
func (h *Handlers) SyncReleases(c *gin.Context) {
	projectID := c.Param("project")
	
	// 检查项目是否存在
	project, err := h.services.Project.GetByID(projectID)
	if err != nil {
		logger.Error("Failed to get project: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	
	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// 执行同步
	if err := h.services.Release.SyncReleases(projectID); err != nil {
		logger.Error("Failed to sync releases: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Releases synced successfully"})
}