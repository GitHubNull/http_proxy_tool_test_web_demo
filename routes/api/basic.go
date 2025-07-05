package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"http_proxy_tool_test_web_demo/routes"

	"github.com/gin-gonic/gin"
)

// BasicAPIModule 基础API模块
type BasicAPIModule struct{}

// RegisterRoutes 注册路由
func (m *BasicAPIModule) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// 基础测试接口
		api.GET("/test", handleTest)
		api.POST("/test", handleTest)
		api.PUT("/test", handleTest)
		api.DELETE("/test", handleTest)
		api.PATCH("/test", handleTest)
		api.HEAD("/test", handleTest)
		api.OPTIONS("/test", handleTest)

		// 状态码测试
		api.GET("/status/:code", handleStatusCode)
		api.POST("/status/:code", handleStatusCode)

		// 延迟测试
		api.GET("/delay/:seconds", handleDelay)
		api.POST("/delay/:seconds", handleDelay)

		// 重定向测试
		api.GET("/redirect/:times", handleRedirect)
		api.GET("/redirect-to", handleRedirectTo)

		// 错误测试
		api.GET("/error", handleError)
		api.GET("/timeout", handleTimeout)
	}
}

// GetPrefix 获取前缀
func (m *BasicAPIModule) GetPrefix() string {
	return "/api"
}

// GetDescription 获取描述
func (m *BasicAPIModule) GetDescription() string {
	return "基础API测试接口"
}

// 通用测试处理器
func handleTest(c *gin.Context) {
	requestInfo := getRequestInfo(c)
	response := routes.CreateSuccessResponse("请求成功", requestInfo)
	c.JSON(http.StatusOK, response)
}

// 状态码测试
func handleStatusCode(c *gin.Context) {
	codeStr := c.Param("code")
	code, err := strconv.Atoi(codeStr)
	if err != nil {
		code = 200
	}

	requestInfo := getRequestInfo(c)
	response := routes.ApiResponse{
		Code:      code,
		Message:   http.StatusText(code),
		Data:      requestInfo,
		Timestamp: time.Now().Unix(),
		RequestID: routes.GenerateRequestID(),
	}

	c.JSON(code, response)
}

// 延迟测试
func handleDelay(c *gin.Context) {
	secondsStr := c.Param("seconds")
	seconds, err := strconv.Atoi(secondsStr)
	if err != nil || seconds < 0 || seconds > 30 {
		seconds = 1
	}

	time.Sleep(time.Duration(seconds) * time.Second)

	requestInfo := getRequestInfo(c)
	message := fmt.Sprintf("延迟 %d 秒后返回", seconds)
	response := routes.CreateSuccessResponse(message, requestInfo)
	c.JSON(http.StatusOK, response)
}

// 重定向测试
func handleRedirect(c *gin.Context) {
	timesStr := c.Param("times")
	times, err := strconv.Atoi(timesStr)
	if err != nil || times < 1 || times > 10 {
		times = 1
	}

	if times == 1 {
		c.Redirect(http.StatusFound, "/api/test")
	} else {
		c.Redirect(http.StatusFound, fmt.Sprintf("/api/redirect/%d", times-1))
	}
}

// 重定向到指定URL
func handleRedirectTo(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		url = "/api/test"
	}
	c.Redirect(http.StatusFound, url)
}

// 错误测试
func handleError(c *gin.Context) {
	response := routes.CreateErrorResponse(500, "这是一个测试错误")
	c.JSON(http.StatusInternalServerError, response)
}

// 超时测试
func handleTimeout(c *gin.Context) {
	time.Sleep(30 * time.Second)
	response := routes.CreateSuccessResponse("超时测试完成", nil)
	c.JSON(http.StatusOK, response)
}

// 获取请求信息
func getRequestInfo(c *gin.Context) *routes.RequestInfo {
	// 读取请求体
	var body interface{}
	if c.Request.Body != nil {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if len(bodyBytes) > 0 {
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "application/json") {
				if err := json.Unmarshal(bodyBytes, &body); err != nil {
					body = string(bodyBytes)
				}
			} else {
				body = string(bodyBytes)
			}
		}
	}

	// 获取请求头
	headers := make(map[string]interface{})
	for name, values := range c.Request.Header {
		if len(values) == 1 {
			headers[name] = values[0]
		} else {
			headers[name] = values
		}
	}

	// 获取查询参数
	query := make(map[string]interface{})
	for key, values := range c.Request.URL.Query() {
		if len(values) == 1 {
			query[key] = values[0]
		} else {
			query[key] = values
		}
	}

	// 获取Cookie
	cookies := make(map[string]string)
	for _, cookie := range c.Request.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}

	return &routes.RequestInfo{
		Method:      c.Request.Method,
		URL:         c.Request.URL.String(),
		Headers:     headers,
		Body:        body,
		Query:       query,
		ClientIP:    c.ClientIP(),
		UserAgent:   c.GetHeader("User-Agent"),
		ContentType: c.GetHeader("Content-Type"),
		Cookies:     cookies,
	}
}
