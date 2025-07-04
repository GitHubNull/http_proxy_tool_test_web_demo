# 使用官方 Go 镜像作为构建环境
FROM golang:1.22-alpine AS builder

# 添加构建参数
ARG BUILD_VERSION=v1.0.0
ARG BUILD_TIME
ARG BUILD_COMMIT

# 设置工作目录
WORKDIR /app

# 设置 Go 环境变量
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download && go mod verify

# 复制源代码
COPY . .

# 构建应用（添加版本信息）
RUN go build \
    -a \
    -installsuffix cgo \
    -ldflags "-X main.Version=${BUILD_VERSION} -X main.BuildTime=${BUILD_TIME} -X main.BuildCommit=${BUILD_COMMIT} -w -s" \
    -o proxy-test-tool \
    .

# 运行时镜像
FROM alpine:latest

# 添加镜像标签
LABEL maintainer="HTTP/WebSocket代理测试工具" \
      version="${BUILD_VERSION}" \
      description="专为测试HTTP(S)代理和WebSocket代理抓包软件而设计的综合测试平台" \
      org.opencontainers.image.title="Proxy Test Tool" \
      org.opencontainers.image.description="HTTP/WebSocket代理测试工具" \
      org.opencontainers.image.version="${BUILD_VERSION}" \
      org.opencontainers.image.created="${BUILD_TIME}"

# 安装运行时依赖
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    wget \
    curl \
    && update-ca-certificates

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN adduser -D -s /bin/sh -u 1001 appuser

# 创建应用目录
WORKDIR /app

# 从构建阶段复制文件
COPY --from=builder /app/proxy-test-tool ./
COPY --from=builder /app/docs ./docs/

# 设置文件权限
RUN chown -R appuser:appuser /app && \
    chmod +x ./proxy-test-tool

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider --timeout=3 http://localhost:8080/api/test || exit 1

# 设置启动命令
ENTRYPOINT ["./proxy-test-tool"] 