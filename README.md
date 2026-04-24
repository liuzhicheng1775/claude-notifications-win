# Claude Notifications Win

Windows toast notifications for Claude Code hooks.

## Features

- **Task Complete Notification** - Get notified when Claude Code completes a task
- **Permission Prompt Notification** - Get notified when Claude Code needs permission
- **Windows Native Toast** - Uses Windows 10/11 native toast notifications
- **Zero Dependencies** - Single binary, no runtime required

## Installation

### Option 1: Claude Code Plugin (Recommended)

1. Add marketplace:
   ```
   /plugin marketplace add YOUR_USERNAME/claude-notifications-win
   ```

2. Install plugin:
   ```
   /plugin install claude-notifications-win@claude-notifications-win
   ```

3. Restart Claude Code

### Option 2: Manual Installation

1. Download the latest `notify.exe` from [Releases](https://github.com/YOUR_USERNAME/claude-notifications-win/releases)
2. Place `notify.exe` in a directory
3. Configure hooks in `~/.claude/settings.json`:
   ```json
   {
     "hooks": {
       "Stop": [{
         "hooks": [{
           "type": "command",
           "command": "path/to/notify.exe stop"
         }]
       }],
       "Notification": [{
         "matcher": "permission_prompt",
         "hooks": [{
           "type": "command",
           "command": "path/to/notify.exe permission"
         }]
       }]
     }
   }
   ```

## Usage

After installation, notifications will automatically appear:

- **Task Complete** - When Claude Code finishes a task
- **Permission Required** - When Claude Code needs your approval

## Configuration

Create `config.json` in the same directory as `notify.exe`:

```json
{
  "notifications": {
    "stop": {
      "enabled": true
    },
    "permission": {
      "enabled": true
    }
  }
}
```

## Development

### Building from Source

```bash
cd src
go mod download
go build -o ../bin/notify.exe .
```

### Project Structure

```
claude-notifications-win/
├── .claude-plugin/
│   └── plugin.json          # Plugin configuration
├── bin/
│   ├── notify.exe          # Built binary
│   ├── hook-wrapper.bat    # Windows hook wrapper
│   └── bootstrap.ps1       # Installation script
├── src/
│   ├── main.go             # Entry point
│   ├── cmd/                 # Command handlers
│   ├── hooks/              # Hook implementations
│   ├── notification/        # Windows toast notifications
│   └── config/              # Configuration
└── README.md
```

## License

MIT
