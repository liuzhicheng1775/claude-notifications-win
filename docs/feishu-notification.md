# 飞书通知使用说明

本文档介绍如何为 claude-notifications-win 配置飞书群通知。配置后，Claude Code 的任务完成和权限审批通知会同时发送到 Windows 气泡通知和飞书群。

---

## 一、创建飞书自定义机器人

1. 打开飞书 PC 端，进入你要接收通知的**群聊**
2. 点右上角 **`···`** -> **设置** -> **群机器人**
3. 点 **添加机器人** -> 选 **自定义机器人**
4. 填机器人名称（如「Claude Code 通知」），可选头像
5. **安全设置**（重要）：
   - 建议勾选 **加签**（最安全）
   - 也可选「自定义关键词」（需包含指定词才发送，不推荐，易漏发）
6. 点 **完成**，页面会显示：
   - **webhook URL**：`https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxx`
   - **secret**（加签模式才有）：一串密钥
7. **复制这两个值**，关掉页面就看不到了

---

## 二、一键配置（推荐）

拿到 webhook 和 secret 后，直接运行向导（PowerShell / cmd 中需要 `.\` 前缀）：

```powershell
.\notify.exe init
```

向导交互流程：

1. `是否配置飞书通知？(y/n)` → 输入 `y`
2. `请输入飞书 webhook URL:` → 粘贴 webhook
3. `请输入加签 secret（可选，回车跳过）:` → 粘贴 secret（没开加签就回车）
4. 向导自动写 `config.json` + 自动配 `~/.claude/settings.json` hooks
5. `是否立即发送测试通知？(y/n)` → 输入 `y` 验证

看到 `2/2 渠道成功`（Windows + 飞书都 ✓）即配置完成。重启 Claude Code 生效。

> `init` 会自适应 `notify.exe` 的实际路径，换机器/移动目录不失效。重复运行不会产生重复 hooks。

---

## 三、手动配置（不走向导）

### 1. 写 config.json

在 `notify.exe` 同目录下创建 `config.json`（已有则编辑）。

**仅飞书（不开加签）**

```json
{
  "notifications": {
    "stop": { "enabled": true },
    "permission": { "enabled": true },
    "feishu": {
      "enabled": true,
      "webhook": "https://open.feishu.cn/open-apis/bot/v2/hook/你的hook_id",
      "secret": ""
    }
  }
}
```

**飞书 + 加签（推荐）**

```json
{
  "notifications": {
    "stop": { "enabled": true },
    "permission": { "enabled": true },
    "feishu": {
      "enabled": true,
      "webhook": "https://open.feishu.cn/open-apis/bot/v2/hook/你的hook_id",
      "secret": "你复制的secret"
    }
  }
}
```

**字段说明**

| 字段 | 说明 |
|------|------|
| `feishu.enabled` | 飞书开关，`false` 或不填则不发飞书 |
| `feishu.webhook` | 机器人 webhook URL，**必填**才能启用 |
| `feishu.secret` | 加签密钥，机器人没开加签就留空 `""` |

> 不配置整个 `feishu` 字段时，只弹 Windows 气泡通知，行为和升级前一样。

### 2. 配置 Claude Code hooks

编辑 `~/.claude/settings.json`（Windows 路径：`C:\Users\你的用户名\.claude\settings.json`）：

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

把 `C:/path/to/notify.exe` 换成你的实际路径（**用正斜杠 `/`**，不要用反斜杠）。

---

## 四、验证

### 用 test 命令一键测试（推荐）

```powershell
.\notify.exe test
```

会向 Windows 通知和已启用的飞书分别发一条测试消息，并汇总：

```
通知渠道测试：
  ✓ Windows
  ✓ 飞书

2/2 渠道成功
```

### 手动测试单个 hook

```bash
notify.exe stop
```

- Windows 应弹出气泡通知
- 飞书群应收到机器人消息：「Claude Code / 任务已完成」

```bash
notify.exe permission
```

飞书应收到：「Claude Code - 需要授权 / 请授权以继续操作」

### 实际使用

重启 Claude Code，正常使用。任务完成或请求授权时会自动通知。

---

## 五、故障排查

| 现象 | 排查 |
|------|------|
| 飞书没收到，但 Windows 弹了 | 检查 `config.json` 的 `feishu.enabled` 是否 `true`，`webhook` 是否完整 |
| 飞书报 `sign match fail` | 机器人开了加签但 `secret` 没填，或填错。重新复制 secret |
| 飞书报 `19021` 等错误码 | webhook URL 错误，或机器人被删了。重新创建机器人 |
| 两个都没收到 | 检查 `notify.exe` 路径和 `settings.json` 配置 |
| Windows 弹了但飞书超时 | 网络问题（需能访问 `open.feishu.cn`），或机器人设置了 IP 白名单 |
| 任务完成不推送 | `~/.claude/settings.json` 缺少 Stop hook，重新跑 `.\notify.exe init` |

查看具体错误：在终端直接运行 `.\notify.exe stop` 或 `.\notify.exe test`，错误会打印到 stderr。

---

## 六、配置文件位置

`config.json` 按以下顺序查找（找到第一个就用）：

1. `%LOCALAPPDATA%\claude-notifications-win\config.json`（`init` 向导默认写这里）
2. `%APPDATA%\claude-notifications-win\config.json`
3. `notify.exe` 同目录下的 `config.json`

`init` 向导默认写到第 1 个位置。手动配置可放任一位置，第 3 个最直观。
