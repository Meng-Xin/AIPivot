// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package models

import (
	"context"
	"time"

	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListModelsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取可用模型列表
func NewListModelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListModelsLogic {
	return &ListModelsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListModelsLogic) ListModels() (resp *types.ModelListResponse, err error) {
	cfg := l.svcCtx.Config.LLM

	chatModels := make([]types.ShowModel, 0, len(cfg.ChatModels))
	for _, m := range cfg.ChatModels {
		chatModels = append(chatModels, types.ShowModel{
			ID:        m.ID,
			Name:      m.Name,
			Type:      "chat",
			Provider:  m.Provider,
			MaxTokens: m.MaxTokens,
			IsDefault: m.ID == cfg.ChatModel,
		})
	}
	// 若配置列表为空，至少返回默认模型
	if len(chatModels) == 0 {
		chatModels = append(chatModels, types.ShowModel{
			ID:        cfg.ChatModel,
			Name:      cfg.ChatModel,
			Type:      "chat",
			IsDefault: true,
		})
	}

	embeddingModels := make([]types.ShowModel, 0, len(cfg.EmbeddingModels))
	for _, m := range cfg.EmbeddingModels {
		embeddingModels = append(embeddingModels, types.ShowModel{
			ID:        m.ID,
			Name:      m.Name,
			Type:      "embedding",
			Provider:  m.Provider,
			IsDefault: m.ID == cfg.EmbeddingModel,
		})
	}
	if len(embeddingModels) == 0 {
		embeddingModels = append(embeddingModels, types.ShowModel{
			ID:        cfg.EmbeddingModel,
			Name:      cfg.EmbeddingModel,
			Type:      "embedding",
			IsDefault: true,
		})
	}

	return &types.ModelListResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data: types.ModelListData{
			ChatModels:      chatModels,
			EmbeddingModels: embeddingModels,
		},
	}, nil
}
