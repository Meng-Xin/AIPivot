package po

import "time"

// Webhook Webhook 配置表持久化对象，对应 webhooks 表。
// 租户注册的回调端点，用于出站事件推送和第三方平台集成。
type Webhook struct {
	ID          int64      `gorm:"primaryKey;column:id"`         // 自增主键
	UUID        string     `gorm:"column:uuid;type:uuid"`        // 对外暴露的唯一标识（UUID v4）
	TenantID    int64      `gorm:"column:tenant_id"`             // 所属租户 ID（级联删除）
	Name        string     `gorm:"column:name"`                  // Webhook 名称（便于管理识别）
	URL         string     `gorm:"column:url"`                   // 回调 URL（HTTPS 推荐）
	Secret      string     `gorm:"column:secret"`                // 签名密钥（HMAC-SHA256，空表示不签名）
	Events      string     `gorm:"column:events;type:jsonb"`     // 订阅事件类型列表（JSON 数组）
	ChannelType string     `gorm:"column:channel_type"`          // 关联渠道类型: webhook / wechat / feishu
	Status      string     `gorm:"column:status"`                // 状态: active / disabled
	RetryCount  int        `gorm:"column:retry_count"`           // 失败重试次数上限
	TimeoutMs   int        `gorm:"column:timeout_ms"`            // 请求超时毫秒数
	LastError   string     `gorm:"column:last_error"`            // 最近推送错误信息（成功时清空）
	LastTrigger *time.Time `gorm:"column:last_trigger"`          // 最近一次触发时间
	CreatedAt   time.Time  `gorm:"column:created_at"`            // 创建时间
	UpdatedAt   time.Time  `gorm:"column:updated_at"`            // 更新时间
}

// TableName 指定 GORM 映射的数据库表名。
func (Webhook) TableName() string {
	return "webhooks"
}
