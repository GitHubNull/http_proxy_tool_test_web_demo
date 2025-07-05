package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogConfig 日志配置
type LogConfig struct {
	LogDir        string   // 日志目录
	LogLevel      LogLevel // 日志级别
	MaxFileSize   int64    // 最大文件大小（MB）
	MaxBackups    int      // 最大备份文件数
	MaxAge        int      // 最大保留天数
	Compress      bool     // 是否压缩旧日志
	EnableConsole bool     // 是否输出到控制台
	DateFormat    string   // 日期格式
	TimeFormat    string   // 时间格式
}

// Logger 日志管理器
type Logger struct {
	config      LogConfig
	currentLog  *os.File
	currentDate string
	mutex       sync.RWMutex
	logger      *log.Logger
}

// NewLogger 创建新的日志管理器
func NewLogger(config LogConfig) (*Logger, error) {
	// 设置默认值
	if config.LogDir == "" {
		config.LogDir = "logs"
	}
	if config.MaxFileSize == 0 {
		config.MaxFileSize = 50 // 50MB
	}
	if config.MaxBackups == 0 {
		config.MaxBackups = 30
	}
	if config.MaxAge == 0 {
		config.MaxAge = 30 // 30天
	}
	if config.DateFormat == "" {
		config.DateFormat = "2006-01-02"
	}
	if config.TimeFormat == "" {
		config.TimeFormat = "2006-01-02 15:04:05"
	}

	// 创建日志目录
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %v", err)
	}

	logger := &Logger{
		config: config,
	}

	// 初始化日志文件
	if err := logger.initLogFile(); err != nil {
		return nil, fmt.Errorf("初始化日志文件失败: %v", err)
	}

	// 启动日志清理协程
	go logger.cleanupRoutine()

	return logger, nil
}

// initLogFile 初始化日志文件
func (l *Logger) initLogFile() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	now := time.Now()
	dateStr := now.Format(l.config.DateFormat)

	// 如果日期变化，需要切换日志文件
	if l.currentDate != dateStr || l.currentLog == nil {
		if l.currentLog != nil {
			l.currentLog.Close()
		}

		// 创建新的日志文件
		logFileName := fmt.Sprintf("server-%s.log", dateStr)
		logFilePath := filepath.Join(l.config.LogDir, logFileName)

		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("打开日志文件失败: %v", err)
		}

		l.currentLog = file
		l.currentDate = dateStr

		// 设置日志输出
		var writer io.Writer = file
		if l.config.EnableConsole {
			writer = io.MultiWriter(os.Stdout, file)
		}

		l.logger = log.New(writer, "", 0)
	}

	return nil
}

// checkRotation 检查是否需要轮转日志
func (l *Logger) checkRotation() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.currentLog == nil {
		return nil
	}

	// 检查日期是否变化
	now := time.Now()
	dateStr := now.Format(l.config.DateFormat)
	if l.currentDate != dateStr {
		return l.initLogFile()
	}

	// 检查文件大小
	fileInfo, err := l.currentLog.Stat()
	if err != nil {
		return err
	}

	// 如果文件大小超过限制，进行轮转
	if fileInfo.Size() > l.config.MaxFileSize*1024*1024 {
		l.currentLog.Close()

		// 重命名当前文件
		oldPath := filepath.Join(l.config.LogDir, fmt.Sprintf("server-%s.log", l.currentDate))
		timestamp := now.Format("15-04-05")
		newPath := filepath.Join(l.config.LogDir, fmt.Sprintf("server-%s-%s.log", l.currentDate, timestamp))

		if err := os.Rename(oldPath, newPath); err != nil {
			return fmt.Errorf("重命名日志文件失败: %v", err)
		}

		// 压缩旧日志文件
		if l.config.Compress {
			go l.compressLogFile(newPath)
		}

		return l.initLogFile()
	}

	return nil
}

