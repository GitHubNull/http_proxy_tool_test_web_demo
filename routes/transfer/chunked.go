package transfer

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"http_proxy_tool_test_web_demo/routes"

	"github.com/gin-gonic/gin"
)

// TransferModule 传输编码测试模块
type TransferModule struct{}

// RegisterRoutes 注册路由
func (m *TransferModule) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/transfer")
	{
		// 分块传输测试
		api.GET("/chunked", handleChunkedTransfer)
		api.POST("/chunked", handleChunkedReceive)
		api.GET("/chunked/stream", handleChunkedStream)
		api.POST("/chunked/upload", handleChunkedUpload)

		// 传输编码测试
		api.GET("/identity", handleIdentityTransfer)
		api.GET("/deflate", handleDeflateTransfer)
		api.GET("/gzip", handleGzipTransfer)

		// 大文件传输测试
		api.GET("/large/:size", handleLargeTransfer)
		api.POST("/large", handleLargeReceive)

		// 流式传输测试
		api.GET("/stream/sse", handleSSETransfer)
		api.GET("/stream/websocket", handleWebSocketTransfer)
	}
}

// GetPrefix 获取前缀
func (m *TransferModule) GetPrefix() string {
	return "/api/transfer"
}

// GetDescription 获取描述
func (m *TransferModule) GetDescription() string {
	return "传输编码和分块传输测试接口"
}

// 分块传输测试
func handleChunkedTransfer(c *gin.Context) {
	chunksStr := c.Query("chunks")
	chunks, err := strconv.Atoi(chunksStr)
	if err != nil || chunks < 1 || chunks > 100 {
		chunks = 5
	}

	delayStr := c.Query("delay")
	delay, err := strconv.Atoi(delayStr)
	if err != nil || delay < 0 || delay > 5000 {
		delay = 500
	}

	// 设置分块传输头
	c.Header("Transfer-Encoding", "chunked")
	c.Header("Content-Type", "application/json")
	c.Status(http.StatusOK)

	// 获取底层的ResponseWriter
	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		response := routes.CreateErrorResponse(500, "不支持分块传输")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// 发送分块数据
	for i := 0; i < chunks; i++ {
		chunk := map[string]interface{}{
			"chunk_id":  i + 1,
			"timestamp": time.Now().Unix(),
			"data":      fmt.Sprintf("这是第%d个分块数据", i+1),
			"size":      len(fmt.Sprintf("这是第%d个分块数据", i+1)),
			"remaining": chunks - i - 1,
		}

		// 写入分块数据
		chunkData := fmt.Sprintf(`{"chunk_id":%d,"timestamp":%d,"data":"这是第%d个分块数据","size":%d,"remaining":%d}`,
			chunk["chunk_id"], chunk["timestamp"], i+1, chunk["size"], chunk["remaining"])

		// 写入分块大小（十六进制）
		fmt.Fprintf(w, "%x\r\n", len(chunkData))
		// 写入分块数据
		fmt.Fprintf(w, "%s\r\n", chunkData)

		flusher.Flush()

		// 延迟发送下一个分块
		if i < chunks-1 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}

	// 结束分块传输
	fmt.Fprintf(w, "0\r\n\r\n")
	flusher.Flush()
}

