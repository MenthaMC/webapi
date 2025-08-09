package handlers

import (
	"webapi/internal/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) GetVersionGroup(c *gin.Context) {
	projectID := c.Param("project")
	family := c.Param("family")

	project, err := h.services.Project.GetByID(projectID)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}
	if project == nil {
		utils.NotFoundResponse(c)
		return
	}

	versionGroupID, err := h.services.VersionGroup.GetVersionGroupID(projectID, family)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}
	if versionGroupID == 0 {
		utils.NotFoundResponse(c)
		return
	}

	_, versionNames, err := h.services.VersionGroup.GetVersionsByGroupID(projectID, versionGroupID)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"project_id":     project.ID,
		"project_name":   project.Name,
		"version_group":  family,
		"versions":       versionNames,
	})
}

func (h *Handlers) GetVersionGroupBuilds(c *gin.Context) {
	projectID := c.Param("project")
	family := c.Param("family")

	project, err := h.services.Project.GetByID(projectID)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}
	if project == nil {
		utils.NotFoundResponse(c)
		return
	}

	versionGroupID, err := h.services.VersionGroup.GetVersionGroupID(projectID, family)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}
	if versionGroupID == 0 {
		utils.NotFoundResponse(c)
		return
	}

	versionIDs, versionNames, err := h.services.VersionGroup.GetVersionsByGroupID(projectID, versionGroupID)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	builds, err := h.services.Build.GetBuildsByVersions(projectID, versionIDs)
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
		"project_id":     project.ID,
		"project_name":   project.Name,
		"version_group":  family,
		"versions":       versionNames,
		"builds":         buildResponses,
	})
}