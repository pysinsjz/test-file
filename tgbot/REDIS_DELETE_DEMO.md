# 🎯 Redis流水删除功能演示

## 功能概述

**Redis流水删除** (`/redisdel`) 是一个完整的企业级流水删除操作流程，完全基于原项目的 `run_del_ratio_pipeline.sh` 脚本逻辑实现。

## 🔄 完整操作流程

### 步骤1: 生成Redis删除命令
- 读取Excel/CSV文件中的用户ID
- 为每个用户生成两条删除命令：
  - `del risk:turnover:req:{userID}`
  - `del risk:turnover:bet:{userID}`

### 步骤2: 创建工作目录
- 创建 `multi-redis` 目录
- 移动生成的命令文件到工作目录

### 步骤3: 文件分割处理
- 自动将命令文件按10,000行分割
- 生成多个 `redis_commands_part_*.txt` 文件
- 创建 `multi-redis-split` 目录存储分割文件

### 步骤4: 创建执行脚本
- 生成 `execute_redis_commands.sh` 脚本
- 包含完整的批量执行逻辑：
  - Redis客户端可用性检查
  - 文件自动发现和统计
  - 批量执行每个分割文件
  - 执行结果反馈
  - 总计统计信息

### 步骤5: ZIP压缩打包
- 将整个分割目录压缩为ZIP文件
- 包含所有分割文件和执行脚本
- 通过Telegram自动发送给用户

## 📦 输出内容

用户将收到一个完整的ZIP压缩包，包含：

```
redis-delete-commands.zip
├── redis_commands_part_0001.txt
├── redis_commands_part_0002.txt
├── redis_commands_part_000N.txt
└── execute_redis_commands.sh
```

## 🚀 使用方法

### 在Telegram中使用
1. 向Bot发送 `/redisdel` 命令
2. 上传包含用户ID的Excel或CSV文件
3. 等待Bot完成5步处理流程
4. 下载生成的ZIP压缩包

### 在服务器上执行
1. 解压下载的ZIP文件
2. 上传到Redis服务器
3. 确保redis-cli可用
4. 运行: `chmod +x execute_redis_commands.sh && ./execute_redis_commands.sh`

## ✨ 技术特色

### 🔥 实时进度反馈
每个步骤都有详细的进度提示：
- 📝 步骤1：生成Redis删除命令...
- 📁 步骤2：创建工作目录...
- ✂️ 步骤3：分割Redis命令文件...
- 📜 步骤4：创建Redis执行脚本...
- 🗜️ 步骤5：压缩文件包...

### 🛡️ 安全可靠
- 文件格式自动验证
- 表头识别和跳过
- 空数据过滤
- 错误处理和回滚
- 资源自动清理

### ⚡ 高性能处理
- 支持大文件处理
- 自动分割避免内存溢出
- 批量命令优化执行
- 并发安全设计

## 📊 处理能力

- **文件大小**: 支持50MB以内的Excel/CSV文件
- **用户数量**: 理论上无限制
- **分割策略**: 每10,000行自动分割
- **执行效率**: 批量执行最大化性能

## 🎉 企业级特性

这个功能完全复制了原项目的企业级流水删除流程：
- ✅ 完整的5步骤流程
- ✅ 自动文件分割和管理
- ✅ 生产就绪的执行脚本
- ✅ ZIP打包便于部署
- ✅ 详细的执行日志和统计

通过Telegram界面，复杂的企业级数据处理操作变得像聊天一样简单！