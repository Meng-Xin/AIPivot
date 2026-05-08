package errorx

import (
	"context"
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type ErrorResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Timestamp int64  `json:"timestamp"`
}

func RegisterErrorHandler() {
	httpx.SetErrorHandlerCtx(func(ctx context.Context, err error) (int, any) {
		var bizErr *BusinessError
		if errors.As(err, &bizErr) {
			return http.StatusOK, bizErr
		}

		logx.WithContext(ctx).Errorf("unexpected error: %v", err)
		return http.StatusInternalServerError, &BusinessError{
			Code: CodeFailed,
			Msg:  "服务器内部错误",
		}
	})
}
