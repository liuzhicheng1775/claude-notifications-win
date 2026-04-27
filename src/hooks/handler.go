package hooks

import (
	"claude-notifications-win/src/config"
	"claude-notifications-win/src/notification"
)

type Handler interface {
	Handle(reason string, taskName string) error
}

type StopHandler struct {
	notifier *notification.WindowsNotifier
	config   *config.Config
}

func NewStopHandler(notifier *notification.WindowsNotifier, cfg *config.Config) *StopHandler {
	return &StopHandler{
		notifier: notifier,
		config:   cfg,
	}
}

func (h *StopHandler) Handle(reason string, taskName string) error {
	title := "Claude Code"
	if taskName != "" {
		// Truncate long task names
		if len(taskName) > 50 {
			taskName = taskName[:47] + "..."
		}
		message := taskName
		return h.notifier.Send(title, message)
	}
	message := "任务已完成"
	if reason != "" {
		message = reason
	}
	return h.notifier.Send(title, message)
}

type PermissionHandler struct {
	notifier *notification.WindowsNotifier
	config   *config.Config
}

func NewPermissionHandler(notifier *notification.WindowsNotifier, cfg *config.Config) *PermissionHandler {
	return &PermissionHandler{
		notifier: notifier,
		config:   cfg,
	}
}

func (h *PermissionHandler) Handle(prompt string, context string) error {
	title := "Claude Code - 需要授权"
	message := "请授权以继续操作"
	if prompt != "" {
		// Truncate long prompts
		if len(prompt) > 100 {
			prompt = prompt[:97] + "..."
		}
		message = prompt
	}
	if context != "" && context != prompt {
		message = context + "\n\n" + message
	}
	return h.notifier.Send(title, message)
}
