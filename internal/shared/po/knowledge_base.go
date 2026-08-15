package po

import "time"

// KnowledgeBase 知识库表持久化对象，对应 knowledge_bases 表。
// 每个租户可创建多个知识库，每个知识库使用独立的 Embedding 模型配置。
type KnowledgeBase struct {
	ID          int64     `gorm:"primaryKey;column:id"`       // 自增主键
	UUID        string    `gorm:"column:uuid;type:uuid"`      // 对外暴露的唯一标识（UUID v4）
	TenantID    int64     `gorm:"column:tenant_id"`           // 所属租户 ID（级联删除）
	Name        string    `gorm:"column:name"`                // 知识库名称
	Description string    `gorm:"column:description"`         // 知识库描述
	Model       string    `gorm:"column:model"`               // Embedding 模型名称
	Dimension   int       `gorm:"column:dimension"`           // 向量维度（与 Embedding 模型对应）
	Status      string    `gorm:"column:status"`              // 状态: active / archived
	DocCount    int       `gorm:"column:doc_count"`           // 文档数量（冗余计数，异步更新）
	ChunkCount  int       `gorm:"column:chunk_count"`         // 切块数量（冗余计数，异步更新）
	Settings          string    `gorm:"column:settings;type:jsonb"`          // 知识库配置（chunk_size / overlap / 检索策略等）
	SuggestedQuestions string  `gorm:"column:suggested_questions;type:jsonb"` // 引导问答 / 快捷回复列表（JSON 数组）
	CreatedAt         time.Time `gorm:"column:created_at"`                    // 创建时间
	UpdatedAt         time.Time `gorm:"column:updated_at"`                    // 更新时间
}

// TableName 指定 GORM 映射的数据库表名。
func (KnowledgeBase) TableName() string {
	return "knowledge_bases"
}
