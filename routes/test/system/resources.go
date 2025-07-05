package system

import (
	"io"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"http_proxy_tool_test_web_demo/routes"

	"github.com/gin-gonic/gin"
)

// SystemModule 系统资源测试模块
type SystemModule struct{}

// RegisterRoutes 注册路由
func (m *SystemModule) RegisterRoutes(r *gin.Engine) {
	test := r.Group("/test")
	{
		// 内存压力测试
		test.GET("/memory", handleMemoryTest)

		// CPU压力测试
		test.GET("/cpu", handleCPUTest)

		// 网络测试
		test.GET("/network", handleNetworkTest)

		// 文件IO测试
		test.GET("/fileio", handleFileIOTest)

		// 数据库连接测试
		test.GET("/database", handleDatabaseTest)

		// 长连接测试
		test.GET("/keepalive", handleKeepAliveTest)

		// 系统信息
		test.GET("/system", handleSystemInfo)
	}
}

// GetPrefix 获取前缀
func (m *SystemModule) GetPrefix() string {
	return "/test"
}

// GetDescription 获取描述
func (m *SystemModule) GetDescription() string {
	return "系统资源测试接口"
}

// 内存压力测试
func handleMemoryTest(c *gin.Context) {
	sizeStr := c.Query("size")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 1000 {
		size = 10 // 默认10MB
	}

	durationStr := c.Query("duration")
	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration < 1 || duration > 300 {
		duration = 10 // 默认10秒
	}

	startTime := time.Now()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	beforeMem := memStats.Alloc

	// 分配内存
	chunks := make([][]byte, 0)
	chunkSize := 1024 * 1024 // 1MB chunks
	totalChunks := size

	for i := 0; i < totalChunks; i++ {
		chunk := make([]byte, chunkSize)
		// 写入随机数据防止编译器优化
		for j := 0; j < chunkSize; j += 1024 {
			chunk[j] = byte(rand.Intn(256)) // #nosec G404 - 用于内存测试填充，非安全敏感
		}
		chunks = append(chunks, chunk)
	}

	// 持续时间
	time.Sleep(time.Duration(duration) * time.Second)

	// 记录分配的块数
	allocatedChunks := len(chunks)
	_ = allocatedChunks // 避免未使用变量警告

	// 释放内存
	for i := range chunks {
		chunks[i] = nil
	}
	chunks = nil //nolint:ineffassign // 明确释放内存引用以帮助垃圾回收
	runtime.GC()

	runtime.ReadMemStats(&memStats)
	afterMem := memStats.Alloc

	memoryStats := map[string]interface{}{
		"requested_size_mb":   size,
		"duration_seconds":    duration,
		"memory_before_bytes": beforeMem,
		"memory_after_bytes":  afterMem,
		"memory_used_bytes":   memStats.TotalAlloc,
		"gc_cycles":           memStats.NumGC,
		"heap_objects":        memStats.HeapObjects,
		"heap_size_bytes":     memStats.HeapSys,
		"execution_time_ms":   time.Since(startTime).Milliseconds(),
	}

	response := routes.CreateSuccessResponse("内存测试完成", memoryStats)
	c.JSON(http.StatusOK, response)
}

