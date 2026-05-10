// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowledge

import (
	"net/http"

	"aipivot/internal/logic/knowledge"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 上传文档到知识库（multipart/form-data）
func UploadDocumentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UploadDocumentRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowledge.NewUploadDocumentLogic(r.Context(), svcCtx)
		resp, err := l.UploadDocument(&req, r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
