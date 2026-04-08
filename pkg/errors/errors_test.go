package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelmForgeError_Error(t *testing.T) {
	// 测试基本错误
	helmErr := New(ErrCodeFileNotFound, "文件未找到")
	assert.Contains(t, helmErr.Error(), string(ErrCodeFileNotFound))
	assert.Contains(t, helmErr.Error(), "文件未找到")

	// 测试包装错误
	originalErr := errors.New("原始错误")
	wrappedErr := Wrap(originalErr, ErrCodeFileRead, "读取失败")
	assert.Contains(t, wrappedErr.Error(), string(ErrCodeFileRead))
	assert.Contains(t, wrappedErr.Error(), "读取失败")
	assert.Contains(t, wrappedErr.Error(), "原始错误")
}

func TestHelmForgeError_Unwrap(t *testing.T) {
	originalErr := errors.New("原始错误")
	wrappedErr := Wrap(originalErr, ErrCodeFileRead, "读取失败")

	assert.Equal(t, originalErr, wrappedErr.Unwrap())
}

func TestHelmForgeError_WithDetails(t *testing.T) {
	helmErr := New(ErrCodeInvalidConfig, "配置无效")
	details := map[string]string{"field": "port", "value": "abc"}
	helmErr.WithDetails(details)

	assert.Equal(t, details, helmErr.Details)
}

func TestErrorConstructors(t *testing.T) {
	// 测试各种错误构造函数
	err := FileNotFound("test.yaml")
	assert.Equal(t, ErrCodeFileNotFound, err.Code)
	assert.Contains(t, err.Message, "test.yaml")

	err = FileReadError("test.yaml", errors.New("权限拒绝"))
	assert.Equal(t, ErrCodeFileRead, err.Code)
	assert.NotNil(t, err.Err)

	err = InvalidConfig("配置项无效")
	assert.Equal(t, ErrCodeInvalidConfig, err.Code)
	assert.Contains(t, err.Message, "配置项无效")

	err = MissingConfig("host")
	assert.Equal(t, ErrCodeMissingConfig, err.Code)
	assert.Contains(t, err.Message, "host")

	err = ConversionError("转换失败")
	assert.Equal(t, ErrCodeConversion, err.Code)

	err = ValidationError("验证失败")
	assert.Equal(t, ErrCodeValidation, err.Code)

	err = ToolUnavailable("kompose")
	assert.Equal(t, ErrCodeToolUnavailable, err.Code)
	assert.Contains(t, err.Message, "kompose")
}

func TestIsHelmForgeError(t *testing.T) {
	// 测试 HelmForge 错误
	helmErr := New(ErrCodeFileNotFound, "文件未找到")
	assert.True(t, IsHelmForgeError(helmErr))

	// 测试普通错误
	normalErr := errors.New("普通错误")
	assert.False(t, IsHelmForgeError(normalErr))
}

func TestGetErrorCode(t *testing.T) {
	// 测试获取错误代码
	helmErr := New(ErrCodeInvalidConfig, "配置无效")
	assert.Equal(t, ErrCodeInvalidConfig, GetErrorCode(helmErr))

	// 测试普通错误
	normalErr := errors.New("普通错误")
	assert.Empty(t, GetErrorCode(normalErr))
}

func TestFormatErrorMessage(t *testing.T) {
	// 测试格式化 HelmForge 错误
	helmErr := New(ErrCodeFileNotFound, "文件未找到")
	formatted := FormatErrorMessage(helmErr)

	assert.Contains(t, formatted, "错误: 文件未找到")
	assert.Contains(t, formatted, "错误代码: FILE_NOT_FOUND")
	assert.Contains(t, formatted, "建议解决方案")

	// 测试普通错误
	normalErr := errors.New("普通错误")
	assert.Equal(t, "普通错误", FormatErrorMessage(normalErr))
}

func TestGetUserFriendlyMessage(t *testing.T) {
	// 测试获取用户友好的消息
	helmErr := New(ErrCodeFileNotFound, "文件未找到")
	friendly := GetUserFriendlyMessage(helmErr)

	assert.Contains(t, friendly, "文件未找到")
	assert.Contains(t, friendly, "提示:")

	// 测试没有建议的错误
	helmErr2 := New(ErrCodeSystemError, "系统错误")
	friendly2 := GetUserFriendlyMessage(helmErr2)

	assert.Contains(t, friendly2, "系统错误")
	// 不应该包含提示
	assert.NotContains(t, friendly2, "提示:")
}

