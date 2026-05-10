package errorx

import (
	"context"
	"encoding/json"
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

		// 参数校验 / 请求解析错误，保留原始信息便于前端提示
		logx.WithContext(ctx).Infof("bad request: %v", err)
		return http.StatusBadRequest, &BusinessError{
			Code: CodeBadRequest,
			Msg:  err.Error(),
		}
	})
}

// WriteNotFoundJSON 写入统一 404 JSON 响应
func WriteNotFoundJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(&BusinessError{
		Code: CodeNotFound,
		Msg:  "route not found",
	})
}

// WriteNotAllowedJSON 写入统一 405 JSON 响应
func WriteNotAllowedJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(&BusinessError{
		Code: CodeNotAllowed,
		Msg:  "method not allowed",
	})
}
