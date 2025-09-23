package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type UserData struct {
	UserID        string
	AdjustAmount  int32
	TurnoverRatio int32
	BetAmount     int32
}

func main() {
	// 读取 CSV 文件
	csvFile, err := os.Open("add-ratio/addRatio.csv")
	if err != nil {
		fmt.Printf("Error opening CSV file: %v\n", err)
		return
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV file: %v\n", err)
		return
	}

	// 创建输出文件
	outputFile, err := os.Create("add-ratio/redis_commands.txt")
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer outputFile.Close()

	// 跳过标题行，处理数据行
	for i, record := range records {
		if i == 0 {
			continue // 跳过标题行
		}

		if len(record) < 5 {
			fmt.Printf("Warning: Record %d has insufficient columns, skipping\n", i+1)
			continue
		}

		// 解析数据
		userID := strings.TrimSpace(record[0])
		adjustAmountFloat, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			fmt.Printf("Error parsing adjust_amount for user %s: %v\n", userID, err)
			continue
		}

		turnoverRatioFloat, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			fmt.Printf("Error parsing turnover_ratio for user %s: %v\n", userID, err)
			continue
		}

		betAmountFloat, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			betAmountFloat = 0.0
		}

		// 转换为 int 类型
		adjustAmount := int64(adjustAmountFloat)
		turnoverRatio := int64(turnoverRatioFloat)
		betAmount := int64(betAmountFloat)

		// 计算 req 值 (adjust_amount * ratio)
		req := adjustAmount * turnoverRatio
		cmd1 := fmt.Sprintf("del risk:turnover:req:{%s} risk:turnover:bet:{%s}\n", userID, userID)
		// 写入文件
		outputFile.WriteString(cmd1)
		if betAmount*100 > req*100 {
			fmt.Printf("Warning: betAmount*100 > req*100 for user %s, skipping\n", userID)
			continue
		}
		// 生成第一个 Redis 命令：设置用户流水要求
		// adjustAmount 需要乘以 100 转换为分
		// "{\"req\":3000000,\"items\":[{\"type\":\"welcome back\",\"bonds\":150000,\"ratio\":20}]}"
		cmd2 := fmt.Sprintf("set risk:turnover:req:{%s} \"{\\\"req\\\":%d,\\\"items\\\":[{\\\"type\\\":\\\"welcome back\\\",\\\"bounds\\\":%d,\\\"ratio\\\":%d}]}\"\n",
			userID, req*100, adjustAmount*100, turnoverRatio)

		// 生成第二个 Redis 命令：设置用户投注流水
		// betAmount 需要乘以 100 转换为分
		cmd3 := fmt.Sprintf("set risk:turnover:bet:{%s} %d\n", userID, betAmount*100)

		outputFile.WriteString(cmd2)
		outputFile.WriteString(cmd3)

		// 每处理100条记录显示进度
		if i%100 == 0 {
			fmt.Printf("Processed %d records...\n", i)
		}
	}

	fmt.Printf("Successfully generated Redis commands for %d users\n", len(records)-1)
	fmt.Println("Commands saved to: add-ratio/redis_commands.txt")
}
