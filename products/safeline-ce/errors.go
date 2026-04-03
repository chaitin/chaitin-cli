package safelinece

import (
	"fmt"
)

// CLIError CLI 错误类型
type CLIError struct {
	Code       int
	Message    string
	Err        error
	StatusCode int // HTTP 状态码（仅 API 错误）
}

func (e *CLIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap 实现 errors.Unwrap
func (e *CLIError) Unwrap() error {
	return e.Err
}

// NewConfigError 创建配置错误
func NewConfigError(msg string, err error) *CLIError {
	return &CLIError{
		Code:    ExitConfigError,
		Message: msg,
		Err:     err,
	}
}

// NewNetworkError 创建网络错误
func NewNetworkError(msg string, err error) *CLIError {
	return &CLIError{
		Code:    ExitNetworkError,
		Message: msg,
		Err:     err,
	}
}

// NewAPIError 创建 API 错误
func NewAPIError(statusCode int, msg string) *CLIError {
	return &CLIError{
		Code:       ExitAPIError,
		Message:    msg,
		StatusCode: statusCode,
	}
}

// IsConfigError 判断是否为配置错误
func IsConfigError(err error) bool {
	if cliErr, ok := err.(*CLIError); ok {
		return cliErr.Code == ExitConfigError
	}
	return false
}

// IsNetworkError 判断是否为网络错误
func IsNetworkError(err error) bool {
	if cliErr, ok := err.(*CLIError); ok {
		return cliErr.Code == ExitNetworkError
	}
	return false
}

// IsAPIError 判断是否为 API 错误
func IsAPIError(err error) bool {
	if cliErr, ok := err.(*CLIError); ok {
		return cliErr.Code == ExitAPIError
	}
	return false
}
