package svc

import (
	"context"
	"fmt"
	"net/http"

	"aipivot/internal/config"
	"aipivot/internal/infra"
	"aipivot/internal/middleware"
	"aipivot/internal/modules/agent"
	"aipivot/internal/modules/agent/tools"
	"aipivot/internal/modules/auth"
	"aipivot/internal/modules/channel/webhook"
	"aipivot/internal/modules/chat"
	"aipivot/internal/modules/knowledge"
	"aipivot/internal/modules/rag"
	"aipivot/internal/observability"
	authrepo "aipivot/internal/repository/auth"
	chatrepo "aipivot/internal/repository/chat"
	flowrepo "aipivot/internal/repository/flow"
	knowledgerepo "aipivot/internal/repository/knowledge"
	skillrepo "aipivot/internal/repository/skill"
	webhookrepo "aipivot/internal/repository/webhook"
	"aipivot/internal/shared/query"
	"aipivot/internal/shared/ratelimit"
	"aipivot/pkg/llm"

	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config       config.Config
	DB           *gorm.DB
	Redis        *redis.Client
	Metrics      *observability.Metrics
	HealthChecks []infra.DependencyCheck
	Shutdown     func(context.Context) error

	// 路由级中间件
	AuthMiddleware   func(http.HandlerFunc) http.HandlerFunc
	AdminMiddleware  func(http.HandlerFunc) http.HandlerFunc // Admin 端点，要求 role=admin
	ApiKeyMiddleware func(http.HandlerFunc) http.HandlerFunc // Open API 使用 API Key 认证

	// Auth Repo
	UserRepo   auth.UserRepository
	TenantRepo auth.TenantRepository
	ApiKeyRepo auth.ApiKeyRepository

	// Knowledge Repo
	KnowledgeBaseRepo knowledge.KBRepository
	DocumentRepo      knowledge.DocumentRepository

	// Knowledge Chunk Repo
	DocumentChunkRepo knowledge.DocChunkRepository

	// Chat Repo
	ConversationRepo chat.ConversationRepository
	MessageRepo      chat.MessageRepository

	// Skill（租户自定义工具）
	SkillRepo skillrepo.Repository

	// Flow（可视化流程定义）
	FlowRepo flowrepo.Repository

	// Webhook
	WebhookRepo     webhook.Repository
	WebhookDelivery *webhook.DeliveryService

	// LLM & RAG
	LLMClient    *llm.Client
	RAGService   *rag.Service
	AsynqClient  *asynq.Client
	TokenLimiter *ratelimit.TokenLimiter // 每租户日 Token 配额限流
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	shutdown, err := observability.InitTracing(context.Background(), c.Telemetry)
	if err != nil {
		return nil, fmt.Errorf("init tracing: %w", err)
	}

	if c.Migration.Enabled {
		if err := infra.RunMigrations(c.Postgres, c.Migration.Path); err != nil {
			return nil, fmt.Errorf("run migrations: %w", err)
		}
	}

	db, err := infra.NewPostgres(c.Postgres)
	if err != nil {
		return nil, fmt.Errorf("init postgres: %w", err)
	}

	redisClient := infra.NewRedis(c.Redis)
	metrics := observability.NewMetrics(prometheus.NewRegistry())

	// DB → Query → Repo（1 步构造链路）
	q := query.Use(db)

	// Auth Repos
	apiKeyRepo := authrepo.NewApiKeyRepo(q)

	// Webhook Repo
	wbRepo := webhookrepo.NewWebhookRepo(q, db)

	// LLM Client (OpenAI-compatible, 支持 One API)
	llmClient := llm.NewClient(c.LLM.BaseURL, c.LLM.APIKey, c.LLM.TimeoutSeconds)

	// Knowledge Repos
	chunkRepo := knowledgerepo.NewDocChunkRepo(q, db)

	// Agent (Function Calling)
	var ag *agent.Agent
	var orchestrator *agent.Orchestrator
	if c.Agent.Enabled {
		registry := agent.NewRegistry()
		registry.Register(tools.NewWeatherTool())
		registry.Register(tools.NewTimeTool())
		registry.Register(tools.NewCalculatorTool())
		registry.Register(tools.NewEscalationTool())
		ag = agent.NewAgent(llmClient, registry, c.Agent.MaxRounds)
		if c.Agent.MultiAgentEnabled {
			orchestrator = agent.NewOrchestrator(llmClient, ag, c.Agent.MaxWorkers)
		}
	}

	// RAG Service（注入 Agent，nil 时退化为纯 LLM）
	ragService := rag.NewService(llmClient, chunkRepo, ag, orchestrator, rag.Config{
		ChatModel:      c.LLM.ChatModel,
		EmbeddingModel: c.LLM.EmbeddingModel,
		MaxTokens:      c.LLM.MaxTokens,
		Temperature:    c.LLM.Temperature,
	})

	// TokenLimiter（Redis 令牌桶，限制每租户日 token 用量）
	tokenLimiter := ratelimit.NewTokenLimiter(redisClient, c.RateLimit.DailyTokenLimit)

	// Asynq Client（用于提交异步任务）
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     c.Redis.Addr,
		Password: c.Redis.Password,
		DB:       c.Redis.DB,
	})

	return &ServiceContext{
		Config:           c,
		DB:               db,
		Redis:            redisClient,
		Metrics:          metrics,
		AuthMiddleware:   middleware.NewAuthMiddleware(c.Auth).Handle,
		AdminMiddleware:  middleware.NewAdminMiddleware(c.Auth).Handle,
		ApiKeyMiddleware: middleware.NewApiKeyMiddleware(apiKeyRepo).Handle,

		// Auth
		UserRepo:   authrepo.NewUserRepo(q),
		TenantRepo: authrepo.NewTenantRepo(q),
		ApiKeyRepo: apiKeyRepo,

		// Knowledge
		KnowledgeBaseRepo: knowledgerepo.NewKBRepo(q),
		DocumentRepo:      knowledgerepo.NewDocumentRepo(q),
		DocumentChunkRepo: chunkRepo,

		// Chat
		ConversationRepo: chatrepo.NewConversationRepo(q),
		MessageRepo:      chatrepo.NewMessageRepo(q),

		// Skill
		SkillRepo: skillrepo.NewSkillRepo(q, db),

		// Flow
		FlowRepo: flowrepo.NewFlowRepo(q),

		// Webhook
		WebhookRepo:     wbRepo,
		WebhookDelivery: webhook.NewDeliveryService(wbRepo),

		// LLM & RAG
		LLMClient:    llmClient,
		RAGService:   ragService,
		AsynqClient:  asynqClient,
		TokenLimiter: tokenLimiter,

		HealthChecks: []infra.DependencyCheck{
			infra.CheckPostgres(db),
			infra.CheckRedis(redisClient),
		},
		Shutdown: func(ctx context.Context) error {
			var shutdownErr error
			if err := asynqClient.Close(); err != nil && shutdownErr == nil {
				shutdownErr = fmt.Errorf("close asynq client: %w", err)
			}
			if err := shutdown(ctx); err != nil && shutdownErr == nil {
				shutdownErr = fmt.Errorf("shutdown tracing: %w", err)
			}
			if err := redisClient.Close(); err != nil && shutdownErr == nil {
				shutdownErr = fmt.Errorf("close redis: %w", err)
			}
			if err := infra.ClosePostgres(db); err != nil && shutdownErr == nil {
				shutdownErr = fmt.Errorf("close postgres: %w", err)
			}
			return shutdownErr
		},
	}, nil
}
