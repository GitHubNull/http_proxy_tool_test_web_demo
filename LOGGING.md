# 日志管理指南

本文档详细介绍项目的日志管理系统，包括配置、使用方法和最佳实践。

## 📋 日志系统概述

项目采用自建的日志管理系统，具备以下特性：

- ✅ **按日期分割**：每天自动创建新的日志文件
- ✅ **大小轮转**：当文件超过指定大小时自动轮转
- ✅ **自动压缩**：旧日志文件自动gzip压缩节省空间
- ✅ **自动清理**：按时间和数量自动清理过期日志
- ✅ **分级记录**：支持DEBUG、INFO、WARN、ERROR、FATAL五个级别
- ✅ **实时监控**：提供日志统计和监控API
- ✅ **双重输出**：同时输出到文件和控制台（可配置）

## 🗂️ 日志文件结构

```
logs/
├── server-2025-01-15.log              # 当天的主日志文件
├── server-2025-01-15-14-30-25.log.gz  # 轮转并压缩的日志文件
├── server-2025-01-14.log.gz           # 昨天的日志（已压缩）
├── server-2025-01-13.log.gz           # 前天的日志（已压缩）
├── nohup.out                          # 系统启动日志
└── ...
```

### 文件命名规则

- **当天日志**：`server-YYYY-MM-DD.log`
- **轮转日志**：`server-YYYY-MM-DD-HH-MM-SS.log.gz`
- **压缩日志**：原文件名 + `.gz` 后缀

## ⚙️ 配置选项

### 命令行参数

程序支持通过命令行参数直接指定日志目录：

```bash
# 使用默认日志目录 (logs/)
./proxy-test-tool

# 指定自定义日志目录
./proxy-test-tool -log-dir /var/log/proxy-test

# 其他常用参数组合
./proxy-test-tool -port 9090 -log-dir /tmp/logs
```

### 环境变量配置

在 `.env` 文件中配置日志相关参数：

```bash
# 日志级别：DEBUG, INFO, WARN, ERROR, FATAL
LOG_LEVEL=info

# 日志目录路径
LOGS_PATH=logs

# 单个日志文件最大大小（MB）
LOG_MAX_SIZE=50

# 最大备份文件数量
LOG_MAX_BACKUPS=30

# 最大保留天数
LOG_MAX_AGE=30

# 是否压缩旧日志文件
LOG_COMPRESS=true

# 是否同时输出到控制台
LOG_CONSOLE=true
```

### 默认配置值

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `LOG_LEVEL` | `INFO` | 日志级别，低于此级别的日志不会记录 |
| `LOGS_PATH` | `logs` | 日志文件存储目录 |
| `LOG_MAX_SIZE` | `50` | 单个日志文件最大大小（MB） |
| `LOG_MAX_BACKUPS` | `30` | 最多保留的备份文件数量 |
| `LOG_MAX_AGE` | `30` | 日志文件最大保留天数 |
| `LOG_COMPRESS` | `true` | 是否自动压缩旧日志文件 |
| `LOG_CONSOLE` | `true` | 是否同时输出到控制台 |

## 🔧 使用方法

### 命令行操作

#### 查看日志

```bash
# 查看当天日志（实时跟踪）
make logs

# 查看所有日志文件
make logs-all

# 查看日志统计信息
make logs-stats

# 实时查看日志（指定行数）
make logs-tail LINES=200
```

#### 搜索日志

```bash
# 搜索日志内容
make logs-search SEARCH="错误关键词"

# 搜索特定IP的访问记录
make logs-search SEARCH="192.168.1.100"

# 搜索特定时间段的日志
make logs-search SEARCH="2025-01-15 14:"
```

#### 维护操作

```bash
# 清理7天前的旧日志
make logs-clean

# 压缩7天前的日志文件
make logs-compress
```

### API接口

#### 获取日志统计信息

```http
GET /api/logs/stats
```

**响应示例：**
```json
{
  "status": "success",
  "data": {
    "total_files": 15,
    "total_size_mb": 125.6,
    "compressed_files": 12,
    "oldest_log_date": "2025-01-01",
    "newest_log_date": "2025-01-15",
    "log_directory": "build/logs",
    "max_file_size_mb": 50,
    "max_backups": 30,
    "max_age_days": 30,
    "compression_enabled": true,
    "current_date": "2025-01-15"
  }
}
```

#### 手动触发日志清理

```http
POST /api/logs/cleanup
```

**响应示例：**
```json
{
  "status": "success",
  "message": "日志清理任务已启动"
}
```

## 📊 日志格式

### 标准日志格式

```
[2025-01-15 14:30:25] [INFO] 服务器启动在端口 8080
[2025-01-15 14:30:25] [INFO] 版本: v1.0.0, 构建时间: 2025-01-15T06:30:00Z
[2025-01-15 14:30:30] [INFO] 新的WebSocket连接: 192.168.1.100:54321
[2025-01-15 14:30:35] [WARN] 请求处理时间过长: 2.5s
[2025-01-15 14:30:40] [ERROR] 数据库连接失败: connection timeout
```

### 日志级别说明

