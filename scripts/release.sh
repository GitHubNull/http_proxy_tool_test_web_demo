#!/bin/bash

# HTTP/WebSocket代理测试工具 - 版本发布脚本
# 使用方法: ./scripts/release.sh v1.0.0

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查参数
if [ $# -eq 0 ]; then
    print_error "请提供版本号"
    echo "使用方法: $0 <version>"
    echo "示例: $0 v1.0.0"
    exit 1
fi

VERSION=$1

# 验证版本号格式 (v1.0.0 格式)
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)*$ ]]; then
    print_error "版本号格式不正确"
    echo "正确格式: v1.0.0 或 v1.0.0-beta"
    exit 1
fi

print_info "准备发布版本: $VERSION"

# 检查是否有未提交的更改
if [ -n "$(git status --porcelain)" ]; then
    print_error "存在未提交的更改，请先提交或暂存"
    git status --short
    exit 1
fi

# 检查当前分支
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ] && [ "$CURRENT_BRANCH" != "master" ]; then
    print_warning "当前不在main/master分支 (当前: $CURRENT_BRANCH)"
    read -p "是否继续? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "发布取消"
        exit 0
    fi
fi

# 拉取最新代码
print_info "拉取最新代码..."
git pull origin $CURRENT_BRANCH

# 检查版本是否已存在
if git tag -l | grep -q "^$VERSION$"; then
    print_error "版本 $VERSION 已存在"
    git tag -l | grep "$VERSION"
    exit 1
fi

# 运行测试
print_info "运行测试..."
if command -v go &> /dev/null; then
    go test ./...
    if [ $? -ne 0 ]; then
        print_error "测试失败，发布取消"
        exit 1
    fi
    print_success "测试通过"
else
    print_warning "Go未安装，跳过测试"
fi

# 运行代码格式检查
print_info "检查代码格式..."
if command -v gofmt &> /dev/null; then
    UNFORMATTED=$(gofmt -l .)
    if [ -n "$UNFORMATTED" ]; then
        print_error "以下文件格式不正确:"
        echo "$UNFORMATTED"
        print_info "运行 'gofmt -w .' 修复格式问题"
        exit 1
    fi
    print_success "代码格式检查通过"
fi

# 生成变更日志
print_info "生成变更日志..."
PREVIOUS_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

if [ -z "$PREVIOUS_TAG" ]; then
    print_warning "未找到上一个版本标签，生成完整变更日志"
    CHANGELOG=$(git log --pretty=format:"- %s (%h)" --reverse)
else
    print_info "上一个版本: $PREVIOUS_TAG"
    CHANGELOG=$(git log --pretty=format:"- %s (%h)" --reverse $PREVIOUS_TAG..HEAD)
fi

# 创建临时变更日志文件
CHANGELOG_FILE="CHANGELOG_$VERSION.md"
cat > $CHANGELOG_FILE << EOF
# 版本 $VERSION

## 发布日期
$(date '+%Y-%m-%d')

## 更改内容

### 🚀 新功能
$(echo "$CHANGELOG" | grep -i -E "(feat|feature|add|新增|添加)" || echo "- 无")

### 🐛 错误修复
$(echo "$CHANGELOG" | grep -i -E "(fix|bug|修复|修改)" || echo "- 无")

### 📝 其他更改
$(echo "$CHANGELOG" | grep -v -i -E "(feat|feature|add|新增|添加|fix|bug|修复|修改)" || echo "- 无")

## 构建信息
- 构建时间: $(date -u +%Y-%m-%dT%H:%M:%SZ)
- 提交哈希: $(git rev-parse HEAD)
- Go版本: $(go version 2>/dev/null || echo "未知")

## 下载
请访问 [GitHub Releases](https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:\/]\([^.]*\).*/\1/')/releases/tag/$VERSION) 下载对应平台的二进制文件。

## Docker
\`\`\`bash
docker pull ghcr.io/$(git config --get remote.origin.url | sed 's/.*github.com[:\/]\([^.]*\).*/\1/'):$VERSION
\`\`\`
EOF

print_success "变更日志生成完成: $CHANGELOG_FILE"

# 显示变更日志预览
echo
print_info "变更日志预览:"
echo "----------------------------------------"
cat $CHANGELOG_FILE
echo "----------------------------------------"
echo

# 确认发布
read -p "确认发布版本 $VERSION? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_info "发布取消"
    rm -f $CHANGELOG_FILE
    exit 0
fi

# 更新版本信息
print_info "更新版本信息..."

# 如果存在package.json或其他版本文件，可以在这里更新
# 例如：sed -i "s/\"version\": \".*\"/\"version\": \"${VERSION#v}\"/" package.json

# 提交版本更新（如果有文件更改）
if [ -n "$(git status --porcelain)" ]; then
    git add .
    git commit -m "chore: bump version to $VERSION"
fi

# 创建标签
print_info "创建Git标签..."
git tag -a $VERSION -m "Release $VERSION

$(cat $CHANGELOG_FILE)"

print_success "标签 $VERSION 创建成功"

# 推送到远程仓库
print_info "推送到远程仓库..."
git push origin $CURRENT_BRANCH
git push origin $VERSION

print_success "版本 $VERSION 推送成功"

# 清理临时文件
rm -f $CHANGELOG_FILE

print_success "发布完成！"
print_info "GitHub Actions将自动构建和发布版本 $VERSION"
print_info "请访问 https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:\/]\([^.]*\).*/\1/')/actions 查看构建状态"

# 可选：自动打开GitHub页面
if command -v xdg-open &> /dev/null; then
    read -p "是否打开GitHub Actions页面? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        xdg-open "https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:\/]\([^.]*\).*/\1/')/actions"
    fi
elif command -v open &> /dev/null; then
    read -p "是否打开GitHub Actions页面? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        open "https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:\/]\([^.]*\).*/\1/')/actions"
    fi
fi 