// 分块接收测试
func handleChunkedReceive(c *gin.Context) {
	// 检查是否为分块传输
	transferEncoding := c.GetHeader("Transfer-Encoding")
	isChunked := strings.Contains(strings.ToLower(transferEncoding), "chunked")

	var receivedData strings.Builder
	var chunks []map[string]interface{}
	chunkCount := 0

	if isChunked {
		// 处理分块传输数据
		reader := bufio.NewReader(c.Request.Body)

		for {
			// 读取分块大小
			line, _, err := reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					break
				}
				response := routes.CreateErrorResponse(400, "读取分块大小失败: "+err.Error())
				c.JSON(http.StatusBadRequest, response)
				return
			}

			// 解析分块大小
			sizeStr := strings.TrimSpace(string(line))
			if sizeStr == "" {
				continue
			}

			size, err := strconv.ParseInt(sizeStr, 16, 64)
			if err != nil {
				response := routes.CreateErrorResponse(400, "解析分块大小失败: "+err.Error())
				c.JSON(http.StatusBadRequest, response)
				return
			}

			// 如果大小为0，表示传输结束
			if size == 0 {
				break
			}

			// 读取分块数据
			chunkData := make([]byte, size)
			_, err = io.ReadFull(reader, chunkData)
			if err != nil {
				response := routes.CreateErrorResponse(400, "读取分块数据失败: "+err.Error())
				c.JSON(http.StatusBadRequest, response)
				return
			}

			// 读取分块结束标记
			reader.ReadLine()

			chunkCount++
			chunks = append(chunks, map[string]interface{}{
				"chunk_id": chunkCount,
				"size":     size,
				"data":     string(chunkData),
			})

			receivedData.Write(chunkData)
		}
	} else {
		// 普通传输方式
		bodyData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			response := routes.CreateErrorResponse(400, "读取请求体失败: "+err.Error())
			c.JSON(http.StatusBadRequest, response)
			return
		}
		receivedData.Write(bodyData)
	}

	result := map[string]interface{}{
		"is_chunked":        isChunked,
		"transfer_encoding": transferEncoding,
		"chunk_count":       chunkCount,
		"total_size":        receivedData.Len(),
		"received_data":     receivedData.String(),
		"chunks":            chunks,
		"content_type":      c.GetHeader("Content-Type"),
		"content_length":    c.GetHeader("Content-Length"),
		"received_at":       time.Now().Unix(),
	}

	response := routes.CreateSuccessResponse("分块传输接收完成", result)
	c.JSON(http.StatusOK, response)
}

// 分块流式传输
func handleChunkedStream(c *gin.Context) {
	durationStr := c.Query("duration")
	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration < 1 || duration > 300 {
		duration = 10
	}

	intervalStr := c.Query("interval")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil || interval < 100 || interval > 10000 {
		interval = 1000
	}

	// 设置分块传输头
	c.Header("Transfer-Encoding", "chunked")
	c.Header("Content-Type", "application/json")
	c.Status(http.StatusOK)

	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		response := routes.CreateErrorResponse(500, "不支持分块传输")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	startTime := time.Now()
	endTime := startTime.Add(time.Duration(duration) * time.Second)
	chunkID := 0

	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	defer ticker.Stop()

	for time.Now().Before(endTime) {
		select {
		case <-ticker.C:
			chunkID++
			chunk := map[string]interface{}{
				"chunk_id":     chunkID,
				"timestamp":    time.Now().Unix(),
				"elapsed_ms":   time.Since(startTime).Milliseconds(),
				"remaining_ms": endTime.Sub(time.Now()).Milliseconds(),
				"data":         fmt.Sprintf("流式数据块 #%d", chunkID),
				"server_time":  time.Now().Format("2006-01-02 15:04:05.000"),
			}

			chunkData := fmt.Sprintf(`{"chunk_id":%d,"timestamp":%d,"elapsed_ms":%d,"remaining_ms":%d,"data":"流式数据块 #%d","server_time":"%s"}`,
				chunkID, chunk["timestamp"], chunk["elapsed_ms"], chunk["remaining_ms"], chunkID, chunk["server_time"])

			// 写入分块
			fmt.Fprintf(w, "%x\r\n", len(chunkData))
			fmt.Fprintf(w, "%s\r\n", chunkData)
			flusher.Flush()
		}
	}

	// 结束分块传输
	fmt.Fprintf(w, "0\r\n\r\n")
	flusher.Flush()
}

