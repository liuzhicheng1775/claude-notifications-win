package cmd

import (
	"fmt"

	"claude-notifications-win/src/config"
	"claude-notifications-win/src/notification"
)

// RunTest 向所有启用的通知渠道发送测试消息，逐个报告结果。
// 会弹真实 Windows 通知 + 发真实飞书消息（验证用）。
func RunTest() error {
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

	type result struct {
		name string
		err  error
	}
	var results []result

	// Windows 通知（始终测试）
	w := notification.NewWindowsNotifier()
	results = append(results, result{"Windows", w.Send("Claude Code", "通知测试")})

	// 飞书（启用且 webhook 非空才测试）
	if cfg != nil && cfg.Notifications.Feishu.Enabled && cfg.Notifications.Feishu.Webhook != "" {
		f := notification.NewFeishuNotifier(
			cfg.Notifications.Feishu.Webhook,
			cfg.Notifications.Feishu.Secret,
		)
		results = append(results, result{"飞书", f.Send("Claude Code", "通知测试")})
	}

	success := 0
	fmt.Println("通知渠道测试：")
	for _, r := range results {
		if r.err != nil {
			fmt.Printf("  ✗ %s: %v\n", r.name, r.err)
		} else {
			fmt.Printf("  ✓ %s\n", r.name)
			success++
		}
	}
	fmt.Printf("\n%d/%d 渠道成功\n", success, len(results))

	return nil
}
