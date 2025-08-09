package handlers

import (
	"net/http"
	"webapi-v2-neo/internal/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) DownloadBuild(c *gin.Context) {
	projectID := c.Param("project")
	versionName := c.Param("version")
	buildIDStr := c.Param("build")
	downloadSource := c.Param("download")

	versionID, err := h.services.Version.GetVersionID(projectID, versionName)
	if err != nil {
		utils.NotFoundResponse(c)
		return
	}

	buildID, err := h.services.Build.ParseBuildID(projectID, versionID, buildIDStr)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	build, err := h.services.Build.GetBuild(projectID, versionID, buildID)
	if err != nil {
		utils.NotFoundResponse(c)
		return
	}

	downloadURL, err := h.services.Download.GetDownloadURL(downloadSource, projectID, build.Tag)
	if err != nil {
		utils.NotFoundResponse(c)
		return
	}

	c.Header("Content-Type", "application/java-archive")
	c.Redirect(http.StatusFound, downloadURL)
}