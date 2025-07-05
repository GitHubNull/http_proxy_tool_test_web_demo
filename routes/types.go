package routes

import (
	"fmt"
	"time"
)

// ApiResponse 通用响应结构
type ApiResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
	RequestID string      `json:"request_id"`
}

// RequestInfo 请求信息结构
type RequestInfo struct {
	Method      string                 `json:"method"`
	URL         string                 `json:"url"`
	Headers     map[string]interface{} `json:"headers"`
	Body        interface{}            `json:"body"`
	Query       map[string]interface{} `json:"query"`
	ClientIP    string                 `json:"client_ip"`
	UserAgent   string                 `json:"user_agent"`
	ContentType string                 `json:"content_type"`
	Cookies     map[string]string      `json:"cookies"`
}

// GenerateRequestID 生成请求ID
func GenerateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// CreateSuccessResponse 创建成功响应
func CreateSuccessResponse(message string, data interface{}) ApiResponse {
	return ApiResponse{
		Code:      200,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: GenerateRequestID(),
	}
}

// CreateErrorResponse 创建错误响应
func CreateErrorResponse(code int, message string) ApiResponse {
	return ApiResponse{
		Code:      code,
		Message:   message,
		Data:      nil,
		Timestamp: time.Now().Unix(),
		RequestID: GenerateRequestID(),
	}
}
