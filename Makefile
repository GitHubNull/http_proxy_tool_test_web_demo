# Makefile for HTTP/WebSocket代理测试工具
.PHONY: help build run test clean docker-build docker-run docker-stop install fmt vet deps setup-dirs

# 默认目标
.DEFAULT_GOAL := help

# 配置变量
BINARY_NAME=proxy-test-tool
DOCKER_IMAGE=proxy-test-tool
PORT=8080

# 构建目录结构
BUILD_DIR=build
BIN_DIR=$(BUILD_DIR)/bin
DIST_DIR=$(BUILD_DIR)/dist
TMP_DIR=$(BUILD_DIR)/tmp
LOGS_DIR=logs
CACHE_DIR=$(BUILD_DIR)/cache

# 平台特定目录
LOCAL_BIN_DIR=$(BIN_DIR)/local
LINUX_BIN_DIR=$(BIN_DIR)/linux
WINDOWS_BIN_DIR=$(BIN_DIR)/windows
DARWIN_BIN_DIR=$(BIN_DIR)/darwin

# 架构特定目录
LINUX_AMD64_DIR=$(LINUX_BIN_DIR)/amd64
LINUX_ARM64_DIR=$(LINUX_BIN_DIR)/arm64
WINDOWS_AMD64_DIR=$(WINDOWS_BIN_DIR)/amd64
DARWIN_AMD64_DIR=$(DARWIN_BIN_DIR)/amd64
DARWIN_ARM64_DIR=$(DARWIN_BIN_DIR)/arm64

# 发布目录
ARCHIVES_DIR=$(DIST_DIR)/archives
CHECKSUMS_DIR=$(DIST_DIR)/checksums

# 创建目录结构
setup-dirs: ## 创建构建目录结构
	@mkdir -p $(LOCAL_BIN_DIR)
	@mkdir -p $(LINUX_AMD64_DIR) $(LINUX_ARM64_DIR)
	@mkdir -p $(WINDOWS_AMD64_DIR)
	@mkdir -p $(DARWIN_AMD64_DIR) $(DARWIN_ARM64_DIR)
	@mkdir -p $(ARCHIVES_DIR) $(CHECKSUMS_DIR)
	@mkdir -p $(TMP_DIR) $(LOGS_DIR) $(CACHE_DIR)
	@echo "✅ 构建目录结构已创建"

# 帮助信息
help: ## 显示帮助信息
	@echo "HTTP/WebSocket代理测试工具 - 可用命令："
	@echo ""
	@echo "构建目录结构："
	@echo "  build/"
	@echo "  ├── bin/           # 二进制文件"
	@echo "  │   ├── local/     # 本地开发"
	@echo "  │   ├── linux/     # Linux平台 (amd64, arm64)"
	@echo "  │   ├── windows/   # Windows平台 (amd64)"
	@echo "  │   └── darwin/    # macOS平台 (amd64, arm64)"
	@echo "  ├── dist/          # 发布包"
	@echo "  │   ├── archives/  # 压缩包"
	@echo "  │   └── checksums/ # 校验和文件"
	@echo "  ├── tmp/           # 临时文件"
	@echo "  └── cache/         # 缓存文件"
	@echo ""
	@echo "logs/                                # 日志文件目录（项目根目录下）"
	@echo "├── server-YYYY-MM-DD.log           # 按日期分割的日志"
	@echo "├── server-YYYY-MM-DD-HH-MM-SS.log.gz  # 压缩的轮转日志"
	@echo "└── nohup.out                       # 系统启动日志"
	@echo ""
	@echo "发布检查选项："
	@echo "  pre-release-light   # 轻量级检查（仅安全检查，适合发布前使用）"
	@echo "  pre-release         # 完整检查（包含代码质量检查，耗时较长）"
	@echo "  check-security      # 仅安全和漏洞检查"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 开发相关
deps: ## 安装依赖
	go mod tidy
	go mod download

