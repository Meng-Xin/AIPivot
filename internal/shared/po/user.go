// Package po 定义持久化对象（Persistent Object），与数据库表结构一一映射。
package po

import "time"

// User 用户表持久化对象，对应 users 表。
// 每个用户归属一个租户（TenantID），同一租户内 Email 唯一。
type User struct {
	ID        int64      `gorm:"primaryKey;column:id"`  // 自增主键
	UUID      string     `gorm:"column:uuid;type:uuid"` // 业务唯一标识（对外暴露）
	TenantID  int64      `gorm:"column:tenant_id"`      // 所属租户 ID
	Email     string     `gorm:"column:email"`          // 登录邮箱（租户内唯一）
	NickName  string     `gorm:"column:nick_name"`      // 用户昵称
	Password  string     `gorm:"column:password"`       // bcrypt 哈希密码
	Role      string     `gorm:"column:role"`           // 角色：admin / member
	Status    string     `gorm:"column:status"`         // 状态：active / disabled
	LastLogin *time.Time `gorm:"column:last_login"`     // 最近登录时间（可为空）
	CreatedAt time.Time  `gorm:"column:created_at"`     // 创建时间
	UpdatedAt time.Time  `gorm:"column:updated_at"`     // 更新时间
}

// TableName 指定 GORM 映射的数据库表名。
func (User) TableName() string {
	return "users"
}