// 分块上传测试
func handleChunkedUpload(c *gin.Context) {
	// 检查是否为分块传输
	transferEncoding := c.GetHeader("Transfer-Encoding")
	isChunked := strings.Contains(strings.ToLower(transferEncoding), "chunked")

	if !isChunked {
		response := routes.CreateErrorResponse(400, "需要分块传输编码")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var totalSize int64
	var chunkCount int
	var chunks []map[string]interface{}

	reader := bufio.NewReader(c.Request.Body)
	startTime := time.Now()

	for {
		// 读取分块大小
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			response := routes.CreateErrorResponse(400, "读取分块失败: "+err.Error())
			c.JSON(http.StatusBadRequest, response)
			return
		}

		sizeStr := strings.TrimSpace(string(line))
		if sizeStr == "" {
			continue
		}

		size, err := strconv.ParseInt(sizeStr, 16, 64)
		if err != nil {
			response := routes.CreateErrorResponse(400, "解析分块大小失败: "+err.Error())
			c.JSON(http.StatusBadRequest, response)
			return
		}

		if size == 0 {
			break
		}

		// 读取分块数据
		chunkData := make([]byte, size)
		_, err = io.ReadFull(reader, chunkData)
		if err != nil {
			response := routes.CreateErrorResponse(400, "读取分块数据失败: "+err.Error())
			c.JSON(http.StatusBadRequest, response)
			return
		}

		// 读取分块结束标记
		reader.ReadLine()

		chunkCount++
		totalSize += size
		chunks = append(chunks, map[string]interface{}{
			"chunk_id":    chunkCount,
			"size":        size,
			"received_at": time.Now().Unix(),
		})
	}

	uploadStats := map[string]interface{}{
		"chunk_count":        chunkCount,
		"total_size":         totalSize,
		"upload_time_ms":     time.Since(startTime).Milliseconds(),
		"average_chunk_size": float64(totalSize) / float64(chunkCount),
		"upload_speed_bps":   float64(totalSize) / time.Since(startTime).Seconds(),
		"chunks":             chunks,
		"content_type":       c.GetHeader("Content-Type"),
		"completed_at":       time.Now().Unix(),
	}

	response := routes.CreateSuccessResponse("分块上传完成", uploadStats)
	c.JSON(http.StatusOK, response)
}

// Identity传输编码测试
func handleIdentityTransfer(c *gin.Context) {
	c.Header("Transfer-Encoding", "identity")

	data := map[string]interface{}{
		"message":   "这是identity传输编码测试",
		"encoding":  "identity",
		"timestamp": time.Now().Unix(),
	}

	response := routes.CreateSuccessResponse("Identity传输编码测试", data)
	c.JSON(http.StatusOK, response)
}

// Deflate传输编码测试
func handleDeflateTransfer(c *gin.Context) {
	c.Header("Transfer-Encoding", "deflate")

	data := map[string]interface{}{
		"message":   "这是deflate传输编码测试",
		"encoding":  "deflate",
		"data":      strings.Repeat("压缩测试数据 ", 50),
		"timestamp": time.Now().Unix(),
	}

	response := routes.CreateSuccessResponse("Deflate传输编码测试", data)
	c.JSON(http.StatusOK, response)
}

// Gzip传输编码测试
func handleGzipTransfer(c *gin.Context) {
	c.Header("Transfer-Encoding", "gzip")

	data := map[string]interface{}{
		"message":   "这是gzip传输编码测试",
		"encoding":  "gzip",
		"data":      strings.Repeat("压缩测试数据 ", 50),
		"timestamp": time.Now().Unix(),
	}

	response := routes.CreateSuccessResponse("Gzip传输编码测试", data)
	c.JSON(http.StatusOK, response)
}

