#!/bin/bash

# Telegram Bot 启动脚本

# 检查BOT_TOKEN环境变量
if [ -z "$BOT_TOKEN" ]; then
    echo "❌ 错误: BOT_TOKEN 环境变量未设置"
    echo "请运行: export BOT_TOKEN='你的Bot Token'"
    exit 1
fi

# 设置默认的临时目录
if [ -z "$TEMP_DIR" ]; then
    export TEMP_DIR="/tmp/tgbot"
fi

echo "🤖 启动Telegram数据处理Bot..."
echo "📂 临时目录: $TEMP_DIR"
echo "🔑 Bot Token: ${BOT_TOKEN:0:10}..."

# 创建临时目录
mkdir -p "$TEMP_DIR"

# 启动Bot
go run main.go