// CPU压力测试
func handleCPUTest(c *gin.Context) {
	workersStr := c.Query("workers")
	workers, err := strconv.Atoi(workersStr)
	if err != nil || workers < 1 || workers > 100 {
		workers = runtime.NumCPU() // 默认CPU核心数
	}

	durationStr := c.Query("duration")
	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration < 1 || duration > 300 {
		duration = 10 // 默认10秒
	}

	startTime := time.Now()
	var wg sync.WaitGroup
	var totalOperations int64
	var mu sync.Mutex

	stopChan := make(chan struct{})

	// 启动工作协程
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			operations := int64(0)

			for {
				select {
				case <-stopChan:
					mu.Lock()
					totalOperations += operations
					mu.Unlock()
					return
				default:
					// 执行一些CPU密集型计算
					for j := 0; j < 1000; j++ {
						_ = j * j * j
					}
					operations++
				}
			}
		}(i)
	}

	// 等待指定时间
	time.Sleep(time.Duration(duration) * time.Second)
	close(stopChan)
	wg.Wait()

	cpuStats := map[string]interface{}{
		"workers":               workers,
		"duration_seconds":      duration,
		"total_operations":      totalOperations,
		"operations_per_second": float64(totalOperations) / float64(duration),
		"operations_per_worker": float64(totalOperations) / float64(workers),
		"execution_time_ms":     time.Since(startTime).Milliseconds(),
		"cpu_cores":             runtime.NumCPU(),
	}

	response := routes.CreateSuccessResponse("CPU测试完成", cpuStats)
	c.JSON(http.StatusOK, response)
}

// 网络测试
func handleNetworkTest(c *gin.Context) {
	sizeStr := c.Query("size")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 {
		size = 1 // 默认1MB
	}

	startTime := time.Now()

	// 生成测试数据
	data := make([]byte, size*1024*1024) // MB
	for i := 0; i < len(data); i += 1024 {
		data[i] = byte(rand.Intn(256)) // #nosec G404 - 用于网络测试数据填充，非安全敏感
	}

	// 模拟网络传输延迟
	networkDelay := rand.Intn(100) + 10 // #nosec G404 - 用于模拟网络延迟，非安全敏感
	time.Sleep(time.Duration(networkDelay) * time.Millisecond)

	networkStats := map[string]interface{}{
		"data_size_mb":      size,
		"data_size_bytes":   len(data),
		"network_delay_ms":  networkDelay,
		"execution_time_ms": time.Since(startTime).Milliseconds(),
		"throughput_mbps":   float64(size) / (float64(networkDelay) / 1000.0),
		"client_ip":         c.ClientIP(),
		"user_agent":        c.GetHeader("User-Agent"),
	}

	response := routes.CreateSuccessResponse("网络测试完成", networkStats)
	c.JSON(http.StatusOK, response)
}