// 大文件传输测试
func handleLargeTransfer(c *gin.Context) {
	sizeStr := c.Param("size")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 1000 {
		size = 10 // 默认10MB
	}

	useChunked := c.Query("chunked") == "true"

	if useChunked {
		c.Header("Transfer-Encoding", "chunked")
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Status(http.StatusOK)

	w := c.Writer
	flusher, ok := w.(http.Flusher)

	// 生成大文件数据
	chunkSize := 1024 * 1024 // 1MB chunks
	totalChunks := size

	for i := 0; i < totalChunks; i++ {
		chunk := make([]byte, chunkSize)
		for j := 0; j < chunkSize; j++ {
			chunk[j] = byte((i + j) % 256)
		}

		if useChunked && ok {
			// 分块传输
			fmt.Fprintf(w, "%x\r\n", len(chunk))
			w.Write(chunk)
			fmt.Fprintf(w, "\r\n")
			flusher.Flush()
		} else {
			// 普通传输
			w.Write(chunk)
			if ok {
				flusher.Flush()
			}
		}
	}

	if useChunked && ok {
		fmt.Fprintf(w, "0\r\n\r\n")
		flusher.Flush()
	}
}

// 大文件接收测试
func handleLargeReceive(c *gin.Context) {
	startTime := time.Now()

	// 检查传输编码
	transferEncoding := c.GetHeader("Transfer-Encoding")
	isChunked := strings.Contains(strings.ToLower(transferEncoding), "chunked")

	var totalSize int64
	var chunkCount int

	if isChunked {
		// 处理分块传输
		reader := bufio.NewReader(c.Request.Body)

		for {
			line, _, err := reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					break
				}
				response := routes.CreateErrorResponse(400, "读取分块失败: "+err.Error())
				c.JSON(http.StatusBadRequest, response)
				return
			}

			sizeStr := strings.TrimSpace(string(line))
			if sizeStr == "" {
				continue
			}

			size, err := strconv.ParseInt(sizeStr, 16, 64)
			if err != nil {
				response := routes.CreateErrorResponse(400, "解析分块大小失败: "+err.Error())
				c.JSON(http.StatusBadRequest, response)
				return
			}

			if size == 0 {
				break
			}

			// 读取分块数据（但不保存到内存）
			_, err = io.CopyN(io.Discard, reader, size)
			if err != nil {
				response := routes.CreateErrorResponse(400, "读取分块数据失败: "+err.Error())
				c.JSON(http.StatusBadRequest, response)
				return
			}

			// 读取分块结束标记
			reader.ReadLine()

			chunkCount++
			totalSize += size
		}
	} else {
		// 普通传输
		size, err := io.Copy(io.Discard, c.Request.Body)
		if err != nil {
			response := routes.CreateErrorResponse(400, "读取数据失败: "+err.Error())
			c.JSON(http.StatusBadRequest, response)
			return
		}
		totalSize = size
	}

	transferTime := time.Since(startTime)
	transferStats := map[string]interface{}{
		"is_chunked":          isChunked,
		"transfer_encoding":   transferEncoding,
		"chunk_count":         chunkCount,
		"total_size":          totalSize,
		"transfer_time_ms":    transferTime.Milliseconds(),
		"transfer_speed_bps":  float64(totalSize) / transferTime.Seconds(),
		"transfer_speed_mbps": (float64(totalSize) / (1024 * 1024)) / transferTime.Seconds(),
		"received_at":         time.Now().Unix(),
	}

	response := routes.CreateSuccessResponse("大文件接收完成", transferStats)
	c.JSON(http.StatusOK, response)
}

// SSE传输测试
func handleSSETransfer(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		response := routes.CreateErrorResponse(500, "不支持SSE")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// 发送SSE数据
	for i := 0; i < 10; i++ {
		data := map[string]interface{}{
			"id":        i + 1,
			"timestamp": time.Now().Unix(),
			"message":   fmt.Sprintf("SSE消息 #%d", i+1),
		}

		fmt.Fprintf(w, "data: %s\n\n", fmt.Sprintf(`{"id":%d,"timestamp":%d,"message":"SSE消息 #%d"}`,
			data["id"], data["timestamp"], i+1))
		flusher.Flush()

		time.Sleep(1 * time.Second)
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

// WebSocket传输测试
func handleWebSocketTransfer(c *gin.Context) {
	// 这里只是一个占位符，实际的WebSocket处理在websocket模块中
	response := routes.CreateSuccessResponse("WebSocket传输测试", map[string]interface{}{
		"message": "请使用 /ws/* 端点进行WebSocket测试",
		"websocket_endpoints": []string{
			"/ws/connect",
			"/ws/echo",
			"/ws/broadcast",
			"/ws/realtime",
		},
	})
	c.JSON(http.StatusOK, response)
}
