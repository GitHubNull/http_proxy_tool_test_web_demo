package performance

import (
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"http_proxy_tool_test_web_demo/routes"

	"github.com/gin-gonic/gin"
)

// PerformanceModule 性能测试模块
type PerformanceModule struct{}

// ConcurrentStats 并发测试统计
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

// RegisterRoutes 注册路由
func (m *PerformanceModule) RegisterRoutes(r *gin.Engine) {
	test := r.Group("/test")
	{
		// 并发测试
		test.GET("/concurrent", handleConcurrentTest)
		test.POST("/concurrent", handleConcurrentTest)

		// 压力测试
		test.GET("/stress", handleStressTest)
		test.POST("/stress", handleStressTest)

		// 批量请求测试
		test.POST("/batch", handleBatchTest)

		// 负载测试
		test.GET("/load", handleLoadTest)

		// 随机延迟测试
		test.GET("/random-delay", handleRandomDelayTest)

		// 并发统计
		test.GET("/stats", handleStatsTest)

		// 重置统计
		test.POST("/reset", handleResetStats)
	}
}

// GetPrefix 获取前缀
func (m *PerformanceModule) GetPrefix() string {
	return "/test"
}

// GetDescription 获取描述
func (m *PerformanceModule) GetDescription() string {
	return "性能和并发测试接口"
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
			processTime := rand.Intn(50) + 10 // #nosec G404 - 用于模拟测试延迟，非安全敏感
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

	response := routes.CreateSuccessResponse("并发测试完成", concurrentStats)
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

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)
	stop := make(chan bool)

	// 启动压力测试协程
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-stop:
					return
				default:
					semaphore <- struct{}{}

					requestStart := time.Now()

					// 模拟处理
					processTime := rand.Intn(100) + 10 // #nosec G404 - 用于模拟测试延迟，非安全敏感
					time.Sleep(time.Duration(processTime) * time.Millisecond)

					responseTime := time.Since(requestStart).Milliseconds()

					atomic.AddInt64(&totalRequests, 1)
					atomic.AddInt64(&successRequests, 1)
					atomic.AddInt64(&totalResponseTime, responseTime)

					if responseTime > atomic.LoadInt64(&maxResponseTime) {
						atomic.StoreInt64(&maxResponseTime, responseTime)
					}
					if responseTime < atomic.LoadInt64(&minResponseTime) {
						atomic.StoreInt64(&minResponseTime, responseTime)
					}

					<-semaphore
				}
			}
		}()
	}

	// 等待测试时间结束
	time.Sleep(time.Until(endTime))
	close(stop)
	wg.Wait()

	actualDuration := time.Since(startTime)
	totalReq := atomic.LoadInt64(&totalRequests)
	avgResponseTime := int64(0)
	if totalReq > 0 {
		avgResponseTime = atomic.LoadInt64(&totalResponseTime) / totalReq
	}

	stressStats := map[string]interface{}{
		"total_requests":      totalReq,
		"success_requests":    atomic.LoadInt64(&successRequests),
		"failed_requests":     0,
		"average_response_ms": avgResponseTime,
		"max_response_ms":     atomic.LoadInt64(&maxResponseTime),
		"min_response_ms":     atomic.LoadInt64(&minResponseTime),
		"requests_per_second": float64(totalReq) / actualDuration.Seconds(),
		"duration_seconds":    actualDuration.Seconds(),
		"concurrency":         concurrency,
	}

	response := routes.CreateSuccessResponse("压力测试完成", stressStats)
	c.JSON(http.StatusOK, response)
}

