# 构建指南

本文档描述了项目的构建系统和目录结构。

## 📁 构建目录结构

项目采用结构化的构建目录，所有编译产物都存放在 `build/` 目录下：

```
build/
├── bin/                    # 二进制文件
│   ├── local/              # 本地开发构建
│   │   ├── proxy-test-tool         # 本地版本
│   │   └── proxy-test-tool-embed   # 嵌入式版本
│   ├── linux/              # Linux平台
│   │   ├── amd64/
│   │   │   └── proxy-test-tool
│   │   └── arm64/
│   │       └── proxy-test-tool
│   ├── windows/            # Windows平台
│   │   └── amd64/
│   │       └── proxy-test-tool.exe
│   └── darwin/             # macOS平台
│       ├── amd64/
│       │   └── proxy-test-tool
│       └── arm64/
│           └── proxy-test-tool
├── dist/                   # 发布包
│   ├── archives/           # 压缩包
│   │   ├── proxy-test-tool-linux-amd64.tar.gz
│   │   ├── proxy-test-tool-linux-arm64.tar.gz
│   │   ├── proxy-test-tool-windows-amd64.zip
│   │   ├── proxy-test-tool-darwin-amd64.tar.gz
│   │   └── proxy-test-tool-darwin-arm64.tar.gz
│   └── checksums/          # 校验和文件
│       ├── sha256sums.txt
│       └── archives-sha256sums.txt
├── tmp/                    # 临时文件
└── cache/                  # 缓存文件
    ├── cpu.prof
    └── mem.prof
```

**注意**：日志文件存储在项目根目录的 `logs/` 目录下，不在 `build/` 目录中：

```
logs/                       # 日志文件目录（项目根目录下）
├── server-YYYY-MM-DD.log   # 按日期分割的日志
├── server-*.log.gz         # 压缩的轮转日志
└── nohup.out               # 系统启动日志
```

## 🔧 构建命令

### 基础构建

```bash
# 构建本地开发版本
make build

# 构建嵌入式版本（静态资源打包）
make build-embed

# 查看帮助信息和目录结构
make help
```

### 平台特定构建

```bash
# Linux构建
make build-linux-amd64    # Linux AMD64
make build-linux-arm64    # Linux ARM64
make build-linux          # 所有Linux版本

# Windows构建
make build-windows-amd64  # Windows AMD64
make build-windows        # 所有Windows版本

# macOS构建
make build-darwin-amd64   # macOS Intel
make build-darwin-arm64   # macOS Apple Silicon
make build-darwin         # 所有macOS版本

# 构建所有平台
make build-all
```

### 发布相关

```bash
# 生成校验和
make checksums

# 创建发布包（包含构建、校验和、打包）
make package

# 查看构建信息
make info
```

### 运行和测试

```bash
# 运行本地构建版本
make run

# 运行嵌入式版本
make run-embed

# 开发模式运行（无需构建）
make dev

# 后台启动服务
make start

# 停止服务
make stop

# 查看服务状态
make status

# 查看日志
make logs

# 查看日志统计
make logs-stats

# 搜索日志内容
make logs-search SEARCH="关键词"

# 清理旧日志
make logs-clean
```

### 清理

```bash
# 清理所有构建文件
make clean

# 只清理二进制文件
make clean-bin

# 只清理发布包
make clean-dist
```

## 🎯 使用场景

### 1. 本地开发

```bash
# 快速构建和运行
make run

# 或者直接运行源码
make dev
```

### 2. 测试特定平台

```bash
# 构建Linux版本
make build-linux-amd64

# 手动运行测试
./build/bin/linux/amd64/proxy-test-tool
```

### 3. 创建发布版本

```bash
# 一键构建所有平台并打包
make package

# 查看发布包
ls -la build/dist/archives/
```

### 4. 性能分析

```bash
# 生成性能分析文件
make profile

# 查看分析文件
ls -la build/cache/
```

## 📦 部署

### Docker部署

```bash
# 构建Docker镜像
make docker-build

# 运行Docker容器
make docker-run

# 使用Docker Compose
make docker-compose-up
```

### 系统安装

```bash
# 安装到系统
make install

# 卸载
make uninstall
```

## 🔍 目录优势

### 1. **组织清晰**
- 所有构建产物集中在 `build/` 目录
- 按平台和架构分类存放
- 根目录保持干净

### 2. **易于管理**
- 可以单独清理特定类型的文件
- 便于CI/CD集成
- 支持并行构建

### 3. **版本控制友好**
- 整个 `build/` 目录被git忽略
- 避免误提交编译产物
- 减少仓库大小

### 4. **开发体验**
- 清晰的构建日志输出
- 详细的帮助信息
- 多种便捷命令

## 🚀 CI/CD集成

GitHub Actions已配置使用新的构建系统：

```yaml
- name: Build for multiple platforms
  run: |
    make build-all
    make checksums
    make package
```

构建产物会自动上传到GitHub Releases。

## 📋 最佳实践

1. **开发时**：使用 `make dev` 进行快速迭代
2. **测试时**：使用 `make build && make run` 测试构建版本
3. **发布前**：使用 `make package` 创建完整发布包
4. **清理时**：定期使用 `make clean` 清理构建文件

## 🔧 自定义配置

在 `Makefile` 中可以修改以下配置：

```makefile
BINARY_NAME=proxy-test-tool  # 二进制文件名
BUILD_DIR=build             # 构建目录
PORT=8080                   # 默认端口
```

## 📞 问题反馈

如果在使用构建系统过程中遇到问题，请：

1. 检查是否有必要的构建工具（Go、make、zip等）
2. 确认Go版本符合要求（Go 1.23+）
3. 查看详细的构建日志输出
4. 在GitHub Issues中提交问题报告 