package po

import "time"

type User struct {
	ID        int64      `gorm:"primaryKey;column:id"`
	UUID      string     `gorm:"column:uuid;type:uuid"`
	TenantID  int64      `gorm:"column:tenant_id"`
	Email     string     `gorm:"column:email"`
	NickName  string     `gorm:"column:nick_name"`
	Password  string     `gorm:"column:password"`
	Role      string     `gorm:"column:role"`
	Status    string     `gorm:"column:status"`
	LastLogin *time.Time `gorm:"column:last_login"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "users"
}
