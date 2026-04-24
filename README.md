# Claude Notifications Win

适用于 Claude Code 的 Windows 气泡通知插件 | [English](./README_en.md)

---

## 功能特性

- ✅ **任务完成通知** - Claude Code 完成任务后自动弹窗提醒
- ✅ **权限审批通知** - 需要授权时弹窗提醒
- ✅ **Windows 原生气泡通知** - 使用 Windows 10/11 原生通知
- ✅ **零依赖** - 单文件运行，无需额外运行时

## 系统要求

- Windows 10 或 Windows 11
- Claude Code

## 安装方法

### 方法一：Claude Code 插件（推荐）

1. 添加插件市场：
   ```
   /plugin marketplace add liuzhicheng1775/claude-notifications-win
   ```

2. 安装插件：
   ```
   /plugin install claude-notifications-win@claude-notifications-win
   ```

3. 重启 Claude Code

### 方法二：手动安装

1. 从 [Releases](https://github.com/liuzhicheng1775/claude-notifications-win/releases) 下载最新的 `notify.exe`
2. 将 `notify.exe` 放到任意目录
3. 在 `~/.claude/settings.json` 中配置 hooks：
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

## 使用方法

安装后，以下通知会自动弹出：

| 通知类型 | 触发时机 |
|---------|---------|
| 任务完成 | Claude Code 完成一个任务 |
| 需要权限 | Claude Code 需要您授权操作 |

## 配置选项

在 `notify.exe` 同目录下创建 `config.json`：

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

## 开发

### 从源码构建

```bash
cd src
go mod download
go build -o ../bin/notify.exe .
```

### 项目结构

```
claude-notifications-win/
├── .claude-plugin/
│   ├── plugin.json          # 插件配置
│   └── marketplace.json     # 市场元数据
├── bin/
│   ├── notify.exe          # 编译后的二进制文件
│   ├── hook-wrapper.bat    # Windows hook 包装器
│   └── bootstrap.ps1       # 安装脚本
├── src/
│   ├── main.go             # 程序入口
│   ├── cmd/                # 命令处理
│   ├── hooks/              # Hook 实现
│   ├── notification/        # Windows 气泡通知
│   └── config/             # 配置管理
└── README.md
```

## 开源协议

MIT License
