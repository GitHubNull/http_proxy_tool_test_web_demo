package main

import (
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// 并发测试统计
type ConcurrentStats struct {
	TotalRequests     int64   `json:"total_requests"`
	SuccessRequests   int64   `json:"success_requests"`
	FailedRequests    int64   `json:"failed_requests"`
	AverageResponse   int64   `json:"average_response_ms"`
	MaxResponse       int64   `json:"max_response_ms"`
	MinResponse       int64   `json:"min_response_ms"`
	RequestsPerSecond float64 `json:"requests_per_second"`
	StartTime         int64   `json:"start_time"`
	EndTime           int64   `json:"end_time"`
	Duration          int64   `json:"duration_ms"`
}

// 全局并发测试状态
var concurrentStats = &ConcurrentStats{}
var statsLock sync.RWMutex

// 初始化测试路由
func initTestRoutes(r *gin.Engine) {
	test := r.Group("/test")
	{
		// 并发测试
		test.GET("/concurrent", handleConcurrentTest)
		test.POST("/concurrent", handleConcurrentTest)

		// 压力测试
		test.GET("/stress", handleStressTest)
		test.POST("/stress", handleStressTest)

		// 内存压力测试
		test.GET("/memory", handleMemoryTest)

		// CPU压力测试
		test.GET("/cpu", handleCPUTest)

		// 批量请求测试
		test.POST("/batch", handleBatchTest)

		// 长连接测试
		test.GET("/keepalive", handleKeepAliveTest)

		// 并发统计
		test.GET("/stats", handleStatsTest)

		// 重置统计
		test.POST("/reset", handleResetStats)

		// 系统信息
		test.GET("/system", handleSystemInfo)

		// 负载测试
		test.GET("/load", handleLoadTest)

		// 随机延迟测试
		test.GET("/random-delay", handleRandomDelayTest)

		// 网络测试
		test.GET("/network", handleNetworkTest)

		// 文件IO测试
		test.GET("/fileio", handleFileIOTest)

		// 数据库连接测试
		test.GET("/database", handleDatabaseTest)
	}
}

// 并发测试处理
func handleConcurrentTest(c *gin.Context) {
	startTime := time.Now()

	// 获取并发参数
	concurrencyStr := c.Query("concurrency")
	concurrency, err := strconv.Atoi(concurrencyStr)
	if err != nil || concurrency < 1 || concurrency > 1000 {
		concurrency = 10
	}

	requestsStr := c.Query("requests")
	requests, err := strconv.Atoi(requestsStr)
	if err != nil || requests < 1 || requests > 10000 {
		requests = 100
	}

	delayStr := c.Query("delay")
	delay, err := strconv.Atoi(delayStr)
	if err != nil || delay < 0 || delay > 5000 {
		delay = 0
	}

	// 更新统计
	statsLock.Lock()
	concurrentStats.TotalRequests = int64(requests)
	concurrentStats.SuccessRequests = 0
	concurrentStats.FailedRequests = 0
	concurrentStats.StartTime = startTime.Unix()
	concurrentStats.MinResponse = 999999
	concurrentStats.MaxResponse = 0
	statsLock.Unlock()

	// 执行并发测试
	var wg sync.WaitGroup
	var totalResponseTime int64

	semaphore := make(chan struct{}, concurrency)

	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(requestID int) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			requestStart := time.Now()

			// 模拟延迟
			if delay > 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}

			// 模拟一些处理
			processTime := rand.Intn(50) + 10
			time.Sleep(time.Duration(processTime) * time.Millisecond)

			responseTime := time.Since(requestStart).Milliseconds()
			atomic.AddInt64(&totalResponseTime, responseTime)

			// 更新统计
			statsLock.Lock()
			concurrentStats.SuccessRequests++
			if responseTime > concurrentStats.MaxResponse {
				concurrentStats.MaxResponse = responseTime
			}
			if responseTime < concurrentStats.MinResponse {
				concurrentStats.MinResponse = responseTime
			}
			statsLock.Unlock()
		}(i)
	}

	wg.Wait()

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// 计算统计
	statsLock.Lock()
	concurrentStats.EndTime = endTime.Unix()
	concurrentStats.Duration = duration.Milliseconds()
	concurrentStats.AverageResponse = totalResponseTime / int64(requests)
	concurrentStats.RequestsPerSecond = float64(requests) / duration.Seconds()
	statsLock.Unlock()

	response := ApiResponse{
		Code:      200,
		Message:   "并发测试完成",
		Data:      concurrentStats,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 压力测试处理
func handleStressTest(c *gin.Context) {
	durationStr := c.Query("duration")
	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration < 1 || duration > 300 {
		duration = 60 // 默认60秒
	}

	concurrencyStr := c.Query("concurrency")
	concurrency, err := strconv.Atoi(concurrencyStr)
	if err != nil || concurrency < 1 || concurrency > 500 {
		concurrency = 20
	}

	startTime := time.Now()
	endTime := startTime.Add(time.Duration(duration) * time.Second)

	var totalRequests int64
	var successRequests int64
	var totalResponseTime int64
	var maxResponseTime int64
	var minResponseTime int64 = 999999

	// 启动压力测试
	var wg sync.WaitGroup
	stopChan := make(chan bool)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case <-stopChan:
					return
				default:
					requestStart := time.Now()

					// 模拟处理
					processTime := rand.Intn(100) + 50
					time.Sleep(time.Duration(processTime) * time.Millisecond)

					responseTime := time.Since(requestStart).Milliseconds()

					atomic.AddInt64(&totalRequests, 1)
					atomic.AddInt64(&successRequests, 1)
					atomic.AddInt64(&totalResponseTime, responseTime)

					if responseTime > maxResponseTime {
						atomic.SwapInt64(&maxResponseTime, responseTime)
					}
					if responseTime < minResponseTime {
						atomic.SwapInt64(&minResponseTime, responseTime)
					}

					// 检查是否超时
					if time.Now().After(endTime) {
						return
					}
				}
			}
		}(i)
	}

	// 等待测试完成
	time.Sleep(time.Duration(duration) * time.Second)
	close(stopChan)
	wg.Wait()

	actualDuration := time.Since(startTime)

	stats := ConcurrentStats{
		TotalRequests:     totalRequests,
		SuccessRequests:   successRequests,
		FailedRequests:    totalRequests - successRequests,
		AverageResponse:   totalResponseTime / totalRequests,
		MaxResponse:       maxResponseTime,
		MinResponse:       minResponseTime,
		RequestsPerSecond: float64(totalRequests) / actualDuration.Seconds(),
		StartTime:         startTime.Unix(),
		EndTime:           time.Now().Unix(),
		Duration:          actualDuration.Milliseconds(),
	}

	response := ApiResponse{
		Code:      200,
		Message:   "压力测试完成",
		Data:      stats,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 内存压力测试
func handleMemoryTest(c *gin.Context) {
	sizeStr := c.Query("size")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 1024 {
		size = 100 // 默认100MB
	}

	durationStr := c.Query("duration")
	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration < 1 || duration > 60 {
		duration = 10 // 默认10秒
	}

	startTime := time.Now()

	// 分配内存
	allocSize := size * 1024 * 1024 // 转换为字节
	data := make([]byte, allocSize)

	// 填充数据
	for i := range data {
		data[i] = byte(i % 256)
	}

	// 保持指定时间
	time.Sleep(time.Duration(duration) * time.Second)

	// 获取内存统计
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	endTime := time.Now()

	result := map[string]interface{}{
		"allocated_size_mb": size,
		"duration_seconds":  duration,
		"heap_alloc_mb":     mem.HeapAlloc / 1024 / 1024,
		"heap_sys_mb":       mem.HeapSys / 1024 / 1024,
		"heap_idle_mb":      mem.HeapIdle / 1024 / 1024,
		"heap_inuse_mb":     mem.HeapInuse / 1024 / 1024,
		"heap_objects":      mem.HeapObjects,
		"gc_runs":           mem.NumGC,
		"gc_pause_total_ns": mem.PauseTotalNs,
		"start_time":        startTime.Unix(),
		"end_time":          endTime.Unix(),
		"test_duration_ms":  endTime.Sub(startTime).Milliseconds(),
	}

	// 清理内存
	data = nil
	runtime.GC()

	response := ApiResponse{
		Code:      200,
		Message:   "内存压力测试完成",
		Data:      result,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// CPU压力测试
func handleCPUTest(c *gin.Context) {
	durationStr := c.Query("duration")
	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration < 1 || duration > 60 {
		duration = 10 // 默认10秒
	}

	coresStr := c.Query("cores")
	cores, err := strconv.Atoi(coresStr)
	if err != nil || cores < 1 || cores > runtime.NumCPU() {
		cores = runtime.NumCPU()
	}

	startTime := time.Now()
	endTime := startTime.Add(time.Duration(duration) * time.Second)

	var wg sync.WaitGroup
	var totalOperations int64

	for i := 0; i < cores; i++ {
		wg.Add(1)
		go func(coreID int) {
			defer wg.Done()

			operations := 0
			for time.Now().Before(endTime) {
				// 执行CPU密集型计算
				for j := 0; j < 10000; j++ {
					_ = j * j * j
				}
				operations++
			}

			atomic.AddInt64(&totalOperations, int64(operations))
		}(i)
	}

	wg.Wait()

	actualDuration := time.Since(startTime)

	result := map[string]interface{}{
		"duration_seconds":      duration,
		"cores_used":            cores,
		"total_operations":      totalOperations,
		"operations_per_second": float64(totalOperations) / actualDuration.Seconds(),
		"start_time":            startTime.Unix(),
		"end_time":              time.Now().Unix(),
		"actual_duration_ms":    actualDuration.Milliseconds(),
		"cpu_count":             runtime.NumCPU(),
	}

	response := ApiResponse{
		Code:      200,
		Message:   "CPU压力测试完成",
		Data:      result,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 批量请求测试
func handleBatchTest(c *gin.Context) {
	var requests []map[string]interface{}

	if err := c.ShouldBindJSON(&requests); err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:      400,
			Message:   "批量请求数据格式错误",
			Data:      nil,
			Timestamp: time.Now().Unix(),
			RequestID: generateRequestID(),
		})
		return
	}

	startTime := time.Now()
	responses := make([]map[string]interface{}, len(requests))

	var wg sync.WaitGroup

	for i, req := range requests {
		wg.Add(1)
		go func(index int, request map[string]interface{}) {
			defer wg.Done()

			requestStart := time.Now()

			// 模拟处理
			if delay, ok := request["delay"].(float64); ok {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}

			responseTime := time.Since(requestStart).Milliseconds()

			responses[index] = map[string]interface{}{
				"request_id":    request["id"],
				"status":        "success",
				"response_time": responseTime,
				"data":          request,
			}
		}(i, req)
	}

	wg.Wait()

	totalDuration := time.Since(startTime)

	result := map[string]interface{}{
		"total_requests":    len(requests),
		"responses":         responses,
		"total_duration_ms": totalDuration.Milliseconds(),
		"start_time":        startTime.Unix(),
		"end_time":          time.Now().Unix(),
	}

	response := ApiResponse{
		Code:      200,
		Message:   "批量请求测试完成",
		Data:      result,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 并发统计
func handleStatsTest(c *gin.Context) {
	statsLock.RLock()
	stats := *concurrentStats
	statsLock.RUnlock()

	response := ApiResponse{
		Code:      200,
		Message:   "并发统计数据",
		Data:      stats,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 重置统计
func handleResetStats(c *gin.Context) {
	statsLock.Lock()
	concurrentStats = &ConcurrentStats{}
	statsLock.Unlock()

	response := ApiResponse{
		Code:      200,
		Message:   "统计数据已重置",
		Data:      nil,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 系统信息
func handleSystemInfo(c *gin.Context) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	systemInfo := map[string]interface{}{
		"go_version":      runtime.Version(),
		"cpu_count":       runtime.NumCPU(),
		"goroutine_count": runtime.NumGoroutine(),
		"memory": map[string]interface{}{
			"alloc_mb":       mem.Alloc / 1024 / 1024,
			"total_alloc_mb": mem.TotalAlloc / 1024 / 1024,
			"sys_mb":         mem.Sys / 1024 / 1024,
			"heap_alloc_mb":  mem.HeapAlloc / 1024 / 1024,
			"heap_sys_mb":    mem.HeapSys / 1024 / 1024,
			"heap_objects":   mem.HeapObjects,
		},
		"gc": map[string]interface{}{
			"num_gc":         mem.NumGC,
			"pause_total_ns": mem.PauseTotalNs,
			"pause_avg_ns":   mem.PauseTotalNs / uint64(mem.NumGC+1),
		},
		"timestamp": time.Now().Unix(),
	}

	response := ApiResponse{
		Code:      200,
		Message:   "系统信息",
		Data:      systemInfo,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 负载测试
func handleLoadTest(c *gin.Context) {
	loadStr := c.Query("load")
	load, err := strconv.Atoi(loadStr)
	if err != nil || load < 1 || load > 100 {
		load = 50 // 默认50%负载
	}

	durationStr := c.Query("duration")
	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration < 1 || duration > 60 {
		duration = 10 // 默认10秒
	}

	startTime := time.Now()

	// 根据负载百分比计算工作量
	workDuration := time.Duration(duration*load/100) * time.Second
	sleepDuration := time.Duration(duration*(100-load)/100) * time.Second

	// 执行工作
	workStart := time.Now()
	for time.Since(workStart) < workDuration {
		// CPU密集型计算
		for i := 0; i < 10000; i++ {
			_ = i * i
		}
	}

	// 休眠
	time.Sleep(sleepDuration)

	endTime := time.Now()

	result := map[string]interface{}{
		"load_percentage":    load,
		"duration_seconds":   duration,
		"work_duration_ms":   workDuration.Milliseconds(),
		"sleep_duration_ms":  sleepDuration.Milliseconds(),
		"actual_duration_ms": endTime.Sub(startTime).Milliseconds(),
		"start_time":         startTime.Unix(),
		"end_time":           endTime.Unix(),
	}

	response := ApiResponse{
		Code:      200,
		Message:   "负载测试完成",
		Data:      result,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 随机延迟测试
func handleRandomDelayTest(c *gin.Context) {
	minStr := c.Query("min")
	min, err := strconv.Atoi(minStr)
	if err != nil || min < 0 || min > 5000 {
		min = 100 // 默认100ms
	}

	maxStr := c.Query("max")
	max, err := strconv.Atoi(maxStr)
	if err != nil || max < min || max > 10000 {
		max = 1000 // 默认1000ms
	}

	startTime := time.Now()

	// 生成随机延迟
	delay := rand.Intn(max-min) + min
	time.Sleep(time.Duration(delay) * time.Millisecond)

	endTime := time.Now()

	result := map[string]interface{}{
		"min_delay_ms":      min,
		"max_delay_ms":      max,
		"actual_delay_ms":   delay,
		"total_duration_ms": endTime.Sub(startTime).Milliseconds(),
		"start_time":        startTime.Unix(),
		"end_time":          endTime.Unix(),
	}

	response := ApiResponse{
		Code:      200,
		Message:   "随机延迟测试完成",
		Data:      result,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 网络测试
func handleNetworkTest(c *gin.Context) {
	sizeStr := c.Query("size")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 10240 {
		size = 1024 // 默认1KB
	}

	// 生成指定大小的数据
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	result := map[string]interface{}{
		"data_size_bytes": size,
		"data_size_kb":    size / 1024,
		"data":            data,
		"client_ip":       c.ClientIP(),
		"user_agent":      c.GetHeader("User-Agent"),
		"timestamp":       time.Now().Unix(),
	}

	response := ApiResponse{
		Code:      200,
		Message:   "网络测试数据",
		Data:      result,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 文件IO测试
func handleFileIOTest(c *gin.Context) {
	// 这里只是模拟文件IO操作，实际不会创建文件
	operationsStr := c.Query("operations")
	operations, err := strconv.Atoi(operationsStr)
	if err != nil || operations < 1 || operations > 1000 {
		operations = 100
	}

	startTime := time.Now()

	// 模拟文件操作
	for i := 0; i < operations; i++ {
		// 模拟文件读写延迟
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	}

	endTime := time.Now()

	result := map[string]interface{}{
		"operations":         operations,
		"total_duration_ms":  endTime.Sub(startTime).Milliseconds(),
		"avg_operation_ms":   endTime.Sub(startTime).Milliseconds() / int64(operations),
		"operations_per_sec": float64(operations) / endTime.Sub(startTime).Seconds(),
		"start_time":         startTime.Unix(),
		"end_time":           endTime.Unix(),
	}

	response := ApiResponse{
		Code:      200,
		Message:   "文件IO测试完成",
		Data:      result,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 数据库连接测试
func handleDatabaseTest(c *gin.Context) {
	connectionsStr := c.Query("connections")
	connections, err := strconv.Atoi(connectionsStr)
	if err != nil || connections < 1 || connections > 100 {
		connections = 10
	}

	durationStr := c.Query("duration")
	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration < 1 || duration > 60 {
		duration = 5
	}

	startTime := time.Now()

	var wg sync.WaitGroup
	results := make([]map[string]interface{}, connections)

	for i := 0; i < connections; i++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()

			connStart := time.Now()

			// 模拟数据库连接和查询
			time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)

			// 模拟查询操作
			queries := 0
			for time.Since(connStart) < time.Duration(duration)*time.Second {
				// 模拟查询延迟
				time.Sleep(time.Duration(rand.Intn(20)+10) * time.Millisecond)
				queries++
			}

			connEnd := time.Now()

			results[connID] = map[string]interface{}{
				"connection_id":   connID,
				"queries":         queries,
				"duration_ms":     connEnd.Sub(connStart).Milliseconds(),
				"queries_per_sec": float64(queries) / connEnd.Sub(connStart).Seconds(),
			}
		}(i)
	}

	wg.Wait()

	endTime := time.Now()

	totalQueries := 0
	for _, result := range results {
		totalQueries += result["queries"].(int)
	}

	result := map[string]interface{}{
		"connections":          connections,
		"duration_seconds":     duration,
		"total_queries":        totalQueries,
		"avg_queries_per_conn": float64(totalQueries) / float64(connections),
		"total_duration_ms":    endTime.Sub(startTime).Milliseconds(),
		"connection_results":   results,
		"start_time":           startTime.Unix(),
		"end_time":             endTime.Unix(),
	}

	response := ApiResponse{
		Code:      200,
		Message:   "数据库连接测试完成",
		Data:      result,
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}

// 长连接测试
func handleKeepAliveTest(c *gin.Context) {
	durationStr := c.Query("duration")
	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration < 1 || duration > 300 {
		duration = 30 // 默认30秒
	}

	intervalStr := c.Query("interval")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil || interval < 1 || interval > 10000 {
		interval = 1000 // 默认1秒
	}

	startTime := time.Now()
	endTime := startTime.Add(time.Duration(duration) * time.Second)

	c.Header("Connection", "keep-alive")
	c.Header("Keep-Alive", "timeout=300, max=1000")

	count := 0
	for time.Now().Before(endTime) {
		count++

		data := map[string]interface{}{
			"sequence":   count,
			"timestamp":  time.Now().Unix(),
			"remaining":  int(endTime.Sub(time.Now()).Seconds()),
			"client_ip":  c.ClientIP(),
			"user_agent": c.GetHeader("User-Agent"),
		}

		if count == 1 {
			response := ApiResponse{
				Code:      200,
				Message:   "长连接测试数据",
				Data:      data,
				Timestamp: time.Now().Unix(),
				RequestID: generateRequestID(),
			}

			c.JSON(http.StatusOK, response)
		}

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}
