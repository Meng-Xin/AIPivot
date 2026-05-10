package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"aipivot/internal/modules/knowledge/repo"
	"aipivot/internal/modules/knowledge/repo/dao"
	"aipivot/pkg/chunker"
	"aipivot/pkg/llm"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logx"
)

// DocumentProcessor 文档异步处理器：读取文档内容 → 切块 → Embedding → 存储到 pgvector。
type DocumentProcessor struct {
	llmClient *llm.Client
	docRepo   repo.DocumentRepository
	chunkRepo repo.DocumentChunkRepository
	kbRepo    repo.KnowledgeBaseRepository
}

func NewDocumentProcessor(
	llmClient *llm.Client,
	docRepo repo.DocumentRepository,
	chunkRepo repo.DocumentChunkRepository,
	kbRepo repo.KnowledgeBaseRepository,
) *DocumentProcessor {
	return &DocumentProcessor{
		llmClient: llmClient,
		docRepo:   docRepo,
		chunkRepo: chunkRepo,
		kbRepo:    kbRepo,
	}
}

func (p *DocumentProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload DocumentProcessPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	logx.WithContext(ctx).Infof("document:process start — docID=%d kbID=%d", payload.DocumentID, payload.KnowledgeBaseID)

	// 1. 标记文档为 processing
	if err := p.docRepo.UpdateStatus(ctx, payload.DocumentID, "processing", ""); err != nil {
		return fmt.Errorf("update doc status to processing: %w", err)
	}

	// 2. 读取文档内容（MVP 阶段仅支持纯文本，file_path 中存储文本内容）
	doc, err := p.docRepo.GetByID(ctx, payload.DocumentID)
	if err != nil || doc == nil {
		p.markFailed(ctx, payload.DocumentID, "文档不存在")
		return fmt.Errorf("get document: %w", err)
	}

	content := doc.FilePath // MVP: 文本内容暂存于 file_path 字段
	if content == "" {
		p.markFailed(ctx, payload.DocumentID, "文档内容为空")
		return fmt.Errorf("empty document content for docID=%d", payload.DocumentID)
	}

	// 3. 切块
	chunks := chunker.SplitText(content, chunker.DefaultConfig())
	if len(chunks) == 0 {
		p.markFailed(ctx, payload.DocumentID, "切块结果为空")
		return fmt.Errorf("chunking produced 0 chunks for docID=%d", payload.DocumentID)
	}

	logx.WithContext(ctx).Infof("document:process chunked — docID=%d chunks=%d", payload.DocumentID, len(chunks))

	// 4. 批量 Embedding（每批最多 20 条防止超时）
	const embeddingBatch = 20
	chunkData := make([]dao.ChunkWithEmbedding, 0, len(chunks))

	for i := 0; i < len(chunks); i += embeddingBatch {
		end := i + embeddingBatch
		if end > len(chunks) {
			end = len(chunks)
		}

		texts := make([]string, end-i)
		for j, c := range chunks[i:end] {
			texts[j] = c.Content
		}

		vectors, err := p.llmClient.Embed(ctx, payload.EmbeddingModel, texts)
		if err != nil {
			p.markFailed(ctx, payload.DocumentID, fmt.Sprintf("embedding 失败: %v", err))
			return fmt.Errorf("embed batch [%d:%d]: %w", i, end, err)
		}

		for j, c := range chunks[i:end] {
			chunkData = append(chunkData, dao.ChunkWithEmbedding{
				DocumentID:      payload.DocumentID,
				KnowledgeBaseID: payload.KnowledgeBaseID,
				TenantID:        payload.TenantID,
				ChunkIndex:      c.Index,
				Content:         c.Content,
				TokenCount:      estimateTokens(c.Content),
				Embedding:       vectors[j],
				Metadata:        "{}",
			})
		}
	}

	// 5. 批量写入 pgvector
	if err := p.chunkRepo.BatchCreateWithEmbedding(ctx, chunkData); err != nil {
		p.markFailed(ctx, payload.DocumentID, fmt.Sprintf("存储切块失败: %v", err))
		return fmt.Errorf("batch create chunks: %w", err)
	}

	// 6. 更新文档状态和计数
	if err := p.docRepo.UpdateStatus(ctx, payload.DocumentID, "completed", ""); err != nil {
		return fmt.Errorf("update doc status to completed: %w", err)
	}

	// 更新知识库 chunk_count（最终一致性，非精确事务）
	totalChunks, _ := p.chunkRepo.CountByKnowledgeBaseID(ctx, payload.KnowledgeBaseID)
	_ = p.kbRepo.Update(ctx, payload.KnowledgeBaseID, map[string]any{
		"chunk_count": totalChunks,
	})

	logx.WithContext(ctx).Infof("document:process done — docID=%d chunks=%d", payload.DocumentID, len(chunkData))

	return nil
}

func (p *DocumentProcessor) markFailed(ctx context.Context, docID int64, errMsg string) {
	if err := p.docRepo.UpdateStatus(ctx, docID, "failed", errMsg); err != nil {
		logx.WithContext(ctx).Errorf("markFailed docID=%d err: %v", docID, err)
	}
}

// estimateTokens 粗略估算 token 数（英文 ~4 chars/token, 中文 ~1.5 chars/token，取中间值 ~2.5）。
func estimateTokens(text string) int {
	runeCount := utf8.RuneCountInString(text)
	return (runeCount*2 + 4) / 5
}
