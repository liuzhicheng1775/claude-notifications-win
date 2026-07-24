package cmd

import (
	"os"
	"reflect"
	"testing"
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

func TestExtractTaskName(t *testing.T) {
	cases := []struct {
		name    string
		subject string
		want    string
	}{
		{"空字符串", "", ""},
		{"带空格前缀的任务标记", "prefix - 任务：实际任务", "实际任务"},
		{"无空格前缀的任务标记", "prefix- 任务：实际任务", "实际任务"},
		{"仅任务标记", "任务：实际任务", "实际任务"},
		{"会话前缀取 ] 之后内容", "[会话:abc] 实际任务", "实际任务"},
		{"纯文本原样返回", "纯任务名", "纯任务名"},
		{"会话加任务标记优先取任务", "[会话:abc] - 任务：实际任务", "实际任务"},
		{"取最后一个任务标记", "任务：旧 - 任务：新", "新"},
		{"会话前缀 ] 后带空格", "[会话 abc]   带空格任务", "带空格任务"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := extractTaskName(c.subject)
			if got != c.want {
				t.Errorf("extractTaskName(%q) = %q, want %q", c.subject, got, c.want)
			}
		})
	}
}

func TestFilterGarbage(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"正常输入原样返回", "正常的任务信息", "正常的任务信息"},
		{"首行含 session 被过滤", "session id abc", ""},
		{"首行含 session 单词被过滤", "session 内容", ""},
		{"多行首行正常但含 session 原样返回", "第一行正常\n第二行 session 内容", "第一行正常\n第二行 session 内容"},
		{"多行 JSON 首行仅花括号被过滤", "{\n  \"session\": \"abc\"\n}", ""},
		{"JSON 不含 session 原样返回", "{\n  \"foo\": \"bar\"\n}", "{\n  \"foo\": \"bar\"\n}"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := filterGarbage(c.input)
			if got != c.want {
				t.Errorf("filterGarbage(%q) = %q, want %q", c.input, got, c.want)
			}
		})
	}
}

func TestGetFlag(t *testing.T) {
	cases := []struct {
		name string
		args []string
		flag string
		want string
	}{
		{"等号形式", []string{"prog", "--reason=hello"}, "--reason", "hello"},
		{"空格形式", []string{"prog", "--reason", "hello"}, "--reason", "hello"},
		{"无匹配返回空", []string{"prog", "--other=x"}, "--reason", ""},
		{"flag 末尾无值返回空", []string{"prog", "--reason"}, "--reason", ""},
		{"多参数中查找", []string{"prog", "--foo", "bar", "--reason=hi"}, "--reason", "hi"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			withArgs(t, c.args)
			got := getFlag(c.flag)
			if got != c.want {
				t.Errorf("getFlag(%q) = %q, want %q", c.flag, got, c.want)
			}
		})
	}
}

func TestGetAllArgs(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want []string
	}{
		{"仅程序名", []string{"prog"}, []string{}},
		{"程序名加命令", []string{"prog", "cmd"}, []string{}},
		{"程序名命令加参数", []string{"prog", "cmd", "a", "b"}, []string{"a", "b"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			withArgs(t, c.args)
			got := getAllArgs()
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("getAllArgs() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestExtractArg(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		prefix string
		want   string
	}{
		{"等号形式", []string{"prog", "cmd", "--reason=hello"}, "--reason", "hello"},
		{"无匹配返回空", []string{"prog", "cmd", "--other=x"}, "--reason", ""},
		{"无参数返回空", []string{"prog", "cmd"}, "--reason", ""},
		{"多参数中查找", []string{"prog", "cmd", "--foo=bar", "--reason=hi"}, "--reason", "hi"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			withArgs(t, c.args)
			got := extractArg(c.prefix)
			if got != c.want {
				t.Errorf("extractArg(%q) = %q, want %q", c.prefix, got, c.want)
			}
		})
	}
}
