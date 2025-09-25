package utils

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Logger 企业级日志记录器
type Logger struct {
	*slog.Logger
	logDir       string
	currentFile  *os.File
	currentDate  string
	maxFileSize  int64
	mu           sync.RWMutex
}

// LogLevel 日志级别
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
)

var (
	globalLogger *Logger
	once         sync.Once
)

// LogConfig 日志配置
type LogConfig struct {
	Level       LogLevel
	LogDir      string
	MaxFileSize int64 // 最大文件大小（字节）
	EnableJSON  bool  // 是否启用JSON格式
}

// InitLogger 初始化全局日志记录器
func InitLogger(config LogConfig) error {
	var err error
	once.Do(func() {
		globalLogger, err = NewLogger(config)
	})
	return err
}

// NewLogger 创建新的日志记录器
func NewLogger(config LogConfig) (*Logger, error) {
	// 默认配置
	if config.LogDir == "" {
		wd, _ := os.Getwd()
		config.LogDir = filepath.Join(wd, "logs")
	}
	if config.MaxFileSize == 0 {
		config.MaxFileSize = 100 * 1024 * 1024 // 100MB
	}

	// 创建日志目录
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %v", err)
	}

	logger := &Logger{
		logDir:      config.LogDir,
		maxFileSize: config.MaxFileSize,
	}

	// 初始化日志文件
	if err := logger.rotateLogFile(); err != nil {
		return nil, fmt.Errorf("初始化日志文件失败: %v", err)
	}

	// 配置slog
	level := logger.parseSlogLevel(config.Level)

	var handler slog.Handler
	if config.EnableJSON {
		handler = slog.NewJSONHandler(logger.currentFile, &slog.HandlerOptions{
			Level: level,
			AddSource: true,
		})
	} else {
		handler = slog.NewTextHandler(logger.currentFile, &slog.HandlerOptions{
			Level: level,
			AddSource: true,
		})
	}

	logger.Logger = slog.New(handler)

	return logger, nil
}

// GetLogger 获取全局日志记录器
func GetLogger() *Logger {
	if globalLogger == nil {
		// 使用默认配置初始化
		InitLogger(LogConfig{
			Level:      LevelInfo,
			EnableJSON: true,
		})
	}
	return globalLogger
}

// rotateLogFile 轮转日志文件
func (l *Logger) rotateLogFile() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	today := time.Now().Format("2006-01-02")

	// 如果是同一天且文件大小未超限，不需要轮转
	if l.currentDate == today && l.currentFile != nil {
		if stat, err := l.currentFile.Stat(); err == nil && stat.Size() < l.maxFileSize {
			return nil
		}
	}

	// 关闭当前文件
	if l.currentFile != nil {
		l.currentFile.Close()
	}

	// 生成新的日志文件名
	baseFileName := fmt.Sprintf("tgbot-%s.log", today)
	logFilePath := filepath.Join(l.logDir, baseFileName)

	// 如果文件已存在且超过大小限制，添加序号
	counter := 1
	for {
		if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
			break
		}

		if stat, err := os.Stat(logFilePath); err == nil && stat.Size() < l.maxFileSize {
			break
		}

		baseFileName = fmt.Sprintf("tgbot-%s-%d.log", today, counter)
		logFilePath = filepath.Join(l.logDir, baseFileName)
		counter++
	}

	// 创建或打开日志文件
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.currentFile = file
	l.currentDate = today

	return nil
}

// parseSlogLevel 解析slog日志级别
func (l *Logger) parseSlogLevel(level LogLevel) slog.Level {
	switch strings.ToUpper(string(level)) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithContext 添加上下文信息
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		Logger:      l.Logger.With(),
		logDir:      l.logDir,
		currentFile: l.currentFile,
		currentDate: l.currentDate,
		maxFileSize: l.maxFileSize,
	}
}

// WithFields 添加结构化字段
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	return &Logger{
		Logger:      l.Logger.With(args...),
		logDir:      l.logDir,
		currentFile: l.currentFile,
		currentDate: l.currentDate,
		maxFileSize: l.maxFileSize,
	}
}

