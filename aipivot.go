package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"aipivot/internal/config"
	"aipivot/internal/handler"
	"aipivot/internal/observability"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/worker"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/aipivot-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	svcCtx, err := svc.NewServiceContext(c)
	if err != nil {
		logx.Errorf("failed to initialize service context: %v", err)
		panic(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := svcCtx.Shutdown(ctx); err != nil {
			logx.Errorf("failed to shutdown service context: %v", err)
		}
	}()

	// 启动 Asynq 异步任务 worker（文档处理 pipeline）
	if c.Worker.Enabled {
		processor := worker.NewDocumentProcessor(
			svcCtx.LLMClient,
			svcCtx.DocumentRepo,
			svcCtx.DocumentChunkRepo,
			svcCtx.KnowledgeBaseRepo,
		)
		workerShutdown, err := worker.StartWorker(c, processor)
		if err != nil {
			logx.Errorf("failed to start worker: %v", err)
		} else {
			defer workerShutdown()
		}
	}

	errorx.RegisterErrorHandler()
	server.Use(observability.Middleware(svcCtx.Metrics, c.Name))
	handler.RegisterHandlers(server, svcCtx)

	fmt.Printf("Starting %s at %s:%d...\n", c.Name, c.Host, c.Port)
	server.Start()
}
