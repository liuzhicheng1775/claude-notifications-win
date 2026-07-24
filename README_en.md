# Claude Notifications Win

Windows toast notification tool for Claude Code - get notified when tasks finish.

[中文](./README.md) | English

> Stop watching the terminal — get an instant pop-up when Claude Code completes a task.

---

## Features

- ✅ **Task Complete Notification** - Get notified when Claude Code completes a task, no need to watch the terminal
- ✅ **Permission Prompt Notification** - Get notified when Claude Code needs permission
- ✅ **Windows Native Toast** - Uses Windows 10/11 native toast notifications
- ✅ **Feishu Notification (Optional)** - Forward to a Feishu group bot, with HMAC signing
- ✅ **Multi-Channel Fan-Out** - Windows + Feishu in parallel; one channel failing won't block others
- ✅ **Zero Dependencies** - Single binary, no runtime required

## Requirements

- Windows 10 or Windows 11
- Claude Code

## Installation

### Option 1: Manual Installation (Recommended)

1. Download the latest `notify.exe` from [Releases](https://github.com/liuzhicheng1775/claude-notifications-win/releases)
2. Run the one-click setup wizard:
   ```bash
   notify.exe init
   ```
   The wizard interactively walks you through everything:
   - Configure Feishu notifications (optional, press Enter to skip)
   - Automatically writes `config.json`
   - Automatically writes Stop / Notification hooks to `~/.claude/settings.json` (skips if already present, no duplicates)
   - Optionally sends a test notification right away

3. Restart Claude Code to take effect.

> `init` auto-detects the actual path of `notify.exe`, so it survives moving directories or switching machines. Re-run `init` to change config, or edit `config.json` directly.

### Option 2: Claude Code Plugin

> **Note**: If you encounter SSH key issues, run this first to configure git to use HTTPS:
> ```bash
> git config --global url."https://github.com/".insteadOf "git@github.com:"
> ```

1. Add marketplace:
   ```
   /plugin marketplace add liuzhicheng1775/claude-notifications-win
   ```

2. Install plugin:
   ```
   /plugin install claude-notifications-win@claude-notifications-win
   ```

3. Restart Claude Code

### Option 3: Manual Configuration (without the wizard)

If you prefer not to use the `init` wizard, you can configure manually:

1. Create `config.json` in the same directory as `notify.exe` (optional, defaults to all enabled):
   ```json
   {
     "notifications": {
       "stop": { "enabled": true },
       "permission": { "enabled": true }
     }
   }
   ```
2. Configure hooks in `~/.claude/settings.json`:
   ```json
   {
     "hooks": {
       "Stop": [{
         "hooks": [{
           "type": "command",
           "command": "C:/path/to/notify.exe stop"
         }]
       }],
       "Notification": [{
         "matcher": "permission_prompt",
         "hooks": [{
           "type": "command",
           "command": "C:/path/to/notify.exe permission"
         }]
       }]
     }
   }
   ```

## Commands

| Command | Purpose |
|---------|---------|
| `notify.exe init` | Interactive setup wizard (Feishu + config.json + hooks in one shot) |
| `notify.exe test` | Send a test message to all enabled channels, report ✓/✗ per channel |
| `notify.exe stop` | Trigger task-complete notification (called by Claude Code Stop hook) |
| `notify.exe permission` | Trigger permission-prompt notification (called by Claude Code Notification hook) |
| `notify.exe version` | Show version |

## Usage

After installation and configuration, notifications appear automatically:

| Notification Type | Trigger |
|-------------------|---------|
| Task Complete | When Claude Code finishes a task |
| Permission Required | When Claude Code needs your approval |

Verify your setup:

```bash
notify.exe test
```

Sends a test message to Windows toast and any enabled Feishu channel, then prints an `N/M channels succeeded` summary.

## Feishu Notification (Optional)

> Recommended: use the `notify.exe init` wizard. Below is the manual configuration reference. See [docs/feishu-notification.md](./docs/feishu-notification.md) for detailed steps.

In addition to Windows toast, notifications can be sent to a Feishu group. Create a custom bot in your Feishu group, obtain the webhook URL (and secret if signature verification is enabled), and add it to `config.json`:

```json
{
  "notifications": {
    "feishu": {
      "enabled": true,
      "webhook": "https://open.feishu.cn/open-apis/bot/v2/hook/your_hook_id",
      "secret": "optional, required only if bot signature verification is enabled"
    }
  }
}
```

When `enabled: true` and `webhook` is non-empty, notifications are sent to both Windows toast and the Feishu group. Without the `feishu` field, only Windows toast is used.

## Development

### Building from Source

```bash
cd src
go mod download
# Version auto-filled from git tag (64-bit)
GOARCH=amd64 go build -ldflags "-X main.version=$(git describe --tags --abbrev=0 | sed 's/^v//')" -o ../bin/notify.exe .
```

### Project Structure

```
claude-notifications-win/
├── bin/
│   ├── notify.exe          # Built binary
│   ├── hook-wrapper.bat    # Windows hook wrapper
│   └── bootstrap.ps1       # Installation script
├── src/
│   ├── main.go             # Entry point
│   ├── cmd/                # Command handlers
│   ├── hooks/              # Hook implementations
│   ├── notification/        # Windows toast notifications
│   └── config/             # Configuration
├── .claude-plugin/
│   ├── plugin.json          # Plugin configuration
│   └── marketplace.json     # Marketplace metadata
└── README.md
```

## License

MIT License
