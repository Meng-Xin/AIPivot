package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	Postgres  PostgresConf
	Redis     RedisConf
	Telemetry TelemetryConf
	Metrics   MetricsConf
	Auth      AuthConf
	Migration MigrationConf
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
