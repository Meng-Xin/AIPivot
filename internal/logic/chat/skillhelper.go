package chat

import (
	"context"

	"aipivot/internal/modules/agent"
	"aipivot/internal/modules/agent/tools"
	"aipivot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// loadTenantSkills 从 SkillRepo 加载租户启用的自定义工具，返回 agent.Tool 列表。
// 加载失败时静默降级（不中断对话），因为工具不可用不应阻断 AI 回复。
func loadTenantSkills(ctx context.Context, svcCtx *svc.ServiceContext, tenantID int64) []agent.Tool {
	if svcCtx.SkillRepo == nil || tenantID == 0 {
		return nil
	}

	skillPOs, err := svcCtx.SkillRepo.GetEnabledByTenant(ctx, tenantID)
	if err != nil {
		logx.WithContext(ctx).Errorf("loadTenantSkills err: %v (degrading to no custom tools)", err)
		return nil
	}

	result := make([]agent.Tool, 0, len(skillPOs))
	for _, s := range skillPOs {
		result = append(result, tools.NewHttpToolFromSkill(s))
	}
	return result
}
