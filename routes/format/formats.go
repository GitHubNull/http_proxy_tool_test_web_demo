package format

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"http_proxy_tool_test_web_demo/routes"

	"github.com/gin-gonic/gin"
)

// FormatModule 格式测试模块
type FormatModule struct{}

// XMLResponse XML响应结构
type XMLResponse struct {
	XMLName xml.Name `xml:"response"`
	Code    int      `xml:"code"`
	Message string   `xml:"message"`
	Data    string   `xml:"data"`
	Time    string   `xml:"timestamp"`
}

// RegisterRoutes 注册路由
func (m *FormatModule) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// 格式响应测试
		api.GET("/json", handleJSON)
		api.GET("/xml", handleXML)
		api.GET("/html", handleHTML)
		api.GET("/text", handleText)
		api.GET("/binary", handleBinary)

		// 格式解析测试
		api.POST("/parse/json", handleParseJSON)
		api.POST("/parse/xml", handleParseXML)
		api.POST("/parse/multipart", handleParseMultipart)
		api.POST("/parse/binary", handleParseBinary)

		// 复杂格式测试
		api.POST("/complex/nested-json", handleNestedJSON)
		api.POST("/complex/large-xml", handleLargeXML)
		api.POST("/complex/mixed-multipart", handleMixedMultipart)

		// 压缩和编码测试
		api.GET("/gzip", handleGzip)
		api.GET("/deflate", handleDeflate)
		api.GET("/base64", handleBase64)

		// 流式数据测试
		api.GET("/stream/:lines", handleStream)
		api.GET("/bytes/:size", handleBytes)
		api.POST("/bytes/:size", handleBytes)
	}
}

// GetPrefix 获取前缀
func (m *FormatModule) GetPrefix() string {
	return "/api"
}

// GetDescription 获取描述
func (m *FormatModule) GetDescription() string {
	return "数据格式处理和解析测试接口"
}

// JSON响应测试
func handleJSON(c *gin.Context) {
	data := map[string]interface{}{
		"string":  "Hello, World!",
		"number":  42,
		"boolean": true,
		"array":   []int{1, 2, 3, 4, 5},
		"object": map[string]interface{}{
			"nested": "value",
			"count":  100,
		},
		"null": nil,
	}

	response := routes.CreateSuccessResponse("JSON响应测试", data)
	c.JSON(http.StatusOK, response)
}

// XML响应测试
func handleXML(c *gin.Context) {
	xmlResp := XMLResponse{
		Code:    200,
		Message: "XML响应测试",
		Data:    "这是XML格式的响应数据",
		Time:    time.Now().Format("2006-01-02 15:04:05"),
	}

	c.XML(http.StatusOK, xmlResp)
}

