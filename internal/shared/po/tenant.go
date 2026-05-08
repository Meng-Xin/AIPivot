package po

import "time"

type Tenant struct {
	ID        int64     `gorm:"primaryKey;column:id"`
	UUID      string    `gorm:"column:uuid;type:uuid"`
	Name      string    `gorm:"column:name"`
	Slug      string    `gorm:"column:slug"`
	Plan      string    `gorm:"column:plan"`
	Status    string    `gorm:"column:status"`
	Settings  string    `gorm:"column:settings;type:jsonb"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Tenant) TableName() string {
	return "tenants"
}
