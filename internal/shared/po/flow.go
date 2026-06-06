package po

import "time"

// Flow 租户可视化流程定义持久化对象，对应 flows 表。
type Flow struct {
	ID          int64     `gorm:"primaryKey;column:id"`             // 自增主键
	UUID        string    `gorm:"column:uuid"`                      // 对外展示 UUID
	TenantID    int64     `gorm:"column:tenant_id"`                 // 所属租户 ID
	Name        string    `gorm:"column:name"`                      // 流程名称（租户内唯一）
	Description string    `gorm:"column:description"`               // 流程描述
	Definition  JSONMap   `gorm:"column:definition;type:jsonb"`     // 画布定义 JSON
	Status      string    `gorm:"column:status"`                    // draft / published / archived
	Version     int       `gorm:"column:version"`                   // 流程版本
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"` // 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"` // 更新时间
}

func (Flow) TableName() string {
	return "flows"
}
