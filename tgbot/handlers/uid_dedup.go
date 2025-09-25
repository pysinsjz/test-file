package handlers

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"
	"tgbot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// processUIDDeduplicate 处理UID去重功能
func (hm *HandlerManager) processUIDDeduplicate(chatID, userID int64, inputFile string, state *UserState) error {
	// 检查文件格式 - 只支持CSV
	if !utils.IsValidFileType(inputFile, []string{".csv"}) {
		return fmt.Errorf("只支持CSV格式的文件")
	}

	// 发送处理开始消息
	progressMsg := tgbotapi.NewMessage(chatID, "🔄 正在读取和统计UID...")
	hm.bot.Send(progressMsg)

	// 打开输入文件
	file, err := hm.fileManager.OpenFile(inputFile)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer hm.fileManager.CloseFile(inputFile)

	// 统计每个uid的出现次数
	uidCounts := make(map[string]int)

	scanner := bufio.NewScanner(file)
	totalLines := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			uidCounts[line]++
			totalLines++

			// 每处理10000行显示一次进度
			if totalLines%10000 == 0 {
				progressMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("🔄 已读取 %d 行数据...", totalLines))
				hm.bot.Send(progressMsg)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件时出错: %v", err)
	}

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

	progressMsg = tgbotapi.NewMessage(chatID, fmt.Sprintf("📊 分析完成！\n📈 总行数: %d\n🔢 不同UID: %d\n✅ 唯一UID: %d\n🔄 重复UID: %d", totalLines, len(uidCounts), uniqueCount, duplicateCount))
	hm.bot.Send(progressMsg)

	// 创建去重后的输出文件
	outputFile := filepath.Join(state.UserDir, "unique_uids.csv")
	outFile, err := hm.fileManager.CreateOutputFile(outputFile)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer hm.fileManager.CloseFile(outputFile)

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

	// 创建去重报告文件
	reportFile := filepath.Join(state.UserDir, "dedup_report.txt")
	report, err := hm.fileManager.CreateOutputFile(reportFile)
	if err != nil {
		return fmt.Errorf("创建报告文件失败: %v", err)
	}
	defer hm.fileManager.CloseFile(reportFile)

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

	// 发送去重后的文件
	hm.sendResultFile(chatID, outputFile, fmt.Sprintf("✅ UID去重完成！\n📄 成功写入 %d 个唯一UID", writtenCount))

	// 发送详细报告
	hm.sendResultFile(chatID, reportFile, fmt.Sprintf("📋 去重报告\n📊 原始数据: %d 行\n🎯 去重后: %d 个唯一UID\n🔄 重复数据: %d 个", totalLines, uniqueCount, duplicateCount))

	return nil
}