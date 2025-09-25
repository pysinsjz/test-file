package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	excelize "github.com/xuri/excelize/v2"
)

// 主函数
func main() {
	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动清理协程
	go func() {
		<-sigChan
		cleanup()
		os.Exit(0)
	}()

	// 确保程序退出时清理资源
	defer cleanup()

	// 如果启动命令 type 参数为 1 或者空执行logTaceParser()
	if len(os.Args) > 1 && os.Args[1] == "1" {
		LogTaceParser()
	}
	// 如果启动命令 type 参数为 2 lockUser()
	if len(os.Args) > 1 && os.Args[1] == "2" {
		lockUser()
	}
	if len(os.Args) > 1 && os.Args[1] == "3" {
		sqlLogParser()
	}
	if len(os.Args) > 1 && os.Args[1] == "4" {
		// 切分 multi-redis包中的文件  1W 行一个文件
		splitMultiRedisFile()
	}
	if len(os.Args) > 1 && os.Args[1] == "5" {
		// KYC审核处理
		kycReviewProcessor()
	}

	// if len(os.Args) > 1 && os.Args[1] == "4" {
	// 	balanceSqlLogParser()
	// }
}

// 全局变量用于跟踪打开的文件
var openFiles []*os.File

// 清理函数
func cleanup() {
	log.Printf("正在清理资源...")
	for _, file := range openFiles {
		if file != nil {
			file.Close()
		}
	}
	openFiles = nil
}

// 注册文件句柄
func registerFile(file *os.File) {
	openFiles = append(openFiles, file)
}

func sqlLogParser() {
	// 查找 sql-log 目录
	sqlLogDir := "sql-log"
	if _, err := os.Stat(sqlLogDir); os.IsNotExist(err) {
		log.Printf("sql-log 目录不存在，尝试创建...")
		err = os.Mkdir(sqlLogDir, 0755)
		if err != nil {
			log.Printf("创建 sql-log 目录失败: %v", err)
			return
		}
		log.Printf("已创建 sql-log 目录，请将日志文件放入该目录")
		return
	}

	// 查找 TXT 文件
	txtFiles, err := filepath.Glob(filepath.Join(sqlLogDir, "*.txt"))
	if err != nil {
		log.Printf("查找 TXT 文件失败: %v", err)
		return
	}

	if len(txtFiles) == 0 {
		log.Printf("sql-log 目录中没有找到 TXT 文件")
		return
	}

	// 创建输出文件
	outputFile, err := os.Create("sql.log")
	if err != nil {
		log.Printf("创建输出文件失败: %v", err)
		return
	}
	defer outputFile.Close()
	registerFile(outputFile)

	var sqlCount int
	uniqueSQLs := make(map[string]bool) // 用于去重的map

	// 遍历每个 TXT 文件
	for _, txtFile := range txtFiles {
		log.Printf("正在处理文件: %s", txtFile)

		file, err := os.Open(txtFile)
		if err != nil {
			log.Printf("打开文件失败: %v", err)
			continue
		}
		registerFile(file)

		scanner := bufio.NewScanner(file)
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
							_, err := outputFile.WriteString(outputLine)
							if err != nil {
								log.Printf("写入输出文件失败: %v", err)
								continue
							}
							sqlCount++
						}
					}
				}
			}
		}

		// 关闭当前文件
		file.Close()

		if err := scanner.Err(); err != nil {
			log.Printf("读取文件时发生错误: %v", err)
		}
	}

	// 清理内存
	uniqueSQLs = nil

	log.Printf("解析完成，共提取 %d 条唯一 SQL 语句，已保存到 sql.log 文件", sqlCount)
}

// 生成SQL的唯一标识，用于去重
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

// 提取表名
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

// 提取字段列表
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

// 提取WHERE条件
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

// 从WHERE条件中提取字段名，忽略参数值
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

// 从条件片段中提取字段名
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

