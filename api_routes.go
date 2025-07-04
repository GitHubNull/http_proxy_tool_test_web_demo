package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 通用响应结构
type ApiResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
	RequestID string      `json:"request_id"`
}

// 请求信息结构
type RequestInfo struct {
	Method      string                 `json:"method"`
	URL         string                 `json:"url"`
	Headers     map[string]interface{} `json:"headers"`
	Body        interface{}            `json:"body"`
	Query       map[string]interface{} `json:"query"`
	ClientIP    string                 `json:"client_ip"`
	UserAgent   string                 `json:"user_agent"`
	ContentType string                 `json:"content_type"`
	Cookies     map[string]string      `json:"cookies"`
}

// 初始化API路由
func initAPIRoutes(r *gin.Engine) {
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

		// 响应格式测试
		api.GET("/json", handleJSON)
		api.GET("/xml", handleXML)
		api.GET("/html", handleHTML)
		api.GET("/text", handleText)
		api.GET("/binary", handleBinary)

		// 文件上传测试
		api.POST("/upload", handleFileUpload)
		api.POST("/upload-multiple", handleMultipleFileUpload)

		// 表单测试
		api.POST("/form", handleForm)
		api.POST("/form-data", handleFormData)

		// 认证测试
		api.GET("/auth/basic", handleBasicAuth)
		api.GET("/auth/bearer", handleBearerAuth)
		api.GET("/auth/digest", handleDigestAuth)

		// Cookie测试
		api.GET("/cookies", handleCookies)
		api.POST("/cookies/set", handleSetCookies)
		api.GET("/cookies/delete", handleDeleteCookies)

		// 请求头测试
		api.GET("/headers", handleHeaders)
		api.POST("/headers", handleHeaders)

		// 压缩测试
		api.GET("/gzip", handleGzip)
		api.GET("/deflate", handleDeflate)

		// 缓存测试
		api.GET("/cache/:seconds", handleCache)
		api.GET("/etag/:etag", handleEtag)

		// 流数据测试
		api.GET("/stream/:lines", handleStream)
		api.GET("/sse", handleSSE)

		// 大数据测试
		api.GET("/bytes/:size", handleBytes)
		api.POST("/bytes/:size", handleBytes)

		// 错误测试
		api.GET("/error", handleError)
		api.GET("/timeout", handleTimeout)
	}
}