// Close 关闭日志记录器
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.currentFile != nil {
		return l.currentFile.Close()
	}
	return nil
}

// LogRequest 记录请求日志
func (l *Logger) LogRequest(userID int64, chatID int64, command string, message string) {
	l.Info("用户请求",
		slog.Int64("user_id", userID),
		slog.Int64("chat_id", chatID),
		slog.String("command", command),
		slog.String("message", message),
		slog.String("timestamp", time.Now().Format(time.RFC3339)),
	)
}

// LogResponse 记录响应日志
func (l *Logger) LogResponse(userID int64, chatID int64, command string, success bool, duration time.Duration, message string) {
	l.Info("处理响应",
		slog.Int64("user_id", userID),
		slog.Int64("chat_id", chatID),
		slog.String("command", command),
		slog.Bool("success", success),
		slog.String("duration", duration.String()),
		slog.String("message", message),
		slog.String("timestamp", time.Now().Format(time.RFC3339)),
	)
}

// LogFileOperation 记录文件操作日志
func (l *Logger) LogFileOperation(userID int64, operation string, filePath string, size int64, success bool, message string) {
	l.Info("文件操作",
		slog.Int64("user_id", userID),
		slog.String("operation", operation),
		slog.String("file_path", SanitizePath(filePath)),
		slog.Int64("file_size", size),
		slog.Bool("success", success),
		slog.String("message", message),
		slog.String("timestamp", time.Now().Format(time.RFC3339)),
	)
}

// LogError 记录错误日志
func (l *Logger) LogError(userID int64, operation string, err error, context map[string]interface{}) {
	args := []any{
		"user_id", userID,
		"operation", operation,
		"error", err.Error(),
		"timestamp", time.Now().Format(time.RFC3339),
	}

	for k, v := range context {
		args = append(args, k, v)
	}

	l.Error("操作错误", args...)
}

// LogPerformance 记录性能指标
func (l *Logger) LogPerformance(operation string, duration time.Duration, itemCount int, userID int64) {
	l.Info("性能指标",
		slog.String("operation", operation),
		slog.String("duration", duration.String()),
		slog.Int("item_count", itemCount),
		slog.Int64("user_id", userID),
		slog.Float64("items_per_second", float64(itemCount)/duration.Seconds()),
		slog.String("timestamp", time.Now().Format(time.RFC3339)),
	)
}

// SanitizePath 清理敏感路径信息
func SanitizePath(path string) string {
	if path == "" {
		return ""
	}

	// 只保留文件名和最后一级目录
	parts := strings.Split(path, string(os.PathSeparator))
	if len(parts) >= 2 {
		return filepath.Join("...", parts[len(parts)-2], parts[len(parts)-1])
	} else if len(parts) == 1 {
		return parts[0]
	}

	return "unknown"
}

// checkRotation 检查是否需要日志轮转
func (l *Logger) checkRotation() {
	if err := l.rotateLogFile(); err != nil {
		// 轮转失败时记录到标准错误输出
		fmt.Fprintf(os.Stderr, "日志轮转失败: %v\n", err)
	}
}

// Write 实现io.Writer接口，支持日志轮转
func (l *Logger) Write(p []byte) (n int, err error) {
	l.checkRotation()
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.currentFile == nil {
		return 0, fmt.Errorf("日志文件未初始化")
	}

	return l.currentFile.Write(p)
}

// GetLogFiles 获取日志文件列表
func (l *Logger) GetLogFiles() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(l.logDir, "tgbot-*.log"))
	if err != nil {
		return nil, err
	}
	return files, nil
}

// CleanupOldLogs 清理旧日志文件（保留最近N天）
func (l *Logger) CleanupOldLogs(keepDays int) error {
	if keepDays <= 0 {
		return nil
	}

	cutoffDate := time.Now().AddDate(0, 0, -keepDays)
	files, err := l.GetLogFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			continue
		}

		if stat.ModTime().Before(cutoffDate) {
			os.Remove(file)
			l.Info("清理旧日志文件", slog.String("file", filepath.Base(file)))
		}
	}

	return nil
}