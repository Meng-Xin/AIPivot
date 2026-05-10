package svc

import (
	"context"
	"fmt"
	"net/http"

	"aipivot/internal/config"
	"aipivot/internal/infra"
	"aipivot/internal/middleware"
	authRepo "aipivot/internal/modules/auth/repo"
	authDao "aipivot/internal/modules/auth/repo/dao"
	chatRepo "aipivot/internal/modules/chat/repo"
	chatDao "aipivot/internal/modules/chat/repo/dao"
	kbRepo "aipivot/internal/modules/knowledge/repo"
	kbDao "aipivot/internal/modules/knowledge/repo/dao"
	"aipivot/internal/observability"
	"aipivot/internal/shared/query"

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

	// Chat Repo
	ConversationRepo chatRepo.ConversationRepository
	MessageRepo      chatRepo.MessageRepository
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

	// Chat DAOs
	conversationDao := chatDao.NewConversationDao(q)
	messageDao := chatDao.NewMessageDao(q)

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

		// Chat
		ConversationRepo: chatRepo.NewConversationRepo(conversationDao),
		MessageRepo:      chatRepo.NewMessageRepo(messageDao),

		HealthChecks: []infra.DependencyCheck{
			infra.CheckPostgres(db),
			infra.CheckRedis(redisClient),
		},
		Shutdown: func(ctx context.Context) error {
			var shutdownErr error
			if err := shutdown(ctx); err != nil {
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
