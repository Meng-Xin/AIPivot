package knowledge

import (
	"context"
	"fmt"
	"strings"

	knowledgemod "aipivot/internal/modules/knowledge"
	"aipivot/internal/shared/query"

	"gorm.io/gorm"
)

// DocChunkRepo 文档切块数据仓储实现。
type DocChunkRepo struct {
	q  *query.Query
	db *gorm.DB
}

func NewDocChunkRepo(q *query.Query, db *gorm.DB) *DocChunkRepo {
	return &DocChunkRepo{q: q, db: db}
}

// BatchCreateWithEmbedding 批量插入带向量的 document chunks。
// 使用 raw SQL 因为 GORM Gen 不直接支持 pgvector 类型。
func (r *DocChunkRepo) BatchCreateWithEmbedding(ctx context.Context, chunks []knowledgemod.ChunkWithEmbedding) error {
	if len(chunks) == 0 {
		return nil
	}

	// 构建批量 INSERT（每批最多 100 条避免 SQL 过长）
	const batchSize = 100
	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		if err := r.batchInsert(ctx, chunks[i:end]); err != nil {
			return err
		}
	}

	return nil
}

func (r *DocChunkRepo) batchInsert(ctx context.Context, chunks []knowledgemod.ChunkWithEmbedding) error {
	var sb strings.Builder
	sb.WriteString(`INSERT INTO document_chunks (document_id, knowledge_base_id, tenant_id, chunk_index, content, token_count, embedding, metadata) VALUES `)

	args := make([]any, 0, len(chunks)*8)
	for i, c := range chunks {
		if i > 0 {
			sb.WriteString(", ")
		}
		paramBase := i * 8
		sb.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d::vector, $%d)",
			paramBase+1, paramBase+2, paramBase+3, paramBase+4,
			paramBase+5, paramBase+6, paramBase+7, paramBase+8))
		args = append(args,
			c.DocumentID, c.KnowledgeBaseID, c.TenantID,
			c.ChunkIndex, c.Content, c.TokenCount,
			vectorToString(c.Embedding), c.Metadata)
	}

	return r.db.WithContext(ctx).Exec(sb.String(), args...).Error
}

// SimilaritySearch 基于 pgvector cosine 相似度搜索最相关的 chunks。
func (r *DocChunkRepo) SimilaritySearch(ctx context.Context, kbID int64, queryEmbedding []float32, topK int) ([]knowledgemod.ChunkSearchResult, error) {
	sql := `
		SELECT id, uuid, document_id, knowledge_base_id, chunk_index, content, token_count, metadata,
		       1 - (embedding <=> $1::vector) AS score
		FROM document_chunks
		WHERE knowledge_base_id = $2 AND embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $3
	`

	var results []knowledgemod.ChunkSearchResult
	err := r.db.WithContext(ctx).Raw(sql, vectorToString(queryEmbedding), kbID, topK).Scan(&results).Error
	return results, err
}

// DeleteByDocumentID 按文档 ID 删除所有切块（级联删除已在 SQL 层保证，此方法用于显式清理）。
func (r *DocChunkRepo) DeleteByDocumentID(ctx context.Context, docID int64) error {
	chunk := r.q.DocumentChunk
	_, err := chunk.WithContext(ctx).Where(chunk.DocumentID.Eq(docID)).Delete()
	return err
}

// CountByKnowledgeBaseID 统计知识库下的切块总数。
func (r *DocChunkRepo) CountByKnowledgeBaseID(ctx context.Context, kbID int64) (int64, error) {
	chunk := r.q.DocumentChunk
	return chunk.WithContext(ctx).Where(chunk.KnowledgeBaseID.Eq(kbID)).Count()
}

// vectorToString 将 float32 切片转为 pgvector 格式字符串 "[0.1,0.2,...]"。
func vectorToString(v []float32) string {
	if len(v) == 0 {
		return "[]"
	}
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = fmt.Sprintf("%g", f)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
