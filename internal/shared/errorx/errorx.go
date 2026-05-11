package errorx

import "fmt"

const (
	CodeSuccess    = 0
	CodeFailed     = -1
	CodeBadRequest = 400
	CodeUnauth     = 401
	CodeForbid     = 403
	CodeNotFound   = 404
	CodeNotAllowed = 405

	// AI-specific error codes
	CodeLLMTimeout     = 1001
	CodeLLMUnavailable = 1002
	CodeTokenExceeded  = 1003
	CodeKnowledgeMiss  = 1004
)

// BusinessError 统一业务错误类型，支持错误包装以保留原始 cause 便于日志追踪。
type BusinessError struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Cause error  `json:"-"` // 原始错误，不序列化到响应
}

func (e *BusinessError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("biz error: code=%d, msg=%s, cause=%v", e.Code, e.Msg, e.Cause)
	}
	return fmt.Sprintf("biz error: code=%d, msg=%s", e.Code, e.Msg)
}

// Unwrap 支持 errors.Is / errors.As 向下穿透到原始错误
func (e *BusinessError) Unwrap() error {
	return e.Cause
}

// Wrap 在已有 BusinessError 上附加原始错误，返回新实例（不修改原对象）
func Wrap(err error, bizErr *BusinessError) *BusinessError {
	return &BusinessError{
		Code:  bizErr.Code,
		Msg:   bizErr.Msg,
		Cause: err,
	}
}

func NewInternalError(msg string) *BusinessError {
	return &BusinessError{Code: CodeFailed, Msg: msg}
}

func NewBusinessError(code int, msg string) *BusinessError {
	return &BusinessError{Code: code, Msg: msg}
}

func NewUnauthError(msg string) *BusinessError {
	return &BusinessError{Code: CodeUnauth, Msg: msg}
}

func NewForbidError(msg string) *BusinessError {
	return &BusinessError{Code: CodeForbid, Msg: msg}
}

func NewNotFoundError(msg string) *BusinessError {
	return &BusinessError{Code: CodeNotFound, Msg: msg}
}
