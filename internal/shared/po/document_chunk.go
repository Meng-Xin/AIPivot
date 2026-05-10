package po

import "time"

// DocumentChunk 文档切块表持久化对象，对应 document_chunks 表。
// 文档被切块后每个片段对应一行，embedding 列存储向量供 pgvector 检索。
// 注意: embedding 字段使用 string 类型映射 pgvector，实际写入/读取通过 SQL 操作。
type DocumentChunk struct {
	ID              int64     `gorm:"primaryKey;column:id"`
	UUID            string    `gorm:"column:uuid;type:uuid"`
	DocumentID      int64     `gorm:"column:document_id"`
	KnowledgeBaseID int64     `gorm:"column:knowledge_base_id"`
	TenantID        int64     `gorm:"column:tenant_id"`
	ChunkIndex      int       `gorm:"column:chunk_index"`
	Content         string    `gorm:"column:content"`
	TokenCount      int       `gorm:"column:token_count"`
	Metadata        string    `gorm:"column:metadata;type:jsonb"`
	CreatedAt       time.Time `gorm:"column:created_at"`
}

func (DocumentChunk) TableName() string {
	return "document_chunks"
}
