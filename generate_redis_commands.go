package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	// 输入文件路径
	inputFile := "del-ratio/user_id.csv"
	// 输出文件路径
	outputFile := "redis_delete_commands.txt"

	// 检查输入文件是否存在
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		fmt.Printf("错误：输入文件 %s 不存在\n", inputFile)
		return
	}

	// 打开输入文件
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("打开文件失败: %v\n", err)
		return
	}
	defer file.Close()

	// 创建输出文件
	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("创建输出文件失败: %v\n", err)
		return
	}
	defer outFile.Close()

	// 创建写入器
	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	fmt.Printf("开始处理文件：%s\n", inputFile)
	fmt.Printf("输出文件：%s\n", outputFile)

	// 计数器
	count := 0
	startTime := time.Now()

	// 创建扫描器逐行读取
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行
		if line == "" {
			continue
		}

		// 生成两个Redis删除命令
		reqCmd := fmt.Sprintf("del risk:turnover:req:{%s}\n", line)
		betCmd := fmt.Sprintf("del risk:turnover:bet:{%s}\n", line)

		// 写入到输出文件
		writer.WriteString(reqCmd)
		writer.WriteString(betCmd)

		count++

		// 每处理10000行显示一次进度
		if count%10000 == 0 {
			fmt.Printf("已处理 %d 个用户ID...\n", count)
		}
	}

	// 检查扫描过程中是否有错误
	if err := scanner.Err(); err != nil {
		fmt.Printf("读取文件时发生错误: %v\n", err)
		return
	}

	// 计算处理时间
	duration := time.Since(startTime)

	fmt.Printf("处理完成！\n")
	fmt.Printf("总共处理了 %d 个用户ID\n", count)
	fmt.Printf("生成了 %d 条Redis命令\n", count*2)
	fmt.Printf("处理时间: %v\n", duration)
	fmt.Printf("输出文件：%s\n", outputFile)

	// 显示输出文件的前几行作为示例
	fmt.Println("\n输出文件前10行示例：")
	showSampleLines(outputFile, 10)
}

// 显示文件的前几行作为示例
func showSampleLines(filename string, lines int) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("无法打开文件进行预览: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() && count < lines {
		fmt.Println(scanner.Text())
		count++
	}
}
