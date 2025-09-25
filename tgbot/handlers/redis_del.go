package handlers

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tgbot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// processRedisDeleteCmds 处理Redis删除命令生成功能 - 完整流水删除操作流程
func (hm *HandlerManager) processRedisDeleteCmds(chatID, userID int64, inputFile string, state *UserState) error {
	startTime := time.Now()

	hm.logger.Info("开始Redis删除操作流程",
		slog.Int64("user_id", userID),
		slog.Int64("chat_id", chatID),
		slog.String("input_file", utils.SanitizePath(inputFile)),
		slog.String("timestamp", startTime.Format(time.RFC3339)),
	)

	// 检查文件格式
	if !utils.IsValidFileType(inputFile, []string{".xlsx", ".csv"}) {
		hm.logger.Warn("不支持的文件格式",
			slog.Int64("user_id", userID),
			slog.String("input_file", utils.SanitizePath(inputFile)),
		)
		return fmt.Errorf("只支持Excel (.xlsx) 或CSV格式的文件")
	}

	// 发送处理开始消息
	progressMsg := tgbotapi.NewMessage(chatID, "🚀 开始执行Redis流水删除操作流程...")
	hm.bot.Send(progressMsg)

	// 步骤1：生成Redis删除命令
	step1Start := time.Now()
	hm.logger.Info("开始步骤1：生成Redis删除命令",
		slog.Int64("user_id", userID),
		slog.String("input_file", utils.SanitizePath(inputFile)),
		slog.String("timestamp", step1Start.Format(time.RFC3339)),
	)

	progressMsg = tgbotapi.NewMessage(chatID, "📝 步骤1：生成Redis删除命令...")
	hm.bot.Send(progressMsg)

	redisCommandsFile := filepath.Join(state.UserDir, "redis_delete_commands.txt")
	totalCount, err := hm.generateRedisDeleteCommands(inputFile, redisCommandsFile)
	if err != nil {
		hm.logger.LogError(userID, "generate_redis_commands", err, map[string]interface{}{
			"input_file": utils.SanitizePath(inputFile),
			"output_file": utils.SanitizePath(redisCommandsFile),
		})
		return fmt.Errorf("生成Redis命令失败: %v", err)
	}

	step1Duration := time.Since(step1Start)
	hm.logger.LogPerformance("redis_generate_commands", step1Duration, totalCount, userID)
	hm.logger.Info("步骤1完成：生成Redis删除命令",
		slog.Int64("user_id", userID),
		slog.Int("total_users", totalCount),
		slog.Int("total_commands", totalCount*2),
		slog.String("duration", step1Duration.String()),
		slog.String("output_file", utils.SanitizePath(redisCommandsFile)),
	)

	progressMsg = tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ 步骤1完成：成功生成 %d 条Redis命令", totalCount*2))
	hm.bot.Send(progressMsg)

	// 步骤2：创建multi-redis目录并移动文件
	step2Start := time.Now()
	hm.logger.Info("开始步骤2：创建工作目录并移动文件",
		slog.Int64("user_id", userID),
		slog.String("timestamp", step2Start.Format(time.RFC3339)),
	)

	progressMsg = tgbotapi.NewMessage(chatID, "📁 步骤2：创建工作目录...")
	hm.bot.Send(progressMsg)

	multiRedisDir := filepath.Join(state.UserDir, "multi-redis")
	err = os.MkdirAll(multiRedisDir, 0755)
	if err != nil {
		hm.logger.LogError(userID, "create_multi_redis_dir", err, map[string]interface{}{
			"target_dir": utils.SanitizePath(multiRedisDir),
		})
		return fmt.Errorf("创建multi-redis目录失败: %v", err)
	}

	// 移动redis命令文件到multi-redis目录
	multiRedisFile := filepath.Join(multiRedisDir, "redis_commands.txt")
	err = hm.copyFile(redisCommandsFile, multiRedisFile)
	if err != nil {
		hm.logger.LogError(userID, "copy_redis_commands", err, map[string]interface{}{
			"source_file": utils.SanitizePath(redisCommandsFile),
			"target_file": utils.SanitizePath(multiRedisFile),
		})
		return fmt.Errorf("移动Redis命令文件失败: %v", err)
	}

	step2Duration := time.Since(step2Start)
	hm.logger.Info("步骤2完成：工作目录创建和文件移动",
		slog.Int64("user_id", userID),
		slog.String("multi_redis_dir", utils.SanitizePath(multiRedisDir)),
		slog.String("redis_file", utils.SanitizePath(multiRedisFile)),
		slog.String("duration", step2Duration.String()),
	)

	progressMsg = tgbotapi.NewMessage(chatID, "✅ 步骤2完成：文件移动成功")
	hm.bot.Send(progressMsg)

	// 步骤3：执行文件分割（调用现有的文件分割功能）
	step3Start := time.Now()
	hm.logger.Info("开始步骤3：分割Redis命令文件",
		slog.Int64("user_id", userID),
		slog.String("source_file", utils.SanitizePath(multiRedisFile)),
		slog.String("timestamp", step3Start.Format(time.RFC3339)),
	)

	progressMsg = tgbotapi.NewMessage(chatID, "✂️ 步骤3：分割Redis命令文件（每10,000行一个文件）...")
	hm.bot.Send(progressMsg)

	splitDir := filepath.Join(state.UserDir, "multi-redis-split")
	err = hm.splitRedisCommandFile(multiRedisFile, splitDir)
	if err != nil {
		hm.logger.LogError(userID, "split_redis_commands", err, map[string]interface{}{
			"source_file": utils.SanitizePath(multiRedisFile),
			"split_dir":   utils.SanitizePath(splitDir),
		})
		return fmt.Errorf("分割文件失败: %v", err)
	}

	step3Duration := time.Since(step3Start)
	hm.logger.Info("步骤3完成：文件分割",
		slog.Int64("user_id", userID),
		slog.String("split_dir", utils.SanitizePath(splitDir)),
		slog.String("duration", step3Duration.String()),
	)

	progressMsg = tgbotapi.NewMessage(chatID, "✅ 步骤3完成：文件分割成功")
	hm.bot.Send(progressMsg)

	// 步骤4：创建执行脚本
	step4Start := time.Now()
	hm.logger.Info("开始步骤4：创建Redis执行脚本",
		slog.Int64("user_id", userID),
		slog.String("split_dir", utils.SanitizePath(splitDir)),
		slog.String("timestamp", step4Start.Format(time.RFC3339)),
	)

	progressMsg = tgbotapi.NewMessage(chatID, "📜 步骤4：创建Redis执行脚本...")
	hm.bot.Send(progressMsg)

	executeScriptPath := filepath.Join(splitDir, "execute_redis_commands.sh")
	err = hm.createExecuteScript(executeScriptPath)
	if err != nil {
		hm.logger.LogError(userID, "create_execute_script", err, map[string]interface{}{
			"script_path": utils.SanitizePath(executeScriptPath),
		})
		return fmt.Errorf("创建执行脚本失败: %v", err)
	}

	step4Duration := time.Since(step4Start)
	hm.logger.Info("步骤4完成：执行脚本创建",
		slog.Int64("user_id", userID),
		slog.String("script_path", utils.SanitizePath(executeScriptPath)),
		slog.String("duration", step4Duration.String()),
	)

	progressMsg = tgbotapi.NewMessage(chatID, "✅ 步骤4完成：执行脚本创建成功")
	hm.bot.Send(progressMsg)

	// 步骤5：压缩整个分割目录
	step5Start := time.Now()
	hm.logger.Info("开始步骤5：压缩文件包",
		slog.Int64("user_id", userID),
		slog.String("split_dir", utils.SanitizePath(splitDir)),
		slog.String("timestamp", step5Start.Format(time.RFC3339)),
	)

	progressMsg = tgbotapi.NewMessage(chatID, "🗜️ 步骤5：压缩文件包...")
	hm.bot.Send(progressMsg)

	zipFilePath := filepath.Join(state.UserDir, "redis-delete-commands.zip")
	zipHelper := utils.NewZipHelper()
	err = zipHelper.CreateZipFromDirectory(splitDir, zipFilePath)
	if err != nil {
		hm.logger.LogError(userID, "create_zip_package", err, map[string]interface{}{
			"split_dir":     utils.SanitizePath(splitDir),
			"zip_file_path": utils.SanitizePath(zipFilePath),
		})
		return fmt.Errorf("压缩文件失败: %v", err)
	}

	step5Duration := time.Since(step5Start)
	hm.logger.Info("步骤5完成：文件压缩",
		slog.Int64("user_id", userID),
		slog.String("zip_file", utils.SanitizePath(zipFilePath)),
		slog.String("duration", step5Duration.String()),
	)

	progressMsg = tgbotapi.NewMessage(chatID, "✅ 步骤5完成：文件压缩成功")
	hm.bot.Send(progressMsg)

	// 发送最终结果
	caption := fmt.Sprintf(`🎉 Redis流水删除操作流程完成！

📊 处理统计：
• 处理用户数: %d
• 生成Redis命令: %d 条
• 分割文件数: 自动分割
• 包含执行脚本: execute_redis_commands.sh

📦 压缩包内容：
• redis_commands_part_*.txt (分割后的命令文件)
• execute_redis_commands.sh (批量执行脚本)

🚀 使用方法：
1. 解压ZIP文件
2. 上传到Redis服务器
3. 运行 ./execute_redis_commands.sh`, totalCount, totalCount*2)

	hm.sendResultFile(chatID, zipFilePath, caption)

	// 记录操作完成日志
	hm.logger.LogPerformance("redis_delete_pipeline", time.Since(startTime), totalCount, userID)

	hm.logger.Info("Redis删除操作流程完成",
		slog.Int64("user_id", userID),
		slog.Int64("chat_id", chatID),
		slog.Int("total_users", totalCount),
		slog.Int("total_commands", totalCount*2),
		slog.String("zip_file", utils.SanitizePath(zipFilePath)),
		slog.String("duration", time.Since(startTime).String()),
		slog.String("timestamp", time.Now().Format(time.RFC3339)),
	)

	return nil
}

