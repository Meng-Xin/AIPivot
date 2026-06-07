package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	Postgres     PostgresConf
	Redis        RedisConf
	Telemetry    TelemetryConf
	Metrics      MetricsConf
	Auth         AuthConf
	Migration    MigrationConf
	LLM          LLMConf
	Worker       WorkerConf
	Agent        AgentConf
	ModelPricing []ModelPricingConf
	RateLimit    RateLimitConf
}

type AuthConf struct {
	AccessSecret string
	AccessExpire int64  `json:",default=86400"`
	Issuer       string `json:",default=aipivot"`
}

type MigrationConf struct {
	Enabled bool   `json:",default=true"`
	Path    string `json:",default=migrations"`
}

type PostgresConf struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	SSLMode      string
	TimeZone     string
	MaxOpenConns int
	MaxIdleConns int
}

type RedisConf struct {
	Addr     string
	Password string `json:",optional"`
	DB       int
}

type TelemetryConf struct {
	ServiceName    string
	Environment    string
	JaegerEndpoint string
	SampleRatio    float64 `json:",default=1"`
}

type MetricsConf struct {
	Enabled bool   `json:",default=true"`
	Path    string `json:",default=/metrics"`
}

// LLMConf LLM API 配置（兼容 OpenAI Chat Completions，也支持 Ark Responses）。
type LLMConf struct {
	BaseURL             string  `json:",default=http://127.0.0.1:3000/v1"` // One API 或 OpenAI endpoint
	Provider            string  `json:",default=openai"`                   // openai / ark
	APIKey              string  `json:",optional"`
	APIKeyEnv           string  `json:",optional"` // 从环境变量读取 API Key，优先级高于 APIKey
	ChatModel           string  `json:",default=gpt-3.5-turbo"`
	EmbeddingModel      string  `json:",default=text-embedding-3-small"`
	EmbeddingDim        int     `json:",default=1536"`
	MaxTokens           int     `json:",default=2048"`
	Temperature         float64 `json:",default=0.7"`
	TimeoutSeconds      int     `json:",default=60"`
	UseResponsesAPI     bool    `json:",default=false"` // 使用 /responses 代替 /chat/completions
	EnableWebSearch     bool    `json:",default=false"` // Ark Responses 内置 web_search
	WebSearchMaxKeyword int     `json:",default=3"`

	// 多模型路由：声明可用模型列表，供前端选择和后端路由
	ChatModels      []ModelOption `json:",optional"`
	EmbeddingModels []ModelOption `json:",optional"`
}

// ModelOption 可用模型选项（配置驱动，由 One API 等网关提供实际路由）
type ModelOption struct {
	ID        string `json:"ID"`                  // 模型标识，传给 LLM 网关（如 gpt-4o / deepseek-chat）
	Name      string `json:"Name"`                // 前端展示名
	Provider  string `json:"Provider,optional"`   // 供应商（openai / deepseek / qwen 等）
	MaxTokens int    `json:"MaxTokens,default=0"` // 模型最大 context 窗口（0 表示未配置）
}

type WorkerConf struct {
	Enabled     bool `json:",default=true"`
	Concurrency int  `json:",default=5"`
}

// ModelPricingConf 模型计费配置，用于分析报表中的费用估算
type ModelPricingConf struct {
	Model string  // 模型标识，需与 LLM.ChatModels[].ID 保持一致
	PerK  float64 // 每千 tokens 估算费用（美元），混合口径（输入+输出均摊）
}

// RateLimitConf 每租户 Token 日配额限流配置
type RateLimitConf struct {
	DailyTokenLimit int64 `json:",default=0"` // 每日 Token 上限（0 = 不限制）
}

// AgentConf Function Calling Agent 配置
type AgentConf struct {
	Enabled           bool `json:",default=true"`  // 是否启用 Agent（false 时退化为纯 LLM）
	MaxRounds         int  `json:",default=5"`     // ReAct 循环最大轮数，防止无限工具调用
	MultiAgentEnabled bool `json:",default=false"` // 是否启用 Orchestrator-Worker 多 Agent 编排
	MaxWorkers        int  `json:",default=3"`     // 单次编排最多拆分的 worker 数
}
