package errors

import (
	"fmt"
)

// ErrorCode 错误代码类型
type ErrorCode string

const (
	// 文件操作错误
	ErrCodeFileNotFound  ErrorCode = "FILE_NOT_FOUND"
	ErrCodeFileRead      ErrorCode = "FILE_READ_ERROR"
	ErrCodeFileWrite     ErrorCode = "FILE_WRITE_ERROR"
	ErrCodeInvalidFormat ErrorCode = "INVALID_FORMAT"

	// 配置错误
	ErrCodeInvalidConfig ErrorCode = "INVALID_CONFIG"
	ErrCodeMissingConfig ErrorCode = "MISSING_CONFIG"

	// 转换错误
	ErrCodeConversion ErrorCode = "CONVERSION_ERROR"
	ErrCodeDependency ErrorCode = "DEPENDENCY_ERROR"

	// 验证错误
	ErrCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrCodeMissingField ErrorCode = "MISSING_FIELD"

	// 外部工具错误
	ErrCodeExternalTool    ErrorCode = "EXTERNAL_TOOL_ERROR"
	ErrCodeToolUnavailable ErrorCode = "TOOL_UNAVAILABLE"

	// 系统错误
	ErrCodeSystemError ErrorCode = "SYSTEM_ERROR"
)

// HelmForgeError 自定义错误类型
type HelmForgeError struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Err     error       `json:"-"` // 原始错误，不序列化
}

// Error 实现 error 接口
func (e *HelmForgeError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现错误包装
func (e *HelmForgeError) Unwrap() error {
	return e.Err
}

// New 创建新的 HelmForge 错误
func New(code ErrorCode, message string) *HelmForgeError {
	return &HelmForgeError{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装现有错误
func Wrap(err error, code ErrorCode, message string) *HelmForgeError {
	return &HelmForgeError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Wrapf 带格式化地包装现有错误
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *HelmForgeError {
	return &HelmForgeError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

// WithDetails 添加错误详细信息
func (e *HelmForgeError) WithDetails(details interface{}) *HelmForgeError {
	e.Details = details
	return e
}

// 文件操作相关错误
func FileNotFound(filename string) *HelmForgeError {
	return New(ErrCodeFileNotFound, fmt.Sprintf("文件未找到: %s", filename))
}

func FileReadError(filename string, err error) *HelmForgeError {
	return Wrap(err, ErrCodeFileRead, fmt.Sprintf("读取文件失败: %s", filename))
}

func FileWriteError(filename string, err error) *HelmForgeError {
	return Wrap(err, ErrCodeFileWrite, fmt.Sprintf("写入文件失败: %s", filename))
}

func InvalidFormat(filename string, format string) *HelmForgeError {
	return New(ErrCodeInvalidFormat, fmt.Sprintf("文件格式无效，期望 %s: %s", format, filename))
}

// 配置相关错误
func InvalidConfig(message string) *HelmForgeError {
	return New(ErrCodeInvalidConfig, message)
}

func MissingConfig(field string) *HelmForgeError {
	return New(ErrCodeMissingConfig, fmt.Sprintf("缺少必需的配置字段: %s", field))
}

// 转换相关错误
func ConversionError(message string) *HelmForgeError {
	return New(ErrCodeConversion, message)
}

func DependencyError(message string) *HelmForgeError {
	return New(ErrCodeDependency, message)
}

// 验证相关错误
func ValidationError(message string) *HelmForgeError {
	return New(ErrCodeValidation, message)
}

func MissingField(field string) *HelmForgeError {
	return New(ErrCodeMissingField, fmt.Sprintf("缺少必需的字段: %s", field))
}

// 外部工具相关错误
func ExternalToolError(tool string, err error) *HelmForgeError {
	return Wrap(err, ErrCodeExternalTool, fmt.Sprintf("外部工具执行失败: %s", tool))
}

func ToolUnavailable(tool string) *HelmForgeError {
	return New(ErrCodeToolUnavailable, fmt.Sprintf("工具不可用: %s", tool))
}

// 系统错误
func SystemError(message string) *HelmForgeError {
	return New(ErrCodeSystemError, message)
}

// IsHelmForgeError 检查是否为 HelmForge 错误
func IsHelmForgeError(err error) bool {
	_, ok := err.(*HelmForgeError)
	return ok
}

// GetErrorCode 获取错误代码
func GetErrorCode(err error) ErrorCode {
	if helmErr, ok := err.(*HelmForgeError); ok {
		return helmErr.Code
	}
	return ""
}
