package errors

import (
	"fmt"
	"strings"
)

// ErrorSuggestion 错误建议结构
type ErrorSuggestion struct {
	Error       string   `json:"error"`
	Suggestions []string `json:"suggestions"`
	Severity    string   `json:"severity"`
}

// FormatErrorMessage 格式化错误消息，提供用户友好的输出
func FormatErrorMessage(err error) string {
	if helmErr, ok := err.(*HelmForgeError); ok {
		return formatHelmForgeError(helmErr)
	}
	return err.Error()
}

// formatHelmForgeError 格式化 HelmForge 错误
func formatHelmForgeError(helmErr *HelmForgeError) string {
	var sb strings.Builder

	// 错误标题
	sb.WriteString("错误: ")
	sb.WriteString(helmErr.Message)
	sb.WriteString("\n")

	// 错误代码
	sb.WriteString("错误代码: ")
	sb.WriteString(string(helmErr.Code))
	sb.WriteString("\n")

	// 原始错误
	if helmErr.Err != nil {
		sb.WriteString("原始错误: ")
		sb.WriteString(helmErr.Err.Error())
		sb.WriteString("\n")
	}

	// 建议
	suggestions := getErrorSuggestions(helmErr.Code)
	if len(suggestions) > 0 {
		sb.WriteString("\n建议解决方案:\n")
		for i, suggestion := range suggestions {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, suggestion))
		}
	}

	return sb.String()
}

// getErrorSuggestions 根据错误代码获取建议
func getErrorSuggestions(code ErrorCode) []string {
	suggestions := map[ErrorCode][]string{
		ErrCodeFileNotFound: {
			"检查文件路径是否正确",
			"确认文件是否存在",
			"检查文件权限",
		},
		ErrCodeFileRead: {
			"检查文件权限",
			"确认文件未被其他进程占用",
			"检查文件系统是否有足够空间",
		},
		ErrCodeInvalidFormat: {
			"检查文件格式是否正确",
			"验证 YAML 语法是否有效",
			"参考示例文件格式",
		},
		ErrCodeInvalidConfig: {
			"检查配置文件中的参数设置",
			"验证必需的配置项是否提供",
			"检查配置值的格式是否正确",
		},
		ErrCodeMissingConfig: {
			"检查配置文件中是否包含所有必需字段",
			"参考配置文档确保完整性",
		},
		ErrCodeConversion: {
			"检查输入配置是否符合转换要求",
			"验证依赖关系是否正确",
			"尝试使用简化的配置重新转换",
		},
		ErrCodeValidation: {
			"检查生成的 Helm Chart 结构",
			"运行 helm lint 查看具体错误",
			"验证 values.yaml 文件格式",
		},
		ErrCodeExternalTool: {
			"检查外部工具是否正确安装",
			"验证工具版本是否兼容",
			"查看工具的详细错误日志",
		},
		ErrCodeToolUnavailable: {
			"安装所需的外部工具",
			"检查系统环境变量",
			"使用工具包的内置功能替代",
		},
	}

	return suggestions[code]
}

// GetUserFriendlyMessage 获取用户友好的错误消息
func GetUserFriendlyMessage(err error) string {
	if helmErr, ok := err.(*HelmForgeError); ok {
		suggestions := getErrorSuggestions(helmErr.Code)
		if len(suggestions) > 0 {
			return fmt.Sprintf("%s\n\n提示: %s", helmErr.Message, suggestions[0])
		}
		return helmErr.Message
	}
	return err.Error()
}

// IsRecoverable 判断错误是否可以恢复
func IsRecoverable(err error) bool {
	if helmErr, ok := err.(*HelmForgeError); ok {
		switch helmErr.Code {
		case ErrCodeFileNotFound,
			ErrCodeInvalidConfig,
			ErrCodeMissingConfig,
			ErrCodeValidation:
			return true
		default:
			return false
		}
	}
	return false
}

// GetErrorSeverity 获取错误严重性级别
func GetErrorSeverity(err error) string {
	if helmErr, ok := err.(*HelmForgeError); ok {
		switch helmErr.Code {
		case ErrCodeSystemError, ErrCodeExternalTool:
			return "CRITICAL"
		case ErrCodeInvalidConfig, ErrCodeMissingConfig:
			return "ERROR"
		case ErrCodeValidation:
			return "WARNING"
		default:
			return "INFO"
		}
	}
	return "ERROR"
}

// RetryableError 表示可以重试的错误接口
type RetryableError interface {
	MaxRetries() int
	ShouldRetry() bool
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries int
	Delay      int // milliseconds
}

// NewRetryableError 创建可重试的错误
func NewRetryableError(err error, config RetryConfig) *RetryableErrorWrapper {
	return &RetryableErrorWrapper{
		err:    err,
		config: config,
	}
}

// RetryableErrorWrapper 可重试错误包装器
type RetryableErrorWrapper struct {
	err     error
	config  RetryConfig
	current int
}

// Error 实现 error 接口
func (r *RetryableErrorWrapper) Error() string {
	return r.err.Error()
}

// Unwrap 实现错误包装
func (r *RetryableErrorWrapper) Unwrap() error {
	return r.err
}

// MaxRetries 返回最大重试次数
func (r *RetryableErrorWrapper) MaxRetries() int {
	return r.config.MaxRetries
}

// ShouldRetry 判断是否应该重试
func (r *RetryableErrorWrapper) ShouldRetry() bool {
	r.current++
	return r.current <= r.config.MaxRetries
}

// Increment 重试计数器
func (r *RetryableErrorWrapper) Increment() {
	r.current++
}
