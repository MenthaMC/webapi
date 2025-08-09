package handlers

import (
	"webapi-v2-neo/internal/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) GetProjects(c *gin.Context) {
	projects, err := h.services.Project.GetAll()
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	c.JSON(200, projects)
}

func (h *Handlers) GetProject(c *gin.Context) {
	projectID := c.Param("project")

	project, err := h.services.Project.GetByID(projectID)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}
	if project == nil {
		utils.NotFoundResponse(c)
		return
	}

	versions, err := h.services.Project.GetVersions(projectID)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	versionGroups, err := h.services.Project.GetVersionGroups(projectID)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"project_id":      project.ID,
		"project_name":    project.Name,
		"versions":        versions,
		"version_groups":  versionGroups,
	})
}