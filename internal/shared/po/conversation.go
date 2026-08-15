package po

import "time"

// Conversation 会话表持久化对象，对应 conversations 表。
// 一次用户与 AI 的完整对话，可关联知识库、可转人工。
type Conversation struct {
	ID              int64      `gorm:"primaryKey;column:id"`       // 自增主键
	UUID            string     `gorm:"column:uuid;type:uuid"`      // 对外暴露的唯一标识（UUID v4）
	TenantID        int64      `gorm:"column:tenant_id"`           // 所属租户 ID（级联删除）
	UserID          *int64     `gorm:"column:user_id"`             // 发起用户 ID（匿名对话可为空）
	KnowledgeBaseID *int64     `gorm:"column:knowledge_base_id"`   // 关联知识库 ID（决定 RAG 检索范围）
	Title           string     `gorm:"column:title"`               // 会话标题（由 LLM 自动生成或用户指定）
	Model           string     `gorm:"column:model"`               // 聊天模型标识（如 gpt-4o），空字符串表示使用系统默认模型
	Status          string     `gorm:"column:status"`              // 状态: active / waiting_human / resolved / closed
	Channel         string     `gorm:"column:channel"`             // 接入渠道: web / api / webhook / widget / wechat / feishu
	ExternalUserID  string     `gorm:"column:external_user_id"`    // 外部渠道用户标识（Widget 访客 / Webhook 用户），空表示已登录用户
	MessageCount    int        `gorm:"column:message_count"`       // 消息数量（冗余计数）
	Summary         string     `gorm:"column:summary"`             // 会话摘要（上下文压缩后的摘要文本）
	Metadata        string     `gorm:"column:metadata;type:jsonb"` // 会话元数据（客户端信息、标签等）
	CreatedAt       time.Time  `gorm:"column:created_at"`          // 创建时间
	UpdatedAt       time.Time  `gorm:"column:updated_at"`          // 更新时间
	ClosedAt        *time.Time `gorm:"column:closed_at"`           // 关闭时间
}

// TableName 指定 GORM 映射的数据库表名。
func (Conversation) TableName() string {
	return "conversations"
}
