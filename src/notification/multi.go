package notification

import (
	"fmt"
	"strings"
)

// MultiNotifier 组合多个 Notifier，向所有渠道同时发送通知。
// 任一渠道成功即返回 nil；全部失败时返回聚合错误。
// 设计意图：单个渠道故障不阻塞其他渠道。
type MultiNotifier struct {
	notifiers []Notifier
}

// NewMultiNotifier 用给定的 notifier 列表构建聚合通知器。
func NewMultiNotifier(notifiers ...Notifier) *MultiNotifier {
	return &MultiNotifier{notifiers: notifiers}
}

// Send 依次向每个 notifier 发送，收集错误。
func (n *MultiNotifier) Send(title, message string) error {
	var errs []string
	success := 0
	for _, not := range n.notifiers {
		if err := not.Send(title, message); err != nil {
			errs = append(errs, err.Error())
			continue
		}
		success++
	}
	if success > 0 {
		return nil
	}
	return fmt.Errorf("all notifiers failed: %s", strings.Join(errs, "; "))
}
