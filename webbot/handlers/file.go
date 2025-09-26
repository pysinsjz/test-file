package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"webbot/processor"

	"github.com/gin-gonic/gin"
)

const MaxFileSize = 50 * 1024 * 1024 // 50MB

// UploadFileHandler 文件上传处理器
func UploadFileHandler(c *gin.Context) {
	functionID := c.PostForm("function")

	// 验证功能是否存在
	function, exists := Functions[functionID]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的功能类型",
		})
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文件上传失败: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// 检查文件大小
	if header.Size > MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("文件过大，最大支持 %d MB", MaxFileSize/1024/1024),
		})
		return
	}

	// 验证文件类型
	if !isValidFileForFunction(header.Filename, functionID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("不支持的文件格式，%s功能要求%s格式", function.Name, function.InputFormat),
		})
		return
	}

	// 生成任务ID
	taskID := generateTaskID()

	// 创建上传目录
	uploadDir := filepath.Join("uploads", taskID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建上传目录失败",
		})
		return
	}

	// 保存文件
	filename := filepath.Join(uploadDir, header.Filename)
	out, err := os.Create(filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "保存文件失败",
		})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "写入文件失败",
		})
		return
	}

	// 创建任务记录
	task := &TaskInfo{
		ID:        taskID,
		Function:  functionID,
		Status:    "pending",
		Progress:  0,
		Message:   "等待处理...",
		InputFile: filename,
		StartTime: time.Now(),
	}
	tasks[taskID] = task

	log.Printf("创建任务 %s: %s - %s", taskID, function.Name, header.Filename)

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"message": "文件上传成功",
	})
}

// ProcessFileHandler 文件处理处理器
func ProcessFileHandler(c *gin.Context) {
	taskID := c.PostForm("task_id")

	task, exists := tasks[taskID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "任务不存在",
		})
		return
	}

	// 开始异步处理
	go processFileAsync(task)

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"message": "开始处理文件",
	})
}

// processFileAsync 异步处理文件
func processFileAsync(task *TaskInfo) {
	defer func() {
		if r := recover(); r != nil {
			task.Status = "failed"
			task.Message = fmt.Sprintf("处理过程中发生错误: %v", r)
			now := time.Now()
			task.EndTime = &now
			log.Printf("任务 %s 处理失败: %v", task.ID, r)
		}
	}()

	// 更新任务状态
	task.Status = "processing"
	task.Progress = 10
	task.Message = "正在初始化..."

	// 创建输出目录
	outputDir := filepath.Join("uploads", task.ID, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		task.Status = "failed"
		task.Message = "创建输出目录失败"
		return
	}

	// 根据功能类型调用对应的处理器
	var outputFiles []string
	var err error

	task.Progress = 30
	task.Message = "正在处理文件..."

	switch task.Function {
	case "logparse":
		outputFiles, err = processor.ProcessLogParse(task.InputFile, outputDir, updateProgress(task))
	case "lockuser":
		outputFiles, err = processor.ProcessLockUser(task.InputFile, outputDir, updateProgress(task))
	case "sqlparse":
		outputFiles, err = processor.ProcessSQLParse(task.InputFile, outputDir, updateProgress(task))
	case "filesplit":
		outputFiles, err = processor.ProcessFileSplit(task.InputFile, outputDir, updateProgress(task))
	case "kycreview":
		outputFiles, err = processor.ProcessKYCReview(task.InputFile, outputDir, updateProgress(task))
	case "redisdel":
		outputFiles, err = processor.ProcessRedisDel(task.InputFile, outputDir, updateProgress(task))
	case "redisadd":
		outputFiles, err = processor.ProcessRedisAdd(task.InputFile, outputDir, updateProgress(task))
	case "uiddedup":
		outputFiles, err = processor.ProcessUIDDedup(task.InputFile, outputDir, updateProgress(task))
	default:
		err = fmt.Errorf("不支持的功能类型: %s", task.Function)
	}

	now := time.Now()
	task.EndTime = &now

	if err != nil {
		task.Status = "failed"
		task.Message = "处理失败: " + err.Error()
		log.Printf("任务 %s 处理失败: %v", task.ID, err)
		return
	}

	// 处理成功
	task.Status = "completed"
	task.Progress = 100
	task.Message = "处理完成"

	// 清理输出文件路径，移除 uploads/ 前缀以适配下载URL
	cleanedOutputFiles := make([]string, len(outputFiles))
	for i, file := range outputFiles {
		if strings.HasPrefix(file, "uploads/") {
			cleanedOutputFiles[i] = file[8:] // 移除 "uploads/" (8个字符)
		} else {
			cleanedOutputFiles[i] = file
		}
	}
	task.OutputFiles = cleanedOutputFiles

	log.Printf("任务 %s 处理成功，输出 %d 个文件", task.ID, len(outputFiles))
}

// updateProgress 创建进度更新函数
func updateProgress(task *TaskInfo) func(progress int, message string) {
	return func(progress int, message string) {
		if progress > task.Progress {
			task.Progress = progress
		}
		if message != "" {
			task.Message = message
		}
	}
}

// ProgressHandler 进度查询处理器
func ProgressHandler(c *gin.Context) {
	taskID := c.Param("taskid")

	task, exists := tasks[taskID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "任务不存在",
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// ResultHandler 结果页面处理器
func ResultHandler(c *gin.Context) {
	taskID := c.Param("taskid")

	task, exists := tasks[taskID]
	if !exists {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "任务不存在",
		})
		return
	}

	function := Functions[task.Function]

	c.HTML(http.StatusOK, "result.html", gin.H{
		"title":    "处理结果",
		"task":     task,
		"function": function,
	})
}

// DownloadHandler 文件下载处理器
func DownloadHandler(c *gin.Context) {
	filePath := c.Param("filepath")

	// 移除开头的斜杠
	if len(filePath) > 0 && filePath[0] == '/' {
		filePath = filePath[1:]
	}

	// 安全检查，防止路径遍历攻击
	if filepath.IsAbs(filePath) || filepath.Clean(filePath) != filePath {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的文件路径",
		})
		return
	}

	// 构建完整的文件路径
	fullPath := filepath.Join("uploads", filePath)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "文件不存在",
		})
		return
	}

	// 设置下载响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(filePath))
	c.File(fullPath)
}

// isValidFileForFunction 检查文件是否适用于特定功能
func isValidFileForFunction(filename, functionID string) bool {
	ext := getFileExtension(filename)

	switch functionID {
	case "logparse", "sqlparse":
		return ext == ".txt"
	case "lockuser", "redisadd", "uiddedup":
		return ext == ".csv"
	case "kycreview", "redisdel":
		return ext == ".csv" || ext == ".xlsx"
	case "filesplit":
		return true // 支持任意格式
	default:
		return false
	}
}