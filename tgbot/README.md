# Telegram Bot 数据处理工具

这个Telegram Bot集成了原项目的所有数据处理功能，可以通过聊天界面便捷地处理各种文件。

## 功能特性

### 已实现的功能
✅ **日志解析** (`/logparse`) - 解析TXT日志文件，提取结构化数据到CSV
✅ **用户锁定** (`/lockuser`) - 从CSV文件生成用户锁定SQL和Redis命令
✅ **SQL解析** (`/sqlparse`) - 智能去重SQL日志解析
✅ **文件分割** (`/filesplit`) - 大文件分割为10K行小文件，自动打包ZIP
✅ **KYC审核** (`/kycreview`) - KYC审核数据处理
✅ **Redis流水删除** (`/redisdel`) - 完整的Redis流水删除操作流程（生成命令→分割文件→创建执行脚本→ZIP打包）
✅ **Redis流水命令** (`/redisadd`) - 生成Redis流水设置命令
✅ **UID去重** (`/uiddedup`) - 用户ID去重处理

## 快速开始

### 环境要求
- Go 1.23.3+
- Telegram Bot Token

### 安装和配置

1. **克隆项目并安装依赖**
```bash
cd tgbot
go mod tidy
```

2. **配置环境变量**
```bash
export BOT_TOKEN="7247480117:AAHqrIcsj8a-4ALsHPslQMhvOp485TxDUCY"
export TEMP_DIR="/tmp/tgbot"  # 可选，默认为/tmp/tgbot
```

3. **运行Bot**
```bash
go run main.go
```

### 使用方法

1. **启动对话** - 向Bot发送 `/start` 查看功能菜单
2. **选择功能** - 点击按钮或输入命令选择需要的功能
3. **上传文件** - 按提示上传相应格式的文件
4. **获取结果** - Bot处理完成后会自动发送结果文件

## 支持的文件格式

| 功能 | 输入格式 | 输出格式 |
|------|----------|----------|
| 日志解析 | TXT | CSV |
| 用户锁定 | CSV | SQL + TXT |
| SQL解析 | TXT | LOG |
| 文件分割 | 任意 | 同原格式 |
| KYC审核 | CSV/XLSX | SQL |
| Redis删除 | CSV/XLSX | TXT |
| Redis流水 | CSV | TXT |
| UID去重 | CSV | CSV |

## 项目结构

```
tgbot/
├── main.go              # Bot主程序
├── config/              # 配置管理
│   └── config.go
├── handlers/            # 功能处理器
│   ├── manager.go       # 消息路由管理
│   ├── processes.go     # 处理流程协调
│   ├── log_parser.go    # 日志解析处理器
│   └── user_lock.go     # 用户锁定处理器
├── utils/               # 工具函数
│   ├── file_manager.go  # 文件管理
│   ├── common.go        # 通用工具
│   └── excel_helper.go  # Excel处理工具
└── go.mod               # Go模块定义
```

## 开发状态

当前项目处于开发阶段，已完成基础架构和部分核心功能。
- ✅ Bot框架和用户界面
- ✅ 文件上传下载机制
- ✅ 日志解析功能
- ✅ 用户锁定功能
- 🚧 其他功能正在逐步实现中

## 安全特性

- 用户文件隔离存储
- 自动文件清理机制
- 文件大小限制 (50MB)
- 优雅的错误处理

## 部署建议

### Docker部署 (推荐)
```dockerfile
FROM golang:1.23.3-alpine
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o bot main.go
CMD ["./bot"]
```

### 系统服务
可以使用systemd等工具将Bot设置为系统服务，确保稳定运行。

---

**注意**: 这是一个数据处理工具，请确保上传的文件不包含敏感信息。所有文件在处理完成后会自动清理。