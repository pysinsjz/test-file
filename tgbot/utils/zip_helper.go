package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ZipHelper ZIP文件处理辅助类
type ZipHelper struct{}

// NewZipHelper 创建ZIP辅助类实例
func NewZipHelper() *ZipHelper {
	return &ZipHelper{}
}

// CreateZipFromDirectory 将整个目录压缩为ZIP文件
func (zh *ZipHelper) CreateZipFromDirectory(sourceDir, zipPath string) error {
	// 创建ZIP文件
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("创建ZIP文件失败: %v", err)
	}
	defer zipFile.Close()

	// 创建ZIP写入器
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 遍历源目录
	err = filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录本身
		if info.IsDir() {
			return nil
		}

		// 计算ZIP内的相对路径
		relPath, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return err
		}
		// 统一使用斜杠作为路径分隔符（ZIP标准）
		relPath = strings.ReplaceAll(relPath, "\\", "/")

		// 创建ZIP内的文件
		zipFileWriter, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// 打开源文件
		srcFile, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// 复制文件内容到ZIP
		_, err = io.Copy(zipFileWriter, srcFile)
		return err
	})

	return err
}

// CreateZipFromFiles 将多个文件压缩为ZIP
func (zh *ZipHelper) CreateZipFromFiles(files []string, zipPath string) error {
	// 创建ZIP文件
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("创建ZIP文件失败: %v", err)
	}
	defer zipFile.Close()

	// 创建ZIP写入器
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, filePath := range files {
		// 获取文件信息
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			continue // 跳过无法访问的文件
		}

		if fileInfo.IsDir() {
			continue // 跳过目录
		}

		// 使用文件名作为ZIP内的路径
		fileName := filepath.Base(filePath)

		// 创建ZIP内的文件
		zipFileWriter, err := zipWriter.Create(fileName)
		if err != nil {
			return err
		}

		// 打开源文件
		srcFile, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// 复制文件内容到ZIP
		_, err = io.Copy(zipFileWriter, srcFile)
		if err != nil {
			return err
		}
	}

	return nil
}