// 批量请求测试
func handleBatchTest(c *gin.Context) {
	type BatchRequest struct {
		Method  string            `json:"method"`
		URL     string            `json:"url"`
		Headers map[string]string `json:"headers"`
		Body    interface{}       `json:"body"`
	}

	var requests []BatchRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		response := routes.CreateErrorResponse(400, "请求格式错误: "+err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if len(requests) > 100 {
		response := routes.CreateErrorResponse(400, "批量请求数量不能超过100个")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	results := make([]map[string]interface{}, len(requests))
	var wg sync.WaitGroup

	for i, req := range requests {
		wg.Add(1)
		go func(index int, request BatchRequest) {
			defer wg.Done()

			startTime := time.Now()

			// 模拟请求处理
			processTime := rand.Intn(200) + 50 // #nosec G404 - 用于模拟测试延迟，非安全敏感
			time.Sleep(time.Duration(processTime) * time.Millisecond)

			responseTime := time.Since(startTime).Milliseconds()

			results[index] = map[string]interface{}{
				"index":         index,
				"method":        request.Method,
				"url":           request.URL,
				"response_time": responseTime,
				"status":        "success",
				"response_code": 200,
			}
		}(i, req)
	}

	wg.Wait()

	batchStats := map[string]interface{}{
		"total_requests": len(requests),
		"results":        results,
		"completed_at":   time.Now().Unix(),
	}

	response := routes.CreateSuccessResponse("批量请求完成", batchStats)
	c.JSON(http.StatusOK, response)
}

// 负载测试
func handleLoadTest(c *gin.Context) {
	qpsStr := c.Query("qps")
	qps, err := strconv.Atoi(qpsStr)
	if err != nil || qps < 1 || qps > 1000 {
		qps = 50
	}

	durationStr := c.Query("duration")
	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration < 1 || duration > 300 {
		duration = 30
	}

	interval := time.Second / time.Duration(qps)
	endTime := time.Now().Add(time.Duration(duration) * time.Second)

	var totalRequests int64
	var successRequests int64

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for time.Now().Before(endTime) {
		<-ticker.C

		atomic.AddInt64(&totalRequests, 1)

		// 模拟请求处理
		go func() {
			processTime := rand.Intn(50) + 10 // #nosec G404 - 用于模拟测试延迟，非安全敏感
			time.Sleep(time.Duration(processTime) * time.Millisecond)
			atomic.AddInt64(&successRequests, 1)
		}()
	}

	// 等待最后的请求完成
	time.Sleep(100 * time.Millisecond)

	loadStats := map[string]interface{}{
		"target_qps":       qps,
		"actual_qps":       float64(atomic.LoadInt64(&totalRequests)) / float64(duration),
		"total_requests":   atomic.LoadInt64(&totalRequests),
		"success_requests": atomic.LoadInt64(&successRequests),
		"duration_seconds": duration,
	}

	response := routes.CreateSuccessResponse("负载测试完成", loadStats)
	c.JSON(http.StatusOK, response)
}

// 随机延迟测试
func handleRandomDelayTest(c *gin.Context) {
	minDelayStr := c.Query("min_delay")
	minDelay, err := strconv.Atoi(minDelayStr)
	if err != nil || minDelay < 0 || minDelay > 5000 {
		minDelay = 10
	}

	maxDelayStr := c.Query("max_delay")
	maxDelay, err := strconv.Atoi(maxDelayStr)
	if err != nil || maxDelay < minDelay || maxDelay > 10000 {
		maxDelay = 1000
	}

	// 生成随机延迟
	delay := rand.Intn(maxDelay-minDelay+1) + minDelay // #nosec G404 - 用于模拟测试延迟，非安全敏感
	time.Sleep(time.Duration(delay) * time.Millisecond)

	delayStats := map[string]interface{}{
		"min_delay_ms":    minDelay,
		"max_delay_ms":    maxDelay,
		"actual_delay_ms": delay,
		"timestamp":       time.Now().Unix(),
	}

	response := routes.CreateSuccessResponse("随机延迟测试完成", delayStats)
	c.JSON(http.StatusOK, response)
}

// 并发统计
func handleStatsTest(c *gin.Context) {
	statsLock.RLock()
	defer statsLock.RUnlock()

	response := routes.CreateSuccessResponse("统计信息获取成功", concurrentStats)
	c.JSON(http.StatusOK, response)
}

// 重置统计
func handleResetStats(c *gin.Context) {
	statsLock.Lock()
	concurrentStats = &ConcurrentStats{}
	statsLock.Unlock()

	response := routes.CreateSuccessResponse("统计信息已重置", nil)
	c.JSON(http.StatusOK, response)
}
