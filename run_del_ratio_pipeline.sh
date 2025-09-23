#!/bin/bash

# 脚本描述：执行Redis命令生成和处理的完整流程
# 作者：自动生成
# 日期：$(date)

set -e  # 如果任何命令失败，脚本将退出

rm -rf multi-redis*/*

echo "开始执行Redis命令生成和处理流程..."

# 步骤1：执行 go generate_redis_commands.go
echo "步骤1：执行 go generate_redis_commands.go"
go run generate_redis_commands.go
if [ $? -eq 0 ]; then
    echo "✓ 步骤1完成：成功生成Redis命令"
else
    echo "✗ 步骤1失败：generate_redis_commands.go执行失败"
    exit 1
fi

# 步骤2：移动redis_delete_commands.txt到multi-redis目录
echo "步骤2：移动redis_delete_commands.txt到multi-redis目录"
mv redis_delete_commands.txt multi-redis/redis_commands.txt
if [ $? -eq 0 ]; then
    echo "✓ 步骤2完成：成功移动redis_delete_commands.txt"
else
    echo "✗ 步骤2失败：文件移动失败"
    exit 1
fi

# 步骤3：执行 go run main.go
echo "步骤3：执行 go run main.go"
go run main.go 4
if [ $? -eq 0 ]; then
    echo "✓ 步骤3完成：main.go执行成功"
else
    echo "✗ 步骤3失败：main.go执行失败"
    exit 1
fi

# 步骤4：复制execute_redis_commands.sh到multi-redis-split目录
echo "步骤4：复制execute_redis_commands.sh到multi-redis-split目录"
cp execute_redis_commands.sh multi-redis-split/
if [ $? -eq 0 ]; then
    echo "✓ 步骤4完成：成功复制execute_redis_commands.sh"
else
    echo "✗ 步骤4失败：文件复制失败"
    exit 1
fi

# 步骤5：压缩multi-redis-split文件夹
echo "步骤5：压缩multi-redis-split文件夹"
zip -r multi-redis-split.zip multi-redis-split/
if [ $? -eq 0 ]; then
    echo "✓ 步骤5完成：成功压缩multi-redis-split文件夹为multi-redis-split.zip"
else
    echo "✗ 步骤5失败：文件夹压缩失败"
    exit 1
fi

echo ""
echo "🎉 所有步骤执行完成！"
echo "生成的压缩文件：multi-redis-split.zip"
