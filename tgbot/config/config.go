package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	BotToken    string
	TempDir     string
	MaxFileSize int64   // 最大文件大小 (字节)
	AdminUsers  []int64 // 管理员用户ID列表
	Log         LogConfig
}

type LogConfig struct {
	Level       string // DEBUG, INFO, WARN, ERROR
	LogDir      string // 日志目录
	MaxFileSize int64  // 日志文件最大大小
	EnableJSON  bool   // 是否启用JSON格式
	KeepDays    int    // 保留日志天数
}

func LoadConfig() *Config {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		botToken = "8485165772:AAH4XePh1c2vL8VwBh_oqhs0ywFY3-cToqE"
	}

	tempDir := os.Getenv("TEMP_DIR")
	if tempDir == "" {
		tempDir = "/tmp/tgbot"
	}

	// 日志配置
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}

	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		wd, _ := os.Getwd()
		logDir = filepath.Join(wd, "logs")
	}

	logMaxSize, _ := strconv.ParseInt(os.Getenv("LOG_MAX_SIZE"), 10, 64)
	if logMaxSize == 0 {
		logMaxSize = 100 * 1024 * 1024 // 100MB
	}

	enableJSON, _ := strconv.ParseBool(os.Getenv("LOG_JSON"))
	if os.Getenv("LOG_JSON") == "" {
		enableJSON = true // 默认启用JSON
	}

	keepDays, _ := strconv.Atoi(os.Getenv("LOG_KEEP_DAYS"))
	if keepDays == 0 {
		keepDays = 30 // 默认保留30天
	}

	return &Config{
		BotToken:    botToken,
		TempDir:     tempDir,
		MaxFileSize: 50 * 1024 * 1024, // 50MB
		AdminUsers:  []int64{},        // 可以通过环境变量扩展
		Log: LogConfig{
			Level:       strings.ToUpper(logLevel),
			LogDir:      logDir,
			MaxFileSize: logMaxSize,
			EnableJSON:  enableJSON,
			KeepDays:    keepDays,
		},
	}
}

// IsAdmin 检查用户是否是管理员
func (c *Config) IsAdmin(userID int64) bool {
	for _, adminID := range c.AdminUsers {
		if adminID == userID {
			return true
		}
	}
	return false
}
