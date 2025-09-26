package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"webbot/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由
	r := gin.Default()

	// 设置模板函数
	r.SetFuncMap(template.FuncMap{
		"base": filepath.Base,
	})

	// 加载 HTML 模板
	r.LoadHTMLGlob("templates/*")

	// 静态文件服务
	r.Static("/static", "./static")
	r.Static("/uploads", "./uploads")

	// 路由设置
	setupRoutes(r)
	host := "0.0.0.0:9088"
	log.Println("🚀 WebBot 服务启动成功!")
	log.Println("📱 访问地址: " + host)
	log.Fatal(r.Run(host))
}

func setupRoutes(r *gin.Engine) {
	// 主页路由
	r.GET("/", handlers.IndexHandler)

	// 功能页面路由
	r.GET("/upload/:function", handlers.UploadPageHandler)
	r.POST("/process/:function", handlers.ProcessFileHandler)
	r.GET("/result/:taskid", handlers.ResultHandler)

	// API 路由
	api := r.Group("/api")
	{
		api.POST("/upload", handlers.UploadFileHandler)
		api.GET("/progress/:taskid", handlers.ProgressHandler)
		api.GET("/download/*filepath", handlers.DownloadHandler)
	}

	// 帮助页面
	r.GET("/help", handlers.HelpHandler)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "webbot",
		})
	})
}
