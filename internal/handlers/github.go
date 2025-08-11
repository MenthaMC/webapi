package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) ProxyGithubApi(c *gin.Context) {
	targetUrl := "https://api.github.com" + c.Param("path")

	target, err := url.Parse(targetUrl)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid target URL")
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Director = func(req *http.Request) {
		req.URL = target
		req.Host = target.Host
		req.Header = c.Request.Header
		req.Header.Set("Authorization", "token "+os.Getenv("GITHUB_TOKEN"))

		req.Method = c.Request.Method
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