// generateRedisDeleteCommands 生成Redis删除命令
func (hm *HandlerManager) generateRedisDeleteCommands(inputFile, outputFile string) (int, error) {
	// 创建输出文件
	file, err := hm.fileManager.CreateOutputFile(outputFile)
	if err != nil {
		return 0, err
	}
	defer hm.fileManager.CloseFile(outputFile)

	totalCount := 0

	// 使用Excel辅助工具处理文件
	excelHelper := utils.NewExcelHelper()

	err = excelHelper.ProcessFileByType(inputFile, func(rows [][]string) error {
		for i, row := range rows {
			if len(row) == 0 {
				continue
			}

			var userID string
			if len(row) > 0 {
				userID = strings.TrimSpace(row[0])
			}

			if userID == "" {
				continue
			}

			// 跳过表头
			if i == 0 && !utils.IsNumeric(userID) {
				continue
			}

			// 生成两个Redis删除命令
			reqCmd := fmt.Sprintf("del risk:turnover:req:{%s}\n", userID)
			betCmd := fmt.Sprintf("del risk:turnover:bet:{%s}\n", userID)

			file.WriteString(reqCmd)
			file.WriteString(betCmd)

			totalCount++
		}
		return nil
	})

	return totalCount, err
}

// splitRedisCommandFile 分割Redis命令文件（专用版本，不创建ZIP）
func (hm *HandlerManager) splitRedisCommandFile(inputFile, outputDir string) error {
	// 创建输出目录
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return err
	}

	// 记录分割开始
	hm.logger.Info("开始分割Redis命令文件",
		slog.String("input_file", utils.SanitizePath(inputFile)),
		slog.String("output_dir", utils.SanitizePath(outputDir)),
	)

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
	outputFileName := fmt.Sprintf("%s/%s_part_%04d%s", outputDir, nameWithoutExt, currentFileIndex, fileExt)
	currentOutputFile, err = hm.fileManager.CreateOutputFile(outputFileName)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}

	totalLines := 0

	// 逐行读取并写入（不插入额外换行符）
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
			outputFileName = fmt.Sprintf("%s/%s_part_%04d%s", outputDir, nameWithoutExt, currentFileIndex, fileExt)
			currentOutputFile, err = hm.fileManager.CreateOutputFile(outputFileName)
			if err != nil {
				return fmt.Errorf("创建输出文件失败: %v", err)
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

	// 记录分割完成（不创建额外的ZIP文件）
	hm.logger.Info("Redis命令文件分割完成",
		slog.Int("total_lines", totalLines),
		slog.Int("split_files", currentFileIndex),
		slog.String("output_dir", utils.SanitizePath(outputDir)),
	)

	return nil
}

