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
	authRepo "aipivot/internal/modules/auth/repo"
	authDao "aipivot/internal/modules/auth/repo/dao"
	chatRepo "aipivot/internal/modules/chat/repo"
	chatDao "aipivot/internal/modules/chat/repo/dao"
	kbRepo "aipivot/internal/modules/knowledge/repo"
	kbDao "aipivot/internal/modules/knowledge/repo/dao"
	"aipivot/internal/modules/rag"
	"aipivot/internal/observability"
	"aipivot/internal/shared/query"
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

	// 路由级中间件（goctl 生成的 routes.go 通过 serverCtx.AuthMiddleware 引用）
	AuthMiddleware func(http.HandlerFunc) http.HandlerFunc

	// Auth Repo
	UserRepo   authRepo.UserRepository
	TenantRepo authRepo.TenantRepository

	// Knowledge Repo
	KnowledgeBaseRepo kbRepo.KnowledgeBaseRepository
	DocumentRepo      kbRepo.DocumentRepository

	// Knowledge Chunk Repo
	DocumentChunkRepo kbRepo.DocumentChunkRepository

	// Chat Repo
	ConversationRepo chatRepo.ConversationRepository
	MessageRepo      chatRepo.MessageRepository

	// LLM & RAG
	LLMClient   *llm.Client
	RAGService  *rag.Service
	AsynqClient *asynq.Client
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

	// DB → Query → DAO → Repo(接口) 组装链路
	q := query.Use(db)

	// Auth DAOs
	userDao := authDao.NewUserDao(q)
	tenantDao := authDao.NewTenantDao(q)

	// Knowledge DAOs
	knowledgeBaseDao := kbDao.NewKnowledgeBaseDao(q)
	documentDao := kbDao.NewDocumentDao(q)
	documentChunkDao := kbDao.NewDocumentChunkDao(q, db)

	// Chat DAOs
	conversationDao := chatDao.NewConversationDao(q)
	messageDao := chatDao.NewMessageDao(q)

	// LLM Client (OpenAI-compatible, 支持 One API)
	llmClient := llm.NewClient(c.LLM.BaseURL, c.LLM.APIKey, c.LLM.TimeoutSeconds)

	// Document Chunk Repo
	chunkRepo := kbRepo.NewDocumentChunkRepo(documentChunkDao)

	// Agent (Function Calling)
	var ag *agent.Agent
	if c.Agent.Enabled {
		registry := agent.NewRegistry()
		registry.Register(tools.NewWeatherTool())
		registry.Register(tools.NewTimeTool())
		registry.Register(tools.NewCalculatorTool())
		registry.Register(tools.NewEscalationTool())
		ag = agent.NewAgent(llmClient, registry, c.Agent.MaxRounds)
	}

	// RAG Service（注入 Agent，nil 时退化为纯 LLM）
	ragService := rag.NewService(llmClient, chunkRepo, ag, rag.Config{
		ChatModel:      c.LLM.ChatModel,
		EmbeddingModel: c.LLM.EmbeddingModel,
		MaxTokens:      c.LLM.MaxTokens,
		Temperature:    c.LLM.Temperature,
	})

	// Asynq Client（用于提交异步任务）
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     c.Redis.Addr,
		Password: c.Redis.Password,
		DB:       c.Redis.DB,
	})

	return &ServiceContext{
		Config:         c,
		DB:             db,
		Redis:          redisClient,
		Metrics:        metrics,
		AuthMiddleware: middleware.NewAuthMiddleware(c.Auth).Handle,

		// Auth
		UserRepo:   authRepo.NewUserRepo(userDao),
		TenantRepo: authRepo.NewTenantRepo(tenantDao),

		// Knowledge
		KnowledgeBaseRepo: kbRepo.NewKnowledgeBaseRepo(knowledgeBaseDao),
		DocumentRepo:      kbRepo.NewDocumentRepo(documentDao),
		DocumentChunkRepo: chunkRepo,

		// Chat
		ConversationRepo: chatRepo.NewConversationRepo(conversationDao),
		MessageRepo:      chatRepo.NewMessageRepo(messageDao),

		// LLM & RAG
		LLMClient:   llmClient,
		RAGService:  ragService,
		AsynqClient: asynqClient,

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
