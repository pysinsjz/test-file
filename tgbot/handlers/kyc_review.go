package handlers

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"tgbot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// processKYCReviewHandler 处理KYC审核处理功能
func (hm *HandlerManager) processKYCReviewHandler(chatID, userID int64, inputFile string, state *UserState) error {
	// 检查文件格式
	if !utils.IsValidFileType(inputFile, []string{".xlsx", ".csv"}) {
		return fmt.Errorf("只支持Excel (.xlsx) 或CSV格式的文件")
	}

	// 发送处理开始消息
	progressMsg := tgbotapi.NewMessage(chatID, "🔄 正在处理KYC审核数据...")
	hm.bot.Send(progressMsg)

	// 获取当前日期用于文件名
	currentTime := time.Now()
	filename := fmt.Sprintf("kyc-%s.sql", currentTime.Format("2006-01-02"))
	outputFile := filepath.Join(state.UserDir, filename)

	// 创建输出文件
	file, err := hm.fileManager.CreateOutputFile(outputFile)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer hm.fileManager.CloseFile(outputFile)

	var sqlCount int

	// 使用Excel辅助工具处理文件
	excelHelper := utils.NewExcelHelper()

	err = excelHelper.ProcessFileByType(inputFile, func(rows [][]string) error {
		// 跳过标题行，处理数据行
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

					_, err := file.WriteString(sql)
					if err != nil {
						return fmt.Errorf("写入SQL语句失败: %v", err)
					}
					sqlCount++
				}
			}

			// 每处理1000行发送进度
			if i%1000 == 0 && i > 0 {
				progressMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("🔄 已处理 %d 行KYC数据，生成 %d 条SQL...", i, sqlCount))
				hm.bot.Send(progressMsg)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("处理文件失败: %v", err)
	}

	// 发送结果文件
	hm.sendResultFile(chatID, outputFile, fmt.Sprintf("✅ KYC审核处理完成！\n📋 共生成 %d 条SQL语句\n📅 文件名: %s", sqlCount, filename))

	return nil
}