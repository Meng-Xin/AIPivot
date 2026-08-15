package po

import "time"

// FlowRun Flow 执行历史持久化对象，对应 flow_runs 表。
type FlowRun struct {
	ID          int64     `gorm:"primaryKey;column:id"`              // 自增主键
	UUID        string    `gorm:"column:uuid"`                       // 对外展示 UUID
	TenantID    int64     `gorm:"column:tenant_id"`                  // 所属租户 ID
	FlowID      int64     `gorm:"column:flow_id"`                    // 关联 Flow ID
	FlowVersion int       `gorm:"column:flow_version"`               // 执行时 Flow 版本快照
	Status      string    `gorm:"column:status"`                     // running / success / failed / timeout
	TriggerType string    `gorm:"column:trigger_type"`               // manual
	Input       JSONMap   `gorm:"column:input;type:jsonb"`          // 执行输入
	Output      string    `gorm:"column:output"`                    // 最终输出
	NodeResults string    `gorm:"column:node_results;type:jsonb"`   // 节点结果快照数组（JSON 数组字符串）
	Error       string    `gorm:"column:error"`                     // 失败原因
	TotalMs     int       `gorm:"column:total_ms"`                  // 总耗时毫秒
	TokenCount  int       `gorm:"column:token_count"`               // Token 消耗
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"` // 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"` // 更新时间
}

func (FlowRun) TableName() string {
	return "flow_runs"
}
