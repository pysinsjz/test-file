package handlers

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"tgbot/config"
	"tgbot/utils"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandlerManager 处理器管理器
type HandlerManager struct {
	bot         *tgbotapi.BotAPI
	config      *config.Config
	fileManager *utils.FileManager
	logger      *utils.Logger
	userStates  sync.Map // 用户状态管理
}

// UserState 用户状态
type UserState struct {
	CurrentCommand string
	UserDir        string
	Data           map[string]interface{}
}

// NewHandlerManager 创建处理器管理器
func NewHandlerManager(bot *tgbotapi.BotAPI, cfg *config.Config, fm *utils.FileManager, logger *utils.Logger) *HandlerManager {
	return &HandlerManager{
		bot:         bot,
		config:      cfg,
		fileManager: fm,
		logger:      logger,
	}
}

// HandleUpdate 处理更新
func (hm *HandlerManager) HandleUpdate(update tgbotapi.Update) {
	startTime := time.Now()

	if update.Message != nil {
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID
		username := update.Message.From.UserName

		hm.logger.Info("收到消息",
			slog.Int64("user_id", userID),
			slog.Int64("chat_id", chatID),
			slog.String("username", username),
			slog.String("message_type", "message"),
			slog.String("timestamp", startTime.Format(time.RFC3339)),
		)

		hm.handleMessage(update.Message)

		hm.logger.LogResponse(userID, chatID, "message_processing", true,
			time.Since(startTime), "消息处理完成")

	} else if update.CallbackQuery != nil {
		userID := update.CallbackQuery.From.ID
		chatID := update.CallbackQuery.Message.Chat.ID
		username := update.CallbackQuery.From.UserName

		hm.logger.Info("收到回调查询",
			slog.Int64("user_id", userID),
			slog.Int64("chat_id", chatID),
			slog.String("username", username),
			slog.String("callback_data", update.CallbackQuery.Data),
			slog.String("timestamp", startTime.Format(time.RFC3339)),
		)

		hm.handleCallbackQuery(update.CallbackQuery)

		hm.logger.LogResponse(userID, chatID, "callback_processing", true,
			time.Since(startTime), "回调查询处理完成")
	}
}

// handleMessage 处理消息
func (hm *HandlerManager) handleMessage(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// 处理命令
	if message.IsCommand() {
		command := message.Command()
		hm.handleCommand(chatID, userID, command, message.CommandArguments())
		return
	}

	// 处理文件
	if message.Document != nil {
		hm.handleDocument(chatID, userID, message.Document)
		return
	}

	// 处理普通文本消息
	hm.handleTextMessage(chatID, userID, message.Text)
}

// handleCommand 处理命令
func (hm *HandlerManager) handleCommand(chatID, userID int64, command, args string) {
	hm.logger.LogRequest(userID, chatID, command, fmt.Sprintf("命令参数: %s", args))

	startTime := time.Now()

	switch command {
	case "start":
		hm.sendStartMessage(chatID)
	case "help":
		hm.sendHelpMessage(chatID)
	case "menu":
		hm.sendMenuMessage(chatID)
	case "logparse":
		hm.startLogParseProcess(chatID, userID)
	case "lockuser":
		hm.startLockUserProcess(chatID, userID)
	case "sqlparse":
		hm.startSQLParseProcess(chatID, userID)
	case "filesplit":
		hm.startFileSplitProcess(chatID, userID)
	case "kycreview":
		hm.startKYCReviewProcess(chatID, userID)
	case "redisdel":
		hm.startRedisDelProcess(chatID, userID)
	case "redisadd":
		hm.startRedisAddProcess(chatID, userID)
	case "uiddedup":
		hm.startUIDDedupProcess(chatID, userID)
	case "status":
		hm.sendStatusMessage(chatID, userID)
	default:
		hm.logger.Warn("未知命令",
			slog.Int64("user_id", userID),
			slog.Int64("chat_id", chatID),
			slog.String("command", command),
			slog.String("timestamp", time.Now().Format(time.RFC3339)),
		)
		msg := tgbotapi.NewMessage(chatID, "未知命令，请输入 /menu 查看功能菜单或 /help 查看帮助")
		hm.bot.Send(msg)
	}

	hm.logger.LogResponse(userID, chatID, command, true,
		time.Since(startTime), fmt.Sprintf("命令 %s 处理完成", command))
}

