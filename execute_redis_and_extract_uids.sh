#!/bin/bash

# Redis命令执行脚本 - 执行redis_commands_*.txt文件并提取成功的用户ID
# 使用方法: ./execute_redis_and_extract_uids.sh <redis_host>  [redis_port] [redis_db]

# 检查参数
if [ $# -lt 1 ]; then
    echo "使用方法: $0 <redis_host>  [redis_port] [redis_db]"
    echo "例如: $0 127.0.0.1  6379 2"
    exit 1
fi

REDIS_HOST=$1
REDIS_PASSWORD=""
REDIS_PORT=${2:-6379}
REDIS_DB=${3:-0}
CURRENT_DIR=$(cd "$(dirname "$0")" && pwd)

echo "开始执行Redis命令并提取成功的用户ID..."
echo "Redis主机: $REDIS_HOST:$REDIS_PORT"
echo "Redis数据库: $REDIS_DB"
echo "当前目录: $CURRENT_DIR"
echo "================================"

# 构建redis-cli命令
REDIS_CLI="redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB"
if [ -n "$REDIS_PASSWORD" ]; then
    REDIS_CLI="$REDIS_CLI -a $REDIS_PASSWORD"
fi

# 创建输出文件
TIMESTAMP=$(date +%Y%m%d)
SUCCESS_UIDS_FILE="successful_uids_${TIMESTAMP}.txt"

echo "成功的用户ID将保存到: $SUCCESS_UIDS_FILE"
echo "================================"

# 统计变量
total_files=0
total_commands=0
successful_uids=0
failed_commands=0

# 获取所有redis_commands_*.txt文件
files=$(ls -1 ${CURRENT_DIR}/redis_commands_*.txt 2>/dev/null | sort -V)

if [ -z "$files" ]; then
    echo "错误: 在当前目录中没有找到redis_commands_*.txt文件"
    exit 1
fi

# 计算总文件数
total_files=$(echo "$files" | wc -l)
echo "找到 $total_files 个Redis命令文件"
echo "================================"

# 记录开始时间
start_time=$(date)

# 逐个处理文件
for file in $files; do
    filename=$(basename "$file")
    echo "正在处理: $filename"
    
    # 检查文件是否存在且不为空
    if [ ! -f "$file" ] || [ ! -s "$file" ]; then
        echo "  ⚠️  文件不存在或为空，跳过"
        continue
    fi
    
    # 逐行执行命令
    while IFS= read -r command; do
        # 跳过空行
        if [ -z "$command" ] || [ "$command" = "" ]; then
            continue
        fi
        
        ((total_commands++))
        
        # 显示进度（每1000条命令显示一次）
        if [ $((total_commands % 1000)) -eq 0 ]; then
            echo "  已处理 $total_commands 条命令，成功提取 $successful_uids 个用户ID..."
        fi
        
        # 执行Redis命令
        if result=$(echo "$command" | $REDIS_CLI); then
            # 检查是否是risk:turnover:req删除命令且执行成功
            if [[ "$command" =~ ^del[[:space:]]+risk:turnover:req:\{([0-9]+)\}$ ]] && [[ "$result" == "1" ]]; then
                uid="${BASH_REMATCH[1]}"
                echo "$uid" >> "$SUCCESS_UIDS_FILE"
                ((successful_uids++))
            fi
        else
            ((failed_commands++))
        fi
        
        # 添加短暂延迟，避免对Redis造成过大压力
        sleep 0.01
        
    done < "$file"
    
    echo "  ✅ 文件处理完成: $filename"
done

# 记录结束时间
end_time=$(date)

echo "================================"
echo "执行完成!"
echo "开始时间: $start_time"
echo "结束时间: $end_time"
echo "================================"
echo "处理文件数: $total_files"
echo "总命令数: $total_commands"
echo "成功提取的用户ID数: $successful_uids"
echo "失败命令数: $failed_commands"
echo "================================"
echo "成功的用户ID文件: $SUCCESS_UIDS_FILE"

if [ $successful_uids -gt 0 ]; then
    echo "🎉 成功提取 $successful_uids 个用户ID!"
    
    # 显示前10个成功的用户ID作为示例
    echo ""
    echo "前10个成功的用户ID示例:"
    head -10 "$SUCCESS_UIDS_FILE" | while read -r uid; do
        echo "  - $uid"
    done
    
    if [ $successful_uids -gt 10 ]; then
        echo "  ... 还有 $((successful_uids - 10)) 个用户ID"
    fi
else
    echo "⚠️  没有成功提取到任何用户ID"
fi

echo ""
