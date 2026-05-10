// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/po"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateApiKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建 API Key
func NewCreateApiKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateApiKeyLogic {
	return &CreateApiKeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateApiKeyLogic) CreateApiKey(req *types.CreateApiKeyRequest) (resp *types.CreateApiKeyResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)

	scopes := req.Scopes
	if len(scopes) == 0 {
		scopes = []string{"chat"}
	}
	scopesJSON, _ := json.Marshal(scopes)

	// 生成 32 字节随机密钥，格式: sk-<hex>
	rawBytes := make([]byte, 32)
	if _, err = rand.Read(rawBytes); err != nil {
		l.Logger.Errorf("CreateApiKey rand err: %v", err)
		return nil, errorx.NewInternalError("生成 API Key 失败")
	}
	rawKey := fmt.Sprintf("sk-%s", hex.EncodeToString(rawBytes))
	keyPrefix := rawKey[:10] // "sk-" + 前 7 位 hex

	h := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(h[:])

	apiKey := &po.ApiKey{
		TenantID:  tenantID,
		Name:      req.Name,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
		Scopes:    string(scopesJSON),
		Status:    "active",
	}

	if err = l.svcCtx.ApiKeyRepo.Create(l.ctx, apiKey); err != nil {
		l.Logger.Errorf("CreateApiKey Create err: %v", err)
		return nil, errorx.NewInternalError("创建 API Key 失败")
	}

	return &types.CreateApiKeyResponse{
		Code:      0,
		Msg:       "创建成功",
		Timestamp: time.Now().Unix(),
		Data: types.CreateApiKeyData{
			ID:        apiKey.ID,
			Name:      apiKey.Name,
			Key:       rawKey, // 仅此一次返回原始密钥
			KeyPrefix: keyPrefix,
			Scopes:    scopes,
		},
	}, nil
}
