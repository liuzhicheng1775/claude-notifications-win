package hooks

import (
	"claude-notifications-win/config"
	"claude-notifications-win/notification"
)

type Handler interface {
	Handle(reason string) error
}

type StopHandler struct {
	notifier *notification.WindowsNotifier
	config    *config.Config
}

func NewStopHandler(notifier *notification.WindowsNotifier, cfg *config.Config) *StopHandler {
	return &StopHandler{
		notifier: notifier,
		config:   cfg,
	}
}

func (h *StopHandler) Handle(reason string) error {
	title := "Claude Code"
	message := "Task completed"
	if reason != "" {
		message = "Task completed: " + reason
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

func (h *PermissionHandler) Handle(prompt string) error {
	title := "Claude Code - Permission Required"
	message := "A permission is required"
	if prompt != "" {
		// Truncate long prompts
		if len(prompt) > 100 {
			prompt = prompt[:97] + "..."
		}
		message = prompt
	}
	return h.notifier.Send(title, message)
}
