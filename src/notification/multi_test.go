package notification

import (
	"errors"
	"strings"
	"testing"
)

// fakeNotifier 用于测试 MultiNotifier 的调用与错误聚合。
type fakeNotifier struct {
	sendErr error
	called  bool
	title   string
	message string
}

func (f *fakeNotifier) Send(title, message string) error {
	f.called = true
	f.title = title
	f.message = message
	return f.sendErr
}

func TestMultiNotifier_AllSuccess(t *testing.T) {
	n1 := &fakeNotifier{}
	n2 := &fakeNotifier{}
	m := NewMultiNotifier(n1, n2)

	err := m.Send("标题", "内容")
	if err != nil {
		t.Errorf("全部成功期望 nil, 实际 %v", err)
	}
	if !n1.called || !n2.called {
		t.Error("两个 notifier 都应被调用")
	}
	if n1.title != "标题" || n1.message != "内容" {
		t.Errorf("参数传递错误: title=%q message=%q", n1.title, n1.message)
	}
}

func TestMultiNotifier_PartialFailure(t *testing.T) {
	n1 := &fakeNotifier{sendErr: errors.New("渠道1失败")}
	n2 := &fakeNotifier{}
	m := NewMultiNotifier(n1, n2)

	err := m.Send("", "")
	if err != nil {
		t.Errorf("任一成功应返回 nil, 实际 %v", err)
	}
	if !n1.called || !n2.called {
		t.Error("即使部分失败，两个都应被调用")
	}
}

func TestMultiNotifier_AllFail(t *testing.T) {
	n1 := &fakeNotifier{sendErr: errors.New("err1")}
	n2 := &fakeNotifier{sendErr: errors.New("err2")}
	m := NewMultiNotifier(n1, n2)

	err := m.Send("", "")
	if err == nil {
		t.Fatal("全部失败应返回错误")
	}
	if !strings.Contains(err.Error(), "err1") || !strings.Contains(err.Error(), "err2") {
		t.Errorf("错误应包含所有子错误, 实际: %v", err)
	}
}

func TestMultiNotifier_Empty(t *testing.T) {
	m := NewMultiNotifier()
	err := m.Send("", "")
	if err == nil {
		t.Fatal("空 notifier 列表应返回错误（无成功渠道）")
	}
}
