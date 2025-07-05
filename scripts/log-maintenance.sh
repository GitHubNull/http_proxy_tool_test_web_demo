#!/bin/bash

# HTTP/WebSocket代理测试工具 - 日志维护脚本
# 用于定期清理和压缩日志文件

set -e

# 默认配置
LOG_DIR="${LOG_DIR:-logs}"
MAX_AGE_DAYS="${MAX_AGE_DAYS:-30}"
COMPRESS_AGE_DAYS="${COMPRESS_AGE_DAYS:-7}"
MAX_SIZE_GB="${MAX_SIZE_GB:-5}"
VERBOSE="${VERBOSE:-true}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    if [ "$VERBOSE" = "true" ]; then
        echo -e "${BLUE}[INFO]${NC} $1"
    fi
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# 显示使用帮助
show_help() {
    cat << EOF
HTTP/WebSocket代理测试工具 - 日志维护脚本

用法: $0 [选项] [操作]

操作:
  stats       显示日志统计信息
  compress    压缩旧日志文件
  cleanup     清理过期日志文件
  maintain    执行完整维护（压缩+清理）
  monitor     检查磁盘使用并报警

选项:
  -d, --log-dir DIR         日志目录 (默认: logs)
  -a, --max-age DAYS        最大保留天数 (默认: 30)
  -c, --compress-age DAYS   压缩天数阈值 (默认: 7)
  -s, --max-size GB         最大目录大小GB (默认: 5)
  -q, --quiet               静默模式
  -v, --verbose             详细输出 (默认)
  -h, --help                显示帮助信息

环境变量:
  LOG_DIR                   日志目录
  MAX_AGE_DAYS             最大保留天数
  COMPRESS_AGE_DAYS        压缩天数阈值
  MAX_SIZE_GB              最大目录大小
  VERBOSE                  是否详细输出

示例:
  $0 stats                              # 显示统计信息
  $0 -d /var/log/app cleanup           # 清理指定目录
  $0 --max-age 7 --max-size 1 maintain # 7天保留期，1GB限制
  
  # 用于crontab的示例
  0 2 * * * /path/to/log-maintenance.sh maintain 2>&1 | logger
EOF
}

# 检查目录是否存在
check_log_dir() {
    if [ ! -d "$LOG_DIR" ]; then
        log_error "日志目录不存在: $LOG_DIR"
        exit 1
    fi
}

# 获取目录大小（GB）
get_dir_size_gb() {
    du -s "$LOG_DIR" 2>/dev/null | awk '{print $1/1024/1024}' || echo "0"
}

# 显示统计信息
show_stats() {
    check_log_dir
    
    log_info "=== 日志统计信息 ==="
    echo "日志目录: $LOG_DIR"
    
    # 文件统计
    total_files=$(find "$LOG_DIR" -name "server-*.log*" -type f | wc -l)
    log_files=$(find "$LOG_DIR" -name "server-*.log" -type f | wc -l)
    gz_files=$(find "$LOG_DIR" -name "server-*.log.gz" -type f | wc -l)
    
    echo "总文件数: $total_files"
    echo "日志文件: $log_files"
    echo "压缩文件: $gz_files"
    
    # 大小统计
    total_size=$(du -sh "$LOG_DIR" 2>/dev/null | cut -f1)
    total_size_gb=$(get_dir_size_gb)
    echo "总大小: $total_size (${total_size_gb}GB)"
    
    # 时间统计
    if [ $total_files -gt 0 ]; then
        oldest=$(find "$LOG_DIR" -name "server-*.log*" -type f -printf '%T+ %p\n' 2>/dev/null | sort | head -1 | cut -d' ' -f2- | xargs basename 2>/dev/null || echo "无")
        newest=$(find "$LOG_DIR" -name "server-*.log*" -type f -printf '%T+ %p\n' 2>/dev/null | sort -r | head -1 | cut -d' ' -f2- | xargs basename 2>/dev/null || echo "无")
        echo "最旧文件: $oldest"
        echo "最新文件: $newest"
    fi
    
    # 检查磁盘使用（使用awk进行浮点数比较）
    if awk "BEGIN {exit !($total_size_gb > $MAX_SIZE_GB)}"; then
        log_warn "日志目录大小超过限制 ${total_size_gb}GB > ${MAX_SIZE_GB}GB"
    else
        log_success "日志目录大小正常 ${total_size_gb}GB <= ${MAX_SIZE_GB}GB"
    fi
}

