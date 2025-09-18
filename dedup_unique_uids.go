package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	// 读取CSV文件
	inputFile := "rm-repeat-uid/uid.csv"
	outputFile := "rm-repeat-uid/unique_uids.csv"

	// 统计每个uid的出现次数
	uidCounts := make(map[string]int)

	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("打开文件失败: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	totalLines := 0

	fmt.Println("正在读取和统计uid...")
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			uidCounts[line]++
			totalLines++
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("读取文件时出错: %v\n", err)
		return
	}

	fmt.Printf("总共读取了 %d 行数据\n", totalLines)
	fmt.Printf("发现 %d 个不同的uid\n", len(uidCounts))

	// 统计重复情况
	uniqueCount := 0
	duplicateCount := 0

	for _, count := range uidCounts {
		if count == 1 {
			uniqueCount++
		} else {
			duplicateCount++
		}
	}

	fmt.Printf("唯一uid数量: %d\n", uniqueCount)
	fmt.Printf("重复uid数量: %d\n", duplicateCount)

	// 创建输出文件，只写入唯一的uid
	outputFileHandle, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("创建输出文件失败: %v\n", err)
		return
	}
	defer outputFileHandle.Close()

	writer := bufio.NewWriter(outputFileHandle)
	defer writer.Flush()

	// 写入唯一的uid
	writtenCount := 0
	for uid, count := range uidCounts {
		if count == 1 {
			_, err := writer.WriteString(uid + "\n")
			if err != nil {
				fmt.Printf("写入文件时出错: %v\n", err)
				return
			}
			writtenCount++
		}
	}

	fmt.Printf("成功写入 %d 个唯一uid到 %s\n", writtenCount, outputFile)

	// 显示一些重复uid的示例
	fmt.Println("\n重复uid示例（前10个）:")
	exampleCount := 0
	for uid, count := range uidCounts {
		if count > 1 && exampleCount < 10 {
			fmt.Printf("uid: %s, 出现次数: %d\n", uid, count)
			exampleCount++
		}
	}
}
