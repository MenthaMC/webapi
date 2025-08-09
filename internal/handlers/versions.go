package handlers

import (
	"strconv"
	"webapi-v2-neo/internal/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) GetVersion(c *gin.Context) {
	projectID := c.Param("project")
	versionName := c.Param("version")

	project, err := h.services.Project.GetByID(projectID)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}
	if project == nil {
		utils.NotFoundResponse(c)
		return
	}

	versionID, err := h.services.Version.GetVersionID(projectID, versionName)
	if err != nil {
		utils.NotFoundResponse(c)
		return
	}

	builds, err := h.services.Build.GetBuildsByVersion(projectID, versionID)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	var buildIDs []int
	for _, build := range builds {
		buildIDs = append(buildIDs, build.BuildID)
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"project_id":   project.ID,
		"project_name": project.Name,
		"version":      versionName,
		"builds":       buildIDs,
	})
}

func (h *Handlers) GetVersionBuilds(c *gin.Context) {
	projectID := c.Param("project")
	versionName := c.Param("version")

	project, err := h.services.Project.GetByID(projectID)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}
	if project == nil {
		utils.NotFoundResponse(c)
		return
	}

	versionID, err := h.services.Version.GetVersionID(projectID, versionName)
	if err != nil {
		utils.NotFoundResponse(c)
		return
	}

	builds, err := h.services.Build.GetBuildsByVersion(projectID, versionID)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	var buildResponses []interface{}
	for _, build := range builds {
		changes, err := h.services.Change.GetChangesByIDs([]int64(build.Changes))
		if err != nil {
			utils.InternalServerErrorResponse(c)
			return
		}

		buildResponse := utils.BuildToBuildResponse(build, changes)
		buildResponses = append(buildResponses, buildResponse)
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"project_id":   project.ID,
		"project_name": project.Name,
		"version":      versionName,
		"builds":       buildResponses,
	})
}

func (h *Handlers) GetBuild(c *gin.Context) {
	projectID := c.Param("project")
	versionName := c.Param("version")
	buildIDStr := c.Param("build")

	project, err := h.services.Project.GetByID(projectID)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}
	if project == nil {
		utils.NotFoundResponse(c)
		return
	}

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

	changes, err := h.services.Change.GetChangesByIDs([]int64(build.Changes))
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	buildResponse := utils.BuildToBuildResponse(*build, changes)

	response := map[string]interface{}{
		"project_id":   project.ID,
		"project_name": project.Name,
		"version":      versionName,
	}

	// 合并 buildResponse 的字段
	response["build"] = buildResponse.Build
	response["time"] = buildResponse.Time
	response["channel"] = buildResponse.Channel
	response["promoted"] = buildResponse.Promoted
	response["changes"] = buildResponse.Changes
	response["downloads"] = buildResponse.Downloads

	utils.SuccessResponse(c, response)
}

func (h *Handlers) GetLatestGroupBuildId(c *gin.Context) {
	projectID := c.Param("project")
	versionName := c.Param("version")

	versionID, err := h.services.Version.GetVersionID(projectID, versionName)
	if err != nil {
		utils.NotFoundResponse(c)
		return
	}

	versionGroupID, err := h.services.Version.GetVersionGroupID(versionID)
	if err != nil {
		utils.NotFoundResponse(c)
		return
	}

	versionIDs, _, err := h.services.Version.GetVersionsByGroupID(projectID, versionGroupID)
	if err != nil {
		utils.NotFoundResponse(c)
		return
	}

	latestBuildID, err := h.services.Version.GetLatestBuildID(projectID, versionIDs)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	c.Header("Content-Type", "text/plain")
	c.String(200, strconv.Itoa(latestBuildID))
}

func (h *Handlers) GetVersionDiffer(c *gin.Context) {
	projectID := c.Param("project")
	versionName := c.Param("version")
	verRef := c.Param("verRef")

	versionID, err := h.services.Version.GetVersionID(projectID, versionName)
	if err != nil {
		utils.NotFoundResponse(c)
		return
	}

	latestBuildID, err := h.services.Version.GetLatestBuildID(projectID, []int{versionID})
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	changeID, err := h.services.Change.GetChangeIDByCommitPrefix(projectID, verRef)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	referredBuildID, err := h.services.Change.GetBuildIDByChange(versionID, changeID)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	diff := latestBuildID - referredBuildID

	c.Header("Content-Type", "text/plain")
	c.String(200, strconv.Itoa(diff))
}