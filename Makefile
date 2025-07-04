# Makefile for HTTP/WebSocket代理测试工具
.PHONY: help build run test clean docker-build docker-run docker-stop install fmt vet deps

# 默认目标
.DEFAULT_GOAL := help

# 配置变量
BINARY_NAME=proxy-test-tool
DOCKER_IMAGE=proxy-test-tool
PORT=8080

# 帮助信息
help: ## 显示帮助信息
	@echo "HTTP/WebSocket代理测试工具 - 可用命令："
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

# 构建相关
build: deps fmt vet ## 构建项目
	go build -o $(BINARY_NAME) .

build-embed: deps fmt vet ## 构建嵌入式版本（静态资源打包到二进制文件中）
	go build -o $(BINARY_NAME)-embed .

build-linux: deps fmt vet ## 构建Linux版本
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux .

build-linux-embed: deps fmt vet ## 构建Linux嵌入式版本
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-embed .

build-windows: deps fmt vet ## 构建Windows版本
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows.exe .

build-windows-embed: deps fmt vet ## 构建Windows嵌入式版本
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-embed.exe .

build-mac: deps fmt vet ## 构建macOS版本
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-mac .

build-mac-embed: deps fmt vet ## 构建macOS嵌入式版本
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-mac-embed .

build-all: build-linux build-windows build-mac ## 构建所有平台版本

build-all-embed: build-linux-embed build-windows-embed build-mac-embed ## 构建所有平台嵌入式版本

# 运行相关
run: build ## 构建并运行
	./$(BINARY_NAME)

dev: ## 开发模式运行
	go run .

# 清理相关
clean: ## 清理构建文件
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -f server.log
	rm -f nohup.out

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
	sudo cp $(BINARY_NAME) /usr/local/bin/

uninstall: ## 从系统卸载
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

# 发布相关
release: clean build-all ## 创建发布版本
	mkdir -p release
	cp $(BINARY_NAME)-linux release/
	cp $(BINARY_NAME)-windows.exe release/
	cp $(BINARY_NAME)-mac release/
	cp README.md release/
	cp LICENSE release/
	tar -czf release/$(BINARY_NAME)-linux.tar.gz -C release $(BINARY_NAME)-linux README.md LICENSE
	zip -j release/$(BINARY_NAME)-windows.zip release/$(BINARY_NAME)-windows.exe release/README.md release/LICENSE
	tar -czf release/$(BINARY_NAME)-mac.tar.gz -C release $(BINARY_NAME)-mac README.md LICENSE

# 服务管理
start: ## 后台启动服务
	nohup ./$(BINARY_NAME) > server.log 2>&1 &
	@echo "服务已启动，日志文件：server.log"

stop: ## 停止服务
	pkill -f $(BINARY_NAME) || true
	@echo "服务已停止"

status: ## 查看服务状态
	@ps aux | grep $(BINARY_NAME) | grep -v grep || echo "服务未运行"

logs: ## 查看日志
	tail -f server.log

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
profile: ## 性能分析
	go run . -cpuprofile=cpu.prof -memprofile=mem.prof

benchmark: ## 基准测试
	go test -bench=. -benchmem

# 文档生成
docs: ## 生成文档
	@echo "API文档可在 http://localhost:$(PORT)/api-docs 查看"
	@echo "用户手册位于 docs/用户使用手册.md" 