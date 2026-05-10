package worker

import (
	"context"
	"fmt"

	"aipivot/internal/config"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logx"
)

// StartWorker 启动 Asynq 异步任务 worker（在独立 goroutine 中运行）。
// 返回 shutdown 函数用于优雅关闭。
func StartWorker(c config.Config, processor *DocumentProcessor) (func(), error) {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     c.Redis.Addr,
			Password: c.Redis.Password,
			DB:       c.Redis.DB,
		},
		asynq.Config{
			Concurrency: c.Worker.Concurrency,
			ErrorHandler: asynq.ErrorHandlerFunc(func(_ context.Context, task *asynq.Task, err error) {
				logx.Errorf("asynq task error: type=%s err=%v", task.Type(), err)
			}),
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeDocumentProcess, processor.ProcessTask)

	// 非阻塞启动 worker（使用 Start 而非 Run，避免 asynq 拦截 OS signal 导致 HTTP server 无法退出）
	if err := srv.Start(mux); err != nil {
		return nil, fmt.Errorf("start asynq worker: %w", err)
	}

	logx.Infof("asynq worker started — concurrency=%d", c.Worker.Concurrency)

	return func() {
		srv.Shutdown()
		logx.Info("asynq worker shutdown")
	}, nil
}

// NewAsynqClient 创建 Asynq 客户端（用于提交任务）。
func NewAsynqClient(c config.Config) *asynq.Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     c.Redis.Addr,
		Password: c.Redis.Password,
		DB:       c.Redis.DB,
	})
	return client
}

// EnqueueDocumentProcess 提交文档处理任务到队列。
func EnqueueDocumentProcess(client *asynq.Client, payload DocumentProcessPayload) error {
	task, err := NewDocumentProcessTask(payload)
	if err != nil {
		return fmt.Errorf("create document process task: %w", err)
	}
	info, err := client.Enqueue(task)
	if err != nil {
		return fmt.Errorf("enqueue document process task: %w", err)
	}
	logx.Infof("document:process enqueued — taskID=%s docID=%d", info.ID, payload.DocumentID)
	return nil
}
