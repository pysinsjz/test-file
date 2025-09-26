package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// FunctionInfo 功能信息结构
type FunctionInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	InputFormat string `json:"input_format"`
	OutputFormat string `json:"output_format"`
	Icon        string `json:"icon"`
	Example     string `json:"example"`
}

// 所有可用功能
var Functions = map[string]FunctionInfo{
	"logparse": {
		ID:          "logparse",
		Name:        "日志解析",
		Description: "从应用程序日志中提取关键信息，生成结构化CSV数据",
		InputFormat: "TXT",
		OutputFormat: "CSV",
		Icon:        "📊",
		Example:     "上传包含用户行为、支付流水等信息的日志文件",
	},
	"lockuser": {
		ID:          "lockuser",
		Name:        "用户锁定",
		Description: "批量生成用户账户锁定的SQL和Redis命令",
		InputFormat: "CSV",
		OutputFormat: "SQL + Redis命令",
		Icon:        "🔒",
		Example:     "第一列包含需要锁定的用户ID",
	},
	"sqlparse": {
		ID:          "sqlparse",
		Name:        "SQL解析",
		Description: "从日志中提取SQL语句并智能去重",
		InputFormat: "TXT",
		OutputFormat: "去重SQL文件",
		Icon:        "🗄️",
		Example:     "包含数据库操作日志的文本文件",
	},
	"filesplit": {
		ID:          "filesplit",
		Name:        "文件分割",
		Description: "将大文件按行数分割成多个小文件",
		InputFormat: "任意格式",
		OutputFormat: "多个小文件",
		Icon:        "✂️",
		Example:     "大型数据文件、Redis命令文件等",
	},
	"kycreview": {
		ID:          "kycreview",
		Name:        "KYC审核",
		Description: "处理KYC（身份验证）审核通过数据",
		InputFormat: "Excel/CSV",
		OutputFormat: "SQL更新语句",
		Icon:        "📋",
		Example:     "包含用户KYC审核结果的表格文件",
	},
	"redisdel": {
		ID:          "redisdel",
		Name:        "Redis删除",
		Description: "生成用户数据的Redis删除命令",
		InputFormat: "Excel/CSV",
		OutputFormat: "Redis命令文件",
		Icon:        "🗑️",
		Example:     "包含需要清理数据的用户ID列表",
	},
	"redisadd": {
		ID:          "redisadd",
		Name:        "Redis增加",
		Description: "生成用户流水要求设置命令",
		InputFormat: "CSV",
		OutputFormat: "Redis设置命令",
		Icon:        "➕",
		Example:     "包含用户ID、金额、比例等字段的CSV文件",
	},
	"uiddedup": {
		ID:          "uiddedup",
		Name:        "UID去重",
		Description: "从用户ID列表中移除重复项",
		InputFormat: "CSV",
		OutputFormat: "去重后的CSV",
		Icon:        "🔄",
		Example:     "包含可能重复用户ID的CSV文件",
	},
}

// TaskInfo 任务信息
type TaskInfo struct {
	ID          string    `json:"id"`
	Function    string    `json:"function"`
	Status      string    `json:"status"` // pending, processing, completed, failed
	Progress    int       `json:"progress"`
	Message     string    `json:"message"`
	InputFile   string    `json:"input_file"`
	OutputFiles []string  `json:"output_files"`
	StartTime   time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
}

// 全局任务存储 (生产环境应使用数据库)
var tasks = make(map[string]*TaskInfo)

// IndexHandler 主页处理器
func IndexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":     "数据处理工具",
		"functions": Functions,
	})
}

// UploadPageHandler 上传页面处理器
func UploadPageHandler(c *gin.Context) {
	functionID := c.Param("function")

	function, exists := Functions[functionID]
	if !exists {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "功能不存在",
		})
		return
	}

	c.HTML(http.StatusOK, "upload.html", gin.H{
		"title":    function.Name,
		"function": function,
	})
}

// HelpHandler 帮助页面处理器
func HelpHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "help.html", gin.H{
		"title":     "使用帮助",
		"functions": Functions,
	})
}

// generateTaskID 生成任务ID
func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

// getFileExtension 获取文件扩展名
func getFileExtension(filename string) string {
	return filepath.Ext(filename)
}

// isValidFileType 检查文件类型是否有效
func isValidFileType(filename string, allowedTypes []string) bool {
	ext := getFileExtension(filename)
	for _, allowedType := range allowedTypes {
		if ext == allowedType {
			return true
		}
	}
	return false
}