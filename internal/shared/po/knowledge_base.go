package po

import "time"

// KnowledgeBase 知识库表持久化对象，对应 knowledge_bases 表。
// 每个租户可创建多个知识库，每个知识库使用独立的 Embedding 模型配置。
type KnowledgeBase struct {
	ID          int64     `gorm:"primaryKey;column:id"`
	UUID        string    `gorm:"column:uuid;type:uuid"`
	TenantID    int64     `gorm:"column:tenant_id"`
	Name        string    `gorm:"column:name"`
	Description string    `gorm:"column:description"`
	Model       string    `gorm:"column:model"`
	Dimension   int       `gorm:"column:dimension"`
	Status      string    `gorm:"column:status"`
	DocCount    int       `gorm:"column:doc_count"`
	ChunkCount  int       `gorm:"column:chunk_count"`
	Settings    string    `gorm:"column:settings;type:jsonb"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (KnowledgeBase) TableName() string {
	return "knowledge_bases"
}
