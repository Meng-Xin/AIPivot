package po

import "time"

// Document 文档表持久化对象，对应 documents 表。
// 上传到知识库的原始文档，经过解析、切块、Embedding 后生成 DocumentChunk。
type Document struct {
	ID              int64     `gorm:"primaryKey;column:id"`       // 自增主键
	UUID            string    `gorm:"column:uuid;type:uuid"`      // 对外暴露的唯一标识（UUID v4）
	KnowledgeBaseID int64     `gorm:"column:knowledge_base_id"`   // 所属知识库 ID（级联删除）
	TenantID        int64     `gorm:"column:tenant_id"`           // 所属租户 ID（冗余字段加速查询）
	Name            string    `gorm:"column:name"`                // 文档名称（原始文件名）
	ContentType     string    `gorm:"column:content_type"`        // 文档 MIME 类型
	FileSize        int64     `gorm:"column:file_size"`           // 文件大小（字节）
	FilePath        string    `gorm:"column:file_path"`           // 文件存储路径（MinIO/OSS）
	ChunkCount      int       `gorm:"column:chunk_count"`         // 切块数量
	Status          string    `gorm:"column:status"`              // 处理状态: pending / processing / completed / failed
	ErrorMsg        string    `gorm:"column:error_msg"`           // 处理失败时的错误信息
	Metadata        string    `gorm:"column:metadata;type:jsonb"` // 文档元数据（标签、来源等）
	CreatedAt       time.Time `gorm:"column:created_at"`          // 创建时间
	UpdatedAt       time.Time `gorm:"column:updated_at"`          // 更新时间
}

// TableName 指定 GORM 映射的数据库表名。
func (Document) TableName() string {
	return "documents"
}
