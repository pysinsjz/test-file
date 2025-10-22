#!/bin/bash

# 部署脚本 - 用于部署 webbot 到远程服务器
# 作者: 自动生成
# 日期: $(date +%Y-%m-%d)

set -e  # 遇到错误立即退出

# 配置变量
REMOTE_HOST="172.32.2.251"
REMOTE_USER="root"
REMOTE_PATH="/app/workspace/webbot"
LOCAL_PROJECT_DIR="./"
BINARY_NAME="webbot"
WEB_PORT="9088"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."

    if ! command -v go &> /dev/null; then
        log_error "Go 未安装或不在 PATH 中"
        exit 1
    fi

    if ! command -v ssh &> /dev/null; then
        log_error "SSH 未安装或不在 PATH 中"
        exit 1
    fi

    if ! command -v scp &> /dev/null; then
        log_error "SCP 未安装或不在 PATH 中"
        exit 1
    fi

    log_info "依赖检查完成"
}

# 步骤1: 打包 Linux 可执行文件
build_binary() {
    log_info "开始打包 Linux 可执行文件..."

    cd "$LOCAL_PROJECT_DIR"

    # 清理之前的构建
    if [ -f "$BINARY_NAME" ]; then
        rm -f "$BINARY_NAME"
        log_info "清理旧的二进制文件"
    fi

    # 构建 Linux 版本
    GOOS=linux GOARCH=amd64 go build -o "$BINARY_NAME" main.go

    if [ -f "$BINARY_NAME" ]; then
        log_info "构建成功: $BINARY_NAME"
        ls -lh "$BINARY_NAME"
    else
        log_error "构建失败"
        exit 1
    fi
}

