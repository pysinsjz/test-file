#!/bin/bash

# Redis批量导入脚本
# 使用方法: ./execute_redis_commands.sh <redis_host>
# 例如: ./execute_redis_commands.sh 127.0.0.1

# 检查参数
if [ $# -eq 0 ]; then
    echo "错误: 请提供Redis主机地址"
    echo "使用方法: $0 <redis_host>"
    echo "例如: $0 127.0.0.1"
    exit 1
fi

REDIS_HOST=$1
REDIS_PASSWORD=$2
REDIS_PORT=6379
CURRENT_DIR=$(cd "$(dirname "$0")" && pwd)

# 创建日志文件（带时间戳）
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
LOG_FILE="${CURRENT_DIR}/redis_import_${TIMESTAMP}.log"

# 日志函数：同时输出到终端和文件
log() {
    echo "$1" | tee -a "$LOG_FILE"
}

log "开始执行Redis命令导入..."
log "Redis主机: $REDIS_HOST"
log "Redis端口: $REDIS_PORT"
log "当前目录: $CURRENT_DIR"
log "日志文件: $LOG_FILE"
log "================================"

# 统计变量
total_files=0
success_files=0
failed_files=0

# 获取所有redis_commands_part_*.txt文件并按数字顺序排序
files=$(ls -1 ${CURRENT_DIR}/redis_commands_part_*.txt 2>/dev/null | sort -V)

if [ -z "$files" ]; then
    log "错误: 在当前目录中没有找到redis_commands_part_*.txt文件"
    exit 1
fi

# 计算总文件数
total_files=$(echo "$files" | wc -l)
log "找到 $total_files 个文件需要处理"
log "================================"

# 逐个处理文件
for file in $files; do
    filename=$(basename "$file")
    log "正在处理: $filename"
    
    # 检查文件是否存在且不为空
    if [ ! -f "$file" ] || [ ! -s "$file" ]; then
        log "  ⚠️  文件不存在或为空，跳过"
        ((failed_files++))
        continue
    fi
    
    # 执行redis命令并记录输出
    log "  开始执行Redis命令..."
    if cat "$file" | redis-cli  -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD"  >> "$LOG_FILE" 2>&1; then
        log "  ✅ 成功导入: $filename"
        ((success_files++))
    else
        log "  ❌ 导入失败: $filename"
        ((failed_files++))
        
        # 询问是否继续
        log "是否继续执行剩余文件? (y/n): "
        read -r response
        if [ "$response" != "y" ] && [ "$response" != "Y" ]; then
            log "用户选择停止执行"
            break
        fi
    fi
    
    # 添加短暂延迟，避免对Redis造成过大压力
    sleep 0.1
done

log "================================"
log "执行完成!"
log "总文件数: $total_files"
log "成功导入: $success_files"
log "失败文件: $failed_files"

if [ $failed_files -eq 0 ]; then
    log "🎉 所有文件都已成功导入Redis!"
    log ""
    log "详细日志已保存到: $LOG_FILE"
    exit 0
else
    log "⚠️  有 $failed_files 个文件导入失败，请检查错误信息"
    log ""
    log "详细日志已保存到: $LOG_FILE"
    exit 1
fi 