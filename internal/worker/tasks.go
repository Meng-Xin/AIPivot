package worker

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

// Task type 常量，所有异步任务类型在此注册。
const (
	TypeDocumentProcess = "document:process"
)

// DocumentProcessPayload 文档处理任务载荷。
type DocumentProcessPayload struct {
	DocumentID      int64  `json:"document_id"`
	KnowledgeBaseID int64  `json:"knowledge_base_id"`
	TenantID        int64  `json:"tenant_id"`
	EmbeddingModel  string `json:"embedding_model"`
	EmbeddingDim    int    `json:"embedding_dim"`
}

// NewDocumentProcessTask 创建文档处理异步任务。
func NewDocumentProcessTask(payload DocumentProcessPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	// 最多重试 3 次，超时 10 分钟（大文档 embedding 耗时较长）
	return asynq.NewTask(TypeDocumentProcess, data,
		asynq.MaxRetry(3),
		asynq.Timeout(10*time.Minute),
	), nil
}
