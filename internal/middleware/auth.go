package middleware

import (
	"net/http"
	"strings"
	"webapi/internal/config"
	"webapi/internal/utils"

	"github.com/gin-gonic/gin"
)

func Authentication(jwtConfig config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Authentication 头
		authHeader := c.GetHeader("Authentication")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Unauthorized",
			})
			c.Abort()
			return
		}

		// 移除 Bearer 前缀（如果存在）
		token := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 验证 JWT
		if err := utils.ValidateJWT(token, jwtConfig); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Unauthorized",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}