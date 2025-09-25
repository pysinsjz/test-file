package utils

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ExcelHelper Excel处理辅助类
type ExcelHelper struct{}

// NewExcelHelper 创建Excel辅助类实例
func NewExcelHelper() *ExcelHelper {
	return &ExcelHelper{}
}

// ReadExcelFile 读取Excel文件的所有行
func (eh *ExcelHelper) ReadExcelFile(filePath string) ([][]string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("无法获取工作表")
	}

	// 读取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// ReadCSVFile 读取CSV文件的所有行
func (eh *ExcelHelper) ReadCSVFile(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

// ProcessFileByType 根据文件类型处理文件
func (eh *ExcelHelper) ProcessFileByType(filePath string, processor func([][]string) error) error {
	ext := GetFileExt(filePath)

	var rows [][]string
	var err error

	if ext == ".xlsx" {
		rows, err = eh.ReadExcelFile(filePath)
	} else if ext == ".csv" {
		rows, err = eh.ReadCSVFile(filePath)
	} else {
		return fmt.Errorf("不支持的文件类型: %s", ext)
	}

	if err != nil {
		return err
	}

	return processor(rows)
}

// IsNumeric 检查字符串是否为数字
func IsNumeric(s string) bool {
	_, err := strconv.Atoi(strings.TrimSpace(s))
	return err == nil
}

// WriteCSV 写入CSV文件
func WriteCSV(filePath string, headers []string, data [][]string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入头部
	if len(headers) > 0 {
		if err := writer.Write(headers); err != nil {
			return err
		}
	}

	// 写入数据
	for _, row := range data {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}