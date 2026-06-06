// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package skills

import (
	"net/http"

	"aipivot/internal/logic/skills"
	"aipivot/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取工具列表
func ListSkillsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := skills.NewListSkillsLogic(r.Context(), svcCtx)
		resp, err := l.ListSkills()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
