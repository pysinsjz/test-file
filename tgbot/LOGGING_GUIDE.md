# 🔍 企业级日志系统

## 功能特性

### ✅ 已实现的日志功能
- **结构化日志记录** - 基于Go 1.21+ `log/slog` 包
- **自动日志轮转** - 按日期和文件大小自动分割
- **多级别日志** - DEBUG、INFO、WARN、ERROR、FATAL
- **JSON格式输出** - 便于日志分析和监控系统集成
- **性能指标记录** - 自动记录处理时间和吞吐量
- **敏感信息脱敏** - 自动处理文件路径等敏感信息
- **自动清理** - 配置保留天数，自动清理旧日志

## 配置选项

通过环境变量配置日志系统：

```bash
LOG_LEVEL=INFO           # 日志级别：DEBUG, INFO, WARN, ERROR
LOG_DIR=./logs          # 日志文件目录
LOG_MAX_SIZE=104857600  # 单个日志文件最大大小 (100MB)
LOG_JSON=true           # true: JSON格式, false: 文本格式
LOG_KEEP_DAYS=30        # 保留日志文件天数
```

## 日志文件结构

```
logs/
├── tgbot-2025-09-25.log      # 当天日志文件
├── tgbot-2025-09-24.log      # 历史日志文件
├── tgbot-2025-09-23-1.log    # 大文件分割后的编号文件
└── tgbot-2025-09-23-2.log
```

## 日志内容示例

### 系统启动日志
```json
{
  "time": "2025-09-25T16:06:11+08:00",
  "level": "INFO",
  "source": {"function": "main.main", "file": "main.go", "line": 37},
  "msg": "Telegram Bot启动",
  "version": "1.0.0",
  "log_level": "INFO",
  "log_dir": "./logs",
  "temp_dir": "/tmp/tgbot"
}
```

### 用户请求日志
```json
{
  "time": "2025-09-25T16:06:25+08:00",
  "level": "INFO",
  "msg": "用户请求",
  "user_id": 12345,
  "chat_id": 12345,
  "command": "/redisdel",
  "message": "命令参数: ",
  "timestamp": "2025-09-25T16:06:25+08:00"
}
```

### 性能指标日志
```json
{
  "time": "2025-09-25T16:08:30+08:00",
  "level": "INFO",
  "msg": "性能指标",
  "operation": "redis_delete_pipeline",
  "duration": "2.547s",
  "item_count": 1000,
  "user_id": 12345,
  "items_per_second": 392.64
}
```

### 错误日志
```json
{
  "time": "2025-09-25T16:10:15+08:00",
  "level": "ERROR",
  "msg": "操作错误",
  "user_id": 12345,
  "operation": "file_upload",
  "error": "文件格式不支持",
  "file_path": ".../user_12345/input.txt"
}
```

## 日志记录覆盖

### 🟢 已集成日志记录的模块
- **系统启动/关闭** - Bot生命周期事件
- **用户交互** - 消息接收、命令处理、回调查询
- **文件操作** - 上传、下载、创建、清理
- **Redis删除流程** - 完整的5步骤处理流程
- **错误处理** - 异常捕获和详细错误信息
- **性能监控** - 处理时间、吞吐量统计

### 🔄 日志轮转机制
- 按日期自动创建新日志文件
- 单文件超过100MB自动分割
- 自动清理超过保留天数的旧日志
- 优雅处理磁盘空间不足等异常情况

### 🛡️ 安全特性
- 敏感信息自动脱敏（文件路径、用户信息）
- 支持生产环境的安全配置
- 日志文件权限控制
- 防止日志注入攻击

## 使用场景

1. **问题排查** - 详细的错误日志和调用栈
2. **性能监控** - 处理时间和吞吐量分析
3. **用户行为分析** - 命令使用统计和用户活跃度
4. **系统监控** - 资源使用情况和健康状态
5. **审计合规** - 完整的操作记录和用户轨迹

## 集成建议

### ELK Stack集成
```bash
# Logstash配置示例
input {
  file {
    path => "/app/logs/tgbot-*.log"
    type => "tgbot"
    codec => "json"
  }
}
```

### 监控告警
```yaml
# Prometheus告警规则示例
- alert: TGBotHighErrorRate
  expr: rate(tgbot_errors_total[5m]) > 0.1
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "TGBot error rate is high"
```

通过这套企业级日志系统，您可以全面监控Bot的运行状态，快速定位问题，并获得详细的性能指标！