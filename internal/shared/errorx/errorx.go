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

type BusinessError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("biz error: code=%d, msg=%s", e.Code, e.Msg)
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
