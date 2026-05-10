package po

import "time"

// Conversation 会话表持久化对象，对应 conversations 表。
// 一次用户与 AI 的完整对话，可关联知识库、可转人工。
type Conversation struct {
	ID              int64      `gorm:"primaryKey;column:id"`
	UUID            string     `gorm:"column:uuid;type:uuid"`
	TenantID        int64      `gorm:"column:tenant_id"`
	UserID          *int64     `gorm:"column:user_id"`
	KnowledgeBaseID *int64     `gorm:"column:knowledge_base_id"`
	Title           string     `gorm:"column:title"`
	Status          string     `gorm:"column:status"`
	Channel         string     `gorm:"column:channel"`
	MessageCount    int        `gorm:"column:message_count"`
	Summary         string     `gorm:"column:summary"`
	Metadata        string     `gorm:"column:metadata;type:jsonb"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	ClosedAt        *time.Time `gorm:"column:closed_at"`
}

func (Conversation) TableName() string {
	return "conversations"
}
