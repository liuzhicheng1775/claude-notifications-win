package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"claude-notifications-win/src/config"
	"claude-notifications-win/src/notification"
)

// withArgs 临时替换 os.Args 并在测试结束后恢复。
// 必须用 t.Cleanup（而非 defer）：defer 会在 withArgs 返回时立即恢复，
// 导致后续调用读到 go test 注入的参数（如 -test.timeout）。
// t.Cleanup 在 Go 1.14 引入，在测试完成时执行。
func withArgs(t *testing.T, args []string) {
	t.Helper()
	saved := os.Args
	os.Args = args
	t.Cleanup(func() { os.Args = saved })
}

// writeTranscript 创建临时 jsonl transcript 文件并返回路径。
// 测试结束自动清理（t.Cleanup, Go 1.14+）。
func writeTranscript(t *testing.T, lines []string) string {
	t.Helper()
	dir, err := ioutil.TempDir("", "transcript-test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	p := filepath.Join(dir, "session.jsonl")
	content := strings.Join(lines, "\n")
	if err := ioutil.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("写入 transcript 失败: %v", err)
	}
	return p
}

// --- extractSessionTitle 测试 ---

func TestExtractSessionTitle_EmptyPath(t *testing.T) {
	if got := extractSessionTitle(""); got != "" {
		t.Errorf("空 path 应返回空字符串, 实际 %q", got)
	}
}

func TestExtractSessionTitle_FileNotExist(t *testing.T) {
	got := extractSessionTitle(filepath.Join(os.TempDir(), "nonexistent-transcript-xxx.jsonl"))
	if got != "" {
		t.Errorf("文件不存在应返回空字符串, 实际 %q", got)
	}
}

func TestExtractSessionTitle_SkipsSnapshotLine(t *testing.T) {
	// 第一行是 file-history-snapshot，应跳过；第二条 user 是真正标题
	snapshot := `{"type":"file-history-snapshot","messageId":"abc","snapshot":{"trackedFileBackups":{}}}`
	user := `{"type":"user","message":{"role":"user","content":"实现飞书通知带会话信息"}}`
	p := writeTranscript(t, []string{snapshot, user})

	got := extractSessionTitle(p)
	want := "实现飞书通知带会话信息"
	if got != want {
		t.Errorf("期望 %q, 实际 %q", want, got)
	}
}

func TestExtractSessionTitle_StringContent(t *testing.T) {
	user := `{"type":"user","message":{"role":"user","content":"这是一条 string 形式的用户消息"}}`
	p := writeTranscript(t, []string{user})

	got := extractSessionTitle(p)
	want := "这是一条 string 形式的用户消息"
	if got != want {
		t.Errorf("期望 %q, 实际 %q", want, got)
	}
}

func TestExtractSessionTitle_ArrayContent(t *testing.T) {
	// content 是 array of blocks，取第一个 type:text 的 text
	user := `{"type":"user","message":{"role":"user","content":[{"type":"text","text":"array 形式的文本"},{"type":"text","text":"第二个 block"}]}}`
	p := writeTranscript(t, []string{user})

	got := extractSessionTitle(p)
	want := "array 形式的文本"
	if got != want {
		t.Errorf("期望 %q, 实际 %q", want, got)
	}
}

func TestExtractSessionTitle_SkipsToolResult(t *testing.T) {
	// 第一条 user 是 tool_result（不是真正用户输入），应跳过
	toolResult := `{"type":"user","message":{"role":"user","content":[{"tool_use_id":"call_xxx","type":"tool_result","content":"命令输出"}]}}`
	realUser := `{"type":"user","message":{"role":"user","content":"真正的用户输入"}}`
	p := writeTranscript(t, []string{toolResult, realUser})

	got := extractSessionTitle(p)
	want := "真正的用户输入"
	if got != want {
		t.Errorf("应跳过 tool_result 取真正 user, 期望 %q, 实际 %q", want, got)
	}
}

func TestExtractSessionTitle_TruncatesLongTitle(t *testing.T) {
	// 超过 50 个 rune 的标题应截断 + "..."
	longText := strings.Repeat("啊", 80) // 80 个中文字符
	user := `{"type":"user","message":{"role":"user","content":"` + longText + `"}}`
	p := writeTranscript(t, []string{user})

	got := extractSessionTitle(p)
	want := strings.Repeat("啊", 50) + "..."
	if got != want {
		if len([]rune(got)) > 51 {
			t.Errorf("应截到 50 rune + ..., 实际长度 %d: %q", len([]rune(got)), got)
		} else {
			t.Errorf("期望 %q..., 实际 %q", want[:30], got)
		}
	}
}

