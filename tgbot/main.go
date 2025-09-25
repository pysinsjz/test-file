package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"tgbot/config"
	"tgbot/handlers"
	"tgbot/utils"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化日志系统
	logConfig := utils.LogConfig{
		Level:       utils.LogLevel(cfg.Log.Level),
		LogDir:      cfg.Log.LogDir,
		MaxFileSize: cfg.Log.MaxFileSize,
		EnableJSON:  cfg.Log.EnableJSON,
	}

	if err := utils.InitLogger(logConfig); err != nil {
		log.Fatalf("初始化日志系统失败: %v", err)
	}

	logger := utils.GetLogger()
	defer logger.Close()

	// 记录系统启动
	logger.Info("Telegram Bot启动",
		slog.String("version", "1.0.0"),
		slog.String("log_level", cfg.Log.Level),
		slog.String("log_dir", cfg.Log.LogDir),
		slog.String("temp_dir", cfg.TempDir),
		slog.String("timestamp", time.Now().Format(time.RFC3339)),
	)

	// 创建Bot实例
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		logger.Error("创建Bot失败",
			slog.String("error", err.Error()),
			slog.String("timestamp", time.Now().Format(time.RFC3339)),
		)
		log.Fatalf("创建Bot失败: %v", err)
	}

	// 设置调试模式 (生产环境应设为false)
	bot.Debug = false

	logger.Info("Bot初始化成功",
		slog.String("bot_username", bot.Self.UserName),
		slog.Int64("bot_id", bot.Self.ID),
		slog.String("timestamp", time.Now().Format(time.RFC3339)),
	)

	log.Printf("Bot已启动，用户名: %s", bot.Self.UserName)

	// 创建文件管理器
	fileManager := utils.NewFileManager(cfg.TempDir)
	defer func() {
		logger.Info("系统关闭：开始清理文件", slog.String("timestamp", time.Now().Format(time.RFC3339)))
		fileManager.CleanupFiles()
		logger.Info("系统关闭：文件清理完成", slog.String("timestamp", time.Now().Format(time.RFC3339)))
	}()

	// 创建处理器管理器
	handlerManager := handlers.NewHandlerManager(bot, cfg, fileManager, logger)

	// 设置更新配置
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// 优雅关闭处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("收到关闭信号",
			slog.String("signal", sig.String()),
			slog.String("timestamp", time.Now().Format(time.RFC3339)),
		)

		log.Println("正在关闭Bot...")

		// 清理旧日志文件
		logger.CleanupOldLogs(cfg.Log.KeepDays)

		fileManager.CleanupFiles()
		bot.StopReceivingUpdates()

		logger.Info("Bot已关闭", slog.String("timestamp", time.Now().Format(time.RFC3339)))
		os.Exit(0)
	}()

	logger.Info("开始处理消息", slog.String("timestamp", time.Now().Format(time.RFC3339)))

	// 处理消息循环
	for update := range updates {
		go func(update tgbotapi.Update) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("处理消息时发生panic",
						slog.Any("panic", r),
						slog.String("timestamp", time.Now().Format(time.RFC3339)),
					)
					log.Printf("处理消息时发生panic: %v", r)
				}
			}()

			handlerManager.HandleUpdate(update)
		}(update)
	}
}