func TestIsRecoverable(t *testing.T) {
	// 测试可恢复的错误
	recoverableErr := New(ErrCodeFileNotFound, "文件未找到")
	assert.True(t, IsRecoverable(recoverableErr))

	recoverableErr2 := New(ErrCodeInvalidConfig, "配置无效")
	assert.True(t, IsRecoverable(recoverableErr2))

	// 测试不可恢复的错误
	nonRecoverableErr := New(ErrCodeSystemError, "系统错误")
	assert.False(t, IsRecoverable(nonRecoverableErr))

	nonRecoverableErr2 := New(ErrCodeExternalTool, "外部工具错误")
	assert.False(t, IsRecoverable(nonRecoverableErr2))

	// 测试普通错误
	normalErr := errors.New("普通错误")
	assert.False(t, IsRecoverable(normalErr))
}

func TestGetErrorSeverity(t *testing.T) {
	// 测试严重性级别
	criticalErr := New(ErrCodeSystemError, "系统错误")
	assert.Equal(t, "CRITICAL", GetErrorSeverity(criticalErr))

	externalToolErr := New(ErrCodeExternalTool, "外部工具错误")
	assert.Equal(t, "CRITICAL", GetErrorSeverity(externalToolErr))

	errorErr := New(ErrCodeInvalidConfig, "配置无效")
	assert.Equal(t, "ERROR", GetErrorSeverity(errorErr))

	validationErr := New(ErrCodeValidation, "验证失败")
	assert.Equal(t, "WARNING", GetErrorSeverity(validationErr))

	// 测试普通错误
	normalErr := errors.New("普通错误")
	assert.Equal(t, "ERROR", GetErrorSeverity(normalErr))
}

func TestRetryableErrorWrapper(t *testing.T) {
	// 测试可重试错误包装器
	originalErr := errors.New("临时错误")
	config := RetryConfig{MaxRetries: 3, Delay: 100}
	retryableErr := NewRetryableError(originalErr, config)

	// 测试错误消息
	assert.Contains(t, retryableErr.Error(), "临时错误")

	// 测试解包
	assert.Equal(t, originalErr, retryableErr.Unwrap())

	// 测试最大重试次数
	assert.Equal(t, 3, retryableErr.MaxRetries())

	// 测试重试逻辑
	assert.True(t, retryableErr.ShouldRetry())  // 第一次
	assert.True(t, retryableErr.ShouldRetry())  // 第二次
	assert.True(t, retryableErr.ShouldRetry())  // 第三次
	assert.False(t, retryableErr.ShouldRetry()) // 第四次，超过最大重试次数

	// 重置重试计数器
	retryableErr.current = 0
	assert.True(t, retryableErr.ShouldRetry())
}

func TestRetryableErrorWrapper_Increment(t *testing.T) {
	originalErr := errors.New("临时错误")
	config := RetryConfig{MaxRetries: 2, Delay: 100}
	retryableErr := NewRetryableError(originalErr, config)

	// 测试递增
	assert.Equal(t, 0, retryableErr.current)
	retryableErr.Increment()
	assert.Equal(t, 1, retryableErr.current)
	retryableErr.Increment()
	assert.Equal(t, 2, retryableErr.current)
}

func TestWrapf(t *testing.T) {
	// 测试带格式化的包装
	originalErr := errors.New("原始错误")
	wrappedErr := Wrapf(originalErr, ErrCodeInvalidFormat, "文件 %s 格式无效", "test.yaml")

	assert.Equal(t, ErrCodeInvalidFormat, wrappedErr.Code)
	assert.Contains(t, wrappedErr.Message, "test.yaml")
	assert.Contains(t, wrappedErr.Message, "格式无效")
	assert.NotNil(t, wrappedErr.Err)
}

func TestGetErrorSuggestions(t *testing.T) {
	// 测试获取错误建议
	suggestions := getErrorSuggestions(ErrCodeFileNotFound)
	assert.NotEmpty(t, suggestions)
	assert.Contains(t, suggestions, "检查文件路径是否正确")

	suggestions2 := getErrorSuggestions(ErrCodeInvalidFormat)
	assert.NotEmpty(t, suggestions2)
	assert.Contains(t, suggestions2, "检查文件格式是否正确")

	// 测试没有建议的错误代码
	suggestions3 := getErrorSuggestions("UNKNOWN_CODE")
	assert.Empty(t, suggestions3)
}
