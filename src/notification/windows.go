package notification

import (
	"github.com/go-toast/toast"
)

type WindowsNotifier struct{}

func NewWindowsNotifier() *WindowsNotifier {
	return &WindowsNotifier{}
}

// Send 弹出 Windows toast 通知。
// 只使用 Title 和 Message，忽略会话字段（toast 文本保持简洁）。
func (n *WindowsNotifier) Send(noti Notification) error {
	t := toast.Notification{
		AppID:   "Claude Code",
		Title:   noti.Title,
		Message: noti.Message,
	}
	return t.Push()
}
