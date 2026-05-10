package po

import "time"

// Document 文档表持久化对象，对应 documents 表。
// 上传到知识库的原始文档，经过解析、切块、Embedding 后生成 DocumentChunk。
type Document struct {
	ID              int64     `gorm:"primaryKey;column:id"`
	UUID            string    `gorm:"column:uuid;type:uuid"`
	KnowledgeBaseID int64     `gorm:"column:knowledge_base_id"`
	TenantID        int64     `gorm:"column:tenant_id"`
	Name            string    `gorm:"column:name"`
	ContentType     string    `gorm:"column:content_type"`
	FileSize        int64     `gorm:"column:file_size"`
	FilePath        string    `gorm:"column:file_path"`
	ChunkCount      int       `gorm:"column:chunk_count"`
	Status          string    `gorm:"column:status"`
	ErrorMsg        string    `gorm:"column:error_msg"`
	Metadata        string    `gorm:"column:metadata;type:jsonb"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (Document) TableName() string {
	return "documents"
}
