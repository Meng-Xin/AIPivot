package response

import "time"

type CommResponse struct {
	Code      int32  `json:"code"`
	Msg       string `json:"msg"`
	Timestamp int64  `json:"timestamp"`
}

func Success(msg string) *CommResponse {
	return &CommResponse{
		Code:      0,
		Msg:       msg,
		Timestamp: time.Now().Unix(),
	}
}

type PageData[T any] struct {
	Total int64 `json:"total"`
	List  []T   `json:"list"`
}

type PageResponse[T any] struct {
	CommResponse
	Data PageData[T] `json:"data"`
}

type DataResponse[T any] struct {
	CommResponse
	Data T `json:"data"`
}
