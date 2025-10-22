package processor

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
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

// processSQLFile 处理SQL文件的具体实现
func processSQLFile(inputFile, outputFile string, callback ProgressCallback) error {
	callback(20, "打开SQL日志文件...")

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

	callback(30, "开始解析SQL语句...")

	var sqlCount int
	uniqueSQLs := make(map[string]bool) // 用于去重的map

	scanner := bufio.NewScanner(inFile)
	// 增加缓冲区大小以处理超长的行
	buf := make([]byte, 0, 1024*1024) // 1MB 缓冲区
	scanner.Buffer(buf, 1024*1024)    // 最大 1MB
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// 查找包含 SQL 信息的行
		if strings.Contains(line, `"sql_INFO":"`) {
			// 提取 SQL 语句
			sqlInfoPrefix := `"sql_INFO":"`
			sqlStart := strings.Index(line, sqlInfoPrefix)
			if sqlStart >= 0 {
				sqlStart += len(sqlInfoPrefix)
				sqlEnd := strings.Index(line[sqlStart:], `"`)
				if sqlEnd > 0 {
					sqlStatement := line[sqlStart : sqlStart+sqlEnd]
					// 解码转义字符
					sqlStatement = strings.ReplaceAll(sqlStatement, `\"`, `"`)
					sqlStatement = strings.ReplaceAll(sqlStatement, `\\`, `\`)

					// 生成SQL的唯一标识（表名、字段、where条件）
					sqlKey := generateSQLKey(sqlStatement)

					// 检查是否已经存在相同的SQL
					if !uniqueSQLs[sqlKey] {
						uniqueSQLs[sqlKey] = true

						// 写入输出文件
						outputLine := fmt.Sprintf("%s\n", sqlStatement)
						_, err := outFile.WriteString(outputLine)
						if err != nil {
							return fmt.Errorf("写入输出文件失败: %v", err)
						}
						sqlCount++
					}
				}
			}
		}

		// 更新进度
		if lineNum%5000 == 0 {
			progress := 30 + (lineNum*60/100000) // 假设最多100000行
			if progress > 90 {
				progress = 90
			}
			callback(progress, fmt.Sprintf("已处理 %d 行，提取 %d 条唯一SQL...", lineNum, sqlCount))
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件时发生错误: %v", err)
	}

	// 清理内存
	uniqueSQLs = nil

	callback(95, fmt.Sprintf("SQL解析完成！总计处理 %d 行日志，提取 %d 条唯一SQL语句", lineNum, sqlCount))
	return nil
}

func processFileSplitLogic(inputFile, outputDir string, callback ProgressCallback) ([]string, error) {
	callback(20, "开始分析文件...")

	// 打开输入文件
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// 增加缓冲区大小以处理超长的行
	buf := make([]byte, 0, 1024*1024) // 1MB 缓冲区
	scanner.Buffer(buf, 1024*1024)    // 最大 1MB

	var currentFileIndex int = 1
	var currentLineCount int = 0
	var currentOutputFile *os.File
	var outputFiles []string

	// 获取原文件名（不含路径）
	baseFileName := filepath.Base(inputFile)
	// 获取原文件扩展名
	fileExt := filepath.Ext(baseFileName)
	// 去掉原文件扩展名
	nameWithoutExt := strings.TrimSuffix(baseFileName, fileExt)

	callback(30, "开始文件分割...")

	// 创建第一个输出文件
	outputFileName := filepath.Join(outputDir, fmt.Sprintf("%s_part_%04d%s", nameWithoutExt, currentFileIndex, fileExt))
	currentOutputFile, err = os.Create(outputFileName)
	if err != nil {
		return nil, fmt.Errorf("创建输出文件失败: %v", err)
	}
	outputFiles = append(outputFiles, outputFileName)

	totalLines := 0

	// 逐行读取并写入
	for scanner.Scan() {
		line := scanner.Text()
		totalLines++
		currentLineCount++

		// 写入当前输出文件
		_, err := currentOutputFile.WriteString(line + "\n")
		if err != nil {
			currentOutputFile.Close()
			return nil, fmt.Errorf("写入文件失败: %v", err)
		}

		// 如果当前文件已达到10000行，创建新文件
		if currentLineCount >= 10000 {
			currentOutputFile.Close()
			currentFileIndex++
			currentLineCount = 0

			// 创建新的输出文件
			outputFileName = filepath.Join(outputDir, fmt.Sprintf("%s_part_%04d%s", nameWithoutExt, currentFileIndex, fileExt))
			currentOutputFile, err = os.Create(outputFileName)
			if err != nil {
				return nil, fmt.Errorf("创建输出文件失败: %v", err)
			}
			outputFiles = append(outputFiles, outputFileName)

			// 更新进度
			progress := 30 + (totalLines*60/100000) // 假设最多100000行
			if progress > 90 {
				progress = 90
			}
			callback(progress, fmt.Sprintf("正在创建第 %d 个分割文件，已处理 %d 行...", currentFileIndex, totalLines))
		}

		// 定期更新进度
		if totalLines%1000 == 0 {
			progress := 30 + (totalLines*60/100000)
			if progress > 90 {
				progress = 90
			}
			callback(progress, fmt.Sprintf("已处理 %d 行数据...", totalLines))
		}
	}

	// 关闭最后一个输出文件
	if currentOutputFile != nil {
		currentOutputFile.Close()
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件时发生错误: %v", err)
	}

	callback(95, fmt.Sprintf("文件分割完成！总计 %d 行数据，分割为 %d 个文件", totalLines, currentFileIndex))
	return outputFiles, nil
}

func processKYCFile(inputFile, outputFile string, callback ProgressCallback) error {
	callback(20, "开始处理KYC审核数据...")

	// 检查文件格式
	ext := strings.ToLower(filepath.Ext(inputFile))
	if ext != ".xlsx" && ext != ".csv" {
		return fmt.Errorf("只支持Excel (.xlsx) 或CSV格式的文件")
	}

	// 创建输出文件
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer outFile.Close()

	callback(30, "正在读取文件数据...")

	var sqlCount int
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	if ext == ".xlsx" {
		// 处理Excel文件
		err = processKYCExcelFile(inputFile, outFile, currentTime, &sqlCount, callback)
	} else if ext == ".csv" {
		// 处理CSV文件
		err = processKYCCSVFile(inputFile, outFile, currentTime, &sqlCount, callback)
	}

	if err != nil {
		return fmt.Errorf("处理文件失败: %v", err)
	}

	callback(95, fmt.Sprintf("KYC审核处理完成！共生成 %d 条SQL语句", sqlCount))
	return nil
}

// processKYCExcelFile 处理Excel格式的KYC文件
func processKYCExcelFile(inputFile string, outFile *os.File, currentTime string, sqlCount *int, callback ProgressCallback) error {
	f, err := excelize.OpenFile(inputFile)
	if err != nil {
		return fmt.Errorf("打开Excel文件失败: %v", err)
	}
	defer f.Close()

	// 获取第一个工作表的名称
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return fmt.Errorf("获取工作表失败")
	}

	// 获取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("读取工作表数据失败: %v", err)
	}

	callback(50, "正在生成SQL语句...")

	// 跳过标题行，处理数据行
	for i, row := range rows {
		if i == 0 {
			continue // 跳过标题行
		}

		// 确保行有足够的列数据
		if len(row) >= 2 {
			// 假设第1列是 user_id，第2列是 id
			userId := strings.TrimSpace(row[0])
			recordId := strings.TrimSpace(row[1])

			// 生成 SQL 语句
			if userId != "" && recordId != "" {
				sql := fmt.Sprintf("UPDATE b_kyc set audit_status = 1,audit_at = '%s' where audit_status = 2 and is_lock = 0 and user_id = %s and id = %s;\n",
					currentTime, userId, recordId)

				_, err := outFile.WriteString(sql)
				if err != nil {
					return fmt.Errorf("写入SQL语句失败: %v", err)
				}
				*sqlCount++
			}
		}

		// 每处理1000行更新进度
		if i%1000 == 0 && i > 0 {
			progress := 50 + (i*40/len(rows))
			if progress > 90 {
				progress = 90
			}
			callback(progress, fmt.Sprintf("已处理 %d 行KYC数据，生成 %d 条SQL...", i, *sqlCount))
		}
	}

	return nil
}

// processKYCCSVFile 处理CSV格式的KYC文件
func processKYCCSVFile(inputFile string, outFile *os.File, currentTime string, sqlCount *int, callback ProgressCallback) error {
	csvFile, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("打开CSV文件失败: %v", err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("读取CSV数据失败: %v", err)
	}

	callback(50, "正在生成SQL语句...")

	// 跳过标题行，处理数据行
	for i, record := range records {
		if i == 0 {
			continue // 跳过标题行
		}

		// 确保行有足够的列数据
		if len(record) >= 2 {
			// 假设第1列是 user_id，第2列是 id
			userId := strings.TrimSpace(record[0])
			recordId := strings.TrimSpace(record[1])

			// 生成 SQL 语句
			if userId != "" && recordId != "" {
				sql := fmt.Sprintf("UPDATE b_kyc set audit_status = 1,audit_at = '%s' where audit_status = 2 and is_lock = 0 and user_id = %s and id = %s;\n",
					currentTime, userId, recordId)

				_, err := outFile.WriteString(sql)
				if err != nil {
					return fmt.Errorf("写入SQL语句失败: %v", err)
				}
				*sqlCount++
			}
		}

		// 每处理1000行更新进度
		if i%1000 == 0 && i > 0 {
			progress := 50 + (i*40/len(records))
			if progress > 90 {
				progress = 90
			}
			callback(progress, fmt.Sprintf("已处理 %d 行KYC数据，生成 %d 条SQL...", i, *sqlCount))
		}
	}

	return nil
}

func processRedisDelLogic(inputFile, outputDir string, callback ProgressCallback) ([]string, error) {
	callback(10, "开始Redis删除命令生成流程...")

	// 检查文件格式
	ext := strings.ToLower(filepath.Ext(inputFile))
	if ext != ".xlsx" && ext != ".csv" {
		return nil, fmt.Errorf("只支持Excel (.xlsx) 或CSV格式的文件")
	}

	// 步骤1：生成Redis删除命令
	callback(20, "步骤1：生成Redis删除命令...")
	redisCommandsFile := filepath.Join(outputDir, "redis_commands.txt")
	outFile, err := os.Create(redisCommandsFile)
	if err != nil {
		return nil, fmt.Errorf("创建Redis命令文件失败: %v", err)
	}
	defer outFile.Close()

	callback(30, "正在读取用户数据...")

	var totalCount int

	if ext == ".xlsx" {
		// 处理Excel文件
		err = processRedisDelExcelFile(inputFile, outFile, &totalCount, callback)
	} else if ext == ".csv" {
		// 处理CSV文件
		err = processRedisDelCSVFile(inputFile, outFile, &totalCount, callback)
	}

	if err != nil {
		return nil, fmt.Errorf("生成Redis命令失败: %v", err)
	}

	outFile.Close() // 确保文件关闭

	callback(60, fmt.Sprintf("步骤1完成：成功生成 %d 条Redis命令", totalCount*2))

	// 步骤2：分割Redis命令文件
	callback(65, "步骤2：分割Redis命令文件（每10,000行一个文件）...")

	// 创建分割目录
	splitDir := filepath.Join(outputDir, "redis-split")
	if err := os.MkdirAll(splitDir, 0755); err != nil {
		return nil, fmt.Errorf("创建分割目录失败: %v", err)
	}

	// 分割文件
	splitFiles, err := splitRedisCommandFile(redisCommandsFile, splitDir, callback)
	if err != nil {
		return nil, fmt.Errorf("分割Redis命令文件失败: %v", err)
	}

	callback(80, fmt.Sprintf("步骤2完成：文件分割为 %d 个部分", len(splitFiles)))

	// 步骤3：复制执行脚本
	callback(85, "步骤3：复制execute_redis_commands.sh脚本...")

	executeScript := "execute_redis_commands.sh"
	scriptSrc := filepath.Join(".", executeScript)
	scriptDst := filepath.Join(splitDir, executeScript)

	err = copyFile(scriptSrc, scriptDst)
	if err != nil {
		return nil, fmt.Errorf("复制执行脚本失败: %v", err)
	}

	// 设置脚本执行权限
	if err := os.Chmod(scriptDst, 0755); err != nil {
		return nil, fmt.Errorf("设置脚本权限失败: %v", err)
	}

	callback(90, "步骤3完成：成功复制执行脚本")

	// 步骤4：压缩分割目录
	callback(92, "步骤4：压缩redis-split文件夹...")

	zipFile := filepath.Join(outputDir, "redis-split.zip")
	err = createZipFile(splitDir, zipFile, callback)
	if err != nil {
		return nil, fmt.Errorf("压缩文件夹失败: %v", err)
	}

	callback(95, "步骤4完成：成功压缩redis-split文件夹")

	// 返回压缩文件路径 - 需要相对于uploads目录的路径
	taskID := filepath.Base(filepath.Dir(outputDir))
	relativeZipPath := filepath.Join(taskID, "output", "redis-split.zip")

	callback(100, fmt.Sprintf("🎉 所有步骤执行完成！处理了 %d 个用户，生成了 %d 条Redis命令，分割为 %d 个文件", totalCount, totalCount*2, len(splitFiles)))

	return []string{relativeZipPath}, nil
}

// processRedisDelExcelFile 处理Excel格式的Redis删除文件
func processRedisDelExcelFile(inputFile string, outFile *os.File, totalCount *int, callback ProgressCallback) error {
	f, err := excelize.OpenFile(inputFile)
	if err != nil {
		return fmt.Errorf("打开Excel文件失败: %v", err)
	}
	defer f.Close()

	// 获取所有工作表名称
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return fmt.Errorf("Excel文件中没有工作表")
	}

	// 使用第一个工作表
	sheetName := sheetList[0]

	// 获取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("读取工作表数据失败: %v", err)
	}

	callback(50, "正在生成Redis删除命令...")

	// 遍历所有行
	for rowIndex, row := range rows {
		// 跳过空行
		if len(row) == 0 {
			continue
		}

		// 获取第一列的值作为用户ID
		var userID string
		if len(row) > 0 {
			userID = strings.TrimSpace(row[0])
		}

		// 跳过空的用户ID
		if userID == "" {
			continue
		}

		// 跳过表头（如果第一行是表头）
		if rowIndex == 0 && !isNumeric(userID) {
			continue
		}

		// 生成两个Redis删除命令
		reqCmd := fmt.Sprintf("del risk:turnover:req:{%s}\n", userID)
		betCmd := fmt.Sprintf("del risk:turnover:bet:{%s}\n", userID)

		_, err := outFile.WriteString(reqCmd)
		if err != nil {
			return fmt.Errorf("写入Redis命令失败: %v", err)
		}

		_, err = outFile.WriteString(betCmd)
		if err != nil {
			return fmt.Errorf("写入Redis命令失败: %v", err)
		}

		*totalCount++

		// 每处理1000行更新进度
		if *totalCount%1000 == 0 {
			progress := 50 + (*totalCount*40/len(rows))
			if progress > 90 {
				progress = 90
			}
			callback(progress, fmt.Sprintf("已处理 %d 个用户ID，生成 %d 条Redis命令...", *totalCount, *totalCount*2))
		}
	}

	return nil
}

// processRedisDelCSVFile 处理CSV格式的Redis删除文件
func processRedisDelCSVFile(inputFile string, outFile *os.File, totalCount *int, callback ProgressCallback) error {
	csvFile, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("打开CSV文件失败: %v", err)
	}
	defer csvFile.Close()

	callback(50, "正在生成Redis删除命令...")

	// 创建扫描器逐行读取
	scanner := bufio.NewScanner(csvFile)
	rowIndex := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行
		if line == "" {
			continue
		}

		// 解析CSV行，只取第一列
		fields := strings.Split(line, ",")
		var userID string
		if len(fields) > 0 {
			userID = strings.TrimSpace(fields[0])
		}

		// 跳过空的用户ID
		if userID == "" {
			continue
		}

		// 跳过表头（如果第一行是表头）
		if rowIndex == 0 && !isNumeric(userID) {
			rowIndex++
			continue
		}

		// 生成两个Redis删除命令
		reqCmd := fmt.Sprintf("del risk:turnover:req:{%s}\n", userID)
		betCmd := fmt.Sprintf("del risk:turnover:bet:{%s}\n", userID)

		_, err := outFile.WriteString(reqCmd)
		if err != nil {
			return fmt.Errorf("写入Redis命令失败: %v", err)
		}

		_, err = outFile.WriteString(betCmd)
		if err != nil {
			return fmt.Errorf("写入Redis命令失败: %v", err)
		}

		*totalCount++

		// 每处理1000行更新进度
		if *totalCount%1000 == 0 {
			progress := 50 + (*totalCount*40/10000) // 假设10000行
			if progress > 90 {
				progress = 90
			}
			callback(progress, fmt.Sprintf("已处理 %d 个用户ID，生成 %d 条Redis命令...", *totalCount, *totalCount*2))
		}

		rowIndex++
	}

	// 检查扫描过程中是否有错误
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件时发生错误: %v", err)
	}

	return nil
}

// isNumeric 检查字符串是否为数字
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func processRedisAddLogic(inputFile, outputFile string, callback ProgressCallback) error {
	callback(20, "开始Redis增加命令生成...")

	// 检查文件格式
	ext := strings.ToLower(filepath.Ext(inputFile))
	if ext != ".csv" {
		return fmt.Errorf("只支持CSV格式的文件")
	}

	// 创建输出文件
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer outFile.Close()

	callback(30, "正在读取CSV数据...")

	// 读取CSV文件
	csvFile, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("打开CSV文件失败: %v", err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("读取CSV数据失败: %v", err)
	}

	callback(50, "正在生成Redis命令...")

	totalCount := 0

	// 跳过标题行，处理数据行
	for i, record := range records {
		if i == 0 {
			continue // 跳过标题行
		}

		if len(record) < 4 {
			continue // 确保有足够的列
		}

		// 解析数据
		userID := strings.TrimSpace(record[0])
		adjustAmountStr := strings.TrimSpace(record[1])
		turnoverRatioStr := strings.TrimSpace(record[2])
		betAmountStr := ""
		if len(record) > 3 {
			betAmountStr = strings.TrimSpace(record[3])
		}

		// 转换数据类型
		adjustAmountFloat, err := strconv.ParseFloat(adjustAmountStr, 64)
		if err != nil {
			continue
		}

		turnoverRatioFloat, err := strconv.ParseFloat(turnoverRatioStr, 64)
		if err != nil {
			continue
		}

		betAmountFloat := 0.0
		if betAmountStr != "" {
			betAmountFloat, _ = strconv.ParseFloat(betAmountStr, 64)
		}

		// 转换为 int 类型
		adjustAmount := int64(adjustAmountFloat)
		turnoverRatio := int64(turnoverRatioFloat)
		betAmount := int64(betAmountFloat)

		// 计算 req 值 (adjust_amount * ratio)
		req := adjustAmount * turnoverRatio

		// 验证数据合法性
		if betAmount*100 > req*100 {
			continue // 跳过不合理的数据
		}

		// 生成三个Redis命令

		// 1. 删除命令
		cmd1 := fmt.Sprintf("del risk:turnover:req:{%s} risk:turnover:bet:{%s}\n", userID, userID)
		_, err = outFile.WriteString(cmd1)
		if err != nil {
			return fmt.Errorf("写入Redis命令失败: %v", err)
		}

		// 2. 设置用户流水要求
		cmd2 := fmt.Sprintf("set risk:turnover:req:{%s} \"{\\\"req\\\":%d,\\\"items\\\":[{\\\"type\\\":\\\"welcome back\\\",\\\"bounds\\\":%d,\\\"ratio\\\":%d}]}\"\n",
			userID, req*100, adjustAmount*100, turnoverRatio)
		_, err = outFile.WriteString(cmd2)
		if err != nil {
			return fmt.Errorf("写入Redis命令失败: %v", err)
		}

		// 3. 设置用户投注流水
		cmd3 := fmt.Sprintf("set risk:turnover:bet:{%s} %d\n", userID, betAmount*100)
		_, err = outFile.WriteString(cmd3)
		if err != nil {
			return fmt.Errorf("写入Redis命令失败: %v", err)
		}

		totalCount++

		// 每处理100条记录更新进度
		if totalCount%100 == 0 {
			progress := 50 + (totalCount*40/len(records))
			if progress > 90 {
				progress = 90
			}
			callback(progress, fmt.Sprintf("已处理 %d 个用户，生成 %d 条命令...", totalCount, totalCount*3))
		}
	}

	callback(95, fmt.Sprintf("Redis增加命令生成完成！共处理 %d 个用户，生成 %d 条Redis命令", totalCount, totalCount*3))
	return nil
}

func processUIDDedupLogic(inputFile, outputFile, reportFile string, callback ProgressCallback) error {
	callback(20, "开始UID去重处理...")

	// 检查文件格式
	ext := strings.ToLower(filepath.Ext(inputFile))
	if ext != ".csv" {
		return fmt.Errorf("只支持CSV格式的文件")
	}

	callback(30, "正在读取和统计UID...")

	// 打开输入文件
	inFile, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("打开输入文件失败: %v", err)
	}
	defer inFile.Close()

	// 统计每个uid的出现次数
	uidCounts := make(map[string]int)

	scanner := bufio.NewScanner(inFile)
	totalLines := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			uidCounts[line]++
			totalLines++

			// 每处理10000行显示一次进度
			if totalLines%10000 == 0 {
				progress := 30 + (totalLines*30/100000) // 假设最多100000行
				if progress > 60 {
					progress = 60
				}
				callback(progress, fmt.Sprintf("已读取 %d 行数据...", totalLines))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件时出错: %v", err)
	}

	callback(70, "正在分析重复情况...")

	// 统计重复情况
	uniqueCount := 0
	duplicateCount := 0
	duplicateExamples := make(map[string]int)
	exampleCount := 0

	for uid, count := range uidCounts {
		if count == 1 {
			uniqueCount++
		} else {
			duplicateCount++
			// 收集前10个重复uid作为示例
			if exampleCount < 10 {
				duplicateExamples[uid] = count
				exampleCount++
			}
		}
	}

	callback(80, fmt.Sprintf("分析完成！总行数: %d, 不同UID: %d, 唯一UID: %d, 重复UID: %d", totalLines, len(uidCounts), uniqueCount, duplicateCount))

	// 创建去重后的输出文件
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer outFile.Close()

	// 写入唯一的uid
	writtenCount := 0
	for uid, count := range uidCounts {
		if count == 1 {
			_, err := outFile.WriteString(uid + "\n")
			if err != nil {
				return fmt.Errorf("写入文件时出错: %v", err)
			}
			writtenCount++
		}
	}

	callback(90, "正在生成去重报告...")

	// 创建去重报告文件
	report, err := os.Create(reportFile)
	if err != nil {
		return fmt.Errorf("创建报告文件失败: %v", err)
	}
	defer report.Close()

	// 写入详细报告
	report.WriteString("UID去重处理报告\n")
	report.WriteString("==================\n\n")
	report.WriteString(fmt.Sprintf("总共读取了 %d 行数据\n", totalLines))
	report.WriteString(fmt.Sprintf("发现 %d 个不同的UID\n", len(uidCounts)))
	report.WriteString(fmt.Sprintf("唯一UID数量: %d\n", uniqueCount))
	report.WriteString(fmt.Sprintf("重复UID数量: %d\n\n", duplicateCount))

	if len(duplicateExamples) > 0 {
		report.WriteString("重复UID示例（前10个）:\n")
		for uid, count := range duplicateExamples {
			report.WriteString(fmt.Sprintf("UID: %s, 出现次数: %d\n", uid, count))
		}
	}

	callback(95, fmt.Sprintf("UID去重完成！成功写入 %d 个唯一UID，原始数据: %d 行，去重后: %d 个唯一UID", writtenCount, totalLines, uniqueCount))
	return nil
}

// SQL解析相关的辅助函数

// generateSQLKey 生成SQL的唯一标识，用于去重
func generateSQLKey(sql string) string {
	// 转换为小写并去除多余空格
	sql = strings.ToLower(strings.TrimSpace(sql))

	// 提取表名
	tableName := extractTableName(sql)

	// 提取字段列表
	fields := extractFields(sql)

	// 提取where条件
	whereCondition := extractWhereCondition(sql)

	// 组合成唯一标识
	return fmt.Sprintf("%s|%s|%s", tableName, fields, whereCondition)
}

// extractTableName 提取表名
func extractTableName(sql string) string {
	// 处理 SELECT 语句
	if strings.HasPrefix(sql, "select") {
		// 查找 FROM 关键字
		fromIndex := strings.Index(sql, " from ")
		if fromIndex > 0 {
			afterFrom := strings.TrimSpace(sql[fromIndex+6:])
			// 查找下一个空格或特殊字符
			endIndex := strings.IndexAny(afterFrom, " \t\n\r")
			if endIndex > 0 {
				return strings.TrimSpace(afterFrom[:endIndex])
			}
			return strings.TrimSpace(afterFrom)
		}
	}
	return ""
}

// extractFields 提取字段列表
func extractFields(sql string) string {
	if strings.HasPrefix(sql, "select") {
		// 查找 FROM 关键字
		fromIndex := strings.Index(sql, " from ")
		if fromIndex > 0 {
			// 提取 SELECT 和 FROM 之间的内容
			selectPart := strings.TrimSpace(sql[6:fromIndex])
			// 去除可能的 DISTINCT 关键字
			selectPart = strings.ReplaceAll(selectPart, "distinct", "")
			return strings.TrimSpace(selectPart)
		}
	}

	return ""
}

// extractWhereCondition 提取WHERE条件
func extractWhereCondition(sql string) string {
	whereIndex := strings.Index(sql, " where ")
	if whereIndex > 0 {
		afterWhere := strings.TrimSpace(sql[whereIndex+7:])
		// 查找可能的 ORDER BY, GROUP BY, LIMIT 等
		orderIndex := strings.Index(afterWhere, " order by ")
		groupIndex := strings.Index(afterWhere, " group by ")
		limitIndex := strings.Index(afterWhere, " limit ")

		// 找到最早出现的结束位置
		endIndex := len(afterWhere)
		if orderIndex > 0 && orderIndex < endIndex {
			endIndex = orderIndex
		}
		if groupIndex > 0 && groupIndex < endIndex {
			endIndex = groupIndex
		}
		if limitIndex > 0 && limitIndex < endIndex {
			endIndex = limitIndex
		}

		whereClause := strings.TrimSpace(afterWhere[:endIndex])

		// 提取字段名，忽略参数值
		return extractFieldNames(whereClause)
	}

	return ""
}

// extractFieldNames 从WHERE条件中提取字段名，忽略参数值
func extractFieldNames(whereClause string) string {
	// 常见的比较操作符
	operators := []string{"=", "!=", "<>", ">", "<", ">=", "<=", "like", "in", "not in", "is", "is not", "between"}

	// 将操作符替换为分隔符，便于分割
	processedClause := strings.ToLower(whereClause)
	for _, op := range operators {
		processedClause = strings.ReplaceAll(processedClause, " "+op+" ", "|")
	}

	// 处理 AND, OR 连接符
	processedClause = strings.ReplaceAll(processedClause, " and ", "|")
	processedClause = strings.ReplaceAll(processedClause, " or ", "|")

	// 分割并提取字段名
	parts := strings.Split(processedClause, "|")
	var fields []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			// 提取字段名（去除可能的表前缀）
			fieldName := extractFieldName(part)
			if fieldName != "" {
				fields = append(fields, fieldName)
			}
		}
	}

	// 去重并排序
	uniqueFields := make(map[string]bool)
	for _, field := range fields {
		uniqueFields[field] = true
	}

	// 转换为切片并排序
	var result []string
	for field := range uniqueFields {
		result = append(result, field)
	}

	// 排序以确保一致性
	sort.Strings(result)

	return strings.Join(result, ",")
}

// extractFieldName 从条件片段中提取字段名
func extractFieldName(condition string) string {
	// 去除可能的括号
	condition = strings.Trim(condition, "()")

	// 去除引号包围的值
	condition = strings.Trim(condition, "'\"")

	// 如果包含点号，取最后一部分（表名.字段名 -> 字段名）
	if strings.Contains(condition, ".") {
		parts := strings.Split(condition, ".")
		if len(parts) > 1 {
			condition = parts[len(parts)-1]
		}
	}

	// 检查是否是有效的字段名（不包含数字开头、特殊字符等）
	if len(condition) > 0 && !strings.ContainsAny(condition, "0123456789") &&
		!strings.ContainsAny(condition, "()[]{}'\"`") {
		return strings.TrimSpace(condition)
	}

	return ""
}

// splitRedisCommandFile 分割Redis命令文件为多个小文件
func splitRedisCommandFile(inputFile, outputDir string, callback ProgressCallback) ([]string, error) {
	// 打开输入文件
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, fmt.Errorf("打开输入文件失败: %v", err)
	}
	defer file.Close()

	var outputFiles []string
	scanner := bufio.NewScanner(file)

	// 增加缓冲区大小以处理超长的行
	buf := make([]byte, 0, 1024*1024) // 1MB 缓冲区
	scanner.Buffer(buf, 1024*1024)    // 最大 1MB

	var currentFileIndex int = 1
	var currentLineCount int = 0
	var currentOutputFile *os.File

	totalLines := 0

	// 创建第一个输出文件
	outputFileName := filepath.Join(outputDir, fmt.Sprintf("redis_commands_part_%04d.txt", currentFileIndex))
	currentOutputFile, err = os.Create(outputFileName)
	if err != nil {
		return nil, fmt.Errorf("创建输出文件失败: %v", err)
	}
	outputFiles = append(outputFiles, outputFileName)

	// 逐行读取并写入
	for scanner.Scan() {
		line := scanner.Text()
		totalLines++
		currentLineCount++

		// 写入当前输出文件
		_, err := currentOutputFile.WriteString(line + "\n")
		if err != nil {
			currentOutputFile.Close()
			return nil, fmt.Errorf("写入文件失败: %v", err)
		}

		// 如果当前文件已达到10000行，创建新文件
		if currentLineCount >= 10000 {
			currentOutputFile.Close()
			currentFileIndex++
			currentLineCount = 0

			// 创建新的输出文件
			outputFileName = filepath.Join(outputDir, fmt.Sprintf("redis_commands_part_%04d.txt", currentFileIndex))
			currentOutputFile, err = os.Create(outputFileName)
			if err != nil {
				return nil, fmt.Errorf("创建输出文件失败: %v", err)
			}
			outputFiles = append(outputFiles, outputFileName)

			// 更新进度 (65% -> 80% 之间)
			progress := 65 + (currentFileIndex*15/100)
			if progress > 80 {
				progress = 80
			}
			callback(progress, fmt.Sprintf("正在创建第 %d 个分割文件，已处理 %d 行...", currentFileIndex, totalLines))
		}
	}

	// 关闭最后一个输出文件
	if currentOutputFile != nil {
		currentOutputFile.Close()
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件时发生错误: %v", err)
	}

	return outputFiles, nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = sourceFile.WriteTo(destFile)
	return err
}

// createZipFile 创建ZIP压缩文件
func createZipFile(sourceDir, zipPath string, callback ProgressCallback) error {
	// 使用绝对路径
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return fmt.Errorf("获取源目录绝对路径失败: %v", err)
	}

	absZipPath, err := filepath.Abs(zipPath)
	if err != nil {
		return fmt.Errorf("获取压缩文件绝对路径失败: %v", err)
	}

	// 切换到源目录的父目录
	parentDir := filepath.Dir(absSourceDir)
	sourceDirName := filepath.Base(absSourceDir)

	// 使用系统zip命令进行压缩
	zipCmd := exec.Command("zip", "-r", absZipPath, sourceDirName)
	zipCmd.Dir = parentDir

	// 执行命令并获取详细错误信息
	output, err := zipCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("压缩文件失败: %v, 输出: %s", err, string(output))
	}

	// 检查压缩文件是否创建成功
	if _, err := os.Stat(absZipPath); os.IsNotExist(err) {
		return fmt.Errorf("压缩文件创建失败，文件不存在: %s", absZipPath)
	}

	return nil
}