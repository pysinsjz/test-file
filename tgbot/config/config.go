package config

import (
	"log"
	"os"
)

type Config struct {
	BotToken    string
	TempDir     string
	MaxFileSize int64 // 最大文件大小 (字节)
	AdminUsers  []int64 // 管理员用户ID列表
}

func LoadConfig() *Config {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN 环境变量未设置")
	}

	tempDir := os.Getenv("TEMP_DIR")
	if tempDir == "" {
		tempDir = "/tmp/tgbot"
	}

	return &Config{
		BotToken:    botToken,
		TempDir:     tempDir,
		MaxFileSize: 50 * 1024 * 1024, // 50MB
		AdminUsers:  []int64{}, // 可以通过环境变量扩展
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