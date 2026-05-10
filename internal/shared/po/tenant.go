package po

import "time"

// Tenant 租户表持久化对象，对应 tenants 表。
// 系统采用多租户隔离，所有业务数据通过 TenantID 关联。
type Tenant struct {
	ID        int64     `gorm:"primaryKey;column:id"`       // 自增主键
	UUID      string    `gorm:"column:uuid;type:uuid"`      // 业务唯一标识（对外暴露）
	Name      string    `gorm:"column:name"`                // 租户名称
	Slug      string    `gorm:"column:slug"`                // 租户标识符（URL 友好，全局唯一）
	Plan      string    `gorm:"column:plan"`                // 订阅计划：free / pro / enterprise
	Status    string    `gorm:"column:status"`              // 状态：active / suspended
	Settings  string    `gorm:"column:settings;type:jsonb"` // 租户配置（JSON 格式）
	CreatedAt time.Time `gorm:"column:created_at"`          // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at"`          // 更新时间
}

// TableName 指定 GORM 映射的数据库表名。
func (Tenant) TableName() string {
	return "tenants"
}
