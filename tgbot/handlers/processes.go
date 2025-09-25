package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// startLogParseProcess 开始日志解析流程
func (hm *HandlerManager) startLogParseProcess(chatID, userID int64) {
	state := &UserState{
		CurrentCommand: "logparse",
		UserDir:        hm.fileManager.CreateUserDir(userID),
		Data:           make(map[string]interface{}),
	}
	hm.setUserState(userID, state)

	msg := tgbotapi.NewMessage(chatID, `📊 *日志解析功能*

请上传TXT格式的日志文件。

🔍 *处理说明：*
• 提取日志中的关键信息
• 生成结构化的CSV数据文件
• 包含用户ID、追踪ID、请求URL等字段

📎 请上传您的日志文件...`)
	msg.ParseMode = "Markdown"
	hm.bot.Send(msg)
}

// startLockUserProcess 开始用户锁定流程
func (hm *HandlerManager) startLockUserProcess(chatID, userID int64) {
	state := &UserState{
		CurrentCommand: "lockuser",
		UserDir:        hm.fileManager.CreateUserDir(userID),
		Data:           make(map[string]interface{}),
	}
	hm.setUserState(userID, state)

	msg := tgbotapi.NewMessage(chatID, `🔒 *用户锁定功能*

请上传包含用户ID的CSV文件。

⚙️ *处理说明：*
• 读取CSV文件第一列的用户ID
• 生成用户锁定的SQL更新语句
• 生成对应的Redis删除命令

📎 请上传您的CSV文件...`)
	msg.ParseMode = "Markdown"
	hm.bot.Send(msg)
}

// startSQLParseProcess 开始SQL解析流程
func (hm *HandlerManager) startSQLParseProcess(chatID, userID int64) {
	state := &UserState{
		CurrentCommand: "sqlparse",
		UserDir:        hm.fileManager.CreateUserDir(userID),
		Data:           make(map[string]interface{}),
	}
	hm.setUserState(userID, state)

	msg := tgbotapi.NewMessage(chatID, `🗄️ *SQL解析功能*

请上传包含SQL信息的TXT日志文件。

🧠 *处理说明：*
• 从日志中提取SQL语句
• 智能去重，保留唯一SQL
• 基于表名、字段、条件进行去重判断

📎 请上传您的日志文件...`)
	msg.ParseMode = "Markdown"
	hm.bot.Send(msg)
}

// startFileSplitProcess 开始文件分割流程
func (hm *HandlerManager) startFileSplitProcess(chatID, userID int64) {
	state := &UserState{
		CurrentCommand: "filesplit",
		UserDir:        hm.fileManager.CreateUserDir(userID),
		Data:           make(map[string]interface{}),
	}
	hm.setUserState(userID, state)

	msg := tgbotapi.NewMessage(chatID, `✂️ *文件分割功能*

请上传需要分割的大文件（支持任意格式）。

📏 *处理说明：*
• 按照每10,000行分割文件
• 保持原文件格式和扩展名
• 生成多个小文件便于处理

📎 请上传您的文件...`)
	msg.ParseMode = "Markdown"
	hm.bot.Send(msg)
}

// startKYCReviewProcess 开始KYC审核流程
func (hm *HandlerManager) startKYCReviewProcess(chatID, userID int64) {
	state := &UserState{
		CurrentCommand: "kycreview",
		UserDir:        hm.fileManager.CreateUserDir(userID),
		Data:           make(map[string]interface{}),
	}
	hm.setUserState(userID, state)

	msg := tgbotapi.NewMessage(chatID, `📋 *KYC审核功能*

请上传Excel或CSV格式的KYC数据文件。

✅ *处理说明：*
• 处理KYC审核通过数据
• 生成审核状态更新的SQL语句
• 按当前日期命名输出文件

📎 请上传您的KYC文件...`)
	msg.ParseMode = "Markdown"
	hm.bot.Send(msg)
}

// startRedisDelProcess 开始Redis删除流程
func (hm *HandlerManager) startRedisDelProcess(chatID, userID int64) {
	state := &UserState{
		CurrentCommand: "redisdel",
		UserDir:        hm.fileManager.CreateUserDir(userID),
		Data:           make(map[string]interface{}),
	}
	hm.setUserState(userID, state)

	msg := tgbotapi.NewMessage(chatID, `🗑️ *Redis删除命令生成*

请上传Excel或CSV格式的用户数据文件。

🔧 *处理说明：*
• 读取文件中的用户ID
• 为每个用户生成两条Redis删除命令
• 删除流水要求和投注流水数据

📎 请上传您的数据文件...`)
	msg.ParseMode = "Markdown"
	hm.bot.Send(msg)
}

// startRedisAddProcess 开始Redis增加流程
func (hm *HandlerManager) startRedisAddProcess(chatID, userID int64) {
	state := &UserState{
		CurrentCommand: "redisadd",
		UserDir:        hm.fileManager.CreateUserDir(userID),
		Data:           make(map[string]interface{}),
	}
	hm.setUserState(userID, state)

	msg := tgbotapi.NewMessage(chatID, `➕ *Redis流水增加命令生成*

请上传包含流水比例数据的CSV文件。

📊 *文件格式要求：*
• 第1列：用户ID
• 第2列：调整金额
• 第3列：流水比例
• 第5列：投注金额

📎 请上传您的CSV文件...`)
	msg.ParseMode = "Markdown"
	hm.bot.Send(msg)
}