// compressLogFile 压缩日志文件
func (l *Logger) compressLogFile(filePath string) {
	inputFile, err := os.Open(filePath)
	if err != nil {
		log.Printf("打开待压缩日志文件失败: %v", err)
		return
	}
	defer inputFile.Close()

	gzipPath := filePath + ".gz"
	outputFile, err := os.Create(gzipPath)
	if err != nil {
		log.Printf("创建压缩文件失败: %v", err)
		return
	}
	defer outputFile.Close()

	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	if _, err := io.Copy(gzipWriter, inputFile); err != nil {
		log.Printf("压缩日志文件失败: %v", err)
		return
	}

	// 删除原文件
	if err := os.Remove(filePath); err != nil {
		log.Printf("删除原日志文件失败: %v", err)
	}
}

// cleanupRoutine 日志清理协程
func (l *Logger) cleanupRoutine() {
	ticker := time.NewTicker(6 * time.Hour) // 每6小时清理一次
	defer ticker.Stop()

	for range ticker.C {
		l.cleanupOldLogs()
	}
}

// cleanupOldLogs 清理旧日志文件
func (l *Logger) cleanupOldLogs() {
	entries, err := os.ReadDir(l.config.LogDir)
	if err != nil {
		log.Printf("读取日志目录失败: %v", err)
		return
	}

	var logFiles []os.FileInfo
	cutoffTime := time.Now().AddDate(0, 0, -l.config.MaxAge)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, "server-") ||
			(!strings.HasSuffix(name, ".log") && !strings.HasSuffix(name, ".log.gz")) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// 按时间删除
		if info.ModTime().Before(cutoffTime) {
			logPath := filepath.Join(l.config.LogDir, name)
			if err := os.Remove(logPath); err != nil {
				log.Printf("删除过期日志文件失败 %s: %v", name, err)
			} else {
				log.Printf("删除过期日志文件: %s", name)
			}
			continue
		}

		logFiles = append(logFiles, info)
	}

	// 按修改时间排序
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].ModTime().After(logFiles[j].ModTime())
	})

	// 按数量删除
	if len(logFiles) > l.config.MaxBackups {
		for _, file := range logFiles[l.config.MaxBackups:] {
			logPath := filepath.Join(l.config.LogDir, file.Name())
			if err := os.Remove(logPath); err != nil {
				log.Printf("删除多余日志文件失败 %s: %v", file.Name(), err)
			} else {
				log.Printf("删除多余日志文件: %s", file.Name())
			}
		}
	}
}

// logWithLevel 按级别记录日志
func (l *Logger) logWithLevel(level LogLevel, format string, args ...interface{}) {
	if level < l.config.LogLevel {
		return
	}

	// 检查是否需要轮转
	if err := l.checkRotation(); err != nil {
		log.Printf("检查日志轮转失败: %v", err)
		return
	}

	timestamp := time.Now().Format(l.config.TimeFormat)
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [%s] %s", timestamp, level.String(), message)

	l.mutex.RLock()
	if l.logger != nil {
		l.logger.Println(logLine)
	}
	l.mutex.RUnlock()
}

// Debug 记录调试日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.logWithLevel(DEBUG, format, args...)
}

// Info 记录信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.logWithLevel(INFO, format, args...)
}

// Warn 记录警告日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.logWithLevel(WARN, format, args...)
}

// Error 记录错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.logWithLevel(ERROR, format, args...)
}

// Fatal 记录致命错误日志并退出
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.logWithLevel(FATAL, format, args...)
	os.Exit(1)
}

