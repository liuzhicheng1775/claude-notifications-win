package notification

import (
	"github.com/go-toast/toast"
)

type WindowsNotifier struct{}

func NewWindowsNotifier() *WindowsNotifier {
	return &WindowsNotifier{}
}

func (n *WindowsNotifier) Send(title, message string) error {
	notification := toast.Notification{
		AppID:   "Claude Code",
		Title:   title,
		Message: message,
	}
	return notification.Push()
}
