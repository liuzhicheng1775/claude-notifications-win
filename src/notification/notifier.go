package notification

// Notification 表示一条待发送的通知。
// Title 和 Message 为必填的展示文本；SessionID 和 SessionTitle
// 为可选的会话上下文，由 hook 解析 Claude Code payload 后填充，
// 供飞书等富文本渠道在消息体里附带会话信息。
// Windows toast 等简洁渠道可忽略会话字段。
type Notification struct {
	Title        string
	Message      string
	SessionID    string
	SessionTitle string
}

// Notifier 抽象通知渠道，便于扩展多种推送方式
// （Windows toast、飞书、企业微信等）。所有通知实现需满足此接口。
type Notifier interface {
	// Send 发送一条通知。实现应忽略自身不关心的字段
	// （例如 Windows toast 只用 Title/Message）。
	Send(n Notification) error
}