# 步骤2: 停止远程 webbot 进程和端口占用
stop_remote_process() {
    log_info "停止远程 webbot 进程和端口占用..."

    ssh "$REMOTE_USER@$REMOTE_HOST" "
        # 查找并停止占用端口 $WEB_PORT 的进程
        PID=\$(lsof -ti:$WEB_PORT 2>/dev/null || true)

        if [ -n \"\$PID\" ]; then
            echo \"找到占用端口 $WEB_PORT 的进程 PID: \$PID\"
            kill -9 \$PID
            sleep 3
            echo \"端口 $WEB_PORT 占用进程已停止\"
        fi

        # 查找并停止 webbot 进程（排除 grep 进程）
        PID=\$(ps aux | grep '[w]ebbot' | grep -v grep | awk '{print \$2}')

        if [ -n \"\$PID\" ]; then
            echo \"找到 webbot 进程 PID: \$PID\"
            kill -9 \$PID

            # 等待进程优雅退出
            sleep 3

            # 检查进程是否还在运行
            if ps -p \$PID > /dev/null 2>&1; then
                echo \"进程未正常退出，强制终止\"
                kill -9 \$PID
            fi

            echo \"webbot 进程已停止\"
        else
            echo \"未找到运行中的 webbot 进程\"
        fi
    "

    log_info "远程进程停止完成"
}

# 步骤3: 上传必要文件到远程服务器
upload_files() {
    log_info "上传文件到远程服务器..."

    # 确保远程目录存在
    ssh "$REMOTE_USER@$REMOTE_HOST" "mkdir -p $REMOTE_PATH"

    # 上传二进制文件
    scp "$LOCAL_PROJECT_DIR/$BINARY_NAME" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/"

    # 上传执行脚本
    scp "$LOCAL_PROJECT_DIR/execute_redis_commands.sh" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/"

    # 上传静态文件目录
    if [ -d "$LOCAL_PROJECT_DIR/static" ]; then
        scp -r "$LOCAL_PROJECT_DIR/static" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/"
    fi

    # 上传模板文件目录
    if [ -d "$LOCAL_PROJECT_DIR/templates" ]; then
        scp -r "$LOCAL_PROJECT_DIR/templates" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/"
    fi

    if [ $? -eq 0 ]; then
        log_info "文件上传成功"
    else
        log_error "文件上传失败"
        exit 1
    fi

    # 设置执行权限
    ssh "$REMOTE_USER@$REMOTE_HOST" "
        chmod +x $REMOTE_PATH/$BINARY_NAME
        chmod +x $REMOTE_PATH/execute_redis_commands.sh
    "

    log_info "文件权限设置完成"
}

# 步骤4: 在服务器后台启动 webbot
start_remote_service() {
    log_info "在服务器后台启动 webbot..."

    ssh "$REMOTE_USER@$REMOTE_HOST" "
        cd $REMOTE_PATH

        # 检查文件是否存在
        if [ ! -f \"$BINARY_NAME\" ]; then
            echo \"错误: $BINARY_NAME 文件不存在\"
            exit 1
        fi

        # 创建uploads目录
        mkdir -p uploads

        # 后台启动服务
        nohup ./$BINARY_NAME > webbot.log 2>&1 &

        # 获取进程ID
        PID=\$!
        echo \"webbot 已启动，PID: \$PID\"

        # 等待一下确保启动成功
        sleep 5

        # 检查进程是否还在运行
        if ps -p \$PID > /dev/null 2>&1; then
            echo \"webbot 启动成功，PID: \$PID\"
        else
            echo \"webbot 启动失败，请检查日志\"
            tail -20 webbot.log
            exit 1
        fi
    "

    if [ $? -eq 0 ]; then
        log_info "webbot 服务启动成功"
    else
        log_error "webbot 服务启动失败"
        exit 1
    fi
}

# 验证部署
verify_deployment() {
    log_info "验证部署..."

    ssh "$REMOTE_USER@$REMOTE_HOST" "
        # 检查进程状态（排除 grep 进程）
        PID=\$(ps aux | grep '[w]ebbot' | grep -v grep | awk '{print \$2}')

        if [ -n \"\$PID\" ]; then
            echo \"✓ webbot 进程正在运行，PID: \$PID\"

            # 检查端口监听
            echo \"检查端口 $WEB_PORT 监听状态...\"
            if netstat -tlnp 2>/dev/null | grep :$WEB_PORT; then
                echo \"✓ 端口 $WEB_PORT 正在监听\"
            else
                echo \"✗ 端口 $WEB_PORT 未监听\"
            fi

            # 显示最近的日志
            echo \"最近的日志内容:\"
            tail -n 10 $REMOTE_PATH/webbot.log 2>/dev/null || echo \"日志文件不存在\"

            # 测试Web服务
            echo \"测试Web服务...\"
            if curl -s http://localhost:$WEB_PORT/ > /dev/null; then
                echo \"✓ Web服务响应正常\"
            else
                echo \"✗ Web服务无响应\"
            fi
        else
            echo \"✗ webbot 进程未运行\"
            exit 1
        fi
    "

    log_info "部署验证完成"
}

# 显示访问信息
show_access_info() {
    log_info "=========================================="
    log_info "🎉 WebBot 部署成功!"
    log_info "=========================================="
    log_info "访问地址: http://$REMOTE_HOST:$WEB_PORT"
    log_info "远程服务器: $REMOTE_USER@$REMOTE_HOST"
    log_info "部署路径: $REMOTE_PATH"
    log_info "二进制文件: $BINARY_NAME"
    log_info "日志文件: $REMOTE_PATH/webbot.log"
    log_info ""
    log_info "🔧 管理命令:"
    log_info "查看日志: ssh $REMOTE_USER@$REMOTE_HOST 'tail -f $REMOTE_PATH/webbot.log'"
    log_info "重启服务: ssh $REMOTE_USER@$REMOTE_HOST 'cd $REMOTE_PATH && pkill webbot && nohup ./webbot > webbot.log 2>&1 &'"
    log_info "=========================================="
}

# 主函数
main() {
    log_info "开始部署 webbot 到 $REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH"
    log_info "=========================================="

    # 检查依赖
    check_dependencies

    # 执行部署步骤
    build_binary
    stop_remote_process
    upload_files
    start_remote_service
    verify_deployment

    # 显示访问信息
    show_access_info
}

# 错误处理
trap 'log_error "部署过程中发生错误，请检查上述输出"; exit 1' ERR

# 执行主函数
main "$@"