// GetLogStats 获取日志统计信息
func (l *Logger) GetLogStats() map[string]interface{} {
	entries, err := os.ReadDir(l.config.LogDir)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("读取日志目录失败: %v", err),
		}
	}

	var totalSize int64
	var fileCount int
	var compressedCount int
	var oldestDate string
	var newestDate string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, "server-") ||
			(!strings.HasSuffix(name, ".log") && !strings.HasSuffix(name, ".log.gz")) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileCount++
		totalSize += info.Size()

		if strings.HasSuffix(name, ".gz") {
			compressedCount++
		}

		// 从文件名提取日期
		if strings.Contains(name, "server-") {
			parts := strings.Split(name, "-")
			if len(parts) >= 2 {
				dateStr := parts[1]
				if len(dateStr) >= 10 {
					dateStr = dateStr[:10]
					if oldestDate == "" || dateStr < oldestDate {
						oldestDate = dateStr
					}
					if newestDate == "" || dateStr > newestDate {
						newestDate = dateStr
					}
				}
			}
		}
	}

	return map[string]interface{}{
		"total_files":         fileCount,
		"total_size_mb":       float64(totalSize) / 1024 / 1024,
		"compressed_files":    compressedCount,
		"oldest_log_date":     oldestDate,
		"newest_log_date":     newestDate,
		"log_directory":       l.config.LogDir,
		"max_file_size_mb":    l.config.MaxFileSize,
		"max_backups":         l.config.MaxBackups,
		"max_age_days":        l.config.MaxAge,
		"compression_enabled": l.config.Compress,
		"current_date":        l.currentDate,
	}
}

// Close 关闭日志器
func (l *Logger) Close() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.currentLog != nil {
		return l.currentLog.Close()
	}
	return nil
}

// 解析日志级别
func parseLogLevel(levelStr string) LogLevel {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

// 全局日志器
var globalLogger *Logger

// InitGlobalLogger 初始化全局日志器
func InitGlobalLogger() error {
	logLevel := parseLogLevel(os.Getenv("LOG_LEVEL"))

	config := LogConfig{
		LogDir:        "logs",
		LogLevel:      logLevel,
		MaxFileSize:   50, // 50MB
		MaxBackups:    30,
		MaxAge:        30, // 30天
		Compress:      true,
		EnableConsole: true,
	}

	// 从环境变量覆盖配置
	if logsPath := os.Getenv("LOGS_PATH"); logsPath != "" {
		config.LogDir = logsPath
	}

	if maxSizeStr := os.Getenv("LOG_MAX_SIZE"); maxSizeStr != "" {
		if maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
			config.MaxFileSize = maxSize
		}
	}

	if maxBackupsStr := os.Getenv("LOG_MAX_BACKUPS"); maxBackupsStr != "" {
		if maxBackups, err := strconv.Atoi(maxBackupsStr); err == nil {
			config.MaxBackups = maxBackups
		}
	}

	if maxAgeStr := os.Getenv("LOG_MAX_AGE"); maxAgeStr != "" {
		if maxAge, err := strconv.Atoi(maxAgeStr); err == nil {
			config.MaxAge = maxAge
		}
	}

	if compressStr := os.Getenv("LOG_COMPRESS"); compressStr != "" {
		config.Compress = strings.ToLower(compressStr) == "true"
	}

	if consoleStr := os.Getenv("LOG_CONSOLE"); consoleStr != "" {
		config.EnableConsole = strings.ToLower(consoleStr) == "true"
	}

	var err error
	globalLogger, err = NewLogger(config)
	if err != nil {
		return err
	}

	// 记录启动日志
	globalLogger.Info("日志系统初始化完成")
	globalLogger.Info("日志配置: 目录=%s, 级别=%s, 最大文件大小=%dMB, 最大备份数=%d, 最大保留天数=%d",
		config.LogDir, config.LogLevel.String(), config.MaxFileSize, config.MaxBackups, config.MaxAge)

	return nil
}

// GetGlobalLogger 获取全局日志器
func GetGlobalLogger() *Logger {
	return globalLogger
}

// 便捷函数
func LogDebug(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(format, args...)
	}
}

func LogInfo(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(format, args...)
	}
}

func LogWarn(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(format, args...)
	}
}

func LogError(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Error(format, args...)
	}
}

func LogFatal(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Fatal(format, args...)
	}
}
