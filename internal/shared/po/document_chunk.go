package po

import "time"

// DocumentChunk 文档切块表持久化对象，对应 document_chunks 表。
// 文档被切块后每个片段对应一行，embedding 列存储向量供 pgvector 检索。
// 注意: embedding 字段使用 string 类型映射 pgvector，实际写入/读取通过 SQL 操作。
type DocumentChunk struct {
	ID              int64     `gorm:"primaryKey;column:id"`       // 自增主键
	UUID            string    `gorm:"column:uuid;type:uuid"`      // 对外暴露的唯一标识（UUID v4）
	DocumentID      int64     `gorm:"column:document_id"`         // 所属文档 ID（级联删除）
	KnowledgeBaseID int64     `gorm:"column:knowledge_base_id"`   // 所属知识库 ID（冗余字段，加速检索过滤）
	TenantID        int64     `gorm:"column:tenant_id"`           // 所属租户 ID（冗余字段，加速租户隔离查询）
	ChunkIndex      int       `gorm:"column:chunk_index"`         // 切块在文档中的顺序索引（从 0 开始）
	Content         string    `gorm:"column:content"`             // 切块文本内容
	TokenCount      int       `gorm:"column:token_count"`         // Token 数量（基于 Embedding 模型 tokenizer 计算）
	Metadata        string    `gorm:"column:metadata;type:jsonb"` // 切块元数据（页码、标题层级等）
	CreatedAt       time.Time `gorm:"column:created_at"`          // 创建时间
}

// TableName 指定 GORM 映射的数据库表名。
func (DocumentChunk) TableName() string {
	return "document_chunks"
}
