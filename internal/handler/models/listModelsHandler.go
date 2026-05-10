// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package models

import (
	"net/http"

	"aipivot/internal/logic/models"
	"aipivot/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取可用模型列表
func ListModelsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := models.NewListModelsLogic(r.Context(), svcCtx)
		resp, err := l.ListModels()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
