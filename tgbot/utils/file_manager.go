package utils

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileManager 文件管理器
type FileManager struct {
	tempDir   string
	openFiles map[string]*os.File
	logger    *Logger
}

// NewFileManager 创建文件管理器
func NewFileManager(tempDir string) *FileManager {
	fm := &FileManager{
		tempDir:   tempDir,
		openFiles: make(map[string]*os.File),
		logger:    GetLogger(), // 获取全局日志记录器
	}

	// 确保临时目录存在
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Printf("创建临时目录失败: %v", err)
		if fm.logger != nil {
			fm.logger.Error("创建临时目录失败",
				slog.String("temp_dir", tempDir),
				slog.String("error", err.Error()),
				slog.String("timestamp", time.Now().Format(time.RFC3339)),
			)
		}
	} else {
		if fm.logger != nil {
			fm.logger.Info("临时目录创建成功",
				slog.String("temp_dir", tempDir),
				slog.String("timestamp", time.Now().Format(time.RFC3339)),
			)
		}
	}

	return fm
}

// CreateUserDir 为用户创建专属目录
func (fm *FileManager) CreateUserDir(userID int64) string {
	userDir := filepath.Join(fm.tempDir, fmt.Sprintf("user_%d_%d", userID, time.Now().Unix()))

	if err := os.MkdirAll(userDir, 0755); err != nil {
		log.Printf("创建用户目录失败: %v", err)
		if fm.logger != nil {
			fm.logger.LogError(userID, "create_user_dir", err, map[string]interface{}{
				"user_dir": SanitizePath(userDir),
			})
		}
		return fm.tempDir
	}

	if fm.logger != nil {
		fm.logger.Info("用户目录创建成功",
			slog.Int64("user_id", userID),
			slog.String("user_dir", SanitizePath(userDir)),
			slog.String("timestamp", time.Now().Format(time.RFC3339)),
		)
	}

	return userDir
}

// SaveUploadedFile 保存上传的文件
func (fm *FileManager) SaveUploadedFile(src io.Reader, destPath string) error {
	destFile, err := os.Create(destPath)
	if err != nil {
		if fm.logger != nil {
			fm.logger.Error("创建目标文件失败",
				slog.String("dest_path", SanitizePath(destPath)),
				slog.String("error", err.Error()),
				slog.String("timestamp", time.Now().Format(time.RFC3339)),
			)
		}
		return fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer destFile.Close()

	// 注册文件句柄
	fm.registerFile(destPath, destFile)

	bytesWritten, err := io.Copy(destFile, src)
	if err != nil {
		if fm.logger != nil {
			fm.logger.Error("复制文件内容失败",
				slog.String("dest_path", SanitizePath(destPath)),
				slog.String("error", err.Error()),
				slog.String("timestamp", time.Now().Format(time.RFC3339)),
			)
		}
		return fmt.Errorf("复制文件内容失败: %v", err)
	}

	if fm.logger != nil {
		fm.logger.Info("文件保存成功",
			slog.String("dest_path", SanitizePath(destPath)),
			slog.Int64("bytes_written", bytesWritten),
			slog.String("timestamp", time.Now().Format(time.RFC3339)),
		)
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
	cleanupCount := 0
	for path, file := range fm.openFiles {
		if file != nil {
			file.Close()
		}
		delete(fm.openFiles, path)
		cleanupCount++
	}

	if fm.logger != nil {
		fm.logger.Info("文件清理完成",
			slog.Int("cleanup_count", cleanupCount),
			slog.String("timestamp", time.Now().Format(time.RFC3339)),
		)
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