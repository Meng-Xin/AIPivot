package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"aipivot/internal/shared/po"

	"github.com/zeromicro/go-zero/core/logx"
)

// EventType Webhook 事件类型常量。
const (
	EventMessageCreated     = "message.created"
	EventConversationClosed = "conversation.closed"
	EventEscalated          = "conversation.escalated"
)

// EventPayload Webhook 推送的事件载荷。
type EventPayload struct {
	Event     string `json:"event"`
	Timestamp int64  `json:"timestamp"`
	Data      any    `json:"data"`
}

// DeliveryService 负责将事件异步推送到租户注册的 Webhook URL。
type DeliveryService struct {
	repo       Repository
	httpClient *http.Client
}

func NewDeliveryService(repo Repository) *DeliveryService {
	return &DeliveryService{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Dispatch 查找订阅了指定事件的所有活跃 Webhook，异步投递。
// 不阻塞调用方：每个 Webhook 在独立 goroutine 中投递。
func (s *DeliveryService) Dispatch(ctx context.Context, tenantID int64, event string, data any) {
	webhooks, err := s.repo.GetActiveByTenantAndEvent(ctx, tenantID, event)
	if err != nil {
		logx.WithContext(ctx).Errorf("WebhookDelivery.Dispatch query err: %v", err)
		return
	}
	if len(webhooks) == 0 {
		return
	}

	payload := &EventPayload{
		Event:     event,
		Timestamp: time.Now().Unix(),
		Data:      data,
	}

	for _, wh := range webhooks {
		go s.deliver(wh, payload)
	}
}

// deliver 执行单次 Webhook 投递，包含重试逻辑。
func (s *DeliveryService) deliver(wh *po.Webhook, payload *EventPayload) {
	body, err := json.Marshal(payload)
	if err != nil {
		logx.Errorf("WebhookDelivery.deliver marshal err: webhookID=%d, err=%v", wh.ID, err)
		return
	}

	var lastErr error
	maxRetries := wh.RetryCount
	if maxRetries <= 0 {
		maxRetries = 1
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// 指数退避：1s, 2s, 4s...
			time.Sleep(time.Duration(1<<uint(attempt-1)) * time.Second)
		}

		req, err := http.NewRequest(http.MethodPost, wh.URL, bytes.NewReader(body))
		if err != nil {
			lastErr = fmt.Errorf("create request: %w", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "AIPivot-Webhook/1.0")
		req.Header.Set("X-Webhook-Event", payload.Event)

		// HMAC-SHA256 签名（签名密钥非空时生效）
		if wh.Secret != "" {
			sig := computeHMAC(body, wh.Secret)
			req.Header.Set("X-Webhook-Signature", sig)
		}

		// 单次请求使用 Webhook 自身的超时配置
		client := s.httpClient
		if wh.TimeoutMs > 0 {
			client = &http.Client{Timeout: time.Duration(wh.TimeoutMs) * time.Millisecond}
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http do: %w", err)
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// 投递成功：清空 last_error，更新 last_trigger
			_ = s.repo.Update(context.Background(), &po.Webhook{
				ID:          wh.ID,
				LastError:   "",
				LastTrigger: timePtr(time.Now()),
			})
			return
		}

		lastErr = fmt.Errorf("http status %d", resp.StatusCode)
	}

	// 所有重试失败：记录错误
	logx.Errorf("WebhookDelivery.deliver failed after %d attempts: webhookID=%d, url=%s, err=%v",
		maxRetries, wh.ID, wh.URL, lastErr)
	_ = s.repo.Update(context.Background(), &po.Webhook{
		ID:          wh.ID,
		LastError:   lastErr.Error(),
		LastTrigger: timePtr(time.Now()),
	})
}

func computeHMAC(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func timePtr(t time.Time) *time.Time {
	return &t
}
