package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
	"webapi/internal/logger"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) ProxyGithubApi(c *gin.Context) {
	// 获取代理目标URL
	githubApiBase := os.Getenv("GITHUB_API_BASE")
	if githubApiBase == "" {
		githubApiBase = "https://api.github.com"
	}
	
	targetUrl := githubApiBase + c.Param("path")
	
	target, err := url.Parse(targetUrl)
	if err != nil {
		logger.Error("无法解析GitHub API URL: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的目标URL",
			"detail": err.Error(),
		})
		return
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(target)
	
	// 设置超时
	proxy.Transport = &http.Transport{
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
		Proxy: getProxyFunc(),
	}
	
	// 配置代理
	proxy.Director = func(req *http.Request) {
		req.URL = target
		req.Host = target.Host
		req.Header = c.Request.Header
		
		// 添加GitHub令牌
		token := os.Getenv("GITHUB_TOKEN")
		if token != "" {
			req.Header.Set("Authorization", "token "+token)
		}
		
		// 保留原始请求方法
		req.Method = c.Request.Method
		
		// 添加用户代理
		req.Header.Set("User-Agent", "MenthaMC-WebAPI/1.0")
	}
	
	// 错误处理
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		logger.Error("GitHub API代理错误: " + err.Error())
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "GitHub API代理错误",
			"detail": err.Error(),
		})
	}

	// 执行代理请求
	proxy.ServeHTTP(c.Writer, c.Request)
}

// 获取代理函数
func getProxyFunc() func(*http.Request) (*url.URL, error) {
	githubProxy := os.Getenv("GITHUB_PROXY")
	if githubProxy == "" {
		return http.ProxyFromEnvironment
	}
	
	proxyURL, err := url.Parse(githubProxy)
	if err != nil {
		logger.Error("无法解析GitHub代理URL: " + err.Error())
		return http.ProxyFromEnvironment
	}
	
	return func(*http.Request) (*url.URL, error) {
		return proxyURL, nil
	}
}
