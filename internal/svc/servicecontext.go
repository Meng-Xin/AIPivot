package svc

import (
	"context"
	"fmt"

	"aipivot/internal/config"
	"aipivot/internal/infra"
	"aipivot/internal/modules/auth/repo"
	"aipivot/internal/modules/auth/repo/dao"
	"aipivot/internal/observability"

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

	// Repo 接口（按 spec 规范使用接口类型，便于 Mock 测试）
	UserRepo   repo.UserRepository
	TenantRepo repo.TenantRepository
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

	// DB → DAO → Repo(接口) 组装链路
	userDao := dao.NewUserDao(db)
	tenantDao := dao.NewTenantDao(db)

	return &ServiceContext{
		Config:     c,
		DB:         db,
		Redis:      redisClient,
		Metrics:    metrics,
		UserRepo:   repo.NewUserRepo(userDao),
		TenantRepo: repo.NewTenantRepo(tenantDao),
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
