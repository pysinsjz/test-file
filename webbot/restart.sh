#!/bin/bash

# WebBot 重启脚本

echo "🔄 正在重启 WebBot 服务..."

# 杀死占用9088端口的进程
echo "📡 检查端口 9088 占用情况..."
PID=$(lsof -ti:9088)
if [ ! -z "$PID" ]; then
    echo "⚡ 发现进程 $PID 占用端口 9088，正在停止..."
    kill -9 $PID
    sleep 2
    echo "✅ 进程已停止"
else
    echo "ℹ️  端口 9088 未被占用"
fi

# 清理可能存在的 webbot 进程
echo "🧹 清理 webbot 进程..."
pkill -f "webbot" || echo "ℹ️  没有找到 webbot 进程"

# 等待进程完全退出
sleep 1

# 构建项目
echo "🔨 构建 WebBot..."
go build
if [ $? -ne 0 ]; then
    echo "❌ 构建失败！"
    exit 1
fi

echo "✅ 构建成功"

# 启动服务
echo "🚀 启动 WebBot 服务..."
nohup ./webbot > webbot.log 2>&1 &

# 等待服务启动
sleep 3

# 检查服务是否成功启动
if lsof -ti:9088 > /dev/null; then
    echo "✅ WebBot 服务启动成功！"
    echo "📱 访问地址: http://0.0.0.0:9088"
    echo "📄 日志文件: webbot.log"
else
    echo "❌ WebBot 服务启动失败！"
    echo "📄 请查看日志文件: webbot.log"
    exit 1
fi