// 文件IO测试
func handleFileIOTest(c *gin.Context) {
	sizeStr := c.Query("size")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 {
		size = 1 // 默认1MB
	}

	startTime := time.Now()

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "fileio_test_*.tmp")
	if err != nil {
		response := routes.CreateErrorResponse(500, "创建临时文件失败: "+err.Error())
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// 写入测试数据
	writeStartTime := time.Now()
	data := make([]byte, size*1024*1024) // MB
	for i := 0; i < len(data); i += 1024 {
		data[i] = byte(rand.Intn(256)) // #nosec G404 - 用于文件IO测试数据填充，非安全敏感
	}

	_, err = tempFile.Write(data)
	if err != nil {
		response := routes.CreateErrorResponse(500, "写入文件失败: "+err.Error())
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	writeTime := time.Since(writeStartTime)

	// 读取测试数据
	readStartTime := time.Now()
	_, _ = tempFile.Seek(0, 0)
	readData := make([]byte, len(data))
	_, err = io.ReadFull(tempFile, readData)
	if err != nil {
		response := routes.CreateErrorResponse(500, "读取文件失败: "+err.Error())
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	readTime := time.Since(readStartTime)

	fileIOStats := map[string]interface{}{
		"file_size_mb":     size,
		"file_size_bytes":  len(data),
		"write_time_ms":    writeTime.Milliseconds(),
		"read_time_ms":     readTime.Milliseconds(),
		"total_time_ms":    time.Since(startTime).Milliseconds(),
		"write_speed_mbps": float64(size) / writeTime.Seconds(),
		"read_speed_mbps":  float64(size) / readTime.Seconds(),
		"temp_file_path":   tempFile.Name(),
	}

	response := routes.CreateSuccessResponse("文件IO测试完成", fileIOStats)
	c.JSON(http.StatusOK, response)
}

// 数据库连接测试
func handleDatabaseTest(c *gin.Context) {
	connectionsStr := c.Query("connections")
	connections, err := strconv.Atoi(connectionsStr)
	if err != nil || connections < 1 || connections > 100 {
		connections = 10 // 默认10个连接
	}

	startTime := time.Now()

	// 模拟数据库连接和查询
	var wg sync.WaitGroup
	var totalQueries int64
	var mu sync.Mutex

	for i := 0; i < connections; i++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()
			queries := int64(0)

			// 模拟连接建立时间
			connectionTime := rand.Intn(100) + 10 // #nosec G404 - 用于模拟连接延迟，非安全敏感
			time.Sleep(time.Duration(connectionTime) * time.Millisecond)

			// 执行一些查询
			for j := 0; j < 10; j++ {
				queryTime := rand.Intn(50) + 5 // #nosec G404 - 用于模拟查询延迟，非安全敏感
				time.Sleep(time.Duration(queryTime) * time.Millisecond)
				queries++
			}

			mu.Lock()
			totalQueries += queries
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	databaseStats := map[string]interface{}{
		"connections":        connections,
		"total_queries":      totalQueries,
		"queries_per_conn":   float64(totalQueries) / float64(connections),
		"execution_time_ms":  time.Since(startTime).Milliseconds(),
		"queries_per_second": float64(totalQueries) / (float64(time.Since(startTime).Milliseconds()) / 1000.0),
		"connection_status":  "simulated",
		"database_type":      "mock",
	}

	response := routes.CreateSuccessResponse("数据库连接测试完成", databaseStats)
	c.JSON(http.StatusOK, response)
}

// 长连接测试
func handleKeepAliveTest(c *gin.Context) {
	durationStr := c.Query("duration")
	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration < 1 || duration > 300 {
		duration = 30 // 默认30秒
	}

	startTime := time.Now()

	// 模拟长连接保持
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	heartbeats := 0
	endTime := time.Now().Add(time.Duration(duration) * time.Second)

	for time.Now().Before(endTime) {
		<-ticker.C
		heartbeats++
		// 模拟心跳包处理
		time.Sleep(10 * time.Millisecond)
	}

	keepAliveStats := map[string]interface{}{
		"duration_seconds":   duration,
		"heartbeats_sent":    heartbeats,
		"heartbeat_interval": 1,
		"connection_status":  "active",
		"execution_time_ms":  time.Since(startTime).Milliseconds(),
		"client_ip":          c.ClientIP(),
		"connection_id":      time.Now().UnixNano(),
	}

	response := routes.CreateSuccessResponse("长连接测试完成", keepAliveStats)
	c.JSON(http.StatusOK, response)
}

// 系统信息
func handleSystemInfo(c *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	systemInfo := map[string]interface{}{
		"go_version":      runtime.Version(),
		"go_os":           runtime.GOOS,
		"go_arch":         runtime.GOARCH,
		"cpu_count":       runtime.NumCPU(),
		"goroutine_count": runtime.NumGoroutine(),
		"memory": map[string]interface{}{
			"alloc_mb":     float64(memStats.Alloc) / 1024 / 1024,
			"total_alloc":  memStats.TotalAlloc,
			"sys_mb":       float64(memStats.Sys) / 1024 / 1024,
			"heap_alloc":   memStats.HeapAlloc,
			"heap_sys":     memStats.HeapSys,
			"heap_objects": memStats.HeapObjects,
		},
		"gc": map[string]interface{}{
			"num_gc":         memStats.NumGC,
			"pause_total_ns": memStats.PauseTotalNs,
			"last_gc":        memStats.LastGC,
		},
		"timestamp": time.Now().Unix(),
	}

	response := routes.CreateSuccessResponse("系统信息获取成功", systemInfo)
	c.JSON(http.StatusOK, response)
}
