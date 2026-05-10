package webhook

import (
	"encoding/json"

	"aipivot/internal/shared/po"
	"aipivot/internal/types"
)

// WebhookPoToShow 将 Webhook PO 转为展示对象。
func WebhookPoToShow(wh *po.Webhook) types.ShowWebhook {
	var events []string
	_ = json.Unmarshal([]byte(wh.Events), &events)

	var lastTrigger int64
	if wh.LastTrigger != nil {
		lastTrigger = wh.LastTrigger.Unix()
	}

	return types.ShowWebhook{
		ID:          wh.ID,
		UUID:        wh.UUID,
		Name:        wh.Name,
		URL:         wh.URL,
		Events:      events,
		ChannelType: wh.ChannelType,
		Status:      wh.Status,
		RetryCount:  wh.RetryCount,
		TimeoutMs:   wh.TimeoutMs,
		LastError:   wh.LastError,
		LastTrigger: lastTrigger,
		CreatedAt:   wh.CreatedAt.Unix(),
		UpdatedAt:   wh.UpdatedAt.Unix(),
	}
}

// WebhookPoListToShowList 批量转换。
func WebhookPoListToShowList(list []*po.Webhook) []types.ShowWebhook {
	result := make([]types.ShowWebhook, 0, len(list))
	for _, wh := range list {
		result = append(result, WebhookPoToShow(wh))
	}
	return result
}
