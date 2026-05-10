package repo

import (
	"context"

	"aipivot/internal/modules/knowledge/repo/dao"
)

type DocumentChunkRepo struct {
	chunkDao *dao.DocumentChunkDao
}

func NewDocumentChunkRepo(chunkDao *dao.DocumentChunkDao) *DocumentChunkRepo {
	return &DocumentChunkRepo{chunkDao: chunkDao}
}

func (r *DocumentChunkRepo) BatchCreateWithEmbedding(ctx context.Context, chunks []dao.ChunkWithEmbedding) error {
	return r.chunkDao.BatchCreateWithEmbedding(ctx, chunks)
}

func (r *DocumentChunkRepo) SimilaritySearch(ctx context.Context, kbID int64, queryEmbedding []float32, topK int) ([]dao.ChunkSearchResult, error) {
	return r.chunkDao.SimilaritySearch(ctx, kbID, queryEmbedding, topK)
}

func (r *DocumentChunkRepo) DeleteByDocumentID(ctx context.Context, docID int64) error {
	return r.chunkDao.DeleteByDocumentID(ctx, docID)
}

func (r *DocumentChunkRepo) CountByKnowledgeBaseID(ctx context.Context, kbID int64) (int64, error) {
	return r.chunkDao.CountByKnowledgeBaseID(ctx, kbID)
}