// 通用测试处理器
func handleTest(c *gin.Context) {
	requestInfo := getRequestInfo(c)

	response := ApiResponse{
		Code:      200,
		Message:   "请求成功",
		Data:      requestInfo,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

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

	response := ApiResponse{
		Code:      code,
		Message:   http.StatusText(code),
		Data:      requestInfo,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
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

	response := ApiResponse{
		Code:      200,
		Message:   fmt.Sprintf("延迟 %d 秒后返回", seconds),
		Data:      requestInfo,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

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

// JSON响应
func handleJSON(c *gin.Context) {
	data := map[string]interface{}{
		"message": "这是一个JSON响应",
		"data": map[string]interface{}{
			"array":   []int{1, 2, 3, 4, 5},
			"object":  map[string]string{"key": "value"},
			"boolean": true,
			"number":  42,
			"null":    nil,
		},
	}

	response := ApiResponse{
		Code:      200,
		Message:   "JSON响应",
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// XML响应
func handleXML(c *gin.Context) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<response>
    <code>200</code>
    <message>这是一个XML响应</message>
    <data>
        <item id="1">项目1</item>
        <item id="2">项目2</item>
    </data>
    <timestamp>` + fmt.Sprintf("%d", time.Now().Unix()) + `</timestamp>
</response>`

	c.Data(http.StatusOK, "application/xml", []byte(xmlData))
}

// HTML响应
func handleHTML(c *gin.Context) {
	htmlData := `<!DOCTYPE html>
<html>
<head>
    <title>HTML响应测试</title>
</head>
<body>
    <h1>这是一个HTML响应</h1>
    <p>时间戳: ` + fmt.Sprintf("%d", time.Now().Unix()) + `</p>
</body>
</html>`

	c.Data(http.StatusOK, "text/html", []byte(htmlData))
}

// 文本响应
func handleText(c *gin.Context) {
	text := "这是一个纯文本响应\n时间戳: " + fmt.Sprintf("%d", time.Now().Unix())
	c.Data(http.StatusOK, "text/plain", []byte(text))
}

// 二进制响应
func handleBinary(c *gin.Context) {
	// 生成一个简单的二进制数据
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	c.Data(http.StatusOK, "application/octet-stream", data)
}

// 文件上传处理
func handleFileUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:      400,
			Message:   "文件上传失败: " + err.Error(),
			Data:      nil,
			Timestamp: time.Now().Unix(),
			RequestID: generateRequestID(),
		})
		return
	}
	defer file.Close()

	// 读取文件内容
	buf := new(bytes.Buffer)
	io.Copy(buf, file)

	data := map[string]interface{}{
		"filename": header.Filename,
		"size":     header.Size,
		"content":  buf.String(),
	}

	response := ApiResponse{
		Code:      200,
		Message:   "文件上传成功",
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 多文件上传处理
func handleMultipleFileUpload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:      400,
			Message:   "多文件上传失败: " + err.Error(),
			Data:      nil,
			Timestamp: time.Now().Unix(),
			RequestID: generateRequestID(),
		})
		return
	}

	files := form.File["files"]
	var filesInfo []map[string]interface{}

	for _, file := range files {
		f, err := file.Open()
		if err != nil {
			continue
		}
		defer f.Close()

		buf := new(bytes.Buffer)
		io.Copy(buf, f)

		fileInfo := map[string]interface{}{
			"filename": file.Filename,
			"size":     file.Size,
			"content":  buf.String(),
		}
		filesInfo = append(filesInfo, fileInfo)
	}

	response := ApiResponse{
		Code:      200,
		Message:   fmt.Sprintf("成功上传 %d 个文件", len(files)),
		Data:      filesInfo,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 表单处理
func handleForm(c *gin.Context) {
	var formData map[string]interface{}

	if err := c.ShouldBind(&formData); err != nil {
		// 手动解析表单数据
		formData = make(map[string]interface{})
		for key, values := range c.Request.PostForm {
			if len(values) == 1 {
				formData[key] = values[0]
			} else {
				formData[key] = values
			}
		}
	}

	response := ApiResponse{
		Code:      200,
		Message:   "表单数据接收成功",
		Data:      formData,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 表单数据处理
func handleFormData(c *gin.Context) {
	var formData map[string]interface{}

	if err := c.ShouldBind(&formData); err != nil {
		formData = make(map[string]interface{})
		for key, values := range c.Request.PostForm {
			if len(values) == 1 {
				formData[key] = values[0]
			} else {
				formData[key] = values
			}
		}
	}

	response := ApiResponse{
		Code:      200,
		Message:   "FormData接收成功",
		Data:      formData,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 基础认证
func handleBasicAuth(c *gin.Context) {
	user, password, hasAuth := c.Request.BasicAuth()

	if !hasAuth {
		c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
		c.JSON(http.StatusUnauthorized, ApiResponse{
			Code:      401,
			Message:   "需要基础认证",
			Data:      nil,
			Timestamp: time.Now().Unix(),
			RequestID: generateRequestID(),
		})
		return
	}

	data := map[string]interface{}{
		"user":          user,
		"password":      password,
		"authenticated": true,
	}

	response := ApiResponse{
		Code:      200,
		Message:   "基础认证成功",
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// Bearer Token认证
func handleBearerAuth(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, ApiResponse{
			Code:      401,
			Message:   "需要Bearer Token认证",
			Data:      nil,
			Timestamp: time.Now().Unix(),
			RequestID: generateRequestID(),
		})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	data := map[string]interface{}{
		"token":         token,
		"authenticated": true,
	}

	response := ApiResponse{
		Code:      200,
		Message:   "Bearer Token认证成功",
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// Digest认证
func handleDigestAuth(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" || !strings.HasPrefix(authHeader, "Digest ") {
		c.Header("WWW-Authenticate", `Digest realm="Restricted", nonce="dcd98b7102dd2f0e8b11d0f600bfb0c093", opaque="5ccc069c403ebaf9f0171e9517f40e41"`)
		c.JSON(http.StatusUnauthorized, ApiResponse{
			Code:      401,
			Message:   "需要Digest认证",
			Data:      nil,
			Timestamp: time.Now().Unix(),
			RequestID: generateRequestID(),
		})
		return
	}

	data := map[string]interface{}{
		"digest_header": authHeader,
		"authenticated": true,
	}

	response := ApiResponse{
		Code:      200,
		Message:   "Digest认证成功",
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// Cookie处理
func handleCookies(c *gin.Context) {
	cookies := make(map[string]string)

	for _, cookie := range c.Request.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}

	response := ApiResponse{
		Code:      200,
		Message:   "Cookie获取成功",
		Data:      cookies,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 设置Cookie
func handleSetCookies(c *gin.Context) {
	var cookieData map[string]string

	if err := c.ShouldBindJSON(&cookieData); err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:      400,
			Message:   "Cookie数据格式错误",
			Data:      nil,
			Timestamp: time.Now().Unix(),
			RequestID: generateRequestID(),
		})
		return
	}

	for name, value := range cookieData {
		c.SetCookie(name, value, 3600, "/", "", false, false)
	}

	response := ApiResponse{
		Code:      200,
		Message:   "Cookie设置成功",
		Data:      cookieData,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 删除Cookie
func handleDeleteCookies(c *gin.Context) {
	cookieName := c.Query("name")
	if cookieName == "" {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:      400,
			Message:   "请指定要删除的Cookie名称",
			Data:      nil,
			Timestamp: time.Now().Unix(),
			RequestID: generateRequestID(),
		})
		return
	}

	c.SetCookie(cookieName, "", -1, "/", "", false, false)

	response := ApiResponse{
		Code:      200,
		Message:   "Cookie删除成功",
		Data:      map[string]string{"deleted": cookieName},
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 请求头处理
func handleHeaders(c *gin.Context) {
	headers := make(map[string]interface{})

	for name, values := range c.Request.Header {
		if len(values) == 1 {
			headers[name] = values[0]
		} else {
			headers[name] = values
		}
	}

	response := ApiResponse{
		Code:      200,
		Message:   "请求头获取成功",
		Data:      headers,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// Gzip压缩
func handleGzip(c *gin.Context) {
	data := map[string]interface{}{
		"message": "这是一个Gzip压缩的响应",
		"data":    strings.Repeat("测试数据", 100),
	}

	response := ApiResponse{
		Code:      200,
		Message:   "Gzip压缩响应",
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.Header("Content-Encoding", "gzip")
	c.JSON(http.StatusOK, response)
}

// Deflate压缩
func handleDeflate(c *gin.Context) {
	data := map[string]interface{}{
		"message": "这是一个Deflate压缩的响应",
		"data":    strings.Repeat("测试数据", 100),
	}

	response := ApiResponse{
		Code:      200,
		Message:   "Deflate压缩响应",
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.Header("Content-Encoding", "deflate")
	c.JSON(http.StatusOK, response)
}

// 缓存测试
func handleCache(c *gin.Context) {
	secondsStr := c.Param("seconds")
	seconds, err := strconv.Atoi(secondsStr)
	if err != nil || seconds < 0 {
		seconds = 60
	}

	c.Header("Cache-Control", fmt.Sprintf("max-age=%d", seconds))
	c.Header("Last-Modified", time.Now().Format(http.TimeFormat))

	response := ApiResponse{
		Code:      200,
		Message:   fmt.Sprintf("缓存 %d 秒", seconds),
		Data:      map[string]interface{}{"cached_at": time.Now().Unix()},
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// ETag测试
func handleEtag(c *gin.Context) {
	etag := c.Param("etag")

	c.Header("ETag", fmt.Sprintf(`"%s"`, etag))

	if c.GetHeader("If-None-Match") == fmt.Sprintf(`"%s"`, etag) {
		c.Status(http.StatusNotModified)
		return
	}

	response := ApiResponse{
		Code:      200,
		Message:   "ETag测试",
		Data:      map[string]interface{}{"etag": etag},
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 流数据
func handleStream(c *gin.Context) {
	linesStr := c.Param("lines")
	lines, err := strconv.Atoi(linesStr)
	if err != nil || lines < 1 || lines > 1000 {
		lines = 10
	}

	c.Header("Content-Type", "text/plain")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	for i := 0; i < lines; i++ {
		c.Writer.WriteString(fmt.Sprintf("第 %d 行数据 - 时间戳: %d\n", i+1, time.Now().Unix()))
		c.Writer.Flush()
		time.Sleep(100 * time.Millisecond)
	}
}

// Server-Sent Events
func handleSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	for i := 0; i < 10; i++ {
		c.Writer.WriteString(fmt.Sprintf("data: {\"message\": \"事件 %d\", \"timestamp\": %d}\n\n", i+1, time.Now().Unix()))
		c.Writer.Flush()
		time.Sleep(1 * time.Second)
	}
}

// 字节数据
func handleBytes(c *gin.Context) {
	sizeStr := c.Param("size")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 1024*1024 {
		size = 1024
	}

	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	c.Data(http.StatusOK, "application/octet-stream", data)
}

// 错误测试
func handleError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, ApiResponse{
		Code:      500,
		Message:   "这是一个测试错误",
		Data:      nil,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	})
}

// 超时测试
func handleTimeout(c *gin.Context) {
	time.Sleep(30 * time.Second)

	response := ApiResponse{
		Code:      200,
		Message:   "超时测试完成",
		Data:      nil,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 获取请求信息
func getRequestInfo(c *gin.Context) *RequestInfo {
	// 读取请求体
	var body interface{}
	if c.Request.Body != nil {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if len(bodyBytes) > 0 {
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "application/json") {
				json.Unmarshal(bodyBytes, &body)
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

	return &RequestInfo{
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

// 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