// HTML响应测试
func handleHTML(c *gin.Context) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>HTML响应测试</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .content { margin: 20px 0; }
        .footer { color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>HTML响应测试</h1>
            <p>这是一个HTML格式的响应页面</p>
        </div>
        <div class="content">
            <h2>测试信息</h2>
            <ul>
                <li>时间: ` + time.Now().Format("2006-01-02 15:04:05") + `</li>
                <li>状态: 成功</li>
                <li>格式: HTML</li>
                <li>编码: UTF-8</li>
            </ul>
        </div>
        <div class="footer">
            <p>HTTP代理测试工具 - HTML格式响应测试</p>
        </div>
    </div>
</body>
</html>`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// 文本响应测试
func handleText(c *gin.Context) {
	text := fmt.Sprintf(`文本响应测试

时间: %s
状态: 成功
格式: 纯文本
编码: UTF-8

这是一个纯文本格式的响应。
支持多行文本内容。
可以包含各种字符和符号。

测试数据:
- 数字: 12345
- 字符串: Hello, World!
- 特殊字符: @#$%%^&*()
- 中文: 你好，世界！
- 日语: こんにちは
- 韩语: 안녕하세요

结束标记: [END]`, time.Now().Format("2006-01-02 15:04:05"))

	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(text))
}

// 二进制响应测试
func handleBinary(c *gin.Context) {
	// 创建一个简单的二进制数据
	data := make([]byte, 1024)
	for i := 0; i < len(data); i++ {
		data[i] = byte(i % 256)
	}

	c.Data(http.StatusOK, "application/octet-stream", data)
}

// JSON解析测试
func handleParseJSON(c *gin.Context) {
	var jsonData interface{}
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		response := routes.CreateErrorResponse(400, "JSON解析失败: "+err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	result := map[string]interface{}{
		"parsed_data": jsonData,
		"data_type":   fmt.Sprintf("%T", jsonData),
		"parse_time":  time.Now().Unix(),
	}

	response := routes.CreateSuccessResponse("JSON解析成功", result)
	c.JSON(http.StatusOK, response)
}

// XML解析测试
func handleParseXML(c *gin.Context) {
	xmlBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response := routes.CreateErrorResponse(400, "读取XML数据失败: "+err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var xmlData interface{}
	if err := xml.Unmarshal(xmlBody, &xmlData); err != nil {
		response := routes.CreateErrorResponse(400, "XML解析失败: "+err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	result := map[string]interface{}{
		"parsed_data":  xmlData,
		"original_xml": string(xmlBody),
		"data_size":    len(xmlBody),
		"parse_time":   time.Now().Unix(),
	}

	response := routes.CreateSuccessResponse("XML解析成功", result)
	c.JSON(http.StatusOK, response)
}

// Multipart解析测试
func handleParseMultipart(c *gin.Context) {
	// 解析multipart表单
	form, err := c.MultipartForm()
	if err != nil {
		response := routes.CreateErrorResponse(400, "Multipart解析失败: "+err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 处理表单字段
	fields := make(map[string]interface{})
	for key, values := range form.Value {
		if len(values) == 1 {
			fields[key] = values[0]
		} else {
			fields[key] = values
		}
	}

	// 处理文件
	files := make(map[string]interface{})
	for key, fileHeaders := range form.File {
		fileInfos := make([]map[string]interface{}, 0)
		for _, fileHeader := range fileHeaders {
			fileInfos = append(fileInfos, map[string]interface{}{
				"filename": fileHeader.Filename,
				"size":     fileHeader.Size,
				"header":   fileHeader.Header,
			})
		}
		files[key] = fileInfos
	}

	result := map[string]interface{}{
		"fields":     fields,
		"files":      files,
		"parse_time": time.Now().Unix(),
	}

	response := routes.CreateSuccessResponse("Multipart解析成功", result)
	c.JSON(http.StatusOK, response)
}

// 二进制解析测试
func handleParseBinary(c *gin.Context) {
	binaryData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response := routes.CreateErrorResponse(400, "读取二进制数据失败: "+err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 分析二进制数据
	analysis := map[string]interface{}{
		"size":         len(binaryData),
		"first_bytes":  binaryData[:min(16, len(binaryData))],
		"last_bytes":   binaryData[max(0, len(binaryData)-16):],
		"content_type": c.GetHeader("Content-Type"),
		"parse_time":   time.Now().Unix(),
	}

	// 检查是否是文本数据
	isText := true
	for _, b := range binaryData[:min(100, len(binaryData))] {
		if b < 32 && b != 9 && b != 10 && b != 13 {
			isText = false
			break
		}
	}

	analysis["is_text"] = isText
	if isText && len(binaryData) > 0 {
		analysis["text_preview"] = string(binaryData[:min(200, len(binaryData))])
	}

	response := routes.CreateSuccessResponse("二进制数据解析成功", analysis)
	c.JSON(http.StatusOK, response)
}

// 嵌套JSON测试
func handleNestedJSON(c *gin.Context) {
	var jsonData interface{}
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		response := routes.CreateErrorResponse(400, "JSON解析失败: "+err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 创建复杂的嵌套响应
	complexData := map[string]interface{}{
		"input": jsonData,
		"metadata": map[string]interface{}{
			"parsed_at":  time.Now().Unix(),
			"request_id": routes.GenerateRequestID(),
			"processing": map[string]interface{}{
				"stage":    "validation",
				"status":   "success",
				"duration": 45,
				"checks": []map[string]interface{}{
					{"name": "structure", "passed": true},
					{"name": "encoding", "passed": true},
					{"name": "size", "passed": true},
				},
			},
		},
		"response": map[string]interface{}{
			"code":    200,
			"message": "复杂JSON处理成功",
			"data": map[string]interface{}{
				"processed": true,
				"items":     []string{"item1", "item2", "item3"},
				"nested": map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": map[string]interface{}{
							"level3": "深层嵌套数据",
						},
					},
				},
			},
		},
	}

	response := routes.CreateSuccessResponse("复杂JSON处理完成", complexData)
	c.JSON(http.StatusOK, response)
}

// 大型XML测试
func handleLargeXML(c *gin.Context) {
	// 创建大型XML响应
	var buffer bytes.Buffer
	buffer.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buffer.WriteString(`<large_xml_response>`)
	buffer.WriteString(`<metadata>`)
	buffer.WriteString(`<timestamp>` + time.Now().Format("2006-01-02T15:04:05Z") + `</timestamp>`)
	buffer.WriteString(`<request_id>` + routes.GenerateRequestID() + `</request_id>`)
	buffer.WriteString(`</metadata>`)
	buffer.WriteString(`<data>`)

	// 生成大量数据
	for i := 0; i < 100; i++ {
		buffer.WriteString(`<item id="` + strconv.Itoa(i) + `">`)
		buffer.WriteString(`<name>Item ` + strconv.Itoa(i) + `</name>`)
		buffer.WriteString(`<value>` + strconv.Itoa(i*10) + `</value>`)
		buffer.WriteString(`<description>这是第` + strconv.Itoa(i) + `个测试项目</description>`)
		buffer.WriteString(`<attributes>`)
		for j := 0; j < 5; j++ {
			buffer.WriteString(`<attr name="attr` + strconv.Itoa(j) + `">value` + strconv.Itoa(j) + `</attr>`)
		}
		buffer.WriteString(`</attributes>`)
		buffer.WriteString(`</item>`)
	}

	buffer.WriteString(`</data>`)
	buffer.WriteString(`</large_xml_response>`)

	c.Data(http.StatusOK, "application/xml; charset=utf-8", buffer.Bytes())
}

// 混合Multipart测试
func handleMixedMultipart(c *gin.Context) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// 添加文本字段
	writer.WriteField("text_field", "这是一个文本字段")
	writer.WriteField("number_field", "12345")
	writer.WriteField("json_field", `{"key": "value", "number": 42}`)

	// 添加文件字段
	fileWriter, _ := writer.CreateFormFile("test_file", "test.txt")
	fileWriter.Write([]byte("这是测试文件的内容\n包含多行数据\n用于测试文件上传"))

	// 添加二进制数据
	binaryWriter, _ := writer.CreateFormFile("binary_file", "binary.dat")
	binaryData := make([]byte, 256)
	for i := 0; i < 256; i++ {
		binaryData[i] = byte(i)
	}
	binaryWriter.Write(binaryData)

	writer.Close()

	c.Data(http.StatusOK, writer.FormDataContentType(), buffer.Bytes())
}

// Gzip压缩测试
func handleGzip(c *gin.Context) {
	data := map[string]interface{}{
		"message": "这是一个Gzip压缩测试",
		"data":    strings.Repeat("压缩测试数据 ", 100),
		"time":    time.Now().Unix(),
	}

	// Gin会自动处理gzip压缩
	c.Header("Content-Encoding", "gzip")
	response := routes.CreateSuccessResponse("Gzip压缩测试", data)
	c.JSON(http.StatusOK, response)
}

// Deflate压缩测试
func handleDeflate(c *gin.Context) {
	data := map[string]interface{}{
		"message": "这是一个Deflate压缩测试",
		"data":    strings.Repeat("压缩测试数据 ", 100),
		"time":    time.Now().Unix(),
	}

	c.Header("Content-Encoding", "deflate")
	response := routes.CreateSuccessResponse("Deflate压缩测试", data)
	c.JSON(http.StatusOK, response)
}

// Base64编码测试
func handleBase64(c *gin.Context) {
	data := "Hello, World! 这是Base64编码测试数据。"

	result := map[string]interface{}{
		"original": data,
		"encoded":  []byte(data), // Gin会自动进行base64编码
		"time":     time.Now().Unix(),
	}

	response := routes.CreateSuccessResponse("Base64编码测试", result)
	c.JSON(http.StatusOK, response)
}

// 流式数据测试
func handleStream(c *gin.Context) {
	linesStr := c.Param("lines")
	lines, err := strconv.Atoi(linesStr)
	if err != nil || lines < 1 || lines > 1000 {
		lines = 10
	}

	c.Stream(func(w io.Writer) bool {
		for i := 0; i < lines; i++ {
			fmt.Fprintf(w, "data: 这是第%d行流式数据 - 时间: %s\n", i+1, time.Now().Format("15:04:05.000"))
			time.Sleep(100 * time.Millisecond)
		}
		return false
	})
}

// 字节数据测试
func handleBytes(c *gin.Context) {
	sizeStr := c.Param("size")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 10000 {
		size = 1024
	}

	data := make([]byte, size)
	for i := 0; i < size; i++ {
		data[i] = byte(i % 256)
	}

	c.Data(http.StatusOK, "application/octet-stream", data)
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
