package po

import "time"

// Message 消息表持久化对象，对应 messages 表。
// 会话中的每条消息，包括用户消息、AI 回复、系统消息。
type Message struct {
	ID             int64     `gorm:"primaryKey;column:id"`
	UUID           string    `gorm:"column:uuid;type:uuid"`
	ConversationID int64     `gorm:"column:conversation_id"`
	TenantID       int64     `gorm:"column:tenant_id"`
	Role           string    `gorm:"column:role"`
	Content        string    `gorm:"column:content"`
	ContentType    string    `gorm:"column:content_type"`
	TokenCount     int       `gorm:"column:token_count"`
	Model          string    `gorm:"column:model"`
	LatencyMs      int       `gorm:"column:latency_ms"`
	Sources        string    `gorm:"column:sources;type:jsonb"`
	Metadata       string    `gorm:"column:metadata;type:jsonb"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (Message) TableName() string {
	return "messages"
}