func lockUser() {
	// 读取根目录csv 包中的 .csv 文件
	// 读取第一列中的 userId切片
	// 拼装 sql UPDATE b_user SET `status` = -1,status_remark = '刷子用户，运营 jello、jerry 申请封禁用户，2025-07-16 16:00:00',updated_at = now() WHERE id = ? and `status` != -1;
	// 写入到当前根目录的 lockUser.sql 文件当中
	// 生成删除用户 redis 的命令
	// del ?

	// 查找 csv 目录
	csvDir := "lock-user-csv"
	if _, err := os.Stat(csvDir); os.IsNotExist(err) {
		log.Printf("csv 目录不存在，尝试创建...")
		err = os.Mkdir(csvDir, 0755)
		if err != nil {
			log.Printf("创建 csv 目录失败: %v", err)
			return
		}
		log.Printf("已创建 csv 目录，请将 CSV 文件放入该目录")
		return
	}

	// 查找 CSV 文件
	csvFiles, err := filepath.Glob(filepath.Join(csvDir, "*.csv"))
	if err != nil {
		log.Printf("查找 CSV 文件失败: %v", err)
		return
	}

	if len(csvFiles) == 0 {
		log.Printf("csv 目录中没有找到 CSV 文件")
		return
	}

	// 读取第一个 CSV 文件
	csvFile := csvFiles[0]
	log.Printf("正在处理文件: %s", csvFile)

	file, err := os.Open(csvFile)
	if err != nil {
		log.Printf("打开 CSV 文件失败: %v", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var userIds []string

	// 读取 CSV 文件，获取第一列的用户 ID
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("读取 CSV 行失败: %v", err)
			continue
		}

		if len(record) > 0 && record[0] != "" {
			userIds = append(userIds, strings.TrimSpace(record[0]))
		}
	}

	if len(userIds) == 0 {
		log.Printf("没有找到有效的用户 ID")
		return
	}

	log.Printf("找到 %d 个用户 ID", len(userIds))

	// 生成 SQL 文件
	sqlContent := generateSQL(userIds)
	err = os.WriteFile("lockUser-db_user库.sql", []byte(sqlContent), 0644)
	if err != nil {
		log.Printf("写入 SQL 文件失败: %v", err)
		return
	}

	// 生成 Redis 命令文件
	redisContent := generateRedisCommands(userIds)
	err = os.WriteFile("lockUser-redis_db0.txt", []byte(redisContent), 0644)
	if err != nil {
		log.Printf("写入 Redis 命令文件失败: %v", err)
		return
	}

	log.Printf("已生成 lockUser.sql 和 lockUser_redis.txt 文件")
}

// 生成 SQL 语句
func generateSQL(userIds []string) string {
	var sqlStatements []string

	for _, userId := range userIds {
		sql := fmt.Sprintf("UPDATE b_user SET `status` = -1,status_remark = '2025/Sep/25 Multiple Accounts Bonus Hunter, KYC script application, do not unlock unless approved by OPS team',updated_at = now() WHERE id = %s and `status` != -1;", userId)
		sqlStatements = append(sqlStatements, sql)
	}

	return strings.Join(sqlStatements, "\n")
}

// 生成 Redis 删除命令
func generateRedisCommands(userIds []string) string {
	var redisCommands []string

	for _, userId := range userIds {
		command := fmt.Sprintf("del %s", userId)
		redisCommands = append(redisCommands, command)
	}

	return strings.Join(redisCommands, "\n")
}

