# Claude Notifications Win

适用于 Claude Code 的任务完成通知提醒插件 - Windows 气泡通知工具 | [English](./README_en.md)

> 让 Claude Code 任务完成时自动弹窗提醒，无需盯着终端，任务结束即时通知。

---

## 功能特性

- ✅ **任务完成通知** - Claude Code 完成任务后自动弹窗提醒，无需盯着终端
- ✅ **权限审批通知** - 需要授权时弹窗提醒
- ✅ **Windows 原生气泡通知** - 使用 Windows 10/11 原生通知
- ✅ **飞书通知（可选）** - 同步推送到飞书群机器人，支持加签校验
- ✅ **多渠道聚合** - Windows + 飞书同时发送，单渠道故障不阻塞其他渠道
- ✅ **零依赖** - 单文件运行，无需额外运行时

## 系统要求

- Windows 10 或 Windows 11
- Claude Code

## 安装方法

### 方法一：手动安装（推荐）

1. 从 [Releases](https://github.com/liuzhicheng1775/claude-notifications-win/releases) 下载最新的 `notify.exe`
2. 双击 `notify.exe` 启动配置向导，或在终端执行（PowerShell / cmd 中需要 `.\` 前缀）：
   ```powershell
   .\notify.exe init
   ```
   向导会交互式引导你完成全部配置：
   - 配置飞书通知（可选，回车跳过）
   - 自动写入 `config.json`
   - 自动把 Stop / Notification hooks 写入 `~/.claude/settings.json`（已存在则跳过，不重复）
   - 询问是否立即发送测试通知

3. 完成后重启 Claude Code 即可生效。

> `init` 会自适应 `notify.exe` 的实际路径，换机器/移动目录不失效。修改配置可重新运行 `init`，或直接编辑 `config.json`。

### 方法二：Claude Code 插件

> **注意**：如果遇到 SSH 密钥问题，可先运行以下命令配置 git 使用 HTTPS：
> ```bash
> git config --global url."https://github.com/".insteadOf "git@github.com:"
> ```

1. 添加插件市场：
   ```
   /plugin marketplace add liuzhicheng1775/claude-notifications-win
   ```

2. 安装插件：
   ```
   /plugin install claude-notifications-win@claude-notifications-win
   ```

3. 重启 Claude Code

### 方法三：手动配置（不使用向导）

如果不走 `init` 向导，也可手动配置：

1. 在 `notify.exe` 同目录下创建 `config.json`（可选，默认全部开启）：
   ```json
   {
     "notifications": {
       "stop": { "enabled": true },
       "permission": { "enabled": true }
     }
   }
   ```
2. 在 `~/.claude/settings.json` 中配置 hooks：
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

## 命令一览

| 命令 | 作用 |
|------|------|
| `.\notify.exe init` | 交互式配置向导（飞书 + config.json + hooks 一键搞定） |
| `.\notify.exe test` | 向所有已启用渠道发测试消息，逐个报告 ✓/✗ |
| `notify.exe stop` | 触发任务完成通知（由 Claude Code Stop hook 调用） |
| `notify.exe permission` | 触发权限审批通知（由 Claude Code Notification hook 调用） |
| `.\notify.exe version` | 显示版本号 |

## 使用方法

安装并配置后，以下通知会自动弹出：

| 通知类型 | 触发时机 |
|---------|---------|
| 任务完成 | Claude Code 完成一个任务 |
| 需要权限 | Claude Code 需要您授权操作 |

验证配置是否生效：

```powershell
.\notify.exe test
```

会向 Windows 通知和已启用的飞书分别发一条测试消息，并汇总 `N/M 渠道成功`。

## 飞书通知（可选）

> 推荐直接用 `.\notify.exe init` 向导配置，下面是手动配置说明。详细步骤见 [docs/feishu-notification.md](./docs/feishu-notification.md)。

除 Windows 气泡通知外，还支持将通知发送到飞书群。在飞书群创建自定义机器人，获取 webhook URL（若开启了签名校验，还需 secret），填入 `config.json`：

```json
{
  "notifications": {
    "feishu": {
      "enabled": true,
      "webhook": "https://open.feishu.cn/open-apis/bot/v2/hook/你的hook_id",
      "secret": "可选，机器人开启签名校验时填写"
    }
  }
}
```

`enabled: true` 且 `webhook` 非空时，通知会同时发送到 Windows 通知和飞书群。不配置 `feishu` 字段时仅使用 Windows 通知。

## 开发

### 从源码构建

```bash
cd src
go mod download
# 版本号从 git tag 自动获取（64 位）
GOARCH=amd64 go build -ldflags "-X main.version=$(git describe --tags --abbrev=0 | sed 's/^v//')" -o ../bin/notify.exe .
```

### 项目结构

```
claude-notifications-win/
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
├── .claude-plugin/
│   ├── plugin.json          # 插件配置
│   └── marketplace.json     # 市场元数据
└── README.md
```

## 开源协议

MIT License
