package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"webapi-v2-neo/internal/logger"
	"webapi-v2-neo/internal/models"
	"webapi-v2-neo/internal/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) CommitBuild(c *gin.Context) {
	var req models.CommitBuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// 解析变更数据
	changesData, err := h.services.Change.ParseChanges(req.Changes)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// 插入变更记录
	changeIDs, err := h.services.Change.InsertChanges(req.ProjectID, changesData)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	// 获取版本ID
	versionID, err := h.services.Version.GetVersionID(req.ProjectID, req.Version)
	if err != nil {
		utils.BadRequestResponse(c, "Version not found")
		return
	}

	// 获取版本组ID
	versionGroupID, err := h.services.Version.GetVersionGroupID(versionID)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	// 获取版本组中的所有版本
	versionIDs, _, err := h.services.Version.GetVersionsByGroupID(req.ProjectID, versionGroupID)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	// 获取最新构建ID
	latestBuildID, err := h.services.Version.GetLatestBuildID(req.ProjectID, versionIDs)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	// 创建新构建
	newBuildID := latestBuildID + 1
	if err := h.services.Build.CreateBuild(req, versionID, newBuildID, changeIDs); err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	// 触发 webhook
	go h.triggerWebhook(req.ProjectID, req.Version, req.Tag)

	utils.SuccessResponse(c, nil)
}

func (h *Handlers) CommitDownloadSource(c *gin.Context) {
	var req models.CommitDownloadSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// 检查下载源是否存在
	exists, err := h.services.Download.DownloadSourceExists(req.Project, req.Tag, req.DownloadSource)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	// 更新或插入下载记录
	if err := h.services.Download.UpsertDownload(req); err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	// 如果是新的下载源，添加到构建的下载源列表
	if !exists {
		if err := h.services.Build.AddDownloadSource(req.Project, req.Tag, req.DownloadSource); err != nil {
			utils.InternalServerErrorResponse(c)
			return
		}
	}

	utils.SuccessResponse(c, nil)
}

func (h *Handlers) DeleteDownloadSource(c *gin.Context) {
	var req models.DeleteDownloadSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// 检查下载源是否存在
	downloadSources, err := h.services.Build.GetDownloadSources(req.Project, req.Tag)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	found := false
	for _, source := range downloadSources {
		if source == req.DownloadSource {
			found = true
			break
		}
	}

	if !found {
		utils.NotFoundResponse(c, fmt.Sprintf("Specified download source %s not found for %s-%s", req.DownloadSource, req.Project, req.Tag))
		return
	}

	// 删除下载记录
	if err := h.services.Download.DeleteDownload(req); err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	// 从构建的下载源列表中移除
	if err := h.services.Build.RemoveDownloadSource(req.Project, req.Tag, req.DownloadSource); err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, nil)
}

func (h *Handlers) triggerWebhook(projectID, version, tag string) {
	if h.config.Webhook.CommitBuildURL == "" {
		return
	}

	// 获取项目仓库信息
	project, err := h.services.Project.GetByID(projectID)
	if err != nil {
		logger.Errorf("Failed to get project repository for webhook: %v", err)
		return
	}

	webhookData := map[string]string{
		"project":    projectID,
		"repository": project.Repo,
		"version":    version,
		"tag":        tag,
	}

	jsonData, err := json.Marshal(webhookData)
	if err != nil {
		logger.Errorf("Failed to marshal webhook data: %v", err)
		return
	}

	resp, err := http.Post(h.config.Webhook.CommitBuildURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Errorf("Failed to trigger commit build webhook %v: %v", webhookData, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		logger.Errorf("Webhook returned error status %d for %v", resp.StatusCode, webhookData)
	}
}