package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) ProxyGithubApi(c *gin.Context) {
	target, _ := url.Parse("https://api.github.com")

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Director = func(req *http.Request) {
		req.Method = c.Request.Method

		req.Header = c.Request.Header.Clone()

		req.Header.Set("Authorization", "token "+os.Getenv("GITHUB_TOKEN"))

		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host

		req.URL.Path = c.Request.URL.Path[len("/github"):]
		req.URL.RawQuery = c.Request.URL.RawQuery
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
