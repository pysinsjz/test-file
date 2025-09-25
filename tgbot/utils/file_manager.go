package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileManager 文件管理器
type FileManager struct {
	tempDir string
	openFiles map[string]*os.File
}

// NewFileManager 创建文件管理器
func NewFileManager(tempDir string) *FileManager {
	fm := &FileManager{
		tempDir: tempDir,
		openFiles: make(map[string]*os.File),
	}

	// 确保临时目录存在
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Printf("创建临时目录失败: %v", err)
	}

	return fm
}

// CreateUserDir 为用户创建专属目录
func (fm *FileManager) CreateUserDir(userID int64) string {
	userDir := filepath.Join(fm.tempDir, fmt.Sprintf("user_%d_%d", userID, time.Now().Unix()))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		log.Printf("创建用户目录失败: %v", err)
		return fm.tempDir
	}
	return userDir
}

// SaveUploadedFile 保存上传的文件
func (fm *FileManager) SaveUploadedFile(src io.Reader, destPath string) error {
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer destFile.Close()

	// 注册文件句柄
	fm.registerFile(destPath, destFile)

	_, err = io.Copy(destFile, src)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %v", err)
	}

	return nil
}

// CreateOutputFile 创建输出文件
func (fm *FileManager) CreateOutputFile(filePath string) (*os.File, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建输出文件失败: %v", err)
	}

	// 注册文件句柄
	fm.registerFile(filePath, file)
	return file, nil
}

// OpenFile 打开文件
func (fm *FileManager) OpenFile(filePath string) (*os.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}

	// 注册文件句柄
	fm.registerFile(filePath, file)
	return file, nil
}

// 注册文件句柄用于管理
func (fm *FileManager) registerFile(path string, file *os.File) {
	fm.openFiles[path] = file
}

// CloseFile 关闭指定文件
func (fm *FileManager) CloseFile(filePath string) {
	if file, exists := fm.openFiles[filePath]; exists {
		file.Close()
		delete(fm.openFiles, filePath)
	}
}

// CleanupFiles 清理所有打开的文件
func (fm *FileManager) CleanupFiles() {
	for path, file := range fm.openFiles {
		if file != nil {
			file.Close()
		}
		delete(fm.openFiles, path)
	}
}

// CleanupUserDir 清理用户目录
func (fm *FileManager) CleanupUserDir(userDir string) {
	// 首先关闭该目录下的所有文件
	for path, file := range fm.openFiles {
		if strings.HasPrefix(path, userDir) {
			if file != nil {
				file.Close()
			}
			delete(fm.openFiles, path)
		}
	}

	// 然后删除整个目录
	if err := os.RemoveAll(userDir); err != nil {
		log.Printf("清理用户目录失败: %v", err)
	}
}

// GetFileExt 获取文件扩展名
func GetFileExt(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

// IsValidFileType 检查文件类型是否有效
func IsValidFileType(filename string, validExts []string) bool {
	ext := GetFileExt(filename)
	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

// GetFileSize 获取文件大小
func GetFileSize(filePath string) (int64, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}