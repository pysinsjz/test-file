#!/bin/bash

# WebBot 启动脚本

echo "🚀 启动 WebBot 数据处理服务..."
echo ""

# 检查 Go 是否已安装
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到 Go 环境，请先安装 Go 1.21+"
    exit 1
fi

# 检查 Go 版本
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
MIN_VERSION="1.21"

if [ "$(printf '%s\n' "$MIN_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$MIN_VERSION" ]; then
    echo "❌ 错误: Go 版本过低，需要 1.21+，当前版本: $GO_VERSION"
    exit 1
fi

# 进入项目目录
cd "$(dirname "$0")"

# 安装依赖
echo "📦 安装依赖..."
go mod tidy

if [ $? -ne 0 ]; then
    echo "❌ 错误: 依赖安装失败"
    exit 1
fi

# 创建必要的目录
mkdir -p uploads
mkdir -p uploads/temp

# 构建项目
echo "🔨 构建项目..."
go build -o webbot .

if [ $? -ne 0 ]; then
    echo "❌ 错误: 项目构建失败"
    exit 1
fi

# 启动服务
echo "✅ 构建完成，启动服务..."
echo ""
echo "📱 访问地址: http://localhost:8080"
echo "❓ 帮助文档: http://localhost:8080/help"
echo ""
echo "按 Ctrl+C 停止服务"
echo ""

./webbot