package analytics

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// ExportReportRequest 报表导出请求参数（直接从 query 解析，不走 goctl types）
type ExportReportRequest struct {
	Days int // 统计最近 N 天（1-90）
}

type ExportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewExportLogic 构建报表导出 Logic
func NewExportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExportLogic {
	return &ExportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ExportCSV 导出 SLA 报表为 CSV 字节流，返回 (data, filename, error)
//
// 报表结构：
//   - 文件头（生成时间/统计范围）
//   - 概览 KPI 汇总
//   - 日粒度趋势明细
//   - 模型用量明细
//   - 会话状态分布
//   - 渠道来源分布
func (l *ExportLogic) ExportCSV(req *ExportReportRequest) ([]byte, string, error) {
	tenantID := auth.TenantIDFromContext(l.ctx)
	if tenantID == 0 {
		return nil, "", errorx.NewUnauthError("unauthenticated")
	}

	days := req.Days
	if days <= 0 || days > 90 {
		days = 30
	}

	// 复用已有 Logic 聚合数据，避免重复 SQL
	ovLogic := NewAnalyticsOverviewLogic(l.ctx, l.svcCtx)
	ovResp, err := ovLogic.AnalyticsOverview()
	if err != nil {
		return nil, "", err
	}
	ov := ovResp.Data

	dailyLogic := NewAnalyticsDailyLogic(l.ctx, l.svcCtx)
	dailyResp, err := dailyLogic.AnalyticsDaily(&types.AnalyticsDailyRequest{Days: days})
	if err != nil {
		return nil, "", err
	}

	buf := &bytes.Buffer{}
	// UTF-8 BOM：让 Excel 正确识别中文编码（无 BOM 时 Excel 会乱码）
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(buf)
	now := time.Now()
	filename := fmt.Sprintf("sla-report-%s-last%dd.csv", now.Format("2006-01-02"), days)

	// ── 报表文件头 ──
	_ = w.Write([]string{"AIPivot SLA 报表"})
	_ = w.Write([]string{"生成时间", now.Format("2006-01-02 15:04:05")})
	_ = w.Write([]string{"统计范围", fmt.Sprintf("最近 %d 天", days)})
	_ = w.Write([]string{}) // 空行分隔

	// ── 概览 KPI ──
	_ = w.Write([]string{"## 概览汇总"})
	_ = w.Write([]string{"指标", "数值"})
	_ = w.Write([]string{"会话总数", strconv.FormatInt(ov.TotalConversations, 10)})
	_ = w.Write([]string{"活跃会话数", strconv.FormatInt(ov.ActiveConversations, 10)})
	_ = w.Write([]string{"AI 回复消息数", strconv.FormatInt(ov.TotalMessages, 10)})
	_ = w.Write([]string{"Token 消耗总量", strconv.FormatInt(ov.TotalTokens, 10)})
	_ = w.Write([]string{"估算费用（USD）", fmt.Sprintf("%.6f", ov.EstimatedCost)})
	_ = w.Write([]string{"AI 解决率 (%)", fmt.Sprintf("%.2f", ov.AIResolveRate)})
	_ = w.Write([]string{"转人工率 (%)", fmt.Sprintf("%.2f", ov.EscalationRate)})
	_ = w.Write([]string{})

	// ── 日粒度趋势 ──
	_ = w.Write([]string{"## 日粒度趋势"})
	_ = w.Write([]string{"日期", "新建会话数", "AI 消息数", "Token 消耗"})
	for _, pt := range dailyResp.Data {
		_ = w.Write([]string{
			pt.Date,
			strconv.FormatInt(pt.ConversationCount, 10),
			strconv.FormatInt(pt.MessageCount, 10),
			strconv.FormatInt(pt.TokenCount, 10),
		})
	}
	_ = w.Write([]string{})

	// ── 模型用量 ──
	_ = w.Write([]string{"## 模型用量明细"})
	_ = w.Write([]string{"模型", "AI 消息数", "Token 消耗", "估算费用（USD）"})
	for _, mu := range ov.ModelUsage {
		_ = w.Write([]string{
			mu.Model,
			strconv.FormatInt(mu.MessageCount, 10),
			strconv.FormatInt(mu.TokenCount, 10),
			fmt.Sprintf("%.6f", mu.EstimatedCost),
		})
	}
	_ = w.Write([]string{})

	// ── 状态分布 ──
	_ = w.Write([]string{"## 会话状态分布"})
	_ = w.Write([]string{"状态", "数量"})
	statusLabel := map[string]string{
		"active":        "进行中",
		"waiting_human": "转人工",
		"closed":        "已关闭",
	}
	for _, st := range ov.ByStatus {
		label := statusLabel[st.Status]
		if label == "" {
			label = st.Status
		}
		_ = w.Write([]string{label, strconv.FormatInt(st.Count, 10)})
	}
	_ = w.Write([]string{})

	// ── 渠道来源 ──
	_ = w.Write([]string{"## 渠道来源分布"})
	_ = w.Write([]string{"渠道", "数量"})
	for _, ch := range ov.ByChannel {
		_ = w.Write([]string{ch.Channel, strconv.FormatInt(ch.Count, 10)})
	}

	w.Flush()
	if err := w.Error(); err != nil {
		l.Logger.Errorf("ExportCSV csv.Writer flush err: %v", err)
		return nil, "", errorx.NewInternalError("报表生成失败")
	}

	return buf.Bytes(), filename, nil
}