- **DEBUG**：调试信息，详细的程序执行流程
- **INFO**：一般信息，正常的程序运行状态
- **WARN**：警告信息，可能的问题或异常情况
- **ERROR**：错误信息，程序运行错误但不影响继续执行
- **FATAL**：致命错误，程序无法继续执行

## 🚀 自动化功能

### 日志轮转

系统会在以下情况自动进行日志轮转：

1. **日期变化**：每天0点自动创建新的日志文件
2. **文件大小**：当文件大小超过`LOG_MAX_SIZE`设置的值时
3. **手动触发**：通过API接口手动触发

### 自动清理

系统每6小时自动执行一次日志清理，清理策略：

1. **按时间清理**：删除超过`LOG_MAX_AGE`天的日志文件
2. **按数量清理**：保留最新的`LOG_MAX_BACKUPS`个备份文件
3. **压缩处理**：自动压缩轮转的日志文件

## 🔍 监控和诊断

### 日志统计监控

定期检查日志统计信息：

```bash
# 查看详细统计
make logs-stats

# 检查日志目录大小
du -sh logs/

# 检查最新日志
ls -lt logs/ | head -5
```

### 常见问题诊断

#### 1. 日志文件过大

**现象**：日志目录占用大量磁盘空间

**解决方案**：
```bash
# 检查大文件
find logs/ -name "*.log*" -size +100M

# 清理旧日志
make logs-clean

# 压缩现有日志
make logs-compress

# 调整配置
echo "LOG_MAX_SIZE=20" >> .env  # 减少单文件大小
echo "LOG_MAX_AGE=7" >> .env    # 减少保留天数
```

#### 2. 日志轮转失败

**现象**：日志文件持续增长，没有进行轮转

**诊断步骤**：
```bash
# 检查文件权限
ls -la logs/

# 检查磁盘空间
df -h

# 查看系统日志
tail -f logs/nohup.out
```

#### 3. 日志丢失

**现象**：找不到某个时间段的日志

**查找步骤**：
```bash
# 检查所有日志文件
make logs-all

# 搜索压缩文件
find logs/ -name "*.gz" -exec zgrep "关键词" {} \;

# 检查系统日志
tail -f /var/log/syslog | grep proxy-test
```

## 📋 最佳实践

### 生产环境配置

```bash
# 生产环境推荐配置
LOG_LEVEL=warn           # 减少日志量
LOG_MAX_SIZE=100         # 增大单文件大小
LOG_MAX_BACKUPS=50       # 增加备份数量
LOG_MAX_AGE=90          # 延长保留时间
LOG_COMPRESS=true        # 必须开启压缩
LOG_CONSOLE=false        # 关闭控制台输出
```

### 开发环境配置

```bash
# 开发环境推荐配置
LOG_LEVEL=debug          # 显示所有日志
LOG_MAX_SIZE=10          # 小文件便于查看
LOG_MAX_BACKUPS=10       # 减少备份数量
LOG_MAX_AGE=7           # 短期保留
LOG_COMPRESS=false       # 便于直接查看
LOG_CONSOLE=true         # 控制台实时显示
```

### 定期维护

建议设置定期维护任务：

```bash
# 添加到crontab（每天凌晨2点清理日志）
0 2 * * * cd /path/to/project && make logs-clean

# 每周压缩日志
0 3 * * 0 cd /path/to/project && make logs-compress

# 每月检查磁盘使用
0 4 1 * * cd /path/to/project && make logs-stats
```

## 🚨 告警配置

### 磁盘空间告警

```bash
#!/bin/bash
# 检查日志目录大小
LOG_DIR="logs"
MAX_SIZE_GB=5

SIZE_GB=$(du -s $LOG_DIR | awk '{print int($1/1024/1024)}')
if [ $SIZE_GB -gt $MAX_SIZE_GB ]; then
    echo "警告：日志目录大小超过限制 ${SIZE_GB}GB > ${MAX_SIZE_GB}GB"
    # 发送告警通知
fi
```

### 日志错误告警

```bash
#!/bin/bash
# 检查错误日志
ERROR_COUNT=$(grep -c "ERROR\|FATAL" logs/server-$(date +%Y-%m-%d).log 2>/dev/null || echo 0)
if [ $ERROR_COUNT -gt 10 ]; then
    echo "警告：发现过多错误日志 $ERROR_COUNT 条"
    # 发送告警通知
fi
```

## 🔧 故障排除

### 常见错误

#### 权限问题

```bash
# 检查权限
ls -la logs/

# 修复权限
chmod 755 logs/
chmod 644 logs/*.log*
```

#### 磁盘空间不足

```bash
# 检查磁盘空间
df -h

# 紧急清理
find logs/ -name "*.log*" -mtime +3 -delete
```

#### 日志系统未初始化

检查环境变量配置和程序启动日志：

```bash
# 检查配置
env | grep LOG_

# 查看启动日志
tail -f logs/nohup.out
```

## 📞 技术支持

如果遇到日志相关问题：

1. **查看日志统计**：`make logs-stats`
2. **检查系统日志**：`tail -f build/logs/nohup.out`
3. **搜索错误信息**：`make logs-search SEARCH="error"`
4. **提交Issue**：在GitHub项目中创建Issue，附上相关日志信息

---

💡 **小贴士**：定期监控日志系统状态，及时发现和解决潜在问题，确保系统稳定运行。 