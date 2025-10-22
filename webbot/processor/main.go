package processor

import (
	"fmt"
	"os"
	"path/filepath"
)

// ProgressCallback 进度回调函数类型
type ProgressCallback func(progress int, message string)

// ProcessLogParse 处理日志解析
func ProcessLogParse(inputFile, outputDir string, callback ProgressCallback) ([]string, error) {
	callback(10, "开始日志解析...")

	// 创建输出文件路径
	outputFile := filepath.Join(outputDir, "parsed_data.csv")

	// 调用实际的日志解析逻辑
	// 这里需要复用 tgbot 中的实际处理逻辑
	err := processLogFile(inputFile, outputFile, callback)
	if err != nil {
		return nil, err
	}

	callback(100, "日志解析完成")
	return []string{outputFile}, nil
}

// ProcessLockUser 处理用户锁定
func ProcessLockUser(inputFile, outputDir string, callback ProgressCallback) ([]string, error) {
	callback(10, "开始用户锁定处理...")

	// 生成输出文件
	sqlFile := filepath.Join(outputDir, "lockUser-db_user库.sql")
	redisFile := filepath.Join(outputDir, "lockUser-redis_db0.txt")

	// 调用实际的用户锁定处理逻辑
	err := processLockUserFile(inputFile, sqlFile, redisFile, callback)
	if err != nil {
		return nil, err
	}

	callback(80, "正在创建压缩包...")

	// 创建压缩包目录
	lockUserDir := filepath.Join(outputDir, "lockuser-files")
	if err := os.MkdirAll(lockUserDir, 0755); err != nil {
		return nil, fmt.Errorf("创建锁定文件目录失败: %v", err)
	}

	// 复制文件到压缩目录
	sqlDst := filepath.Join(lockUserDir, "lockUser-db_user库.sql")
	redisDst := filepath.Join(lockUserDir, "lockUser-redis_db0.txt")

	if err := copyFile(sqlFile, sqlDst); err != nil {
		return nil, fmt.Errorf("复制SQL文件失败: %v", err)
	}

	if err := copyFile(redisFile, redisDst); err != nil {
		return nil, fmt.Errorf("复制Redis文件失败: %v", err)
	}

	callback(90, "正在压缩文件...")

	// 压缩整个目录
	zipFile := filepath.Join(outputDir, "lockuser-files.zip")
	err = createZipFile(lockUserDir, zipFile, callback)
	if err != nil {
		return nil, fmt.Errorf("压缩文件失败: %v", err)
	}

	// 返回压缩文件的相对路径
	taskID := filepath.Base(filepath.Dir(outputDir))
	relativeZipPath := filepath.Join(taskID, "output", "lockuser-files.zip")

	callback(100, "用户锁定处理完成，文件已打包压缩")
	return []string{relativeZipPath}, nil
}

// ProcessSQLParse 处理SQL解析
func ProcessSQLParse(inputFile, outputDir string, callback ProgressCallback) ([]string, error) {
	callback(10, "开始SQL解析...")

	outputFile := filepath.Join(outputDir, "parsed_sql.log")

	err := processSQLFile(inputFile, outputFile, callback)
	if err != nil {
		return nil, err
	}

	callback(100, "SQL解析完成")
	return []string{outputFile}, nil
}

// ProcessFileSplit 处理文件分割
func ProcessFileSplit(inputFile, outputDir string, callback ProgressCallback) ([]string, error) {
	callback(10, "开始文件分割...")

	outputFiles, err := processFileSplitLogic(inputFile, outputDir, callback)
	if err != nil {
		return nil, err
	}

	callback(100, "文件分割完成")
	return outputFiles, nil
}

// ProcessKYCReview 处理KYC审核
func ProcessKYCReview(inputFile, outputDir string, callback ProgressCallback) ([]string, error) {
	callback(10, "开始KYC审核处理...")

	// 生成带日期的文件名
	outputFile := filepath.Join(outputDir, fmt.Sprintf("kyc-%s.sql", getCurrentDateString()))

	err := processKYCFile(inputFile, outputFile, callback)
	if err != nil {
		return nil, err
	}

	callback(100, "KYC审核处理完成")
	return []string{outputFile}, nil
}

// ProcessRedisDel 处理Redis删除
func ProcessRedisDel(inputFile, outputDir string, callback ProgressCallback) ([]string, error) {
	callback(10, "开始Redis删除命令生成...")

	outputFiles, err := processRedisDelLogic(inputFile, outputDir, callback)
	if err != nil {
		return nil, err
	}

	callback(100, "Redis删除命令生成完成")
	return outputFiles, nil
}

// ProcessRedisAdd 处理Redis增加
func ProcessRedisAdd(inputFile, outputDir string, callback ProgressCallback) ([]string, error) {
	callback(10, "开始Redis增加命令生成...")

	outputFile := filepath.Join(outputDir, "redis_add_commands.txt")

	err := processRedisAddLogic(inputFile, outputFile, callback)
	if err != nil {
		return nil, err
	}

	callback(100, "Redis增加命令生成完成")
	return []string{outputFile}, nil
}

// ProcessUIDDedup 处理UID去重
func ProcessUIDDedup(inputFile, outputDir string, callback ProgressCallback) ([]string, error) {
	callback(10, "开始UID去重处理...")

	outputFile := filepath.Join(outputDir, "dedup_uids.csv")
	reportFile := filepath.Join(outputDir, "dedup_report.txt")

	err := processUIDDedupLogic(inputFile, outputFile, reportFile, callback)
	if err != nil {
		return nil, err
	}

	callback(80, "正在创建压缩包...")

	// 创建压缩包目录
	dedupDir := filepath.Join(outputDir, "dedup-files")
	if err := os.MkdirAll(dedupDir, 0755); err != nil {
		return nil, fmt.Errorf("创建去重文件目录失败: %v", err)
	}

	// 复制文件到压缩目录
	uidDst := filepath.Join(dedupDir, "dedup_uids.csv")
	reportDst := filepath.Join(dedupDir, "dedup_report.txt")

	if err := copyFile(outputFile, uidDst); err != nil {
		return nil, fmt.Errorf("复制去重文件失败: %v", err)
	}

	if err := copyFile(reportFile, reportDst); err != nil {
		return nil, fmt.Errorf("复制报告文件失败: %v", err)
	}

	callback(90, "正在压缩文件...")

	// 压缩整个目录
	zipFile := filepath.Join(outputDir, "dedup-files.zip")
	err = createZipFile(dedupDir, zipFile, callback)
	if err != nil {
		return nil, fmt.Errorf("压缩文件失败: %v", err)
	}

	// 返回压缩文件的相对路径
	taskID := filepath.Base(filepath.Dir(outputDir))
	relativeZipPath := filepath.Join(taskID, "output", "dedup-files.zip")

	callback(100, "UID去重处理完成，文件已打包压缩")
	return []string{relativeZipPath}, nil
}