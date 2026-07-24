package notification

// Notifier 抽象通知渠道，便于扩展多种推送方式
// （Windows toast、飞书、企业微信等）。所有通知实现需满足此接口。
type Notifier interface {
	// Send 发送一条通知。title 为标题，message 为正文。
	Send(title, message string) error
}
