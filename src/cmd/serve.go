package cmd

import (
	"bufio"
	"claude-notifications-win/src/config"
	"claude-notifications-win/src/hooks"
	"claude-notifications-win/src/notification"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// ClaudeHookPayload 对应 Claude Code 通过 stdin 传给 hook 的 JSON。
// 字段名来自官方文档（https://docs.claude.com/en/docs/claude-code/hooks）。
// Stop hook 传 session_id/transcript_path/cwd/hook_event_name/stop_hook_active；
// Notification hook 额外传 message/title/notification_type。
type ClaudeHookPayload struct {
	HookEventName    string `json:"hook_event_name"`
	SessionID        string `json:"session_id"`
	TranscriptPath   string `json:"transcript_path"`
	Cwd              string `json:"cwd,omitempty"`
	StopHookActive   bool   `json:"stop_hook_active,omitempty"`
	Message          string `json:"message,omitempty"`
	Title            string `json:"title,omitempty"`
	NotificationType string `json:"notification_type,omitempty"`
}

// HandleStopHook 处理 Stop hook：解析 stdin payload，提取会话信息，发通知。
// 解析失败或 stdin 为空时仍发默认通知（不阻塞 Claude Code 流程）。
func HandleStopHook() error {
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		cfg = &config.Config{}
	}

	notifier := BuildNotifier(cfg)
	handler := hooks.NewStopHandler(notifier, cfg)

	payload, err := readPayload()
	if err != nil {
		// 解析失败仍发通知，用默认消息
		return handler.Handle(notification.Notification{
			Title:   "Claude Code",
			Message: "任务已完成",
		})
	}

	// 会话标题从 transcript 第一条 user message 推导，
	// 仅在会话信息块里展示（不重复填到 message，避免飞书消息重复显示）
	var sessionTitle string
	if payload.TranscriptPath != "" {
		sessionTitle = extractSessionTitle(payload.TranscriptPath)
	}

	return handler.Handle(notification.Notification{
		Title:        "Claude Code",
		Message:      "任务已完成",
		SessionID:    payload.SessionID,
		SessionTitle: sessionTitle,
	})
}

// HandlePermissionHook 处理 Notification hook（permission_prompt）。
// 优先用 payload.Message 作为提示文本，fallback 到默认。
func HandlePermissionHook() error {
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		cfg = &config.Config{}
	}

	notifier := BuildNotifier(cfg)
	handler := hooks.NewPermissionHandler(notifier, cfg)

	payload, err := readPayload()
	if err != nil {
		return handler.Handle(notification.Notification{
			Title:   "Claude Code - 需要授权",
			Message: "请授权以继续操作",
		})
	}

	message := payload.Message
	if message == "" {
		message = "请授权以继续操作"
	}

	var sessionTitle string
	if payload.TranscriptPath != "" {
		sessionTitle = extractSessionTitle(payload.TranscriptPath)
	}

	return handler.Handle(notification.Notification{
		Title:        "Claude Code - 需要授权",
		Message:      message,
		SessionID:    payload.SessionID,
		SessionTitle: sessionTitle,
	})
}

// readPayload 从 stdin 读取 Claude Code hook JSON payload。
// stdin 为空（终端直跑）时返回空 payload 与 nil error。
// 非 JSON 输入返回错误（hook 配置异常时也能感知）。
func readPayload() (*ClaudeHookPayload, error) {
	stat, _ := os.Stdin.Stat()
	// 没有管道输入（终端直接运行）时返回空 payload
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return &ClaudeHookPayload{}, nil
	}

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("read stdin: %w", err)
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return &ClaudeHookPayload{}, nil
	}

	var payload ClaudeHookPayload
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return nil, fmt.Errorf("parse hook payload: %w", err)
	}
	return &payload, nil
}

// extractSessionTitle 从 transcript jsonl 文件提取会话标题。
// 找第一条 type:"user" 且非 tool_result 的消息，取文本前 50 字符 + "..."。
// 文件不存在/解析失败/找不到 user 消息 -> 返回空字符串（不阻塞通知）。
//
// transcript jsonl 每行是一个 JSON 对象，user message 的 content 可能是：
//   - string 形式：{"content": "文本"}
//   - array 形式：{"content": [{"type":"text","text":"..."}, {"type":"tool_result",...}]}
//
// 第一行通常是 file-history-snapshot，要跳过。
// tool_result block 不是真正的用户输入，也要跳过。
func extractSessionTitle(transcriptPath string) string {
	if transcriptPath == "" {
		return ""
	}
	f, err := os.Open(transcriptPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	// jsonl 单行可能含 base64 图片等，缓冲区扩到 10MB
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)

	linesScanned := 0
	for scanner.Scan() {
		linesScanned++
		// 限制扫描行数避免大文件
		if linesScanned > 100 {
			break
		}
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry struct {
			Type    string          `json:"type"`
			Message json.RawMessage `json:"message"`
		}
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		if entry.Type != "user" {
			continue
		}

		text := extractUserText(entry.Message)
		if text == "" {
			continue
		}
		return truncateTitle(text)
	}
	return ""
}

// extractUserText 从 user message 的 message 字段提取文本。
// message 字段结构: {"role":"user","content": <string | []block>}
// 跳过 tool_result block（不是真正用户输入）。
func extractUserText(messageRaw json.RawMessage) string {
	if len(messageRaw) == 0 {
		return ""
	}

	var msg struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(messageRaw, &msg); err != nil {
		return ""
	}

	// 尝试 string 形式
	var s string
	if err := json.Unmarshal(msg.Content, &s); err == nil {
		return strings.TrimSpace(s)
	}

	// 尝试 array of blocks 形式
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(msg.Content, &blocks); err == nil {
		for _, b := range blocks {
			// 跳过 tool_result，只取 text block
			if b.Type == "text" && strings.TrimSpace(b.Text) != "" {
				return strings.TrimSpace(b.Text)
			}
		}
	}
	return ""
}

// xmlTagRegexp 匹配 XML 风格标签，用于剥掉 slash command 注入的
// <command-message>...</command-message> 等标签，保留标签内文本。
var xmlTagRegexp = regexp.MustCompile(`<[^>]+>`)

// truncateTitle 清理并截断会话标题：
//  1. 剥掉 XML 标签（如 <command-name>/resume-session</command-name> -> /resume-session）
//  2. 压缩连续空白（含换行）为单空格
//  3. 截断到 50 个 rune，超出加 "..."
func truncateTitle(s string) string {
	s = strings.TrimSpace(s)
	// 剥 XML 标签，保留内容（slash command 注入的 <command-name> 等）
	s = xmlTagRegexp.ReplaceAllString(s, "")
	// 压缩连续空白（含换行）为单空格
	s = strings.Join(strings.Fields(s), " ")
	if s == "" {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= 50 {
		return s
	}
	return string(runes[:50]) + "..."
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
