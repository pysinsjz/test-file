package handlers

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"strings"
	"tgbot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// processLogParse 处理日志解析
func (hm *HandlerManager) processLogParse(chatID, userID int64, inputFile string, state *UserState) error {
	// 检查输入文件是否是TXT格式
	if !utils.IsValidFileType(inputFile, []string{".txt"}) {
		return fmt.Errorf("只支持TXT格式的日志文件")
	}

	// 创建输出文件
	outputFile := filepath.Join(state.UserDir, "data.csv")
	file, err := hm.fileManager.CreateOutputFile(outputFile)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer hm.fileManager.CloseFile(outputFile)

	// 创建CSV写入器
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入CSV头部
	headers := []string{
		"logTime", "sign", "requestUrl", "userId", "traceId",
		"paySerialNumber", "paySerialNo", "requestReferenceNumber",
		"user_id", "lot_number", "phone", "verifyCode", "userIp",
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("写入CSV头部失败: %v", err)
	}

	// 打开输入文件
	inputFileHandle, err := hm.fileManager.OpenFile(inputFile)
	if err != nil {
		return fmt.Errorf("打开输入文件失败: %v", err)
	}
	defer hm.fileManager.CloseFile(inputFile)

	// 发送处理进度更新 - 我们现在不跟踪具体的消息ID，而是发送新消息
	progressMsg := tgbotapi.NewMessage(chatID, "🔄 正在解析日志文件...")
	hm.bot.Send(progressMsg)

	scanner := bufio.NewScanner(inputFileHandle)
	// 增加缓冲区大小以处理超长的行
	buf := make([]byte, 0, 1024*1024) // 1MB 缓冲区
	scanner.Buffer(buf, 1024*1024)    // 最大 1MB

	lineNum := 0
	processedLines := 0

	for scanner.Scan() {
		lineNum++
		logStr := scanner.Text()

		// 解析日志行
		row := hm.parseLogLine(logStr)

		// 只有当至少有一个字段不为空时才写入
		if hm.hasValidData(row) {
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("写入CSV行失败: %v", err)
			}
			processedLines++
		}

		// 每处理1000行发送一次进度消息
		if lineNum%1000 == 0 {
			progress := fmt.Sprintf("🔄 已处理 %d 行，有效数据 %d 条...", lineNum, processedLines)
			progressMsg := tgbotapi.NewMessage(chatID, progress)
			hm.bot.Send(progressMsg)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件时发生错误: %v", err)
	}

	// 完成处理，发送结果文件
	hm.sendResultFile(chatID, outputFile, fmt.Sprintf("✅ 日志解析完成！\n📊 总计处理 %d 行，提取有效数据 %d 条", lineNum, processedLines))

	return nil
}

// parseLogLine 解析单行日志
func (hm *HandlerManager) parseLogLine(logStr string) []string {
	var sign, requestUrl, logTime, userId, traceId,
		paySerialNumber, paySerialNo, requestReferenceNumber,
		user_id, lot_number, phone,
		verifyCode, userIp string

	// 查找sign的位置
	signStart := strings.Index(logStr, `"sign":["`) + 9
	if signStart > 9 && len(logStr) >= signStart+32 {
		sign = logStr[signStart : signStart+32] // sign固定长度为32
	}

	// 查找verifyCode的位置
	verifyCodeStart := strings.Index(logStr, `"verifyCode":"`) + 14
	if verifyCodeStart > 14 && len(logStr) >= verifyCodeStart+6 {
		verifyCode = logStr[verifyCodeStart : verifyCodeStart+6] // verifyCode固定长度为6
	}

	// 查找requestUrl的位置
	requestUrlStart := strings.Index(logStr, `"requestUrl":"`) + 14
	if requestUrlStart > 14 {
		requestUrlEnd := strings.Index(logStr[requestUrlStart:], `","`)
		if requestUrlEnd > 0 {
			requestUrl = logStr[requestUrlStart : requestUrlStart+requestUrlEnd]
		}
	}

	// 查找logTime的位置 - 取前32个字符
	if len(logStr) >= 32 {
		logTime = logStr[0:32]
	}

	// 查找userId的位置
	userIdStart := strings.Index(logStr, `"userId":"`) + 10
	if userIdStart > 9 && len(logStr) >= userIdStart+8 {
		userId = logStr[userIdStart : userIdStart+8] // userId固定长度为8
	}
	if userId == "" {
		userId = "00000000"
	}

	// 查找user_id的位置
	user_idStart := strings.Index(logStr, `"user_id":`) + 9
	if user_idStart > 8 {
		user_id = logStr[user_idStart : user_idStart+9]
	}

	// 查找lot_number的位置
	lot_numberStart := strings.Index(logStr, `\"lot_number\":\"`) + 16
	if lot_numberStart > 12 && len(logStr) >= lot_numberStart+33 {
		lot_number = logStr[lot_numberStart : lot_numberStart+33]
	}

	// 查找phone的位置
	phoneStart := strings.Index(logStr, `"phone":`) + 8
	if phoneStart > 7 && len(logStr) >= phoneStart+11 {
		phone = logStr[phoneStart : phoneStart+11]
	}

	// 查找traceId的位置
	traceIdStart := strings.Index(logStr, `"traceId":"`) + 11
	if traceIdStart > 10 && len(logStr) >= traceIdStart+36 {
		traceId = logStr[traceIdStart : traceIdStart+36]
	}

	// 查找paySerialNumber的位置
	paySerialNumberStart := strings.Index(logStr, `"paySerialNumber":"`) + 19
	if paySerialNumberStart > 18 && len(logStr) >= paySerialNumberStart+16 {
		paySerialNumber = logStr[paySerialNumberStart : paySerialNumberStart+16]
	}

	// 查找paySerialNo的位置
	paySerialNoStart := strings.Index(logStr, `"paySerialNo":"`) + 15
	if paySerialNoStart > 14 && len(logStr) >= paySerialNoStart+16 {
		paySerialNo = logStr[paySerialNoStart : paySerialNoStart+16]
	}

	// 查找requestReferenceNumber的位置
	requestReferenceNumberStart := strings.Index(logStr, `"requestReferenceNumber":"`) + 26
	if requestReferenceNumberStart > 25 && len(logStr) >= requestReferenceNumberStart+36 {
		requestReferenceNumber = logStr[requestReferenceNumberStart : requestReferenceNumberStart+36]
	} else {
		requestReferenceNumberStart = strings.Index(logStr, `"Request-Reference-No":"`) + 24
		if requestReferenceNumberStart > 23 && len(logStr) >= requestReferenceNumberStart+36 {
			requestReferenceNumber = logStr[requestReferenceNumberStart : requestReferenceNumberStart+36]
		}
	}

	// 创建CSV行数据
	return []string{
		logTime, sign, requestUrl, userId, traceId,
		paySerialNumber, paySerialNo, requestReferenceNumber,
		user_id, lot_number, phone, verifyCode, userIp,
	}
}

// hasValidData 检查行是否包含有效数据
func (hm *HandlerManager) hasValidData(row []string) bool {
	for _, field := range row {
		if strings.TrimSpace(field) != "" {
			return true
		}
	}
	return false
}

// sendResultFile 发送结果文件
func (hm *HandlerManager) sendResultFile(chatID int64, filePath, caption string) {
	// 创建文档消息
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(filePath))
	doc.Caption = caption

	// 发送文件
	_, err := hm.bot.Send(doc)
	if err != nil {
		// 如果发送失败，发送错误消息
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ 文件发送失败: %v\n\n%s", err, caption))
		hm.bot.Send(msg)
	}
}