// sendStartMessage 发送欢迎消息
func (hm *HandlerManager) sendStartMessage(chatID int64) {
	welcomeText := `🤖 *数据处理Bot*

欢迎使用多功能数据处理Bot！我可以帮您处理各种数据文件。

📋 *可用功能：*
• /logparse - 日志追踪解析
• /lockuser - 用户锁定操作
• /sqlparse - SQL日志解析
• /filesplit - 文件分割
• /kycreview - KYC审核处理
• /redisdel - Redis流水清零命令生成
• /redisadd - Redis流水增加命令生成
• /uiddedup - UID去重处理

💡 *使用方法：*
1. 选择您需要的功能命令
2. 按提示上传相应的文件
3. 等待处理完成并下载结果

输入 /help 获取详细帮助信息。`

	msg := tgbotapi.NewMessage(chatID, welcomeText)
	msg.ParseMode = "Markdown"

	// 创建内联键盘
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 日志解析", "cmd_logparse"),
			tgbotapi.NewInlineKeyboardButtonData("🔒 用户锁定", "cmd_lockuser"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗄️ SQL解析", "cmd_sqlparse"),
			tgbotapi.NewInlineKeyboardButtonData("✂️ 文件分割", "cmd_filesplit"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 KYC审核", "cmd_kycreview"),
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Redis流水清零命令", "cmd_redisdel"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Redis增加", "cmd_redisadd"),
			tgbotapi.NewInlineKeyboardButtonData("🔄 UID去重", "cmd_uiddedup"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ 帮助", "cmd_help"),
			tgbotapi.NewInlineKeyboardButtonData("📈 状态", "cmd_status"),
		),
	)
	msg.ReplyMarkup = keyboard

	hm.bot.Send(msg)
}

// sendHelpMessage 发送帮助信息
func (hm *HandlerManager) sendHelpMessage(chatID int64) {
	helpText := `📚 *详细帮助文档*

*🔧 各功能详细说明：*

*1. 📊 日志解析 (/logparse)*
• 输入：TXT格式的日志文件
• 输出：CSV格式的结构化数据
• 功能：提取关键信息如用户ID、追踪ID等

*2. 🔒 用户锁定 (/lockuser)*
• 输入：包含用户ID的CSV文件
• 输出：SQL更新语句 + Redis删除命令
• 功能：批量生成用户锁定操作命令

*3. 🗄️ SQL解析 (/sqlparse)*
• 输入：包含SQL信息的TXT日志文件
• 输出：去重后的SQL语句文件
• 功能：智能去重，提取唯一SQL语句

*4. ✂️ 文件分割 (/filesplit)*
• 输入：大文件（任意格式）
• 输出：按10,000行分割的小文件
• 功能：将大文件拆分成易处理的小文件

*5. 📋 KYC审核 (/kycreview)*
• 输入：Excel或CSV格式的KYC数据
• 输出：KYC审核通过的SQL更新语句
• 功能：批量处理KYC审核结果

*6. 🗑️ Redis流水清零命令 (/redisdel)*
• 输入：Excel或CSV格式的用户数据
• 输出：Redis删除命令文件
• 功能：生成流水删除命令

*7. ➕ Redis增加 (/redisadd)*
• 输入：包含流水比例数据的CSV文件
• 输出：Redis设置命令文件
• 功能：生成用户流水要求设置命令

*8. 🔄 UID去重 (/uiddedup)*
• 输入：包含用户ID的CSV文件
• 输出：去重后的唯一用户ID文件
• 功能：移除重复的用户ID

*📝 使用提示：*
• 文件大小限制：50MB
• 支持的格式：TXT, CSV, XLSX
• 处理过程中请耐心等待
• 大文件处理可能需要几分钟时间

*🚀 快速访问：*
• 输入 /menu 随时显示功能菜单
• 处理完成后会自动返回菜单

有问题请联系管理员。`

	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ParseMode = "Markdown"
	hm.bot.Send(msg)
}

// handleCallbackQuery 处理回调查询
func (hm *HandlerManager) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	chatID := callbackQuery.Message.Chat.ID
	userID := callbackQuery.From.ID
	data := callbackQuery.Data

	// 回答回调查询
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	hm.bot.Request(callback)

	// 处理回调数据
	if strings.HasPrefix(data, "cmd_") {
		command := strings.TrimPrefix(data, "cmd_")
		hm.handleCommand(chatID, userID, command, "")
	}
}

