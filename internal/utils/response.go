package utils

import (
	"net/http"
	"webapi/internal/models"

	"github.com/gin-gonic/gin"
)

func SuccessResponse(c *gin.Context, data interface{}) {
	response := gin.H{"code": 200}
	
	if data != nil {
		// 如果 data 是 map，则合并到 response 中
		if dataMap, ok := data.(map[string]interface{}); ok {
			for k, v := range dataMap {
				response[k] = v
			}
		} else {
			// 否则直接设置 data
			response["data"] = data
		}
	}
	
	c.JSON(http.StatusOK, response)
}

func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"code": statusCode,
		"msg":  message,
	})
}

func NotFoundResponse(c *gin.Context, message ...string) {
	msg := "Not Found"
	if len(message) > 0 {
		msg = message[0]
	}
	ErrorResponse(c, http.StatusNotFound, msg)
}

func BadRequestResponse(c *gin.Context, message ...string) {
	msg := "Bad Request"
	if len(message) > 0 {
		msg = message[0]
	}
	ErrorResponse(c, http.StatusBadRequest, msg)
}

func UnauthorizedResponse(c *gin.Context) {
	ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
}

func InternalServerErrorResponse(c *gin.Context) {
	ErrorResponse(c, http.StatusInternalServerError, "An error occurred")
}

func BuildToBuildResponse(build models.Build, changes []models.ChangeResponse) models.BuildResponse {
	channel := "default"
	if build.Experimental {
		channel = "experimental"
	}

	downloads := make(map[string]models.DownloadInfo)
	for _, source := range build.DownloadSources {
		downloads[source] = models.DownloadInfo{
			Name:   build.JarName,
			SHA256: build.SHA256,
		}
	}

	return models.BuildResponse{
		Build:     build.BuildID,
		Time:      build.Time.Format("2006-01-02T15:04:05.000Z"),
		Channel:   channel,
		Promoted:  false, // 根据原始逻辑，promoted 总是 false
		Changes:   changes,
		Downloads: downloads,
	}
}