func LogTaceParser() {
	// 获取 ./logs 目录下所有的 TXT 文件
	files, err := filepath.Glob("./logs/*.txt")
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		return
	}

	// 删除旧的输出文件
	os.Remove("data.csv")

	// 创建输出文件
	outputFile, err := os.Create("data.csv")
	if err != nil {
		fmt.Printf("创建输出文件失败: %v\n", err)
		return
	}
	defer outputFile.Close()
	registerFile(outputFile)

	// 创建CSV写入器
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// 写入CSV头部
	headers := []string{
		"logTime", "sign", "requestUrl", "userId", "traceId",
		"paySerialNumber", "paySerialNo", "requestReferenceNumber",
		"user_id", "lot_number", "phone", "verifyCode", "userIp",
	}
	if err := writer.Write(headers); err != nil {
		fmt.Printf("写入CSV头部失败: %v\n", err)
		return
	}

	// 遍历每个 TXT 文件
	for _, file := range files {
		// fmt.Printf("处理文件: %s\n", file)
		f, err := os.Open(file)
		if err != nil {
			fmt.Printf("打开文件失败 %s: %v\n", file, err)
			continue
		}

		scanner := bufio.NewScanner(f)
		// 增加缓冲区大小以处理超长的行
		buf := make([]byte, 0, 1024*1024) // 1MB 缓冲区
		scanner.Buffer(buf, 1024*1024)    // 最大 1MB
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			logStr := scanner.Text()
			// fmt.Printf("第 %d 行内容: %s\n", lineNum, logStr)
			// 2025-06-18T20:58:32.084920+08:00 [INFO] /opt/www/pokerapi/app/Http/Middleware/AuthLog.php:75 pid:197 clent_ip:64.226.56.215 api_header:{"x-forwarded-for":["172.71.210.102, 192.112.4.50"],"host":["io.playgame.zone"],"content-length":["66"],"x-request-id":["5164003be271c0aef81315e03d9a7d90"],"x-forwarded-host":["io.playgame.zone"],"x-forwarded-port":["443"],"x-forwarded-proto":["https"],"x-forwarded-scheme":["https"],"x-scheme":["https"],"x-original-forwarded-for":["64.226.56.215"],"cf-ray":["951af56e56b884eb-HKG"],"cf-worker":["playgame.zone"],"accept-encoding":["gzip, br"],"ts":["1750251511"],"true-client-ip":["64.226.56.215"],"cf-visitor":["{\"scheme\":\"https\"}"],"cf-ipcountry":["PH"],"terminal":["16"],"sign":["455e00c850a52054628a02510343ae0d"],"sec-fetch-site":["cross-site"],"sec-fetch-mode":["cors"],"content-type":["application\/json"],"cf-ew-via":["15"],"cdn-loop":["cloudflare; loops=1; subreqs=1"],"accept-language":["en-US,en;q=0.9,ar-AE;q=0.8,ar;q=0.7"],"accept":["*\/*"],"referer":["https:\/\/gcash-spintime.playgame.zone\/"],"user-agent":["Mozilla\/5.0 (Linux; Android 10; K) AppleWebKit\/537.36 (KHTML, like Gecko) SamsungBrowser\/27.0 Chrome\/125.0.0.0 Mobile Safari\/537.36"],"cf-connecting-ip":["64.226.56.215"],"origin":["https:\/\/gcash-spintime.playgame.zone"],"platform":["glife-h5"],"priority":["u=1, i"],"sec-ch-ua":["\"Chromium\";v=\"125\", \"Not.A\/Brand\";v=\"24\", \"Samsung Internet\";v=\"27.0\""],"sec-ch-ua-mobile":["?1"],"sec-ch-ua-platform":["\"Android\""],"sec-fetch-dest":["empty"]} api_params:##{"token":"27657618:1750247701:0:329c92ba398d056f9cf6eb446c2fb15b"} {"requestUrl":"/api/activity/vip-daily-rebate","terminalName":"GCashSpin","userId":"27657618","traceId":"73002223-648a-483d-a623-3cb119baa6b4","swTraceId":"","userIp":"64.226.56.215"}
			var sign, requestUrl, logTime, userId, traceId,
				paySerialNumber, paySerialNo, requestReferenceNumber,
				user_id, lot_number, phone,
				verifyCode, userIp string
			_ = userIp
			_ = verifyCode
			_ = logTime
			_ = sign
			_ = requestUrl
			_ = traceId
			_ = paySerialNumber
			_ = paySerialNo
			_ = requestReferenceNumber
			_ = user_id
			_ = lot_number
			_ = phone
			_ = userId
			_ = user_id
			_ = lot_number
			_ = phone
			_ = logTime
			_ = sign
			_ = requestUrl
			_ = traceId
			_ = paySerialNumber
			_ = paySerialNo
			_ = requestReferenceNumber
			// 查找sign的位置
			signStart := strings.Index(logStr, `"sign":["`) + 9
			if signStart > 9 {
				sign = logStr[signStart : signStart+32] // sign固定长度为32
			}
			verifyCodeStart := strings.Index(logStr, `"verifyCode":"`) + 14
			if verifyCodeStart > 14 {
				verifyCode = logStr[verifyCodeStart : verifyCodeStart+6] // verifyCode固定长度为4
			}
			// 查找requestUrl的位置
			requestUrlStart := strings.Index(logStr, `"requestUrl":"`) + 14
			if requestUrlStart > 14 {
				requestUrlEnd := strings.Index(logStr[requestUrlStart:], `","`)
				if requestUrlEnd > 0 {
					requestUrl = logStr[requestUrlStart : requestUrlStart+requestUrlEnd]
				}
			}
			// 查找logTime的位置
			logTime = logStr[0:32] // logTime固定长度为19
			// 查找userId的位置
			userIdStart := strings.Index(logStr, `"userId":"`) + 10
			if userIdStart > 9 {
				userId = logStr[userIdStart : userIdStart+8] // userId固定长度为8
			}
			if userId == "" {
				userId = "00000000"
			}

			// 查找user_id的位置
			user_idStart := strings.Index(logStr, `"user_id":`) + 9
			if user_idStart > 8 {
				user_id = logStr[user_idStart : user_idStart+9] // user_id固定长度为8
			}
			// 查找user_id的位置
			user_idNumStart := strings.Index(logStr, `"userId":`) + 9

			if user_idNumStart > 8 {
				user_id = logStr[user_idNumStart : user_idNumStart+8] // user_id固定长度为8
				fmt.Printf("user_idNumStart: %v, user_id: %s\n", user_idNumStart, user_id)
			}
			// \"lot_number\":\"3677c41d61854cbd8956598ee09ccedd\"
			// 查找lot_number的位置
			lot_numberStart := strings.Index(logStr, `\"lot_number\":\"`) + 16
			if lot_numberStart > 12 {
				lot_number = logStr[lot_numberStart : lot_numberStart+33] // lot_number固定长度为32
			}
			// 查找phone的位置 ,"phone":"9757915506"
			phoneStart := strings.Index(logStr, `"phone":`) + 8
			if phoneStart > 7 {
				phone = logStr[phoneStart : phoneStart+11] // phone固定长度为11
			}

			// 查找traceId的位置
			traceIdStart := strings.Index(logStr, `"traceId":"`) + 11
			if traceIdStart > 10 {
				traceId = logStr[traceIdStart : traceIdStart+36] // traceId固定长度为36
			}

			// 查找paySerialNumber的位置
			paySerialNumberStart := strings.Index(logStr, `"paySerialNumber":"`) + 19
			if paySerialNumberStart > 18 {
				paySerialNumber = logStr[paySerialNumberStart : paySerialNumberStart+16] // paySerialNumber固定长度为16
			}
			// 查找paySerialNo的位置
			paySerialNoStart := strings.Index(logStr, `"paySerialNo":"`) + 15
			if paySerialNoStart > 14 {
				paySerialNo = logStr[paySerialNoStart : paySerialNoStart+16] // paySerialNo固定长度为16
			}

			// 查找requestReferenceNumber的位置
			requestReferenceNumberStart := strings.Index(logStr, `"requestReferenceNumber":"`) + 26
			if requestReferenceNumberStart > 25 {
				requestReferenceNumber = logStr[requestReferenceNumberStart : requestReferenceNumberStart+36] // requestReferenceNumber固定长度为36
			} else {
				requestReferenceNumberStart = strings.Index(logStr, `"Request-Reference-No":"`) + 24
				if requestReferenceNumberStart > 23 {
					requestReferenceNumber = logStr[requestReferenceNumberStart : requestReferenceNumberStart+36] // requestReferenceNumber固定长度为36
				}
			}

			// 只有当至少有一个字段不为空时才写入
			if userId != "" || traceId != "" || paySerialNumber != "" || paySerialNo != "" || requestReferenceNumber != "" || user_id != "" || lot_number != "" || phone != "" || logTime != "" || sign != "" || requestUrl != "" {
				// 创建CSV行数据
				row := []string{
					logTime, sign, requestUrl, userId, traceId,
					paySerialNumber, paySerialNo, requestReferenceNumber,
					user_id, lot_number, phone, verifyCode, userIp,
				}

				// 写入CSV行
				if err := writer.Write(row); err != nil {
					fmt.Printf("写入CSV行失败: %v\n", err)
				}
			}

			// if userId != "" || traceId != "" || paySerialNumber != "" || paySerialNo != "" || requestReferenceNumber != "" || user_id != "" || lot_number != "" || phone != "" || logTime != "" || sign != "" || requestUrl != "" {
			// 	fmt.Printf("logTime: %s, sign: %s, requestUrl: %s，userId: %s, traceId: %s, paySerialNumber: %s, paySerialNo: %s, requestReferenceNumber: %s, user_id: %s, lot_number: %s, phone: %s , verifyCode: %s \n",
			// 		logTime, sign, requestUrl, userId, traceId, paySerialNumber, paySerialNo, requestReferenceNumber, user_id, lot_number, phone, verifyCode)
			// }
			// fmt.Printf("%s,", userId)
		}

		// 关闭当前文件
		f.Close()

		if err := scanner.Err(); err != nil {
			fmt.Printf("读取文件时发生错误: %v\n", err)
		}
	}

	fmt.Printf("解析完成，数据已导出到 data.csv 文件\n")
}

