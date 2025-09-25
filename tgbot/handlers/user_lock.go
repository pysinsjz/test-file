package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"tgbot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// processUserLock 处理用户锁定功能
func (hm *HandlerManager) processUserLock(chatID, userID int64, inputFile string, state *UserState) error {
	// 检查输入文件是否是CSV格式
	if !utils.IsValidFileType(inputFile, []string{".csv"}) {
		return fmt.Errorf("只支持CSV格式的文件")
	}

	// 发送处理开始消息
	progressMsg := tgbotapi.NewMessage(chatID, "🔄 正在读取用户ID...")
	hm.bot.Send(progressMsg)

	// 打开CSV文件
	file, err := hm.fileManager.OpenFile(inputFile)
	if err != nil {
		return fmt.Errorf("打开CSV文件失败: %v", err)
	}
	defer hm.fileManager.CloseFile(inputFile)

	reader := csv.NewReader(file)

	var userIds []string

	// 读取CSV文件，获取第一列的用户ID
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取CSV行失败: %v", err)
		}

		if len(record) > 0 && record[0] != "" {
			userIds = append(userIds, strings.TrimSpace(record[0]))
		}
	}

	if len(userIds) == 0 {
		return fmt.Errorf("没有找到有效的用户ID")
	}

	// 更新进度
	progressMsg = tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ 找到 %d 个用户ID，正在生成命令...", len(userIds)))
	hm.bot.Send(progressMsg)

	// 生成SQL文件
	sqlContent := hm.generateLockUserSQL(userIds)
	sqlFile := filepath.Join(state.UserDir, "lockUser-db_user库.sql")
	err = hm.writeStringToFile(sqlFile, sqlContent)
	if err != nil {
		return fmt.Errorf("写入SQL文件失败: %v", err)
	}

	// 生成Redis命令文件
	redisContent := hm.generateLockUserRedis(userIds)
	redisFile := filepath.Join(state.UserDir, "lockUser-redis_db0.txt")
	err = hm.writeStringToFile(redisFile, redisContent)
	if err != nil {
		return fmt.Errorf("写入Redis命令文件失败: %v", err)
	}

	// 发送SQL文件
	hm.sendResultFile(chatID, sqlFile, fmt.Sprintf("✅ 用户锁定SQL文件生成完成！\n👤 处理了 %d 个用户", len(userIds)))

	// 发送Redis文件
	hm.sendResultFile(chatID, redisFile, fmt.Sprintf("✅ Redis删除命令文件生成完成！\n🗑️ 包含 %d 条删除命令", len(userIds)))

	return nil
}

// generateLockUserSQL 生成用户锁定SQL语句
func (hm *HandlerManager) generateLockUserSQL(userIds []string) string {
	var sqlStatements []string

	for _, userId := range userIds {
		sql := fmt.Sprintf("UPDATE b_user SET `status` = -1,status_remark = '2025/Sep/25 Multiple Accounts Bonus Hunter, KYC script application, do not unlock unless approved by OPS team',updated_at = now() WHERE id = %s and `status` != -1;", userId)
		sqlStatements = append(sqlStatements, sql)
	}

	return strings.Join(sqlStatements, "\n")
}

// generateLockUserRedis 生成用户Redis删除命令
func (hm *HandlerManager) generateLockUserRedis(userIds []string) string {
	var redisCommands []string

	for _, userId := range userIds {
		command := fmt.Sprintf("del %s", userId)
		redisCommands = append(redisCommands, command)
	}

	return strings.Join(redisCommands, "\n")
}

// writeStringToFile 将字符串写入文件
func (hm *HandlerManager) writeStringToFile(filePath, content string) error {
	file, err := hm.fileManager.CreateOutputFile(filePath)
	if err != nil {
		return err
	}
	defer hm.fileManager.CloseFile(filePath)

	_, err = file.WriteString(content)
	return err
}