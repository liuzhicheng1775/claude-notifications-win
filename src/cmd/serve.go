package cmd

import (
	"bufio"
	"claude-notifications-win/src/config"
	"claude-notifications-win/src/hooks"
	"claude-notifications-win/src/notification"
	"encoding/json"
	"os"
	"strings"
)

type ClaudeHookPayload struct {
	Type      string `json:"type"`
	Subject   string `json:"subject"`
	StopType  string `json:"stopType,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

func HandleStopHook() error {
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

	notifier := BuildNotifier(cfg)
	handler := hooks.NewStopHandler(notifier, cfg)

	// Get optional reason from args
	reason := getFlag("--reason")

	// Read from stdin for task info from Claude Code
	taskName := readStdin()

	return handler.Handle(reason, taskName)
}

func HandlePermissionHook() error {
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

	notifier := BuildNotifier(cfg)
	handler := hooks.NewPermissionHandler(notifier, cfg)

	// Get prompt from args
	prompt := getFlag("--prompt")

	// Read from stdin for additional context
	context := readStdin()

	return handler.Handle(prompt, context)
}

func readStdin() string {
	// Check if stdin has data
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		reader := bufio.NewReader(os.Stdin)
		var sb strings.Builder
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			sb.Write(scanner.Bytes())
			sb.WriteByte('\n')
		}
		input := strings.TrimSpace(sb.String())
		if input != "" {
			// Try to parse as JSON first
			var payload ClaudeHookPayload
			if err := json.Unmarshal([]byte(input), &payload); err == nil {
				if payload.Subject != "" {
					return extractTaskName(payload.Subject)
				}
				if payload.Reason != "" {
					return payload.Reason
				}
				// Valid JSON but no useful content - return empty to skip notification
				return ""
			}
			// Return raw input if not JSON, but filter out session-like garbage
			return filterGarbage(input)
		}
	}
	return ""
}

// filterGarbage removes unwanted noise from raw input
func filterGarbage(input string) string {
	// If input looks like session ID garbage, return empty
	if strings.Contains(input, "session id") || strings.Contains(input, "session") {
		// Check if it's mostly session garbage
		lines := strings.Split(input, "\n")
		if len(lines) > 0 && strings.Contains(lines[0], "session") {
			return ""
		}
	}
	// If input is mostly JSON-like garbage, return empty
	if strings.HasPrefix(strings.TrimSpace(input), "{") && strings.Contains(input, "\"session") {
		return ""
	}
	return input
}

// extractTaskName extracts the meaningful task name from a potentially noisy subject string
func extractTaskName(subject string) string {
	// Pattern: "[会话:xxx] ... - 任务：xxx" or "xxx - 任务：xxx"
	// Extract the part after " - 任务：" or just after " - " if it looks like a task
	if idx := strings.LastIndex(subject, " - 任务："); idx != -1 {
		return subject[idx+len(" - 任务："):]
	}
	if idx := strings.LastIndex(subject, "- 任务："); idx != -1 {
		return subject[idx+len("- 任务："):]
	}
	// If no clear task marker, try to find "任务：" anywhere
	if idx := strings.Index(subject, "任务："); idx != -1 {
		return subject[idx+len("任务："):]
	}
	// If subject starts with session pattern, try to extract after ]
	// Pattern: "[会话:xxx] xxx" or "[会话 xxx] xxx"
	if idx := strings.Index(subject, "]"); idx != -1 && idx < len(subject)-1 {
		return strings.TrimSpace(subject[idx+1:])
	}
	// Return cleaned subject, removing leading brackets/prefixes
	return strings.TrimSpace(subject)
}

func getFlag(name string) string {
	for i, arg := range os.Args {
		if strings.HasPrefix(arg, name+"=") {
			return strings.TrimPrefix(arg, name+"=")
		}
		if arg == name && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return ""
}

func getAllArgs() []string {
	// Get all args after command name
	if len(os.Args) < 3 {
		return []string{}
	}
	return os.Args[2:]
}

func extractArg(prefix string) string {
	for _, arg := range getAllArgs() {
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix+"=")
		}
	}
	return ""
}

// BuildNotifier 根据配置组装通知器。
// Windows 通知始终启用；飞书在 enabled 且 webhook 非空时启用。
// 多渠道时用 MultiNotifier 聚合，任一渠道成功即视为成功。
func BuildNotifier(cfg *config.Config) notification.Notifier {
	var notifiers []notification.Notifier
	notifiers = append(notifiers, notification.NewWindowsNotifier())
	if cfg != nil && cfg.Notifications.Feishu.Enabled && cfg.Notifications.Feishu.Webhook != "" {
		notifiers = append(notifiers, notification.NewFeishuNotifier(
			cfg.Notifications.Feishu.Webhook, cfg.Notifications.Feishu.Secret))
	}
	if len(notifiers) == 1 {
		return notifiers[0]
	}
	return notification.NewMultiNotifier(notifiers...)
}
