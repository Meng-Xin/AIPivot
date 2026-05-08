package po

import "time"

type ApiKey struct {
	ID        int64      `gorm:"primaryKey;column:id"`
	TenantID  int64      `gorm:"column:tenant_id"`
	Name      string     `gorm:"column:name"`
	KeyHash   string     `gorm:"column:key_hash"`
	KeyPrefix string     `gorm:"column:key_prefix"`
	Scopes    string     `gorm:"column:scopes;type:jsonb"`
	Status    string     `gorm:"column:status"`
	LastUsed  *time.Time `gorm:"column:last_used"`
	ExpiresAt *time.Time `gorm:"column:expires_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (ApiKey) TableName() string {
	return "api_keys"
}
