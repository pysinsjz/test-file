package handlers

import (
	"bufio"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"tgbot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// processSQLLogParse 处理SQL日志解析功能
func (hm *HandlerManager) processSQLLogParse(chatID, userID int64, inputFile string, state *UserState) error {
	// 检查输入文件是否是TXT格式
	if !utils.IsValidFileType(inputFile, []string{".txt"}) {
		return fmt.Errorf("只支持TXT格式的日志文件")
	}

	// 发送处理开始消息
	progressMsg := tgbotapi.NewMessage(chatID, "🔄 正在解析SQL日志文件...")
	hm.bot.Send(progressMsg)

	// 创建输出文件
	outputFile := filepath.Join(state.UserDir, "sql.log")
	file, err := hm.fileManager.CreateOutputFile(outputFile)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer hm.fileManager.CloseFile(outputFile)

	// 打开输入文件
	inputFileHandle, err := hm.fileManager.OpenFile(inputFile)
	if err != nil {
		return fmt.Errorf("打开输入文件失败: %v", err)
	}
	defer hm.fileManager.CloseFile(inputFile)

	var sqlCount int
	uniqueSQLs := make(map[string]bool) // 用于去重的map

	scanner := bufio.NewScanner(inputFileHandle)
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
					sqlKey := hm.generateSQLKey(sqlStatement)

					// 检查是否已经存在相同的SQL
					if !uniqueSQLs[sqlKey] {
						uniqueSQLs[sqlKey] = true

						// 写入输出文件
						outputLine := fmt.Sprintf("%s\n", sqlStatement)
						_, err := file.WriteString(outputLine)
						if err != nil {
							return fmt.Errorf("写入输出文件失败: %v", err)
						}
						sqlCount++
					}
				}
			}
		}

		// 每处理5000行发送一次进度消息
		if lineNum%5000 == 0 {
			progress := fmt.Sprintf("🔄 已处理 %d 行，提取 %d 条唯一SQL...", lineNum, sqlCount)
			progressMsg := tgbotapi.NewMessage(chatID, progress)
			hm.bot.Send(progressMsg)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件时发生错误: %v", err)
	}

	// 清理内存
	uniqueSQLs = nil

	// 发送结果文件
	hm.sendResultFile(chatID, outputFile, fmt.Sprintf("✅ SQL解析完成！\n📊 总计处理 %d 行日志，提取 %d 条唯一SQL语句", lineNum, sqlCount))

	return nil
}

// generateSQLKey 生成SQL的唯一标识，用于去重
func (hm *HandlerManager) generateSQLKey(sql string) string {
	// 转换为小写并去除多余空格
	sql = strings.ToLower(strings.TrimSpace(sql))

	// 提取表名
	tableName := hm.extractTableName(sql)

	// 提取字段列表
	fields := hm.extractFields(sql)

	// 提取where条件
	whereCondition := hm.extractWhereCondition(sql)

	// 组合成唯一标识
	return fmt.Sprintf("%s|%s|%s", tableName, fields, whereCondition)
}

// extractTableName 提取表名
func (hm *HandlerManager) extractTableName(sql string) string {
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
func (hm *HandlerManager) extractFields(sql string) string {
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
func (hm *HandlerManager) extractWhereCondition(sql string) string {
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
		return hm.extractFieldNames(whereClause)
	}

	return ""
}

// extractFieldNames 从WHERE条件中提取字段名，忽略参数值
func (hm *HandlerManager) extractFieldNames(whereClause string) string {
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
			fieldName := hm.extractFieldName(part)
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
func (hm *HandlerManager) extractFieldName(condition string) string {
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