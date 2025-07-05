package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"http_proxy_tool_test_web_demo/routes"
	"http_proxy_tool_test_web_demo/routes/api"
	"http_proxy_tool_test_web_demo/routes/format"
	"http_proxy_tool_test_web_demo/routes/test/performance"
	"http_proxy_tool_test_web_demo/routes/test/system"
	"http_proxy_tool_test_web_demo/routes/transfer"
)

// setupTestRouter 创建测试用的Gin路由器
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
	}))

	// 版本信息API
	r.GET("/api/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":    version,
			"build_time": buildTime,
			"timestamp":  time.Now().Unix(),
		})
	})

	// 创建路由管理器并注册所有模块
	routeManager := routes.NewRouteManager()
	routeManager.RegisterModule(&api.BasicAPIModule{})
	routeManager.RegisterModule(&format.FormatModule{})
	routeManager.RegisterModule(&performance.PerformanceModule{})
	routeManager.RegisterModule(&system.SystemModule{})
	routeManager.RegisterModule(&transfer.TransferModule{})

	// 初始化所有路由模块
	routeManager.InitializeRoutes(r)

	return r
}

// TestVersionAPI 测试版本信息API
func TestVersionAPI(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/version", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "version")
	assert.Contains(t, w.Body.String(), "build_time")
	assert.Contains(t, w.Body.String(), "timestamp")
}

// TestAPITest 测试基础API接口
func TestAPITest(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	// 检查响应包含正确的JSON结构
	assert.Contains(t, w.Body.String(), `"code":200`)
	assert.Contains(t, w.Body.String(), "method")
	assert.Contains(t, w.Body.String(), "url")
}

// TestFormatJSON 测试JSON格式接口
func TestFormatJSON(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/json", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"code":200`)
	assert.Contains(t, w.Body.String(), "string")
	assert.Contains(t, w.Body.String(), "number")
}

// TestTransferChunked 测试分块传输接口
func TestTransferChunked(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/transfer/chunked?chunks=2&delay=100", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Header().Get("Transfer-Encoding"), "chunked")
}

// TestSystemInfo 测试系统信息接口
func TestSystemInfo(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test/system", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	// 检查响应包含正确的JSON结构
	assert.Contains(t, w.Body.String(), `"code":200`)
	assert.Contains(t, w.Body.String(), "cpu_cores")
	assert.Contains(t, w.Body.String(), "memory")
}

// TestPerformanceConcurrent 测试并发测试接口
func TestPerformanceConcurrent(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test/concurrent?concurrency=2&requests=5", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"code":200`)
	assert.Contains(t, w.Body.String(), "total_requests")
	assert.Contains(t, w.Body.String(), "success_requests")
}

// TestCORS 测试CORS配置
func TestCORS(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	router.ServeHTTP(w, req)

	assert.Equal(t, 204, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

// BenchmarkAPITest API性能基准测试
func BenchmarkAPITest(b *testing.B) {
	router := setupTestRouter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		router.ServeHTTP(w, req)
	}
}

// BenchmarkSystemInfo 系统信息接口性能基准测试
func BenchmarkSystemInfo(b *testing.B) {
	router := setupTestRouter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test/system", nil)
		router.ServeHTTP(w, req)
	}
}

// BenchmarkTransferChunked 分块传输性能基准测试
func BenchmarkTransferChunked(b *testing.B) {
	router := setupTestRouter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/transfer/chunked?chunks=1", nil)
		router.ServeHTTP(w, req)
	}
}