// 获取用户状态
func (hm *HandlerManager) getUserState(userID int64) *UserState {
	if state, ok := hm.userStates.Load(userID); ok {
		return state.(*UserState)
	}
	return &UserState{
		Data: make(map[string]interface{}),
	}
}

// 设置用户状态
func (hm *HandlerManager) setUserState(userID int64, state *UserState) {
	hm.userStates.Store(userID, state)
}

// 清除用户状态
func (hm *HandlerManager) clearUserState(userID int64) {
	hm.userStates.Delete(userID)
}

// handleTextMessage 处理文本消息
func (hm *HandlerManager) handleTextMessage(chatID, userID int64, text string) {
	state := hm.getUserState(userID)

	if state.CurrentCommand == "" {
		// 创建带菜单按钮的消息
		msg := tgbotapi.NewMessage(chatID, "请选择一个功能开始使用")
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📋 功能菜单", "cmd_menu"),
				tgbotapi.NewInlineKeyboardButtonData("❓ 帮助", "cmd_help"),
			),
		)
		msg.ReplyMarkup = keyboard
		hm.bot.Send(msg)
		return
	}

	// 根据当前命令状态处理文本
	// 这里可以扩展处理用户输入的参数等
}

// handleDocument 处理文档
func (hm *HandlerManager) handleDocument(chatID, userID int64, document *tgbotapi.Document) {
	state := hm.getUserState(userID)

	if state.CurrentCommand == "" {
		msg := tgbotapi.NewMessage(chatID, "请先选择要使用的功能，然后再上传文件")
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📋 功能菜单", "cmd_menu"),
				tgbotapi.NewInlineKeyboardButtonData("❓ 帮助", "cmd_help"),
			),
		)
		msg.ReplyMarkup = keyboard
		hm.bot.Send(msg)
		return
	}

	// 检查文件大小
	if int64(document.FileSize) > hm.config.MaxFileSize {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("文件过大！最大支持 %s\n\n请重新选择功能或上传较小的文件：", utils.FormatFileSize(hm.config.MaxFileSize)))
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📋 返回菜单", "cmd_menu"),
			),
		)
		msg.ReplyMarkup = keyboard
		hm.bot.Send(msg)
		return
	}

	// 根据当前命令处理文件
	hm.processUploadedFile(chatID, userID, document, state)
}

// sendStatusMessage 发送状态信息
func (hm *HandlerManager) sendStatusMessage(chatID, userID int64) {
	state := hm.getUserState(userID)

	statusText := "📊 *当前状态*\n\n"
	if state.CurrentCommand == "" {
		statusText += "• 当前没有进行中的任务\n"
		statusText += "• 输入 /start 开始使用功能"
	} else {
		statusText += fmt.Sprintf("• 当前功能: %s\n", state.CurrentCommand)
		statusText += "• 等待文件上传或处理中..."
	}

	msg := tgbotapi.NewMessage(chatID, statusText)
	msg.ParseMode = "Markdown"
	hm.bot.Send(msg)
}

// sendMenuMessage 发送功能菜单
func (hm *HandlerManager) sendMenuMessage(chatID int64) {
	menuText := `📋 *功能菜单*

请选择您需要的功能：

💡 *快速访问：*
随时输入 /menu 可重新显示此菜单`

	msg := tgbotapi.NewMessage(chatID, menuText)
	msg.ParseMode = "Markdown"

	// 创建内联键盘，复用现有布局
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 日志解析", "cmd_logparse"),
			tgbotapi.NewInlineKeyboardButtonData("🔒 用户锁定", "cmd_lockuser"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗄️ SQL解析", "cmd_sqlparse"),
			tgbotapi.NewInlineKeyboardButtonData("✂️ 文件分割", "cmd_filesplit"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 KYC审核", "cmd_kycreview"),
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Redis流水清零命令", "cmd_redisdel"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Redis增加", "cmd_redisadd"),
			tgbotapi.NewInlineKeyboardButtonData("🔄 UID去重", "cmd_uiddedup"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ 帮助", "cmd_help"),
			tgbotapi.NewInlineKeyboardButtonData("📈 状态", "cmd_status"),
		),
	)
	msg.ReplyMarkup = keyboard

	hm.bot.Send(msg)
}
