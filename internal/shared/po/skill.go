package po

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSONMap 是 JSONB 字段的通用 Go 类型，支持 GORM 读写。
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	b, err := json.Marshal(j)
	return string(b), err
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = JSONMap{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return fmt.Errorf("JSONMap.Scan: unsupported type %T", value)
	}
	return json.Unmarshal(bytes, j)
}

// Skill 租户自定义工具持久化对象，对应 skills 表。
// 每条记录定义一个 Agent 可调用的 HTTP 回调工具（Function Calling Skill）。
type Skill struct {
	ID          int64     `gorm:"primaryKey;column:id"`              // 自增主键
	TenantID    int64     `gorm:"column:tenant_id"`                  // 所属租户 ID（级联删除）
	Name        string    `gorm:"column:name"`                       // 工具唯一名称（租户内唯一，对应 function.name）
	Description string    `gorm:"column:description"`                // 工具描述，帮助 LLM 判断何时调用
	Parameters  JSONMap   `gorm:"column:parameters;type:jsonb"`      // JSON Schema 参数定义（type/properties/required）
	Endpoint    string    `gorm:"column:endpoint"`                   // HTTP 回调端点 URL
	Method      string    `gorm:"column:method"`                     // HTTP 方法: GET / POST
	Headers     JSONMap   `gorm:"column:headers;type:jsonb"`         // 附加请求头 JSON 对象（如认证 Token）
	TimeoutMs   int       `gorm:"column:timeout_ms"`                 // 请求超时毫秒数（默认 5000）
	Enabled     bool      `gorm:"column:enabled"`                    // 是否启用（false 时不注册到 Agent）
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"` // 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"` // 更新时间
}

// TableName 指定 GORM 映射的数据库表名。
func (Skill) TableName() string {
	return "skills"
}