// 切分 multi-redis 目录中的文件，每1万行一个文件
func splitMultiRedisFile() {
	// 查找 multi-redis 目录
	multiRedisDir := "multi-redis"
	if _, err := os.Stat(multiRedisDir); os.IsNotExist(err) {
		log.Printf("multi-redis 目录不存在，尝试创建...")
		err = os.Mkdir(multiRedisDir, 0755)
		if err != nil {
			log.Printf("创建 multi-redis 目录失败: %v", err)
			return
		}
		log.Printf("已创建 multi-redis 目录，请将需要切分的文件放入该目录")
		return
	}

	// 查找所有文件（不限文件类型）
	files, err := filepath.Glob(filepath.Join(multiRedisDir, "*"))
	if err != nil {
		log.Printf("查找文件失败: %v", err)
		return
	}

	if len(files) == 0 {
		log.Printf("multi-redis 目录中没有找到文件")
		return
	}

	// 创建输出目录
	outputDir := "multi-redis-split"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.Mkdir(outputDir, 0755)
		if err != nil {
			log.Printf("创建输出目录失败: %v", err)
			return
		}
	}

	// 处理每个文件
	for _, filePath := range files {
		// 检查是否是文件（而不是目录）
		fileInfo, err := os.Stat(filePath)
		if err != nil || fileInfo.IsDir() {
			continue
		}

		log.Printf("正在处理文件: %s", filePath)

		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("打开文件失败: %v", err)
			continue
		}
		registerFile(file)

		scanner := bufio.NewScanner(file)
		// 增加缓冲区大小以处理超长的行
		buf := make([]byte, 0, 1024*1024) // 1MB 缓冲区
		scanner.Buffer(buf, 1024*1024)    // 最大 1MB

		var currentFileIndex int = 1
		var currentLineCount int = 0
		var currentOutputFile *os.File

		// 获取原文件名（不含路径）
		baseFileName := filepath.Base(filePath)
		// 获取原文件扩展名
		fileExt := filepath.Ext(baseFileName)
		// 去掉原文件扩展名
		nameWithoutExt := strings.TrimSuffix(baseFileName, fileExt)

		// 创建第一个输出文件
		outputFileName := fmt.Sprintf("%s/%s_part_%04d%s", outputDir, nameWithoutExt, currentFileIndex, fileExt)
		currentOutputFile, err = os.Create(outputFileName)
		if err != nil {
			log.Printf("创建输出文件失败: %v", err)
			file.Close()
			continue
		}
		registerFile(currentOutputFile)
		log.Printf("创建输出文件: %s", outputFileName)

		totalLines := 0
		// 第一行插入换行符
		_, err = currentOutputFile.WriteString("\n")
		// 逐行读取并写入
		for scanner.Scan() {
			line := scanner.Text()
			totalLines++
			currentLineCount++

			// 写入当前输出文件
			_, err := currentOutputFile.WriteString(line + "\n")
			if err != nil {
				log.Printf("写入文件失败: %v", err)
				break
			}

			// 如果当前文件已达到1万行，创建新文件
			if currentLineCount >= 10000 {
				currentOutputFile.Close()
				currentFileIndex++
				currentLineCount = 0

				// 创建新的输出文件
				outputFileName = fmt.Sprintf("%s/%s_part_%04d%s", outputDir, nameWithoutExt, currentFileIndex, fileExt)
				currentOutputFile, err = os.Create(outputFileName)
				_, err = currentOutputFile.WriteString("\n")
				if err != nil {
					log.Printf("创建输出文件失败: %v", err)
					break
				}
				registerFile(currentOutputFile)
				log.Printf("创建输出文件: %s", outputFileName)
			}
		}

		// 关闭最后一个输出文件
		if currentOutputFile != nil {
			currentOutputFile.Close()
		}

		// 关闭输入文件
		file.Close()

		if err := scanner.Err(); err != nil {
			log.Printf("读取文件时发生错误: %v", err)
		}

		log.Printf("文件 %s 处理完成，总共 %d 行，切分为 %d 个文件", baseFileName, totalLines, currentFileIndex)
	}

	log.Printf("所有文件处理完成，输出文件保存在 %s 目录中", outputDir)
}

