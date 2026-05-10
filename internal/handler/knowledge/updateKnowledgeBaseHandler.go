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

// 更新知识库
func UpdateKnowledgeBaseHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateKnowledgeBaseRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowledge.NewUpdateKnowledgeBaseLogic(r.Context(), svcCtx)
		resp, err := l.UpdateKnowledgeBase(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
