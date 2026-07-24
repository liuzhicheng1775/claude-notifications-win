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

## 二、配置 config.json

在 `notify.exe` 同目录下创建 `config.json`（已有则编辑）。

### 仅飞书（不开加签）

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

### 飞书 + 加签（推荐）

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

### 字段说明

| 字段 | 说明 |
|------|------|
| `feishu.enabled` | 飞书开关，`false` 或不填则不发飞书 |
| `feishu.webhook` | 机器人 webhook URL，**必填**才能启用 |
| `feishu.secret` | 加签密钥，机器人没开加签就留空 `""` |

> 不配置整个 `feishu` 字段时，只弹 Windows 气泡通知，行为和升级前一样。

---

## 三、配置 Claude Code hooks

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

### 手动测试（立即看效果）

打开 cmd 或 PowerShell，执行：

```bash
notify.exe stop
```

- Windows 应弹出气泡通知
- 飞书群应收到机器人消息：「Claude Code / 任务已完成」

测试权限通知：

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

查看具体错误：在终端直接运行 `notify.exe stop`，错误会打印到 stderr。

---

## 六、配置文件位置

`config.json` 按以下顺序查找（找到第一个就用）：

1. `%LOCALAPPDATA%\claude-notifications-win\config.json`
2. `%APPDATA%\claude-notifications-win\config.json`
3. `notify.exe` 同目录下的 `config.json`

推荐放 `notify.exe` 同目录（第 3 个位置），最直观。
