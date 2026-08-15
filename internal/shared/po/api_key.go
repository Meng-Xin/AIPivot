package po

import (
	"time"

	"github.com/lib/pq"
)

// ApiKey API 密钥表持久化对象，对应 api_keys 表。
// 每个密钥归属一个租户，仅存储哈希值，原始密钥只在创建时返回一次。
// master 密钥用于服务端调用（无域名限制）；public 密钥（pk_ 前缀）可嵌入
// 前端 Chat Widget，受 allowed_origins 域名白名单与 knowledge_base_id 绑定约束。
type ApiKey struct {
	ID              int64      `gorm:"primaryKey;column:id"`                  // 自增主键
	TenantID        int64      `gorm:"column:tenant_id"`                      // 所属租户 ID
	Name            string     `gorm:"column:name"`                           // 密钥名称（便于管理识别）
	KeyHash         string     `gorm:"column:key_hash"`                       // 密钥哈希（SHA-256）
	KeyPrefix       string     `gorm:"column:key_prefix"`                     // 密钥前缀（用于列表展示，如 sk-abc...）
	Scopes          string     `gorm:"column:scopes;type:jsonb"`              // 权限范围（JSON 数组）
	Status          string     `gorm:"column:status"`                         // 状态：active / revoked
	KeyType         string     `gorm:"column:key_type"`                       // 密钥类型: master / public
	AllowedOrigins  pq.StringArray `gorm:"column:allowed_origins;type:text[]"` // 域名白名单（仅 public key 生效）
	KnowledgeBaseID *int64     `gorm:"column:knowledge_base_id"`              // 绑定的知识库 ID（仅 public key）
	LastUsed        *time.Time `gorm:"column:last_used"`                      // 最近使用时间（可为空）
	ExpiresAt       *time.Time `gorm:"column:expires_at"`                     // 过期时间（可为空，空表示永不过期）
	CreatedAt       time.Time  `gorm:"column:created_at"`                     // 创建时间
	UpdatedAt       time.Time  `gorm:"column:updated_at"`                     // 更新时间
}

// TableName 指定 GORM 映射的数据库表名。
func (ApiKey) TableName() string {
	return "api_keys"
}
