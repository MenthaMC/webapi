package handlers

import (
	"io/ioutil"
	"net/http"
	"path/filepath"
	"webapi-v2-neo/internal/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) RedirectToAPI(c *gin.Context) {
	c.Redirect(http.StatusFound, "/v2/docs")
}

func (h *Handlers) ServeFavicon(c *gin.Context) {
	faviconPath := filepath.Join("public", "favicon.ico")
	data, err := ioutil.ReadFile(faviconPath)
	if err != nil {
		utils.NotFoundResponse(c)
		return
	}

	c.Header("Content-Type", "image/x-icon")
	c.Data(http.StatusOK, "image/x-icon", data)
}

func (h *Handlers) ServeDocs(c *gin.Context) {
	docsPath := filepath.Join("public", "docs.html")
	data, err := ioutil.ReadFile(docsPath)
	if err != nil {
		utils.NotFoundResponse(c)
		return
	}

	c.Header("Content-Type", "text/html")
	c.Data(http.StatusOK, "text/html", data)
}

func (h *Handlers) ServeAPISpec(c *gin.Context) {
	apiPath := filepath.Join("public", "api-v2.json")
	data, err := ioutil.ReadFile(apiPath)
	if err != nil {
		utils.NotFoundResponse(c)
		return
	}

	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", data)
}

func (h *Handlers) Handle404(c *gin.Context) {
	utils.NotFoundResponse(c)
}