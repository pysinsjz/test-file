package handlers

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tgbot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// processMultiFileSplit 处理文件分割功能
func (hm *HandlerManager) processMultiFileSplit(chatID, userID int64, inputFile string, state *UserState) error {
	// 发送处理开始消息
	progressMsg := tgbotapi.NewMessage(chatID, "🔄 正在分析文件...")
	hm.bot.Send(progressMsg)

	// 打开输入文件
	file, err := hm.fileManager.OpenFile(inputFile)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer hm.fileManager.CloseFile(inputFile)

	scanner := bufio.NewScanner(file)
	// 增加缓冲区大小以处理超长的行
	buf := make([]byte, 0, 1024*1024) // 1MB 缓冲区
	scanner.Buffer(buf, 1024*1024)    // 最大 1MB

	var currentFileIndex int = 1
	var currentLineCount int = 0
	var currentOutputFile *os.File

	// 获取原文件名（不含路径）
	baseFileName := filepath.Base(inputFile)
	// 获取原文件扩展名
	fileExt := filepath.Ext(baseFileName)
	// 去掉原文件扩展名
	nameWithoutExt := strings.TrimSuffix(baseFileName, fileExt)

	// 创建第一个输出文件
	outputFileName := fmt.Sprintf("%s/%s_part_%04d%s", state.UserDir, nameWithoutExt, currentFileIndex, fileExt)
	currentOutputFile, err = hm.fileManager.CreateOutputFile(outputFileName)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}

	progressMsg = tgbotapi.NewMessage(chatID, fmt.Sprintf("📝 创建分割文件: %s_part_%04d%s", nameWithoutExt, currentFileIndex, fileExt))
	hm.bot.Send(progressMsg)

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
			return fmt.Errorf("写入文件失败: %v", err)
		}

		// 如果当前文件已达到1万行，创建新文件
		if currentLineCount >= 10000 {
			hm.fileManager.CloseFile(outputFileName)
			currentFileIndex++
			currentLineCount = 0

			// 创建新的输出文件
			outputFileName = fmt.Sprintf("%s/%s_part_%04d%s", state.UserDir, nameWithoutExt, currentFileIndex, fileExt)
			currentOutputFile, err = hm.fileManager.CreateOutputFile(outputFileName)
			if err != nil {
				return fmt.Errorf("创建输出文件失败: %v", err)
			}
			_, err = currentOutputFile.WriteString("\n")
			if err != nil {
				return fmt.Errorf("写入文件头失败: %v", err)
			}

			// 每创建新文件时发送进度
			if currentFileIndex%5 == 0 {
				progressMsg = tgbotapi.NewMessage(chatID, fmt.Sprintf("📝 正在创建第 %d 个分割文件，已处理 %d 行...", currentFileIndex, totalLines))
				hm.bot.Send(progressMsg)
			}
		}
	}

	// 关闭最后一个输出文件
	if currentOutputFile != nil {
		hm.fileManager.CloseFile(outputFileName)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件时发生错误: %v", err)
	}

	// 创建压缩文件（如果有多个文件）
	if currentFileIndex > 1 {
		zipFileName := fmt.Sprintf("%s/%s_split_files.zip", state.UserDir, nameWithoutExt)

		// 使用真正的ZIP压缩功能
		splitFilesDir := state.UserDir
		zipHelper := utils.NewZipHelper()
		err = zipHelper.CreateZipFromDirectory(splitFilesDir, zipFileName)
		if err != nil {
			return fmt.Errorf("创建压缩文件失败: %v", err)
		}

		// 发送压缩文件
		hm.sendResultFile(chatID, zipFileName, fmt.Sprintf("✅ 文件分割完成！\n📄 总计 %d 行数据\n📦 分割为 %d 个文件", totalLines, currentFileIndex))
	} else {
		// 只有一个文件，直接发送
		singleFile := fmt.Sprintf("%s/%s_part_%04d%s", state.UserDir, nameWithoutExt, 1, fileExt)
		hm.sendResultFile(chatID, singleFile, fmt.Sprintf("✅ 文件处理完成！\n📄 总计 %d 行数据（无需分割）", totalLines))
	}

	return nil
}