package dao

import (
	"context"
	"fmt"
	"strings"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DocumentChunkDao struct {
	q  *query.Query
	db *gorm.DB
}

func NewDocumentChunkDao(q *query.Query, db *gorm.DB) *DocumentChunkDao {
	return &DocumentChunkDao{q: q, db: db}
}

func (d *DocumentChunkDao) WithTx(tx *query.Query) *DocumentChunkDao {
	return &DocumentChunkDao{q: tx, db: d.db}
}

// BatchCreateWithEmbedding 批量插入带向量的 document chunks。
// 使用 raw SQL 因为 GORM Gen 不直接支持 pgvector 类型。
func (d *DocumentChunkDao) BatchCreateWithEmbedding(ctx context.Context, chunks []ChunkWithEmbedding) error {
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
		if err := d.batchInsert(ctx, chunks[i:end]); err != nil {
			return err
		}
	}

	return nil
}

func (d *DocumentChunkDao) batchInsert(ctx context.Context, chunks []ChunkWithEmbedding) error {
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

	if err := d.db.WithContext(ctx).Exec(sb.String(), args...).Error; err != nil {
		logx.WithContext(ctx).Errorf("DocumentChunkDao.BatchCreateWithEmbedding err: %v", err)
		return err
	}

	return nil
}

// SimilaritySearch 基于 pgvector cosine 相似度搜索最相关的 chunks。
func (d *DocumentChunkDao) SimilaritySearch(ctx context.Context, kbID int64, queryEmbedding []float32, topK int) ([]ChunkSearchResult, error) {
	sql := `
		SELECT id, uuid, document_id, knowledge_base_id, chunk_index, content, token_count, metadata,
		       1 - (embedding <=> $1::vector) AS score
		FROM document_chunks
		WHERE knowledge_base_id = $2 AND embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $3
	`

	var results []ChunkSearchResult
	err := d.db.WithContext(ctx).Raw(sql, vectorToString(queryEmbedding), kbID, topK).Scan(&results).Error
	if err != nil {
		logx.WithContext(ctx).Errorf("DocumentChunkDao.SimilaritySearch err: %v", err)
		return nil, err
	}

	return results, nil
}

// DeleteByDocumentID 按文档 ID 删除所有切块（级联删除已在 SQL 层保证，此方法用于显式清理）。
func (d *DocumentChunkDao) DeleteByDocumentID(ctx context.Context, docID int64) error {
	chunk := d.q.DocumentChunk
	_, err := chunk.WithContext(ctx).Where(chunk.DocumentID.Eq(docID)).Delete()
	if err != nil {
		logx.WithContext(ctx).Errorf("DocumentChunkDao.DeleteByDocumentID err: %v", err)
		return err
	}
	return nil
}

// CountByKnowledgeBaseID 统计知识库下的切块总数。
func (d *DocumentChunkDao) CountByKnowledgeBaseID(ctx context.Context, kbID int64) (int64, error) {
	chunk := d.q.DocumentChunk
	count, err := chunk.WithContext(ctx).Where(chunk.KnowledgeBaseID.Eq(kbID)).Count()
	if err != nil {
		logx.WithContext(ctx).Errorf("DocumentChunkDao.CountByKnowledgeBaseID err: %v", err)
		return 0, err
	}
	return count, nil
}

// ========== 辅助类型 ==========

// ChunkWithEmbedding 带向量的切块数据，用于批量写入。
type ChunkWithEmbedding struct {
	DocumentID      int64
	KnowledgeBaseID int64
	TenantID        int64
	ChunkIndex      int
	Content         string
	TokenCount      int
	Embedding       []float32
	Metadata        string
}

// ChunkSearchResult 相似度搜索结果。
type ChunkSearchResult struct {
	ID              int64   `gorm:"column:id"`
	UUID            string  `gorm:"column:uuid"`
	DocumentID      int64   `gorm:"column:document_id"`
	KnowledgeBaseID int64   `gorm:"column:knowledge_base_id"`
	ChunkIndex      int     `gorm:"column:chunk_index"`
	Content         string  `gorm:"column:content"`
	TokenCount      int     `gorm:"column:token_count"`
	Metadata        string  `gorm:"column:metadata"`
	Score           float64 `gorm:"column:score"`
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

// 保留未使用的 po 引用以供后续 GORM Gen 查询
var _ = po.DocumentChunk{}
