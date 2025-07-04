package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	// 版本信息变量，将在构建时注入
	Version   = "dev"
	BuildTime = "unknown"
	BuildCommit = "unknown"
)

// 嵌入静态资源文件和模板文件
//go:embed static/*
var staticFS embed.FS

//go:embed templates/*
var templatesFS embed.FS

func main() {
	// 定义命令行参数
	var (
		showVersion = flag.Bool("version", false, "显示版本信息")
		showHelp    = flag.Bool("help", false, "显示帮助信息")
		port        = flag.String("port", "", "服务器端口")
	)
	flag.Parse()

	// 处理版本信息
	if *showVersion {
		fmt.Printf("HTTP/WebSocket代理测试工具\n")
		fmt.Printf("版本: %s\n", Version)
		fmt.Printf("构建时间: %s\n", BuildTime)
		fmt.Printf("提交哈希: %s\n", BuildCommit)
		os.Exit(0)
	}

	// 处理帮助信息
	if *showHelp {
		fmt.Printf("HTTP/WebSocket代理测试工具\n\n")
		fmt.Printf("使用方法:\n")
		fmt.Printf("  %s [选项]\n\n", os.Args[0])
		fmt.Printf("选项:\n")
		fmt.Printf("  -version    显示版本信息\n")
		fmt.Printf("  -help       显示帮助信息\n")
		fmt.Printf("  -port       设置服务器端口 (默认: 8080)\n\n")
		fmt.Printf("示例:\n")
		fmt.Printf("  %s -port 9090\n", os.Args[0])
		fmt.Printf("  %s -version\n", os.Args[0])
		os.Exit(0)
	}

	// 创建Gin引擎
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 静态文件服务 - 使用嵌入的文件系统
	staticSubFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal("无法创建静态文件子系统:", err)
	}
	r.StaticFS("/static", http.FS(staticSubFS))
	
	// 模板文件 - 使用嵌入的文件系统
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		log.Fatal("无法解析模板文件:", err)
	}
	r.SetHTMLTemplate(tmpl)

	// 主页
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":       "HTTP/WebSocket代理测试工具",
			"version":     Version,
			"buildTime":   BuildTime,
			"buildCommit": BuildCommit,
		})
	})

	// API文档页面
	r.GET("/api-docs", func(c *gin.Context) {
		c.HTML(http.StatusOK, "api-docs.html", gin.H{
			"title":       "API文档",
			"version":     Version,
			"buildTime":   BuildTime,
			"buildCommit": BuildCommit,
		})
	})

	// 版本信息API
	r.GET("/api/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"version":     Version,
				"buildTime":   BuildTime,
				"buildCommit": BuildCommit,
				"goVersion":   fmt.Sprintf("%s", os.Getenv("GO_VERSION")),
			},
		})
	})

	// 初始化路由
	initAPIRoutes(r)
	initWebSocketRoutes(r)
	initTestRoutes(r)

	// 获取端口
	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "8080"
	}
	
	// 如果命令行指定了端口，使用命令行参数
	if *port != "" {
		serverPort = *port
	}

	log.Printf("服务器启动在端口 %s", serverPort)
	log.Printf("版本: %s, 构建时间: %s", Version, BuildTime)
	log.Fatal(r.Run(":" + serverPort))
} 