// createExecuteScript 创建Redis命令执行脚本
func (hm *HandlerManager) createExecuteScript(scriptPath string) error {
	script := `#!/bin/bash

# Redis命令批量执行脚本
# 自动生成时间: ` + time.Now().Format("2006-01-02 15:04:05") + `

echo "🚀 开始执行Redis删除命令..."
echo "⏰ 开始时间: $(date)"

# 统计文件数量
file_count=$(ls redis_commands_part_*.txt 2>/dev/null | wc -l)
echo "📁 找到 $file_count 个命令文件"

if [ $file_count -eq 0 ]; then
    echo "❌ 没有找到Redis命令文件"
    exit 1
fi

# 确保Redis客户端可用
if ! command -v redis-cli &> /dev/null; then
    echo "❌ redis-cli 命令不可用，请确保Redis客户端已安装"
    exit 1
fi

# 执行每个命令文件
counter=0
total_commands=0

for file in redis_commands_part_*.txt; do
    if [ -f "$file" ]; then
        counter=$((counter + 1))
        commands_in_file=$(wc -l < "$file")
        total_commands=$((total_commands + commands_in_file))

        echo "📝 正在执行第 $counter 个文件: $file (包含 $commands_in_file 条命令)"

        # 执行Redis命令文件
        redis-cli < "$file"

        if [ $? -eq 0 ]; then
            echo "✅ $file 执行成功"
        else
            echo "❌ $file 执行失败"
        fi
    fi
done

echo ""
echo "🎉 批量执行完成！"
echo "📊 总计执行了 $total_commands 条Redis命令"
echo "⏰ 结束时间: $(date)"
`

	file, err := hm.fileManager.CreateOutputFile(scriptPath)
	if err != nil {
		return err
	}
	defer hm.fileManager.CloseFile(scriptPath)

	_, err = file.WriteString(script)
	return err
}

// copyFile 复制文件
func (hm *HandlerManager) copyFile(src, dst string) error {
	srcFile, err := hm.fileManager.OpenFile(src)
	if err != nil {
		return err
	}
	defer hm.fileManager.CloseFile(src)

	dstFile, err := hm.fileManager.CreateOutputFile(dst)
	if err != nil {
		return err
	}
	defer hm.fileManager.CloseFile(dst)

	// 简单的文件复制
	buffer := make([]byte, 1024*1024) // 1MB buffer
	for {
		n, err := srcFile.Read(buffer)
		if n == 0 || err != nil {
			break
		}
		dstFile.Write(buffer[:n])
	}

	return nil
}