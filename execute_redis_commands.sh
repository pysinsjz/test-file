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
REDIS_PORT=$3
CURRENT_DIR=$(cd "$(dirname "$0")" && pwd)

echo "开始执行Redis命令导入..."
echo "Redis主机: $REDIS_HOST"
echo "Redis端口: $REDIS_PORT"
echo "当前目录: $CURRENT_DIR"
echo "================================"

# 统计变量
total_files=0
success_files=0
failed_files=0

# 获取所有redis_commands_part_*.txt文件并按数字顺序排序（排除已完成的文件）
files=$(ls -1 ${CURRENT_DIR}/redis_commands_part_*.txt 2>/dev/null | grep -v "_done\.txt$" | sort -V)

if [ -z "$files" ]; then
    echo "错误: 在当前目录中没有找到redis_commands_part_*.txt文件"
    exit 1
fi

# 计算总文件数
total_files=$(echo "$files" | wc -l)
echo "找到 $total_files 个文件需要处理"
echo "================================"

# 逐个处理文件
for file in $files; do
    filename=$(basename "$file")
    echo "正在处理: $filename"
    
    # 检查文件是否存在且不为空
    if [ ! -f "$file" ] || [ ! -s "$file" ]; then
        echo "  ⚠️  文件不存在或为空，跳过"
        ((failed_files++))
        continue
    fi
    
    # 执行redis命令
    if cat "$file" | redis-cli  -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" -n 0; then
        echo "  ✅ 成功导入: $filename"
        # 重命名文件为 {filename}_done
        file_dir=$(dirname "$file")
        new_filename="${filename}_done"
        new_filepath="${file_dir}/${new_filename}"
        if mv "$file" "$new_filepath"; then
            echo "  📝 文件已重命名为: $new_filename"
        else
            echo "  ⚠️  文件重命名失败: $filename"
        fi
        ((success_files++))
    else
        echo "  ❌ 导入失败: $filename"
        ((failed_files++))
        
        # 询问是否继续
        echo "是否继续执行剩余文件? (y/n): "
        read -r response
        if [ "$response" != "y" ] && [ "$response" != "Y" ]; then
            echo "用户选择停止执行"
            break
        fi
    fi
    
    # 添加短暂延迟，避免对Redis造成过大压力
    sleep 1
done

echo "================================"
echo "执行完成!"
echo "总文件数: $total_files"
echo "成功导入: $success_files"
echo "失败文件: $failed_files"

if [ $failed_files -eq 0 ]; then
    echo "🎉 所有文件都已成功导入Redis!"
    exit 0
else
    echo "⚠️  有 $failed_files 个文件导入失败，请检查错误信息"
    exit 1
fi 