func TestExtractSessionTitle_CompressesWhitespace(t *testing.T) {
	// 含换行和多空白应压缩为单空格
	user := `{"type":"user","message":{"role":"user","content":"第一行\n\n\n第二行   多空格"}}`
	p := writeTranscript(t, []string{user})

	got := extractSessionTitle(p)
	want := "第一行 第二行 多空格"
	if got != want {
		t.Errorf("期望 %q, 实际 %q", want, got)
	}
}

func TestExtractSessionTitle_NoUserMessage(t *testing.T) {
	// 全是 assistant 消息，找不到 user，返回空
	assistant := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"hi"}]}}`
	p := writeTranscript(t, []string{assistant, assistant})

	got := extractSessionTitle(p)
	if got != "" {
		t.Errorf("无 user 消息应返回空, 实际 %q", got)
	}
}

func TestExtractSessionTitle_SkipsAssistantAndFindsUser(t *testing.T) {
	assistant := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"你好"}]}}`
	user := `{"type":"user","message":{"role":"user","content":"用户的问题"}}`
	p := writeTranscript(t, []string{assistant, assistant, user})

	got := extractSessionTitle(p)
	want := "用户的问题"
	if got != want {
		t.Errorf("期望 %q, 实际 %q", want, got)
	}
}

// --- truncateTitle 测试 ---

func TestTruncateTitle(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"空字符串", "", ""},
		{"纯空白", "   \n\t  ", ""},
		{"短文本不截断", "短标题", "短标题"},
		{"压缩多空白", "a   b\n\nc", "a b c"},
		{"正好 50 rune 不截断", strings.Repeat("a", 50), strings.Repeat("a", 50)},
		{"51 rune 截断加省略号", strings.Repeat("a", 51), strings.Repeat("a", 50) + "..."},
		{"中文按 rune 截断", strings.Repeat("中", 60), strings.Repeat("中", 50) + "..."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := truncateTitle(c.in)
			if got != c.want {
				t.Errorf("truncateTitle(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// --- extractUserText 测试 ---

func TestExtractUserText(t *testing.T) {
	cases := []struct {
		name    string
		json    string
		want    string
	}{
		{
			name: "string content",
			json: `{"role":"user","content":"纯文本"}`,
			want: "纯文本",
		},
		{
			name: "array content 取第一个 text",
			json: `{"role":"user","content":[{"type":"text","text":"第一段"},{"type":"text","text":"第二段"}]}`,
			want: "第一段",
		},
		{
			name: "跳过 tool_result 取 text",
			json: `{"role":"user","content":[{"type":"tool_result","content":"输出"},{"type":"text","text":"真正输入"}]}`,
			want: "真正输入",
		},
		{
			name: "全是 tool_result 返回空",
			json: `{"role":"user","content":[{"type":"tool_result","content":"输出"}]}`,
			want: "",
		},
		{
			name: "空 raw 返回空",
			json: ``,
			want: "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := extractUserText(json.RawMessage(c.json))
			if got != c.want {
				t.Errorf("extractUserText(%s) = %q, want %q", c.json, got, c.want)
			}
		})
	}
}

// 防止 lint 警告：保留 withArgs 供未来测试使用
var _ = withArgs

func TestBuildNotifier_WindowsOnly(t *testing.T) {
	cfg := &config.Config{}
	n := BuildNotifier(cfg)
	if _, ok := n.(*notification.WindowsNotifier); !ok {
		t.Errorf("飞书未启用时期望 *WindowsNotifier, 实际 %T", n)
	}
}

func TestBuildNotifier_WithFeishu(t *testing.T) {
	cfg := &config.Config{}
	cfg.Notifications.Feishu.Enabled = true
	cfg.Notifications.Feishu.Webhook = "https://open.feishu.cn/open-apis/bot/v2/hook/test"
	n := BuildNotifier(cfg)
	if _, ok := n.(*notification.MultiNotifier); !ok {
		t.Errorf("飞书启用时期望 *MultiNotifier, 实际 %T", n)
	}
}

func TestBuildNotifier_FeishuEnabledButNoWebhook(t *testing.T) {
	cfg := &config.Config{}
	cfg.Notifications.Feishu.Enabled = true
	// webhook 为空，不应启用飞书
	n := BuildNotifier(cfg)
	if _, ok := n.(*notification.WindowsNotifier); !ok {
		t.Errorf("webhook 空时应只返回 Windows, 实际 %T", n)
	}
}

func TestBuildNotifier_NilConfig(t *testing.T) {
	n := BuildNotifier(nil)
	if _, ok := n.(*notification.WindowsNotifier); !ok {
		t.Errorf("nil cfg 应只返回 Windows, 实际 %T", n)
	}
}