// KYC审核处理函数
func kycReviewProcessor() {
	// 查找 kyc-review 目录
	kycDir := "kyc-review"
	if _, err := os.Stat(kycDir); os.IsNotExist(err) {
		log.Printf("kyc-review 目录不存在，尝试创建...")
		err = os.Mkdir(kycDir, 0755)
		if err != nil {
			log.Printf("创建 kyc-review 目录失败: %v", err)
			return
		}
		log.Printf("已创建 kyc-review 目录，请将 KYC 文件放入该目录")
		return
	}

	// 查找 Excel 或 CSV 文件
	xlsxFiles, _ := filepath.Glob(filepath.Join(kycDir, "*.xlsx"))
	csvFiles, _ := filepath.Glob(filepath.Join(kycDir, "*.csv"))

	var allFiles []string
	allFiles = append(allFiles, xlsxFiles...)
	allFiles = append(allFiles, csvFiles...)

	if len(allFiles) == 0 {
		log.Printf("kyc-review 目录中没有找到 Excel 或 CSV 文件")
		return
	}

	// 获取当前日期用于文件名
	currentTime := time.Now()
	filename := fmt.Sprintf("kyc-%s.sql", currentTime.Format("2006-01-02"))

	// 创建输出文件
	outputFile, err := os.Create(filename)
	if err != nil {
		log.Printf("创建输出文件失败: %v", err)
		return
	}
	defer outputFile.Close()
	registerFile(outputFile)

	var sqlCount int

	// 处理每个文件
	for _, filePath := range allFiles {
		log.Printf("正在处理文件: %s", filePath)

		ext := filepath.Ext(filePath)
		if ext == ".xlsx" {
			count, err := processExcelFile(filePath, outputFile)
			if err != nil {
				log.Printf("处理 Excel 文件失败: %v", err)
				continue
			}
			sqlCount += count
		} else if ext == ".csv" {
			count, err := processCSVFile(filePath, outputFile)
			if err != nil {
				log.Printf("处理 CSV 文件失败: %v", err)
				continue
			}
			sqlCount += count
		}
	}

	log.Printf("KYC审核处理完成，共生成 %d 条 SQL 语句，已保存到 %s 文件", sqlCount, filename)
}

