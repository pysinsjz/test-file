package processor

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"
)

// getCurrentDateString 获取当前日期字符串
func getCurrentDateString() string {
	return time.Now().Format("2006-01-02")
}

// processLogFile 处理日志文件的具体实现
func processLogFile(inputFile, outputFile string, callback ProgressCallback) error {
	callback(20, "打开日志文件...")

	// 打开输入文件
	inFile, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("打开输入文件失败: %v", err)
	}
	defer inFile.Close()

	// 创建输出文件
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer outFile.Close()

	callback(30, "开始解析日志...")

	// 创建CSV写入器
	writer := csv.NewWriter(outFile)
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

	scanner := bufio.NewScanner(inFile)
	lineNum := 0
	processedLines := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// 解析日志行
		row := parseLogLine(line)
		if hasValidData(row) {
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("写入CSV行失败: %v", err)
			}
			processedLines++
		}

		// 更新进度
		if lineNum%1000 == 0 {
			progress := 30 + (lineNum*60/10000) // 假设最多10000行
			if progress > 90 {
				progress = 90
			}
			callback(progress, fmt.Sprintf("已处理 %d 行，有效数据 %d 条", lineNum, processedLines))
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件时发生错误: %v", err)
	}

	callback(95, fmt.Sprintf("日志解析完成，总计处理 %d 行，提取有效数据 %d 条", lineNum, processedLines))
	return nil
}

// parseLogLine 解析单行日志 (复用 tgbot 的逻辑)
func parseLogLine(logStr string) []string {
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
func hasValidData(row []string) bool {
	for _, field := range row {
		if strings.TrimSpace(field) != "" {
			return true
		}
	}
	return false
}

// processLockUserFile 处理用户锁定文件
func processLockUserFile(inputFile, sqlFile, redisFile string, callback ProgressCallback) error {
	callback(20, "读取用户ID列表...")

	// 读取CSV文件
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("打开CSV文件失败: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var userIds []string

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("读取CSV失败: %v", err)
	}

	// 提取第一列的用户ID
	for _, record := range records {
		if len(record) > 0 && record[0] != "" {
			userIds = append(userIds, strings.TrimSpace(record[0]))
		}
	}

	if len(userIds) == 0 {
		return fmt.Errorf("未找到有效的用户ID")
	}

	callback(40, fmt.Sprintf("找到 %d 个用户ID，生成SQL语句...", len(userIds)))

	// 生成SQL文件
	sqlContent := generateLockUserSQL(userIds)
	err = os.WriteFile(sqlFile, []byte(sqlContent), 0644)
	if err != nil {
		return fmt.Errorf("写入SQL文件失败: %v", err)
	}

	callback(70, "生成Redis命令...")

	// 生成Redis命令文件
	redisContent := generateLockUserRedis(userIds)
	err = os.WriteFile(redisFile, []byte(redisContent), 0644)
	if err != nil {
		return fmt.Errorf("写入Redis文件失败: %v", err)
	}

	callback(95, "用户锁定处理完成")
	return nil
}

// generateLockUserSQL 生成用户锁定SQL语句
func generateLockUserSQL(userIds []string) string {
	var sqlStatements []string

	for _, userId := range userIds {
		sql := fmt.Sprintf("UPDATE b_user SET `status` = -1,status_remark = '%s Multiple Accounts Bonus Hunter, KYC script application, do not unlock unless approved by OPS team',updated_at = now() WHERE id = %s and `status` != -1;",
			time.Now().Format("2006/Jan/02"), userId)
		sqlStatements = append(sqlStatements, sql)
	}

	return strings.Join(sqlStatements, "\n")
}

// generateLockUserRedis 生成用户Redis删除命令
func generateLockUserRedis(userIds []string) string {
	var redisCommands []string

	for _, userId := range userIds {
		command := fmt.Sprintf("del %s", userId)
		redisCommands = append(redisCommands, command)
	}

	return strings.Join(redisCommands, "\n")
}

// 其他处理函数的占位符实现
func processSQLFile(inputFile, outputFile string, callback ProgressCallback) error {
	callback(50, "SQL解析处理中...")
	// TODO: 实现SQL解析逻辑
	return fmt.Errorf("SQL解析功能正在开发中")
}

func processFileSplitLogic(inputFile, outputDir string, callback ProgressCallback) ([]string, error) {
	callback(50, "文件分割处理中...")
	// TODO: 实现文件分割逻辑
	return nil, fmt.Errorf("文件分割功能正在开发中")
}

func processKYCFile(inputFile, outputFile string, callback ProgressCallback) error {
	callback(50, "KYC审核处理中...")
	// TODO: 实现KYC审核逻辑
	return fmt.Errorf("KYC审核功能正在开发中")
}

func processRedisDelLogic(inputFile, outputDir string, callback ProgressCallback) ([]string, error) {
	callback(50, "Redis删除命令生成中...")
	// TODO: 实现Redis删除逻辑
	return nil, fmt.Errorf("Redis删除功能正在开发中")
}

func processRedisAddLogic(inputFile, outputFile string, callback ProgressCallback) error {
	callback(50, "Redis增加命令生成中...")
	// TODO: 实现Redis增加逻辑
	return fmt.Errorf("Redis增加功能正在开发中")
}

func processUIDDedupLogic(inputFile, outputFile, reportFile string, callback ProgressCallback) error {
	callback(50, "UID去重处理中...")
	// TODO: 实现UID去重逻辑
	return fmt.Errorf("UID去重功能正在开发中")
}