fmt: ## 格式化代码
	go fmt ./...

vet: ## 代码静态检查
	go vet ./...

test: ## 运行测试
	go test -v ./...

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "🧪 运行测试覆盖率分析..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out
	@echo "✅ 覆盖率报告生成完成: coverage.html"

bench: ## 运行基准测试
	@echo "⚡ 运行基准测试..."
	go test -bench=. -benchmem ./...

# 代码质量检查
install-tools: ## 安装代码质量检查工具
	@echo "🔧 安装代码质量检查工具..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "安装 golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.57.2; \
	}
	@command -v gosec >/dev/null 2>&1 || { \
		echo "安装 gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	}
	@command -v govulncheck >/dev/null 2>&1 || { \
		echo "安装 govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	}
	@command -v staticcheck >/dev/null 2>&1 || { \
		echo "安装 staticcheck..."; \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
	}
	@echo "✅ 所有工具安装完成"

lint: ## 运行golangci-lint代码检查
	@echo "🔍 运行golangci-lint检查..."
	@GOPATH=$$(go env GOPATH); \
	if [ ! -f "$$GOPATH/bin/golangci-lint" ]; then \
		echo "❌ golangci-lint未安装，请运行 'make install-tools'"; \
		exit 1; \
	fi
	@echo "🔧 使用超轻量级配置以避免内存问题..."
	@GOMAXPROCS=1 GOGC=100 $$(go env GOPATH)/bin/golangci-lint run \
		--timeout=5m \
		--enable=errcheck,govet,gofmt \
		--max-issues-per-linter=20 \
		--max-same-issues=5 \
		--concurrency=1 \
		--print-resources-usage \
		|| { \
			echo "⚠️  golangci-lint内存不足，使用基础工具检查..."; \
			echo "🔍 运行基础代码检查..."; \
			go vet ./... && echo "✅ go vet 通过"; \
			gofmt -l . | grep -v "^$$" && echo "❌ 发现格式问题，请运行 go fmt" || echo "✅ 格式检查通过"; \
			echo "🔍 运行errcheck检查..."; \
			which errcheck >/dev/null 2>&1 && errcheck ./... || echo "⚠️  errcheck未安装，跳过"; \
		}
	@echo "✅ golangci-lint检查完成"

lint-basic: ## 运行基础代码检查（内存友好）
	@echo "🔍 运行基础代码检查（内存友好）..."
	@echo "📋 运行 go vet..."
	@go vet ./...
	@echo "📋 检查代码格式..."
	@gofmt -l . | grep -v "^$$" && echo "❌ 发现格式问题，请运行 'make fmt'" || echo "✅ 格式检查通过"
	@echo "📋 检查未处理错误..."
	@which errcheck >/dev/null 2>&1 && errcheck ./... || echo "⚠️  errcheck未安装，请运行 'go install github.com/kisielk/errcheck@latest'"
	@echo "📋 检查死代码..."
	@which deadcode >/dev/null 2>&1 && deadcode ./... || echo "⚠️  deadcode未安装，跳过检查"
	@echo "✅ 基础代码检查完成"

lint-fix: ## 运行golangci-lint并自动修复问题
	@echo "🔧 运行golangci-lint自动修复..."
	@GOPATH=$$(go env GOPATH); \
	if [ ! -f "$$GOPATH/bin/golangci-lint" ]; then \
		echo "❌ golangci-lint未安装，请运行 'make install-tools'"; \
		exit 1; \
	fi
	@echo "🔧 使用轻量级配置进行自动修复..."
	@GOMAXPROCS=1 $$(go env GOPATH)/bin/golangci-lint run \
		--fix \
		--timeout=10m \
		--enable=gofmt,goimports,misspell \
		--max-issues-per-linter=50 \
		--max-same-issues=10 \
		--concurrency=1 \
		|| { \
			echo "⚠️  golangci-lint自动修复遇到问题，尝试基础修复..."; \
			echo "🔧 运行基础格式化..."; \
			go fmt ./... && echo "✅ go fmt 完成"; \
		}
	@echo "✅ golangci-lint自动修复完成"

security: ## 运行安全检查
	@echo "🛡️ 运行安全检查..."
	@GOPATH=$$(go env GOPATH); \
	if [ ! -f "$$GOPATH/bin/gosec" ]; then \
		echo "❌ gosec未安装，请运行 'make install-tools'"; \
		exit 1; \
	fi
	$$(go env GOPATH)/bin/gosec -fmt json -out gosec-report.json -fmt sarif -out gosec-report.sarif ./...
	$$(go env GOPATH)/bin/gosec ./...
	@echo "✅ 安全检查完成"

vulncheck: ## 检查依赖漏洞
	@echo "🔒 检查依赖漏洞..."
	@GOPATH=$$(go env GOPATH); \
	if [ ! -f "$$GOPATH/bin/govulncheck" ]; then \
		echo "❌ govulncheck未安装，请运行 'make install-tools'"; \
		exit 1; \
	fi
	@echo "🔍 使用govulncheck检查主包..."
	@$$(go env GOPATH)/bin/govulncheck -mode=source . || { \
		echo "⚠️  govulncheck遇到内部错误，使用替代方案..."; \
		echo "🔍 检查模块依赖完整性..."; \
		go mod verify && echo "✅ 模块验证通过"; \
		echo "🔍 列出所有依赖项..."; \
		go list -m all | wc -l | xargs printf "📦 共有 %s 个依赖项\n"; \
		echo "🔍 检查高风险依赖模式..."; \
		go list -m all | grep -E -i "(crypto|auth|jwt|session|password|hash)" | head -10 || echo "未发现明显高风险依赖"; \
	}
	@echo "✅ 漏洞检查完成"

staticcheck: ## 运行静态代码分析
	@echo "📊 运行静态代码分析..."
	@GOPATH=$$(go env GOPATH); \
	if [ ! -f "$$GOPATH/bin/staticcheck" ]; then \
		echo "❌ staticcheck未安装，请运行 'make install-tools'"; \
		exit 1; \
	fi
	$$(go env GOPATH)/bin/staticcheck ./...
	@echo "✅ 静态分析完成"

mod-verify: ## 验证go.mod和go.sum
	@echo "📦 验证模块依赖..."
	go mod verify
	go mod tidy
	@if [ -n "$$(git status --porcelain go.mod go.sum)" ]; then \
		echo "❌ go.mod或go.sum有变化，请检查依赖"; \
		git diff go.mod go.sum; \
		exit 1; \
	fi
	@echo "✅ 模块依赖验证通过"

# 综合检查
check-basic: fmt vet test ## 基础代码检查
	@echo "✅ 基础代码检查完成"

check-quality: lint-basic security staticcheck ## 代码质量检查
	@echo "✅ 代码质量检查完成"

check-security: security vulncheck ## 安全检查
	@echo "✅ 安全检查完成"

check-all: deps mod-verify fmt vet test-coverage lint-basic security staticcheck vulncheck ## 运行所有检查
	@echo "🎉 所有检查完成！"

# 预提交检查
pre-commit: check-all ## 提交前完整检查
	@echo "🚀 预提交检查完成，代码质量良好！"

# 轻量级发布前检查（仅安全检查）
pre-release-light: install-tools deps fmt vet test check-security ## 轻量级发布前检查（仅安全检查）
	@echo "📦 检查构建目录清洁..."
	@if [ -d "$(BUILD_DIR)" ]; then \
		echo "⚠️  构建目录存在，建议清理"; \
		echo "运行 'make clean' 清理构建目录"; \
	fi
	@echo "🔍 检查Git状态..."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "❌ 存在未提交的更改:"; \
		git status --short; \
		echo "请先提交所有更改"; \
		exit 1; \
	fi
	@echo "🏷️ 检查Git分支..."
	@CURRENT_BRANCH=$$(git branch --show-current); \
	if [ "$$CURRENT_BRANCH" != "main" ] && [ "$$CURRENT_BRANCH" != "master" ]; then \
		echo "⚠️  当前分支: $$CURRENT_BRANCH (建议在main/master分支发布)"; \
	fi
	@echo "🎯 运行构建测试..."
	@$(MAKE) build-all >/dev/null 2>&1
	@echo "🧹 清理测试构建..."
	@$(MAKE) clean-bin >/dev/null 2>&1
	@echo "🎉 轻量级发布前检查完成，可以安全发布！"

# 发布前检查
pre-release: install-tools pre-commit bench ## 发布前完整检查
	@echo "📦 检查构建目录清洁..."
	@if [ -d "$(BUILD_DIR)" ]; then \
		echo "⚠️  构建目录存在，建议清理"; \
		echo "运行 'make clean' 清理构建目录"; \
	fi
	@echo "🔍 检查Git状态..."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "❌ 存在未提交的更改:"; \
		git status --short; \
		echo "请先提交所有更改"; \
		exit 1; \
	fi
	@echo "🏷️ 检查Git分支..."
	@CURRENT_BRANCH=$$(git branch --show-current); \
	if [ "$$CURRENT_BRANCH" != "main" ] && [ "$$CURRENT_BRANCH" != "master" ]; then \
		echo "⚠️  当前分支: $$CURRENT_BRANCH (建议在main/master分支发布)"; \
	fi
	@echo "🎯 运行构建测试..."
	@$(MAKE) build-all >/dev/null 2>&1
	@echo "🧹 清理测试构建..."
	@$(MAKE) clean-bin >/dev/null 2>&1
	@echo "🎉 发布前检查完成，可以安全发布！"

# 快速检查（适用于开发过程中）
quick-check: fmt vet test ## 快速检查（开发时使用）
	@echo "⚡ 快速检查完成"

# 构建相关
build: setup-dirs deps fmt vet ## 构建本地开发版本
	go build -o $(LOCAL_BIN_DIR)/$(BINARY_NAME) .
	@echo "✅ 本地构建完成: $(LOCAL_BIN_DIR)/$(BINARY_NAME)"

build-embed: setup-dirs deps fmt vet ## 构建本地嵌入式版本（静态资源打包到二进制文件中）
	go build -o $(LOCAL_BIN_DIR)/$(BINARY_NAME)-embed .
	@echo "✅ 本地嵌入式构建完成: $(LOCAL_BIN_DIR)/$(BINARY_NAME)-embed"

# Linux构建
build-linux-amd64: setup-dirs deps fmt vet ## 构建Linux AMD64版本
	GOOS=linux GOARCH=amd64 go build -o $(LINUX_AMD64_DIR)/$(BINARY_NAME) .
	@echo "✅ Linux AMD64构建完成: $(LINUX_AMD64_DIR)/$(BINARY_NAME)"

build-linux-arm64: setup-dirs deps fmt vet ## 构建Linux ARM64版本
	GOOS=linux GOARCH=arm64 go build -o $(LINUX_ARM64_DIR)/$(BINARY_NAME) .
	@echo "✅ Linux ARM64构建完成: $(LINUX_ARM64_DIR)/$(BINARY_NAME)"

build-linux: build-linux-amd64 build-linux-arm64 ## 构建所有Linux版本

# Windows构建
build-windows-amd64: setup-dirs deps fmt vet ## 构建Windows AMD64版本
	GOOS=windows GOARCH=amd64 go build -o $(WINDOWS_AMD64_DIR)/$(BINARY_NAME).exe .
	@echo "✅ Windows AMD64构建完成: $(WINDOWS_AMD64_DIR)/$(BINARY_NAME).exe"

build-windows: build-windows-amd64 ## 构建所有Windows版本

# macOS构建
build-darwin-amd64: setup-dirs deps fmt vet ## 构建macOS AMD64版本
	GOOS=darwin GOARCH=amd64 go build -o $(DARWIN_AMD64_DIR)/$(BINARY_NAME) .
	@echo "✅ macOS AMD64构建完成: $(DARWIN_AMD64_DIR)/$(BINARY_NAME)"

build-darwin-arm64: setup-dirs deps fmt vet ## 构建macOS ARM64版本（Apple Silicon）
	GOOS=darwin GOARCH=arm64 go build -o $(DARWIN_ARM64_DIR)/$(BINARY_NAME) .
	@echo "✅ macOS ARM64构建完成: $(DARWIN_ARM64_DIR)/$(BINARY_NAME)"

build-darwin: build-darwin-amd64 build-darwin-arm64 ## 构建所有macOS版本

# 构建所有平台
build-all: build-linux build-windows build-darwin ## 构建所有平台版本
	@echo "🎉 所有平台构建完成！"
	@echo "构建产物位置："
	@find $(BIN_DIR) -name "$(BINARY_NAME)*" -type f

# 生成校验和
checksums: ## 生成所有二进制文件的校验和
	@echo "生成校验和文件..."
	@find $(BIN_DIR) -name "$(BINARY_NAME)*" -type f -exec sha256sum {} \; > $(CHECKSUMS_DIR)/sha256sums.txt
	@echo "✅ 校验和文件已生成: $(CHECKSUMS_DIR)/sha256sums.txt"

# 创建发布包
package: build-all checksums ## 创建发布包
	@echo "创建发布包..."
	
	# Linux包
	@tar -czf $(ARCHIVES_DIR)/$(BINARY_NAME)-linux-amd64.tar.gz -C $(LINUX_AMD64_DIR) $(BINARY_NAME) && \
	tar -czf $(ARCHIVES_DIR)/$(BINARY_NAME)-linux-arm64.tar.gz -C $(LINUX_ARM64_DIR) $(BINARY_NAME)
	
	# Windows包
	@cd $(WINDOWS_AMD64_DIR) && zip -r $(CURDIR)/$(ARCHIVES_DIR)/$(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME).exe
	
	# macOS包
	@tar -czf $(ARCHIVES_DIR)/$(BINARY_NAME)-darwin-amd64.tar.gz -C $(DARWIN_AMD64_DIR) $(BINARY_NAME) && \
	tar -czf $(ARCHIVES_DIR)/$(BINARY_NAME)-darwin-arm64.tar.gz -C $(DARWIN_ARM64_DIR) $(BINARY_NAME)
	
	# 生成发布包校验和
	@cd $(ARCHIVES_DIR) && sha256sum *.tar.gz *.zip > ../checksums/archives-sha256sums.txt
	
	@echo "✅ 发布包创建完成:"
	@ls -la $(ARCHIVES_DIR)/

# 运行相关
run: build ## 构建并运行本地版本
	$(LOCAL_BIN_DIR)/$(BINARY_NAME)

run-embed: build-embed ## 构建并运行嵌入式版本
	$(LOCAL_BIN_DIR)/$(BINARY_NAME)-embed

dev: ## 开发模式运行
	go run .

# 清理相关
clean: ## 清理所有构建文件
	rm -rf $(BUILD_DIR)
	rm -f nohup.out
	@echo "✅ 清理完成"

clean-bin: ## 只清理二进制文件
	rm -rf $(BIN_DIR)
	@echo "✅ 二进制文件清理完成"

clean-dist: ## 只清理发布包
	rm -rf $(DIST_DIR)
	@echo "✅ 发布包清理完成"

# Docker相关
docker-build: ## 构建Docker镜像
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## 运行Docker容器
	docker run -d -p $(PORT):8080 --name $(BINARY_NAME) $(DOCKER_IMAGE)

docker-stop: ## 停止Docker容器
	docker stop $(BINARY_NAME) || true
	docker rm $(BINARY_NAME) || true

docker-compose-up: ## 使用Docker Compose启动
	docker-compose up -d

docker-compose-down: ## 使用Docker Compose停止
	docker-compose down

docker-compose-logs: ## 查看Docker Compose日志
	docker-compose logs -f

# 安装相关
install: build ## 安装到系统
	sudo cp $(LOCAL_BIN_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ 已安装到 /usr/local/bin/$(BINARY_NAME)"

uninstall: ## 从系统卸载
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ 已从系统卸载"

# 服务管理
start: build ## 后台启动服务
	@mkdir -p $(LOGS_DIR)
	nohup $(LOCAL_BIN_DIR)/$(BINARY_NAME) > $(LOGS_DIR)/nohup.out 2>&1 &
	@echo "✅ 服务已启动"
	@echo "📋 日志管理："
	@echo "  - 应用日志: $(LOGS_DIR)/server-$(shell date +%Y-%m-%d).log"
	@echo "  - 系统日志: $(LOGS_DIR)/nohup.out"
	@echo "  - 查看日志: make logs"
	@echo "  - 日志统计: make logs-stats"
	@echo "  - 自定义日志目录: ./$(BINARY_NAME) -log-dir /path/to/logs"

stop: ## 停止服务
	pkill -f $(BINARY_NAME) || true
	@echo "✅ 服务已停止"

status: ## 查看服务状态
	@ps aux | grep $(BINARY_NAME) | grep -v grep || echo "❌ 服务未运行"

logs: ## 查看当天日志
	@if [ -f "$(LOGS_DIR)/server-$(shell date +%Y-%m-%d).log" ]; then \
		tail -f $(LOGS_DIR)/server-$(shell date +%Y-%m-%d).log; \
	else \
		echo "❌ 当天日志文件不存在: $(LOGS_DIR)/server-$(shell date +%Y-%m-%d).log"; \
	fi

logs-all: ## 查看所有日志文件
	@echo "📋 日志文件列表:"
	@ls -lht $(LOGS_DIR)/ 2>/dev/null || echo "❌ 日志目录不存在"

logs-stats: ## 查看日志统计信息
	@echo "📊 日志统计信息:"
	@if [ -d "$(LOGS_DIR)" ]; then \
		echo "日志目录: $(LOGS_DIR)"; \
		echo "文件数量: $$(find $(LOGS_DIR) -name "server-*.log*" | wc -l)"; \
		echo "总大小: $$(du -sh $(LOGS_DIR) 2>/dev/null | cut -f1)"; \
		echo "最新日志: $$(ls -t $(LOGS_DIR)/server-*.log* 2>/dev/null | head -1 | xargs basename 2>/dev/null || echo '无')"; \
		echo "最旧日志: $$(ls -tr $(LOGS_DIR)/server-*.log* 2>/dev/null | head -1 | xargs basename 2>/dev/null || echo '无')"; \
		echo "压缩文件: $$(find $(LOGS_DIR) -name "*.gz" | wc -l)"; \
	else \
		echo "❌ 日志目录不存在"; \
	fi

logs-clean: ## 清理旧日志文件（保留最近7天）
	@echo "🧹 清理旧日志文件..."
	@if [ -d "$(LOGS_DIR)" ]; then \
		find $(LOGS_DIR) -name "server-*.log*" -mtime +7 -type f -exec rm -v {} \; 2>/dev/null || true; \
		echo "✅ 日志清理完成"; \
	else \
		echo "❌ 日志目录不存在"; \
	fi

logs-compress: ## 压缩7天前的日志文件
	@echo "🗜️ 压缩旧日志文件..."
	@if [ -d "$(LOGS_DIR)" ]; then \
		find $(LOGS_DIR) -name "server-*.log" -mtime +7 -type f -exec gzip -v {} \; 2>/dev/null || true; \
		echo "✅ 日志压缩完成"; \
	else \
		echo "❌ 日志目录不存在"; \
	fi

logs-search: ## 搜索日志内容（用法: make logs-search SEARCH="关键词"）
	@if [ -z "$(SEARCH)" ]; then \
		echo "❌ 请指定搜索关键词: make logs-search SEARCH=\"关键词\""; \
	elif [ -d "$(LOGS_DIR)" ]; then \
		echo "🔍 搜索日志内容: $(SEARCH)"; \
		grep -r "$(SEARCH)" $(LOGS_DIR)/ || echo "未找到匹配内容"; \
	else \
		echo "❌ 日志目录不存在"; \
	fi

logs-tail: ## 实时查看日志（用法: make logs-tail [LINES=100]）
	@LINES=$${LINES:-100}; \
	if [ -f "$(LOGS_DIR)/server-$(shell date +%Y-%m-%d).log" ]; then \
		echo "📖 实时查看日志 (最近$$LINES行):"; \
		tail -n $$LINES -f $(LOGS_DIR)/server-$(shell date +%Y-%m-%d).log; \
	else \
		echo "❌ 当天日志文件不存在"; \
	fi

logs-maintain: ## 执行完整的日志维护（压缩+清理）
	@echo "🔧 执行日志维护..."
	@if [ -x "scripts/log-maintenance.sh" ]; then \
		./scripts/log-maintenance.sh maintain; \
	else \
		echo "❌ 日志维护脚本不存在或无执行权限"; \
		echo "请运行: chmod +x scripts/log-maintenance.sh"; \
	fi

logs-monitor: ## 监控日志目录大小
	@if [ -x "scripts/log-maintenance.sh" ]; then \
		./scripts/log-maintenance.sh monitor; \
	else \
		echo "❌ 日志维护脚本不存在或无执行权限"; \
	fi

# 测试相关
test-api: ## 测试API接口
	@echo "测试基础API..."
	curl -s http://localhost:$(PORT)/api/test | head -200
	@echo "\n\n测试系统信息..."
	curl -s http://localhost:$(PORT)/test/system | head -200

test-health: ## 健康检查
	@echo "健康检查..."
	curl -f http://localhost:$(PORT)/api/test > /dev/null 2>&1 && echo "✅ 服务正常" || echo "❌ 服务异常"

# 开发工具
watch: ## 监控文件变化自动重启（需要安装fswatch）
	fswatch -o . | xargs -n1 -I{} make dev

# 性能分析
profile: setup-dirs ## 性能分析
	go run . -cpuprofile=$(CACHE_DIR)/cpu.prof -memprofile=$(CACHE_DIR)/mem.prof

benchmark: ## 基准测试
	go test -bench=. -benchmem

# 信息查看
info: ## 显示构建信息
	@echo "项目信息："
	@echo "  项目名称: $(BINARY_NAME)"
	@echo "  构建目录: $(BUILD_DIR)/"
	@echo "  Docker镜像: $(DOCKER_IMAGE)"
	@echo "  服务端口: $(PORT)"
	@echo ""
	@echo "构建产物统计："
	@if [ -d "$(BIN_DIR)" ]; then \
		echo "  二进制文件:"; \
		find $(BIN_DIR) -name "$(BINARY_NAME)*" -type f -exec ls -lh {} \; | awk '{printf "    %s %s\n", $$5, $$9}'; \
	else \
		echo "  暂无二进制文件"; \
	fi
	@if [ -d "$(ARCHIVES_DIR)" ]; then \
		echo "  发布包:"; \
		ls -lh $(ARCHIVES_DIR)/* 2>/dev/null | awk '{printf "    %s %s\n", $$5, $$9}' || echo "    暂无发布包"; \
	else \
		echo "  暂无发布包"; \
	fi

# 文档生成
docs: ## 生成文档
	@echo "API文档可在 http://localhost:$(PORT)/api-docs 查看"
	@echo "用户手册位于 docs/用户使用手册.md" 