package po

import "time"

// Message 消息表持久化对象，对应 messages 表。
// 会话中的每条消息，包括用户消息、AI 回复、系统消息。
type Message struct {
	ID             int64     `gorm:"primaryKey;column:id"`       // 自增主键
	UUID           string    `gorm:"column:uuid;type:uuid"`      // 对外暴露的唯一标识（UUID v4）
	ConversationID int64     `gorm:"column:conversation_id"`     // 所属会话 ID（级联删除）
	TenantID       int64     `gorm:"column:tenant_id"`           // 所属租户 ID（冗余字段，加速查询）
	Role           string    `gorm:"column:role"`                // 角色: user / assistant / system
	Content        string    `gorm:"column:content"`             // 消息内容
	ContentType    string    `gorm:"column:content_type"`        // 内容类型: text / image / file
	TokenCount     int       `gorm:"column:token_count"`         // Token 消耗数量
	Model          string    `gorm:"column:model"`               // 生成此消息的 LLM 模型名称（仅 assistant 消息）
	LatencyMs      int       `gorm:"column:latency_ms"`          // LLM 响应延迟（毫秒，仅 assistant 消息）
	Sources        string    `gorm:"column:sources;type:jsonb"`  // 知识来源引用（RAG 检索命中的 chunk UUID 列表）
	Metadata       string    `gorm:"column:metadata;type:jsonb"` // 消息元数据（评价、标注等）
	Rating         string    `gorm:"column:rating"`              // 访客评分：up / down / 空=未评分
	RatingFeedback string    `gorm:"column:rating_feedback"`     // 负评文字反馈（仅 rating=down 时可能有值）
	CreatedAt      time.Time `gorm:"column:created_at"`          // 创建时间
}

// TableName 指定 GORM 映射的数据库表名。
func (Message) TableName() string {
	return "messages"
}