// startUIDDedupProcess 开始UID去重流程
func (hm *HandlerManager) startUIDDedupProcess(chatID, userID int64) {
	state := &UserState{
		CurrentCommand: "uiddedup",
		UserDir:        hm.fileManager.CreateUserDir(userID),
		Data:           make(map[string]interface{}),
	}
	hm.setUserState(userID, state)

	msg := tgbotapi.NewMessage(chatID, `🔄 *UID去重功能*

请上传包含用户ID的CSV文件。

🎯 *处理说明：*
• 统计每个UID的出现次数
• 只保留唯一出现的UID
• 生成去重报告和清理后的文件

📎 请上传您的UID文件...`)
	msg.ParseMode = "Markdown"
	hm.bot.Send(msg)
}

// processUploadedFile 处理上传的文件
func (hm *HandlerManager) processUploadedFile(chatID, userID int64, document *tgbotapi.Document, state *UserState) {
	// 发送处理开始消息
	processingMsg := tgbotapi.NewMessage(chatID, "📥 正在下载文件...")
	sentMsg, _ := hm.bot.Send(processingMsg)

	// 下载文件
	fileConfig := tgbotapi.FileConfig{FileID: document.FileID}
	file, err := hm.bot.GetFile(fileConfig)
	if err != nil {
		hm.updateMessage(chatID, sentMsg.MessageID, "❌ 下载文件失败: "+err.Error())
		hm.clearUserState(userID)
		return
	}

	// 获取文件URL并下载
	fileURL := file.Link(hm.bot.Token)
	localFilePath := fmt.Sprintf("%s/%s", state.UserDir, document.FileName)

	// 这里需要实现文件下载逻辑
	err = hm.downloadFile(fileURL, localFilePath)
	if err != nil {
		hm.updateMessage(chatID, sentMsg.MessageID, "❌ 保存文件失败: "+err.Error())
		hm.clearUserState(userID)
		return
	}

	// 更新消息为处理中
	hm.updateMessage(chatID, sentMsg.MessageID, "⚙️ 正在处理文件，请稍等...")

	// 根据命令类型处理文件
	go func() {
		defer func() {
			if r := recover(); r != nil {
				hm.updateMessage(chatID, sentMsg.MessageID, fmt.Sprintf("❌ 处理过程中发生错误: %v", r))
			}
			// 清理用户目录
			hm.fileManager.CleanupUserDir(state.UserDir)
			hm.clearUserState(userID)
		}()

		var err error
		switch state.CurrentCommand {
		case "logparse":
			err = hm.processLogParse(chatID, userID, localFilePath, state)
		case "lockuser":
			err = hm.processLockUser(chatID, userID, localFilePath, state)
		case "sqlparse":
			err = hm.processSQLParse(chatID, userID, localFilePath, state)
		case "filesplit":
			err = hm.processFileSplit(chatID, userID, localFilePath, state)
		case "kycreview":
			err = hm.processKYCReview(chatID, userID, localFilePath, state)
		case "redisdel":
			err = hm.processRedisDel(chatID, userID, localFilePath, state)
		case "redisadd":
			err = hm.processRedisAdd(chatID, userID, localFilePath, state)
		case "uiddedup":
			err = hm.processUIDDedup(chatID, userID, localFilePath, state)
		default:
			err = fmt.Errorf("未知的命令类型: %s", state.CurrentCommand)
		}

		if err != nil {
			hm.updateMessage(chatID, sentMsg.MessageID, "❌ 处理失败: "+err.Error())
		}
	}()
}

// updateMessage 更新消息内容
func (hm *HandlerManager) updateMessage(chatID int64, messageID int, text string) {
	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	editMsg.ParseMode = "Markdown"
	hm.bot.Send(editMsg)
}

// downloadFile 下载文件的辅助函数
func (hm *HandlerManager) downloadFile(url, localPath string) error {
	// 使用http包下载文件
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP状态错误: %d", resp.StatusCode)
	}

	// 创建本地文件
	out, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("创建本地文件失败: %v", err)
	}
	defer out.Close()

	// 复制文件内容
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}

	return nil
}

// 以下是各个处理功能的占位符实现，将逐步完善

// processLockUser 处理用户锁定
func (hm *HandlerManager) processLockUser(chatID, userID int64, inputFile string, state *UserState) error {
	return hm.processUserLock(chatID, userID, inputFile, state)
}

// processSQLParse 处理SQL解析
func (hm *HandlerManager) processSQLParse(chatID, userID int64, inputFile string, state *UserState) error {
	return fmt.Errorf("SQL解析功能待实现")
}

// processFileSplit 处理文件分割
func (hm *HandlerManager) processFileSplit(chatID, userID int64, inputFile string, state *UserState) error {
	return fmt.Errorf("文件分割功能待实现")
}

// processKYCReview 处理KYC审核
func (hm *HandlerManager) processKYCReview(chatID, userID int64, inputFile string, state *UserState) error {
	return fmt.Errorf("KYC审核功能待实现")
}

// processRedisDel 处理Redis删除命令生成
func (hm *HandlerManager) processRedisDel(chatID, userID int64, inputFile string, state *UserState) error {
	return fmt.Errorf("Redis删除命令生成功能待实现")
}

// processRedisAdd 处理Redis增加命令生成
func (hm *HandlerManager) processRedisAdd(chatID, userID int64, inputFile string, state *UserState) error {
	return fmt.Errorf("Redis增加命令生成功能待实现")
}

// processUIDDedup 处理UID去重
func (hm *HandlerManager) processUIDDedup(chatID, userID int64, inputFile string, state *UserState) error {
	return fmt.Errorf("UID去重功能待实现")
}