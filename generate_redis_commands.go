package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func main() {
	// 输入目录路径
	inputDir := "del-ratio"
	// 输出文件路径
	outputFile := "redis_delete_commands.txt"

	// 检查输入目录是否存在
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		fmt.Printf("错误：输入目录 %s 不存在\n", inputDir)
		return
	}

	// 查找所有 .csv 和 .xlsx 文件
	var files []string
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".csv" || ext == ".xlsx" {
				files = append(files, path)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("扫描目录时发生错误: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Printf("在目录 %s 中没有找到 .csv 或 .xlsx 文件\n", inputDir)
		return
	}

	fmt.Printf("找到 %d 个文件需要处理:\n", len(files))
	for _, file := range files {
		fmt.Printf("  - %s\n", file)
	}

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

	fmt.Printf("\n开始处理文件...\n")
	fmt.Printf("输出文件：%s\n", outputFile)

	// 计数器
	totalCount := 0
	startTime := time.Now()

	// 处理每个文件
	for _, inputFile := range files {
		fmt.Printf("\n正在处理文件：%s\n", inputFile)

		fileCount, err := processFile(inputFile, writer)
		if err != nil {
			fmt.Printf("处理文件 %s 时发生错误: %v\n", inputFile, err)
			continue
		}

		totalCount += fileCount
		fmt.Printf("文件 %s 处理完成，处理了 %d 个用户ID\n", inputFile, fileCount)
	}

	// 计算处理时间
	duration := time.Since(startTime)

	fmt.Printf("\n所有文件处理完成！\n")
	fmt.Printf("总共处理了 %d 个用户ID\n", totalCount)
	fmt.Printf("生成了 %d 条Redis命令\n", totalCount*2)
	fmt.Printf("处理时间: %v\n", duration)
	fmt.Printf("输出文件：%s\n", outputFile)

	// 显示输出文件的前几行作为示例
	fmt.Println("\n输出文件前10行示例：")
	showSampleLines(outputFile, 10)
}

// processFile 处理单个文件，返回处理的用户ID数量
func processFile(inputFile string, writer *bufio.Writer) (int, error) {
	ext := strings.ToLower(filepath.Ext(inputFile))

	if ext == ".xlsx" {
		return processExcelFile(inputFile, writer)
	} else if ext == ".csv" {
		return processCSVFile(inputFile, writer)
	} else {
		return 0, fmt.Errorf("不支持的文件格式: %s", ext)
	}
}

// processCSVFile 处理CSV文件
func processCSVFile(inputFile string, writer *bufio.Writer) (int, error) {
	// 打开输入文件
	file, err := os.Open(inputFile)
	if err != nil {
		return 0, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 计数器
	count := 0

	// 创建扫描器逐行读取
	scanner := bufio.NewScanner(file)
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

		// 生成两个Redis删除命令
		reqCmd := fmt.Sprintf("del risk:turnover:req:{%s}\n", userID)
		betCmd := fmt.Sprintf("del risk:turnover:bet:{%s}\n", userID)

		// 写入到输出文件
		writer.WriteString(reqCmd)
		writer.WriteString(betCmd)

		count++

		// 每处理10000行显示一次进度
		if count%10000 == 0 {
			fmt.Printf("  已处理 %d 个用户ID...\n", count)
		}
	}

	// 检查扫描过程中是否有错误
	if err := scanner.Err(); err != nil {
		return count, fmt.Errorf("读取文件时发生错误: %v", err)
	}

	return count, nil
}

// processExcelFile 处理Excel文件
func processExcelFile(inputFile string, writer *bufio.Writer) (int, error) {
	// 打开Excel文件
	f, err := excelize.OpenFile(inputFile)
	if err != nil {
		return 0, fmt.Errorf("打开Excel文件失败: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("关闭Excel文件时发生错误: %v\n", err)
		}
	}()

	// 获取所有工作表名称
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return 0, fmt.Errorf("Excel文件中没有工作表")
	}

	// 使用第一个工作表
	sheetName := sheetList[0]
	fmt.Printf("  正在处理工作表: %s\n", sheetName)

	// 获取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return 0, fmt.Errorf("读取Excel工作表失败: %v", err)
	}

	// 计数器
	count := 0

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
			fmt.Printf("  跳过表头行: %s\n", userID)
			continue
		}

		// 生成两个Redis删除命令
		reqCmd := fmt.Sprintf("del risk:turnover:req:{%s}\n", userID)
		betCmd := fmt.Sprintf("del risk:turnover:bet:{%s}\n", userID)

		// 写入到输出文件
		writer.WriteString(reqCmd)
		writer.WriteString(betCmd)

		count++

		// 每处理10000行显示一次进度
		if count%10000 == 0 {
			fmt.Printf("  已处理 %d 个用户ID...\n", count)
		}
	}

	return count, nil
}

// isNumeric 检查字符串是否为数字
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
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