// 处理 Excel 文件
func processExcelFile(filePath string, outputFile *os.File) (int, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return 0, fmt.Errorf("无法获取工作表")
	}

	// 读取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return 0, err
	}

	// 跳过标题行，处理数据行
	sqlCount := 0
	for i, row := range rows {
		// 跳过标题行（假设第一行是标题）
		if i == 0 {
			continue
		}

		// 确保行有足够的列数据
		if len(row) >= 2 {
			// 假设第1列是 user_id，第2列是 id（根据实际Excel结构调整）
			userId := strings.TrimSpace(row[0])
			recordId := strings.TrimSpace(row[1])

			// 生成 SQL 语句
			if userId != "" && recordId != "" {
				sql := fmt.Sprintf("UPDATE b_kyc set audit_status = 1,audit_at = '%s' where audit_status = 2 and is_lock = 0 and user_id = %s and id = %s;\n",
					time.Now().Format("2006-01-02 15:04:05"), userId, recordId)

				_, err := outputFile.WriteString(sql)
				if err != nil {
					return sqlCount, err
				}
				sqlCount++
			}
		}
	}

	return sqlCount, nil
}

// 处理 CSV 文件
func processCSVFile(filePath string, outputFile *os.File) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// 读取所有记录
	records, err := reader.ReadAll()
	if err != nil {
		return 0, err
	}

	// 跳过标题行，处理数据行
	sqlCount := 0
	for i, record := range records {
		// 跳过标题行（假设第一行是标题）
		if i == 0 {
			continue
		}

		// 确保行有足够的列数据
		if len(record) >= 2 {
			// 假设第1列是 user_id，第2列是 id
			userId := strings.TrimSpace(record[0])
			recordId := strings.TrimSpace(record[1])

			// 生成 SQL 语句
			if userId != "" && recordId != "" {
				sql := fmt.Sprintf("UPDATE b_kyc set audit_status = 1,audit_at = '%s' where audit_status = 2 and is_lock = 0 and user_id = %s and id = %s;\n",
					time.Now().Format("2006-01-02 15:04:05"), userId, recordId)

				_, err := outputFile.WriteString(sql)
				if err != nil {
					return sqlCount, err
				}
				sqlCount++
			}
		}
	}

	return sqlCount, nil
}
