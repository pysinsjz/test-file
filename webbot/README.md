# WebBot - 数据处理Web工具

> 基于 Go + Gin 的数据处理Web应用，提供友好的界面来处理各种数据文件

## ✨ 功能特色

### 🎯 核心功能
- **日志解析** - 从应用程序日志中提取结构化数据
- **用户锁定** - 批量生成用户账户锁定命令
- **SQL解析** - 从日志中提取并去重SQL语句
- **文件分割** - 将大文件按行数分割成小文件
- **KYC审核** - 处理身份验证审核数据
- **Redis操作** - 生成Redis删除/添加命令
- **UID去重** - 从用户ID列表中移除重复项

### 🌟 界面特点
- **拖拽上传** - 支持文件拖拽上传，操作便捷
- **实时进度** - 处理过程实时显示进度和状态
- **响应式设计** - 适配桌面和移动设备
- **友好提示** - 详细的使用说明和错误提示

## 🚀 快速开始

### 环境要求
- Go 1.21+
- 现代浏览器 (Chrome/Firefox/Safari/Edge)

### 安装依赖
```bash
cd webbot
go mod tidy
```

### 启动服务
```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动

### 访问应用
1. 打开浏览器访问 `http://localhost:8080`
2. 选择需要的功能
3. 上传文件并等待处理
4. 下载结果文件

## 📁 项目结构

```
webbot/
├── main.go              # 应用入口
├── go.mod               # 依赖管理
├── handlers/            # HTTP处理器
│   ├── base.go         # 基础功能和数据结构
│   └── file.go         # 文件处理相关
├── templates/           # HTML模板
│   ├── layout.html     # 基础布局
│   ├── index.html      # 首页
│   ├── upload.html     # 上传页面
│   ├── result.html     # 结果页面
│   ├── help.html       # 帮助页面
│   └── error.html      # 错误页面
├── static/             # 静态资源
│   ├── css/
│   │   └── style.css   # 样式文件
│   └── js/
│       └── main.js     # JavaScript功能
├── processor/          # 数据处理逻辑
│   ├── main.go        # 处理器接口
│   └── impl.go        # 具体实现
├── utils/              # 工具函数
│   └── file.go        # 文件操作工具
└── uploads/            # 临时文件目录
```

## 🔧 API 接口

### 文件上传
```
POST /api/upload
Content-Type: multipart/form-data

Parameters:
- file: 要处理的文件
- function: 功能类型 (logparse/lockuser/sqlparse等)

Response:
{
  "task_id": "task_1234567890",
  "message": "文件上传成功"
}
```

### 查询进度
```
GET /api/progress/:taskid

Response:
{
  "id": "task_1234567890",
  "function": "logparse",
  "status": "processing",
  "progress": 50,
  "message": "正在处理...",
  "start_time": "2025-01-01T12:00:00Z"
}
```

### 下载文件
```
GET /api/download/:filename
```

## 📝 支持的文件格式

| 功能 | 输入格式 | 输出格式 | 文件大小限制 |
|------|----------|----------|--------------|
| 日志解析 | TXT | CSV | 50MB |
| 用户锁定 | CSV | SQL + Redis命令 | 50MB |
| SQL解析 | TXT | SQL文件 | 50MB |
| 文件分割 | 任意格式 | 多个小文件 | 50MB |
| KYC审核 | Excel/CSV | SQL文件 | 50MB |
| Redis删除 | Excel/CSV | Redis命令 | 50MB |
| Redis增加 | CSV | Redis命令 | 50MB |
| UID去重 | CSV | CSV + 报告 | 50MB |

## 🛠️ 开发说明

### 添加新功能
1. 在 `handlers/base.go` 的 `Functions` map 中添加功能定义
2. 在 `processor/main.go` 中添加处理函数
3. 在 `processor/impl.go` 中实现具体逻辑
4. 更新模板和静态资源（如需要）

### 自定义样式
编辑 `static/css/style.css` 文件来自定义界面样式

### 自定义功能
编辑 `static/js/main.js` 文件来添加前端交互功能

## 🔒 安全特性

- **文件类型验证** - 严格验证上传文件的类型
- **文件大小限制** - 防止过大文件占用系统资源
- **路径安全检查** - 防止目录遍历攻击
- **自动清理** - 处理完成后自动清理临时文件
- **会话隔离** - 每个用户的文件独立存储

## 📊 性能优化

- **异步处理** - 文件处理采用异步方式，不阻塞界面
- **进度反馈** - 实时显示处理进度，提升用户体验
- **资源管理** - 自动管理文件句柄和内存使用
- **缓存优化** - 静态资源使用浏览器缓存

## 🐛 故障排除

### 常见问题

**Q: 文件上传失败**
A: 检查文件大小是否超过50MB，文件格式是否正确

**Q: 处理时间过长**
A: 大文件需要更多时间，请耐心等待或刷新页面查看状态

**Q: 下载文件失败**
A: 可能文件已过期，请重新处理

### 日志查看
应用日志会输出到控制台，包含详细的处理信息和错误信息

## 📞 技术支持

- **问题反馈**: 创建 GitHub Issue
- **功能建议**: 提交 Pull Request
- **技术讨论**: 联系开发团队

## 📄 许可证

本项目采用 MIT 许可证，详情请查看 LICENSE 文件。

---

**💡 提示**: 这个Web版本提供了比Telegram Bot更友好的用户界面，同时保持了所有原有功能的完整性。