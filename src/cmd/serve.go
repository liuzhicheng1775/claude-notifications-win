package cmd

import (
	"claude-notifications-win/config"
	"claude-notifications-win/hooks"
	"claude-notifications-win/notification"
	"os"
	"strings"
)

func HandleStopHook() error {
	notifier := notification.NewWindowsNotifier()
	cfg, err := config.Load()
	if err != nil {
		// Use default config if load fails
		cfg = &config.Config{}
	}

	handler := hooks.NewStopHandler(notifier, cfg)

	// Get optional reason from args
	reason := getFlag("--reason")

	return handler.Handle(reason)
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

	return handler.Handle(prompt)
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
