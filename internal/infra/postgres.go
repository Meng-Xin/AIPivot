package infra

import (
	"context"
	"fmt"
	"time"

	"aipivot/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres(conf config.PostgresConf) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(postgresDSN(conf)), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get postgres raw db: %w", err)
	}

	if conf.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(conf.MaxOpenConns)
	}
	if conf.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
	}

	return db, nil
}

func CheckPostgres(db *gorm.DB) DependencyCheck {
	return DependencyCheck{
		Name: "postgres",
		Check: func(ctx context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return fmt.Errorf("get postgres raw db: %w", err)
			}

			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			if err := sqlDB.PingContext(pingCtx); err != nil {
				return fmt.Errorf("ping postgres: %w", err)
			}

			return nil
		},
	}
}

func ClosePostgres(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get postgres raw db: %w", err)
	}

	return sqlDB.Close()
}

func postgresDSN(conf config.PostgresConf) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		conf.Host,
		conf.User,
		conf.Password,
		conf.Database,
		conf.Port,
		conf.SSLMode,
		conf.TimeZone,
	)
}
