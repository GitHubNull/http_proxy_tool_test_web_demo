package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"http_proxy_tool_test_web_demo/routes"
	"http_proxy_tool_test_web_demo/routes/api"
	"http_proxy_tool_test_web_demo/routes/format"
	"http_proxy_tool_test_web_demo/routes/test/performance"
	"http_proxy_tool_test_web_demo/routes/test/system"
	"http_proxy_tool_test_web_demo/routes/transfer"
)

var (
	version     string = "dev"
	buildTime   string = "unknown"
	port               = flag.String("port", "8080", "服务器端口")
	logDir             = flag.String("log-dir", "logs", "日志目录")
	showVersion        = flag.Bool("version", false, "显示版本信息")
	showHelp           = flag.Bool("help", false, "显示帮助信息")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("版本: %s\n", version)
		fmt.Printf("构建时间: %s\n", buildTime)
		os.Exit(0)
	}

	if *showHelp {
		fmt.Println("HTTP代理测试工具")
		fmt.Println("")
		fmt.Println("选项:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// 初始化日志系统
	initLogger(*logDir)

	// 创建Gin引擎
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 静态文件服务
	r.Static("/static", "./static")

	// 加载HTML模板
	r.SetHTMLTemplate(template.Must(template.ParseGlob("templates/*.html")))

	// 创建路由管理器
	routeManager := routes.NewRouteManager()

	// 注册所有模块
	routeManager.RegisterModule(&api.BasicAPIModule{})
	routeManager.RegisterModule(&format.FormatModule{})
	routeManager.RegisterModule(&performance.PerformanceModule{})
	routeManager.RegisterModule(&system.SystemModule{})
	routeManager.RegisterModule(&transfer.TransferModule{})

	// 基础路由
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":     "HTTP代理测试工具",
			"version":   version,
			"buildTime": buildTime,
		})
	})

	// API文档路由（带返回主页链接）
	r.GET("/api-docs", func(c *gin.Context) {
		modules := routeManager.GetRegisteredModules()

		// 构建API文档数据
		apiGroups := []map[string]interface{}{
			{
				"name":        "基础API测试",
				"prefix":      "/api",
				"description": "基础的HTTP方法测试、状态码测试、延迟测试等",
				"endpoints": []map[string]string{
					{"method": "GET/POST/PUT/DELETE", "path": "/api/test", "desc": "通用HTTP方法测试"},
					{"method": "GET/POST", "path": "/api/status/:code", "desc": "HTTP状态码测试"},
					{"method": "GET/POST", "path": "/api/delay/:seconds", "desc": "延迟响应测试"},
					{"method": "GET", "path": "/api/redirect/:times", "desc": "重定向测试"},
					{"method": "GET", "path": "/api/redirect-to", "desc": "重定向到指定URL"},
					{"method": "GET", "path": "/api/error", "desc": "错误响应测试"},
					{"method": "GET", "path": "/api/timeout", "desc": "超时测试"},
				},
			},
			{
				"name":        "格式处理测试",
				"prefix":      "/api",
				"description": "支持JSON、XML、HTML、文本、二进制等多种数据格式",
				"endpoints": []map[string]string{
					{"method": "GET", "path": "/api/json", "desc": "JSON格式响应"},
					{"method": "GET", "path": "/api/xml", "desc": "XML格式响应"},
					{"method": "GET", "path": "/api/html", "desc": "HTML格式响应"},
					{"method": "GET", "path": "/api/text", "desc": "纯文本格式响应"},
					{"method": "GET", "path": "/api/binary", "desc": "二进制格式响应"},
					{"method": "POST", "path": "/api/parse/json", "desc": "JSON解析测试"},
					{"method": "POST", "path": "/api/parse/xml", "desc": "XML解析测试"},
					{"method": "POST", "path": "/api/parse/multipart", "desc": "Multipart解析测试"},
					{"method": "POST", "path": "/api/parse/binary", "desc": "二进制数据解析测试"},
					{"method": "GET", "path": "/api/gzip", "desc": "Gzip压缩测试"},
					{"method": "GET", "path": "/api/deflate", "desc": "Deflate压缩测试"},
					{"method": "GET", "path": "/api/stream/:lines", "desc": "流式数据测试"},
				},
			},
			{
				"name":        "传输编码测试",
				"prefix":      "/api/transfer",
				"description": "分块传输(chunked)、大文件传输、SSE等传输测试",
				"endpoints": []map[string]string{
					{"method": "GET", "path": "/api/transfer/chunked", "desc": "分块传输测试"},
					{"method": "POST", "path": "/api/transfer/chunked", "desc": "分块接收测试"},
					{"method": "GET", "path": "/api/transfer/chunked/stream", "desc": "分块流式传输"},
					{"method": "POST", "path": "/api/transfer/chunked/upload", "desc": "分块上传测试"},
					{"method": "GET", "path": "/api/transfer/large/:size", "desc": "大文件传输测试"},
					{"method": "POST", "path": "/api/transfer/large", "desc": "大文件接收测试"},
					{"method": "GET", "path": "/api/transfer/stream/sse", "desc": "SSE流式传输"},
				},
			},
			{
				"name":        "性能测试",
				"prefix":      "/test",
				"description": "并发测试、压力测试、负载测试等性能相关测试",
				"endpoints": []map[string]string{
					{"method": "GET/POST", "path": "/test/concurrent", "desc": "并发测试"},
					{"method": "GET/POST", "path": "/test/stress", "desc": "压力测试"},
					{"method": "POST", "path": "/test/batch", "desc": "批量请求测试"},
					{"method": "GET", "path": "/test/load", "desc": "负载测试"},
					{"method": "GET", "path": "/test/random-delay", "desc": "随机延迟测试"},
					{"method": "GET", "path": "/test/stats", "desc": "获取测试统计"},
					{"method": "POST", "path": "/test/reset", "desc": "重置测试统计"},
				},
			},
			{
				"name":        "系统资源测试",
				"prefix":      "/test",
				"description": "内存、CPU、网络、文件IO、数据库等系统资源测试",
				"endpoints": []map[string]string{
					{"method": "GET", "path": "/test/memory", "desc": "内存压力测试"},
					{"method": "GET", "path": "/test/cpu", "desc": "CPU压力测试"},
					{"method": "GET", "path": "/test/network", "desc": "网络测试"},
					{"method": "GET", "path": "/test/fileio", "desc": "文件IO测试"},
					{"method": "GET", "path": "/test/database", "desc": "数据库连接测试"},
					{"method": "GET", "path": "/test/keepalive", "desc": "长连接测试"},
					{"method": "GET", "path": "/test/system", "desc": "系统信息查询"},
				},
			},
		}

		c.HTML(http.StatusOK, "api-docs.html", gin.H{
			"title":       "API接口文档",
			"version":     version,
			"buildTime":   buildTime,
			"modules":     modules,
			"apiGroups":   apiGroups,
			"currentTime": time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	// 版本信息API
	r.GET("/api/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":    version,
			"build_time": buildTime,
			"timestamp":  time.Now().Unix(),
		})
	})

	// 初始化所有路由模块
	routeManager.InitializeRoutes(r)

	// 启动服务器
	serverAddr := ":" + *port
	log.Printf("服务器启动在端口 %s", *port)
	log.Printf("版本: %s, 构建时间: %s", version, buildTime)
	log.Printf("访问 http://localhost:%s 查看主页", *port)
	log.Printf("访问 http://localhost:%s/api-docs 查看API文档", *port)

	if err := r.Run(serverAddr); err != nil {
		log.Fatal(err)
	}
}

// 初始化日志系统（简化版本）
func initLogger(logDir string) {
	// 创建日志目录
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("创建日志目录失败: %v", err)
		return
	}

	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("日志系统初始化完成，日志目录: %s", logDir)
}
