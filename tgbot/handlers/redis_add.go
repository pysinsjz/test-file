package handlers

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"tgbot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// processRedisAddCmds 处理Redis流水增加命令生成功能
func (hm *HandlerManager) processRedisAddCmds(chatID, userID int64, inputFile string, state *UserState) error {
	// 检查文件格式 - 只支持CSV
	if !utils.IsValidFileType(inputFile, []string{".csv"}) {
		return fmt.Errorf("只支持CSV格式的文件")
	}

	// 发送处理开始消息
	progressMsg := tgbotapi.NewMessage(chatID, "🔄 正在生成Redis流水设置命令...")
	hm.bot.Send(progressMsg)

	// 创建输出文件
	outputFile := filepath.Join(state.UserDir, "redis_add_commands.txt")
	file, err := hm.fileManager.CreateOutputFile(outputFile)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer hm.fileManager.CloseFile(outputFile)

	// 计数器
	totalCount := 0
	startTime := time.Now()

	// 使用Excel辅助工具处理CSV文件
	excelHelper := utils.NewExcelHelper()

	err = excelHelper.ProcessFileByType(inputFile, func(rows [][]string) error {
		// 跳过标题行，处理数据行
		for i, row := range rows {
			if i == 0 {
				continue // 跳过标题行
			}

			if len(row) < 5 {
				continue // 确保有足够的列
			}

			// 解析数据
			userID := strings.TrimSpace(row[0])
			adjustAmountStr := strings.TrimSpace(row[1])
			turnoverRatioStr := strings.TrimSpace(row[2])
			betAmountStr := ""
			if len(row) > 4 {
				betAmountStr = strings.TrimSpace(row[4])
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
			file.WriteString(cmd1)

			// 2. 设置用户流水要求
			cmd2 := fmt.Sprintf("set risk:turnover:req:{%s} \"{\\\"req\\\":%d,\\\"items\\\":[{\\\"type\\\":\\\"welcome back\\\",\\\"bounds\\\":%d,\\\"ratio\\\":%d}]}\"\n",
				userID, req*100, adjustAmount*100, turnoverRatio)
			file.WriteString(cmd2)

			// 3. 设置用户投注流水
			cmd3 := fmt.Sprintf("set risk:turnover:bet:{%s} %d\n", userID, betAmount*100)
			file.WriteString(cmd3)

			totalCount++

			// 每处理100条记录显示进度
			if totalCount%100 == 0 {
				progressMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("🔄 已处理 %d 个用户，生成 %d 条命令...", totalCount, totalCount*3))
				hm.bot.Send(progressMsg)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("处理文件失败: %v", err)
	}

	// 计算处理时间
	duration := time.Since(startTime)

	// 发送结果文件
	hm.sendResultFile(chatID, outputFile, fmt.Sprintf("✅ Redis流水命令生成完成！\n👤 处理了 %d 个用户\n⚙️ 生成了 %d 条Redis命令\n⏱️ 处理时间: %v", totalCount, totalCount*3, duration))

	return nil
}