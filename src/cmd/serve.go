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
	notifier := notification.NewWindowsNotifier()
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

	handler := hooks.NewStopHandler(notifier, cfg)

	// Get optional reason from args
	reason := getFlag("--reason")

	// Read from stdin for task info from Claude Code
	taskName := readStdin()

	return handler.Handle(reason, taskName)
}

func HandlePermissionHook() error {
	notifier := notification.NewWindowsNotifier()
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

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
					return payload.Subject
				}
				if payload.Reason != "" {
					return payload.Reason
				}
			}
			// Return raw input if not JSON
			return input
		}
	}
	return ""
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
