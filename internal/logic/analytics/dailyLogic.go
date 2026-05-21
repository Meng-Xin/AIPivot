package analytics

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnalyticsDailyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAnalyticsDailyLogic 获取日粒度对话统计趋势
func NewAnalyticsDailyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AnalyticsDailyLogic {
	return &AnalyticsDailyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type dailyConvRow struct {
	Date  string `gorm:"column:date"`
	Count int64  `gorm:"column:count"`
}

type dailyMsgRow struct {
	Date         string `gorm:"column:date"`
	MessageCount int64  `gorm:"column:message_count"`
	TokenCount   int64  `gorm:"column:token_count"`
}

func (l *AnalyticsDailyLogic) AnalyticsDaily(req *types.AnalyticsDailyRequest) (*types.AnalyticsDailyResponse, error) {
	tenantID := auth.TenantIDFromContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("unauthenticated")
	}

	days := req.Days
	if days <= 0 {
		days = 7
	}

	db := l.svcCtx.DB.WithContext(l.ctx)

	// 使用参数化 interval 避免 SQL 注入
	var convRows []dailyConvRow
	if err := db.Raw(
		"SELECT TO_CHAR(DATE(created_at), 'YYYY-MM-DD') AS date, COUNT(*) AS count FROM conversations WHERE tenant_id = ? AND created_at >= NOW() - (? * INTERVAL '1 day') GROUP BY DATE(created_at) ORDER BY date ASC",
		tenantID, days,
	).Scan(&convRows).Error; err != nil {
		l.Logger.Errorf("AnalyticsDaily query convs err: %v", err)
		return nil, errorx.NewInternalError("查询分析数据失败")
	}

	var msgRows []dailyMsgRow
	if err := db.Raw(
		"SELECT TO_CHAR(DATE(created_at), 'YYYY-MM-DD') AS date, COUNT(*) AS message_count, COALESCE(SUM(token_count), 0) AS token_count FROM messages WHERE tenant_id = ? AND role = 'assistant' AND created_at >= NOW() - (? * INTERVAL '1 day') GROUP BY DATE(created_at) ORDER BY date ASC",
		tenantID, days,
	).Scan(&msgRows).Error; err != nil {
		l.Logger.Errorf("AnalyticsDaily query messages err: %v", err)
		return nil, errorx.NewInternalError("查询分析数据失败")
	}

	// 合并两个查询结果到 date-keyed map
	type dayData struct {
		convCount  int64
		msgCount   int64
		tokenCount int64
	}
	dayMap := make(map[string]*dayData, days)
	for _, row := range convRows {
		dayMap[row.Date] = &dayData{convCount: row.Count}
	}
	for _, row := range msgRows {
		if d, ok := dayMap[row.Date]; ok {
			d.msgCount = row.MessageCount
			d.tokenCount = row.TokenCount
		} else {
			dayMap[row.Date] = &dayData{msgCount: row.MessageCount, tokenCount: row.TokenCount}
		}
	}

	// 按时间顺序填充所有日期（补零），确保前端图表无断点
	now := time.Now()
	result := make([]types.DailyStatPoint, 0, days)
	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		point := types.DailyStatPoint{Date: date}
		if d, ok := dayMap[date]; ok {
			point.ConversationCount = d.convCount
			point.MessageCount = d.msgCount
			point.TokenCount = d.tokenCount
		}
		result = append(result, point)
	}

	return &types.AnalyticsDailyResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data:      result,
	}, nil
}
