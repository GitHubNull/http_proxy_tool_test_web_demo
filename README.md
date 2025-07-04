# HTTP/WebSocket代理测试工具

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey.svg)](https://github.com/gin-gonic/gin)

一个专为测试HTTP(S)代理和WebSocket代理抓包软件而设计的综合测试平台。

## 🚀 项目简介

这个工具提供了丰富的测试场景，包括各种HTTP方法、响应格式、WebSocket连接类型、高并发测试等，帮助开发者和测试人员验证代理软件的功能和性能。

## ✨ 功能特性

### 🌐 HTTP测试功能
- ✅ 支持所有HTTP方法（GET、POST、PUT、DELETE、PATCH、HEAD、OPTIONS）
- ✅ 多种响应格式（JSON、XML、HTML、文本、二进制）
- ✅ 认证测试（Basic、Bearer、Digest）
- ✅ 文件上传（单文件、多文件）
- ✅ Cookie管理和会话测试
- ✅ 自定义请求头处理
- ✅ 压缩测试（Gzip、Deflate）
- ✅ 缓存和ETag测试
- ✅ 延迟和超时模拟
- ✅ 重定向测试
- ✅ 流数据和SSE（Server-Sent Events）

### 🔌 WebSocket测试功能
- ✅ 基础连接测试
- ✅ 回声（Echo）测试
- ✅ 广播测试
- ✅ 实时数据推送
- ✅ 心跳检测
- ✅ 二进制数据传输
- ✅ 聊天室模拟
- ✅ 性能测试

### 🏃 性能测试功能
- ✅ 并发测试（支持1-1000并发）
- ✅ 压力测试（支持1-300秒持续测试）
- ✅ 内存压力测试
- ✅ CPU密集型测试
- ✅ 网络测试
- ✅ 文件IO测试
- ✅ 系统信息监控

### 🎨 前端界面
- ✅ 现代化响应式设计
- ✅ Bootstrap + jQuery支持
- ✅ 四个主要标签页（HTTP测试、WebSocket测试、性能测试、系统信息）
- ✅ 实时结果展示
- ✅ 测试统计
- ✅ 数据导出功能

## 🛠️ 技术栈

- **后端**: Go 1.23+ + Gin框架 + Gorilla WebSocket
- **前端**: HTML5 + CSS3 + JavaScript + Bootstrap 5 + jQuery
- **依赖管理**: Go Modules
- **跨域支持**: CORS中间件
- **代码质量**: golangci-lint（使用默认配置）
- **容器化**: Docker + Docker Compose
- **CI/CD**: GitHub Actions

## 💎 代码质量

本项目使用 golangci-lint 进行代码质量检查，采用默认配置以确保：

- ✅ **无语法错误**: 使用 `govet` 检查Go代码语法
- ✅ **错误处理**: 使用 `errcheck` 确保错误得到适当处理
- ✅ **代码简化**: 使用 `gosimple` 和 `staticcheck` 优化代码
- ✅ **未使用变量**: 使用 `unused` 和 `ineffassign` 检查
- ✅ **代码格式**: 使用 `gofmt` 和 `goimports` 保持一致的代码风格

### 本地代码检查

```bash
# 安装 golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行代码检查
golangci-lint run

# 自动修复部分问题
golangci-lint run --fix
```

### CI/CD集成

项目在GitHub Actions中集成了完整的CI/CD流程：

- **代码质量检查**: golangci-lint 自动化检查
- **单元测试**: Go 1.23和1.24多版本测试
- **安全扫描**: 使用 gosec 进行安全检查
- **依赖检查**: 使用 govulncheck 检查漏洞
- **性能测试**: 自动化基准测试
- **Docker构建**: 多平台容器构建测试
- **集成测试**: 端到端功能测试

## 📦 安装与部署

### 系统要求

- Go 1.23 或更高版本
- 支持的操作系统：Linux、macOS、Windows
- 内存：最低 512MB，推荐 1GB 以上
- 磁盘空间：100MB 以上

### 快速开始

#### 方式一：下载预编译版本（推荐）
```bash
# 下载嵌入式版本（静态资源已打包，仅需一个二进制文件）
wget https://github.com/username/proxy-test-tool/releases/latest/download/proxy-test-tool-linux-amd64
chmod +x proxy-test-tool-linux-amd64
./proxy-test-tool-linux-amd64
```

#### 方式二：从源码构建
1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd http_proxy_tool_test_web_demo
   ```

2. **安装依赖**
   ```bash
   go mod tidy
   ```

3. **构建项目**
   ```bash
   # 构建嵌入式版本（推荐）- 静态资源打包到二进制文件中
   make build-embed
   ./proxy-test-tool-embed
   
   # 或构建标准版本
   make build
   ./proxy-test-tool
   ```

4. **访问界面**
   打开浏览器访问 `http://localhost:8080`

#### 💡 嵌入式构建的优势
- ✅ **单文件部署**：所有静态资源（HTML、CSS、JS）打包到二进制文件中
- ✅ **无需额外文件**：不需要单独部署static和templates目录
- ✅ **便于分发**：只需拷贝一个二进制文件即可运行
- ✅ **容器化友好**：Docker镜像更小，部署更简单

### Docker部署

1. **构建Docker镜像**
   ```bash
   docker build -t proxy-test-tool .
   ```

2. **运行容器**
   ```bash
   docker run -d -p 8080:8080 --name proxy-test proxy-test-tool
   ```

## 🎯 使用指南

### 基础HTTP测试

1. 在"HTTP测试"标签页中选择HTTP方法
2. 输入请求URL（如 `/api/test`）
3. 配置请求头和请求体（可选）
4. 点击"发送请求"按钮
5. 查看右侧的响应结果

### WebSocket连接测试

1. 切换到"WebSocket测试"标签页
2. 选择连接类型（如"基础连接"）
3. 点击"连接"按钮建立WebSocket连接
4. 在消息输入框中输入消息
5. 点击"发送"按钮发送消息
6. 在消息区域查看收发的消息

### 性能测试

1. 切换到"性能测试"标签页
2. 配置并发参数（并发数、请求总数）
3. 点击"开始并发测试"按钮
4. 查看测试结果和统计信息

## 🔧 配置选项

环境变量配置：

- `PORT`: 服务端口（默认：8080）
- `GIN_MODE`: 运行模式（debug/release，默认：debug）

## 📋 API接口

### HTTP测试接口

- `GET/POST/PUT/DELETE/PATCH/HEAD/OPTIONS /api/test` - 基础测试
- `GET/POST /api/status/:code` - 状态码测试
- `GET/POST /api/delay/:seconds` - 延迟测试
- `GET /api/redirect/:times` - 重定向测试
- `GET /api/json` - JSON响应测试
- `GET /api/xml` - XML响应测试
- `POST /api/upload` - 文件上传测试
- `GET /api/auth/basic` - Basic认证测试
- `GET /api/cookies` - Cookie测试
- `GET /api/gzip` - 压缩测试

### WebSocket接口

- `GET /ws/connect` - 基础连接
- `GET /ws/echo` - 回声测试
- `GET /ws/broadcast` - 广播测试
- `GET /ws/realtime` - 实时数据推送
- `GET /ws/heartbeat` - 心跳检测
- `GET /ws/binary` - 二进制数据传输
- `GET /ws/chat` - 聊天室模拟
- `GET /ws/performance` - 性能测试

### 性能测试接口

- `GET/POST /test/concurrent` - 并发测试
- `GET/POST /test/stress` - 压力测试
- `GET /test/memory` - 内存测试
- `GET /test/cpu` - CPU测试
- `GET /test/system` - 系统信息

详细的API文档请访问：`http://localhost:8080/api-docs`

## 📖 文档

- [用户使用手册](docs/用户使用手册.md) - 详细的使用说明
- [API文档](http://localhost:8080/api-docs) - 在线API文档

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证。详情请查看 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [Gin](https://gin-gonic.com/) - Go Web框架
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket支持
- [Bootstrap](https://getbootstrap.com/) - 前端UI框架
- [jQuery](https://jquery.com/) - JavaScript库

## 📞 支持

如果你有任何问题或建议，请：

1. 查看 [文档](docs/用户使用手册.md)
2. 搜索已有的 [Issues](https://github.com/your-repo/issues)
3. 创建新的 [Issue](https://github.com/your-repo/issues/new)

---

⭐ 如果这个项目对你有帮助，请给个Star！ 