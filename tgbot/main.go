package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"tgbot/config"
	"tgbot/handlers"
	"tgbot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 创建Bot实例
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalf("创建Bot失败: %v", err)
	}

	// 设置调试模式 (生产环境应设为false)
	bot.Debug = false
	log.Printf("Bot已启动，用户名: %s", bot.Self.UserName)

	// 创建文件管理器
	fileManager := utils.NewFileManager(cfg.TempDir)
	defer fileManager.CleanupFiles()

	// 创建处理器管理器
	handlerManager := handlers.NewHandlerManager(bot, cfg, fileManager)

	// 设置更新配置
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// 优雅关闭处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("正在关闭Bot...")
		fileManager.CleanupFiles()
		bot.StopReceivingUpdates()
		os.Exit(0)
	}()

	// 处理消息循环
	for update := range updates {
		go func(update tgbotapi.Update) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("处理消息时发生panic: %v", r)
				}
			}()

			handlerManager.HandleUpdate(update)
		}(update)
	}
}