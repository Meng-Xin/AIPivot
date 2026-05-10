package channel

// ChannelType 渠道类型标识，标记消息来源或会话接入渠道。
type ChannelType string

const (
	Web     ChannelType = "web"     // 管理后台/前端 Chat Widget（JWT 认证）
	API     ChannelType = "api"     // 外部系统通过 REST API 接入（API Key 认证）
	Webhook ChannelType = "webhook" // 第三方平台 Webhook 回调接入
	Wechat  ChannelType = "wechat"  // 微信（Phase 3）
	Feishu  ChannelType = "feishu"  // 飞书（Phase 3）
)

// Valid 校验渠道类型是否合法。
func (c ChannelType) Valid() bool {
	switch c {
	case Web, API, Webhook, Wechat, Feishu:
		return true
	}
	return false
}

func (c ChannelType) String() string {
	return string(c)
}

// Adapter 渠道适配器接口，不同渠道实现各自的消息收发逻辑。
// MVP 阶段 Web 和 API 渠道不需要 Adapter（直接走 HTTP handler），
// Webhook/Wechat/Feishu 等需要适配器处理平台特定的消息格式。
type Adapter interface {
	// Type 返回渠道类型标识。
	Type() ChannelType

	// ValidateRequest 校验渠道入站请求的合法性（如签名验证）。
	ValidateRequest(payload []byte, signature string) bool

	// ParseInbound 解析渠道入站消息为统一 InboundMessage 格式。
	ParseInbound(payload []byte) (*InboundMessage, error)

	// FormatOutbound 将统一 OutboundMessage 转为渠道特定的响应格式。
	FormatOutbound(msg *OutboundMessage) ([]byte, error)
}

// InboundMessage 渠道入站消息的统一格式（外部平台 → AIPivot）。
type InboundMessage struct {
	ChannelType    ChannelType       // 渠道类型
	ExternalUserID string            // 外部平台用户标识
	Content        string            // 消息内容
	ContentType    string            // 内容类型: text / image / file
	Metadata       map[string]string // 渠道特定的元数据
}

// OutboundMessage 渠道出站消息的统一格式（AIPivot → 外部平台）。
type OutboundMessage struct {
	ConversationID int64             // 会话 ID
	MessageID      string            // 消息 UUID
	Role           string            // 角色: assistant / system
	Content        string            // 消息内容
	ContentType    string            // 内容类型
	Model          string            // 使用的模型
	Sources        []string          // RAG 来源引用
	Metadata       map[string]string // 附加元数据
}