# 压缩旧日志
compress_logs() {
    check_log_dir
    
    log_info "=== 压缩旧日志文件 ==="
    log_info "压缩 $COMPRESS_AGE_DAYS 天前的日志文件..."
    
    compressed_count=0
    while IFS= read -r -d '' file; do
        if [ -f "$file" ]; then
            log_info "压缩文件: $(basename "$file")"
            if gzip -v "$file" 2>/dev/null; then
                ((compressed_count++))
            else
                log_error "压缩失败: $file"
            fi
        fi
    done < <(find "$LOG_DIR" -name "server-*.log" -mtime +$COMPRESS_AGE_DAYS -type f -print0 2>/dev/null)
    
    if [ $compressed_count -eq 0 ]; then
        log_info "没有需要压缩的文件"
    else
        log_success "成功压缩 $compressed_count 个文件"
    fi
}

# 清理过期日志
cleanup_logs() {
    check_log_dir
    
    log_info "=== 清理过期日志文件 ==="
    log_info "删除 $MAX_AGE_DAYS 天前的日志文件..."
    
    deleted_count=0
    while IFS= read -r -d '' file; do
        if [ -f "$file" ]; then
            log_info "删除文件: $(basename "$file")"
            if rm -f "$file"; then
                ((deleted_count++))
            else
                log_error "删除失败: $file"
            fi
        fi
    done < <(find "$LOG_DIR" -name "server-*.log*" -mtime +$MAX_AGE_DAYS -type f -print0 2>/dev/null)
    
    if [ $deleted_count -eq 0 ]; then
        log_info "没有需要清理的文件"
    else
        log_success "成功删除 $deleted_count 个文件"
    fi
}

# 完整维护
maintain_logs() {
    log_info "=== 开始日志维护 ==="
    echo "配置信息:"
    echo "  日志目录: $LOG_DIR"
    echo "  压缩阈值: $COMPRESS_AGE_DAYS 天"
    echo "  保留期限: $MAX_AGE_DAYS 天"
    echo "  大小限制: $MAX_SIZE_GB GB"
    echo ""
    
    # 显示维护前统计
    show_stats
    echo ""
    
    # 执行压缩
    compress_logs
    echo ""
    
    # 执行清理
    cleanup_logs
    echo ""
    
    # 显示维护后统计
    log_info "=== 维护完成后统计 ==="
    show_stats
    
    log_success "日志维护完成"
}

# 监控磁盘使用
monitor_logs() {
    check_log_dir
    
    total_size_gb=$(get_dir_size_gb)
    
    if awk "BEGIN {exit !($total_size_gb > $MAX_SIZE_GB)}"; then
        log_error "警告：日志目录大小超过限制!"
        echo "当前大小: ${total_size_gb}GB"
        echo "限制大小: ${MAX_SIZE_GB}GB"
        echo "建议执行: $0 maintain"
        exit 1
    else
        log_success "日志目录大小正常: ${total_size_gb}GB <= ${MAX_SIZE_GB}GB"
    fi
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -d|--log-dir)
                LOG_DIR="$2"
                shift 2
                ;;
            -a|--max-age)
                MAX_AGE_DAYS="$2"
                shift 2
                ;;
            -c|--compress-age)
                COMPRESS_AGE_DAYS="$2"
                shift 2
                ;;
            -s|--max-size)
                MAX_SIZE_GB="$2"
                shift 2
                ;;
            -q|--quiet)
                VERBOSE="false"
                shift
                ;;
            -v|--verbose)
                VERBOSE="true"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            stats|compress|cleanup|maintain|monitor)
                ACTION="$1"
                shift
                ;;
            *)
                log_error "未知参数: $1"
                echo "使用 $0 --help 查看帮助信息"
                exit 1
                ;;
        esac
    done
}

# 主函数
main() {
    # 检查依赖
    if ! command -v find >/dev/null 2>&1; then
        log_error "缺少依赖: find"
        exit 1
    fi
    
    if ! command -v gzip >/dev/null 2>&1; then
        log_error "缺少依赖: gzip"
        exit 1
    fi
    
    # 解析参数
    ACTION=""
    parse_args "$@"
    
    # 默认操作
    if [ -z "$ACTION" ]; then
        ACTION="stats"
    fi
    
    # 执行操作
    case $ACTION in
        stats)
            show_stats
            ;;
        compress)
            compress_logs
            ;;
        cleanup)
            cleanup_logs
            ;;
        maintain)
            maintain_logs
            ;;
        monitor)
            monitor_logs
            ;;
        *)
            log_error "未知操作: $ACTION"
            echo "使用 $0 --help 查看帮助信息"
            exit 1
            ;;
    esac
}

# 脚本入口
main "$@" 