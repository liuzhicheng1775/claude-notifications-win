package hooks

import (
	"claude-notifications-win/src/config"
	"claude-notifications-win/src/notification"
)

// Handler 抽象 hook 处理器，接收已构造好的 Notification 并转发给 notifier。
type Handler interface {
	Handle(n notification.Notification) error
}

// StopHandler 处理 Claude Code Stop hook（任务完成）。
type StopHandler struct {
	notifier notification.Notifier
	config   *config.Config
}

func NewStopHandler(notifier notification.Notifier, cfg *config.Config) *StopHandler {
	return &StopHandler{
		notifier: notifier,
		config:   cfg,
	}
}

// Handle 直接转发 Notification 给 notifier。
// serve.go 负责 payload 解析、构造 Notification（含会话信息提取）。
func (h *StopHandler) Handle(n notification.Notification) error {
	// 截断过长的 message，避免飞书消息爆炸
	if len(n.Message) > 200 {
		n.Message = n.Message[:197] + "..."
	}
	if n.Message == "" {
		n.Message = "任务已完成"
	}
	if n.Title == "" {
		n.Title = "Claude Code"
	}
	return h.notifier.Send(n)
}

// PermissionHandler 处理 Claude Code Notification hook（permission_prompt）。
type PermissionHandler struct {
	notifier notification.Notifier
	config   *config.Config
}

func NewPermissionHandler(notifier notification.Notifier, cfg *config.Config) *PermissionHandler {
	return &PermissionHandler{
		notifier: notifier,
		config:   cfg,
	}
}

// Handle 转发授权请求通知。message 为空时使用默认提示。
func (h *PermissionHandler) Handle(n notification.Notification) error {
	if n.Title == "" {
		n.Title = "Claude Code - 需要授权"
	}
	if n.Message == "" {
		n.Message = "请授权以继续操作"
	}
	// 截断过长的 message
	if len(n.Message) > 200 {
		n.Message = n.Message[:197] + "..."
	}
	return h.notifier.Send(n)
}
