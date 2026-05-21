package analytics

import (
	"context"
	"math"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnalyticsOverviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAnalyticsOverviewLogic 获取对话分析概览
func NewAnalyticsOverviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AnalyticsOverviewLogic {
	return &AnalyticsOverviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 中间扫描结构体（GORM Raw SQL Scan 目标）
type convStatusRow struct {
	Status string `gorm:"column:status"`
	Count  int64  `gorm:"column:count"`
}

type channelRow struct {
	Channel string `gorm:"column:channel"`
	Count   int64  `gorm:"column:count"`
}

type modelMsgRow struct {
	Model        string `gorm:"column:model"`
	MessageCount int64  `gorm:"column:message_count"`
	TokenCount   int64  `gorm:"column:token_count"`
}

func (l *AnalyticsOverviewLogic) AnalyticsOverview() (*types.AnalyticsOverviewResponse, error) {
	tenantID := auth.TenantIDFromContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("unauthenticated")
	}

	db := l.svcCtx.DB.WithContext(l.ctx)

	// 按状态聚合会话数
	var statusRows []convStatusRow
	if err := db.Raw(
		"SELECT status, COUNT(*) AS count FROM conversations WHERE tenant_id = ? GROUP BY status",
		tenantID,
	).Scan(&statusRows).Error; err != nil {
		l.Logger.Errorf("AnalyticsOverview query status err: %v", err)
		return nil, errorx.NewInternalError("查询分析数据失败")
	}

	var totalConvs, activeConvs, closedConvs, waitingConvs int64
	byStatus := make([]types.ConvStatusStat, 0, len(statusRows))
	for _, row := range statusRows {
		totalConvs += row.Count
		switch row.Status {
		case "active":
			activeConvs = row.Count
		case "closed":
			closedConvs = row.Count
		case "waiting_human":
			waitingConvs = row.Count
		}
		byStatus = append(byStatus, types.ConvStatusStat{Status: row.Status, Count: row.Count})
	}

	// 按渠道聚合会话数
	var channelRows []channelRow
	if err := db.Raw(
		"SELECT channel, COUNT(*) AS count FROM conversations WHERE tenant_id = ? GROUP BY channel",
		tenantID,
	).Scan(&channelRows).Error; err != nil {
		l.Logger.Errorf("AnalyticsOverview query channel err: %v", err)
		return nil, errorx.NewInternalError("查询分析数据失败")
	}

	byChannel := make([]types.ChannelStat, 0, len(channelRows))
	for _, row := range channelRows {
		byChannel = append(byChannel, types.ChannelStat{Channel: row.Channel, Count: row.Count})
	}

	// 按模型聚合 AI 回复消息数和 token 数（仅 assistant 消息）
	var modelRows []modelMsgRow
	if err := db.Raw(
		"SELECT model, COUNT(*) AS message_count, COALESCE(SUM(token_count), 0) AS token_count FROM messages WHERE tenant_id = ? AND role = 'assistant' GROUP BY model",
		tenantID,
	).Scan(&modelRows).Error; err != nil {
		l.Logger.Errorf("AnalyticsOverview query messages err: %v", err)
		return nil, errorx.NewInternalError("查询分析数据失败")
	}

	// 构建模型定价查找表（来自配置）
	priceMap := make(map[string]float64, len(l.svcCtx.Config.ModelPricing))
	for _, p := range l.svcCtx.Config.ModelPricing {
		priceMap[p.Model] = p.PerK
	}

	var totalMessages, totalTokens int64
	var totalCost float64
	modelUsage := make([]types.ModelUsageStat, 0, len(modelRows))
	for _, row := range modelRows {
		totalMessages += row.MessageCount
		totalTokens += row.TokenCount
		cost := float64(row.TokenCount) * priceMap[row.Model] / 1000.0
		totalCost += cost
		modelUsage = append(modelUsage, types.ModelUsageStat{
			Model:         row.Model,
			MessageCount:  row.MessageCount,
			TokenCount:    row.TokenCount,
			EstimatedCost: roundFloat(cost, 6),
		})
	}

	// AI 解决率 = closed / total；转人工率 = waiting_human / total（百分比，保留两位小数）
	var resolveRate, escalationRate float64
	if totalConvs > 0 {
		resolveRate = roundFloat(float64(closedConvs)/float64(totalConvs)*100, 2)
		escalationRate = roundFloat(float64(waitingConvs)/float64(totalConvs)*100, 2)
	}

	return &types.AnalyticsOverviewResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data: types.AnalyticsOverviewData{
			TotalConversations:  totalConvs,
			ActiveConversations: activeConvs,
			TotalMessages:       totalMessages,
			TotalTokens:         totalTokens,
			EstimatedCost:       roundFloat(totalCost, 6),
			AIResolveRate:       resolveRate,
			EscalationRate:      escalationRate,
			ByStatus:            byStatus,
			ByChannel:           byChannel,
			ModelUsage:          modelUsage,
		},
	}, nil
}

func roundFloat(v float64, precision int) float64 {
	p := math.Pow(10, float64(precision))
	return math.Round(v*p) / p
}
