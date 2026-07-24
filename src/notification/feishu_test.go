package notification

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// feishuTestServer 创建模拟飞书 webhook 的测试服务器。
// 返回的 *string 记录最后收到的请求体。
func feishuTestServer(t *testing.T, status int, respBody string) (*httptest.Server, *string) {
	t.Helper()
	var lastBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		lastBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(respBody))
	}))
	return server, &lastBody
}

func TestFeishuNotifier_SendBasic(t *testing.T) {
	server, lastBody := feishuTestServer(t, http.StatusOK, `{"code":0,"msg":"success"}`)
	defer server.Close()

	n := NewFeishuNotifier(server.URL, "")
	if err := n.Send("标题", "内容"); err != nil {
		t.Fatalf("Send 返回意外错误: %v", err)
	}

	var req feishuRequest
	if err := json.Unmarshal([]byte(*lastBody), &req); err != nil {
		t.Fatalf("解析请求体失败: %v", err)
	}
	if req.MsgType != "text" {
		t.Errorf("msg_type 期望 text, 实际 %q", req.MsgType)
	}
	if req.Content.Text != "标题\n内容" {
		t.Errorf("content.text 期望 %q, 实际 %q", "标题\n内容", req.Content.Text)
	}
	if req.Sign != "" || req.Timestamp != "" {
		t.Errorf("无 secret 时不应包含 sign/timestamp")
	}
}

func TestFeishuNotifier_SendWithSign(t *testing.T) {
	server, lastBody := feishuTestServer(t, http.StatusOK, `{"code":0,"msg":"success"}`)
	defer server.Close()

	secret := "mysecret"
	n := NewFeishuNotifier(server.URL, secret)
	if err := n.Send("标题", "内容"); err != nil {
		t.Fatalf("Send 返回意外错误: %v", err)
	}

	var req feishuRequest
	if err := json.Unmarshal([]byte(*lastBody), &req); err != nil {
		t.Fatalf("解析请求体失败: %v", err)
	}
	if req.Timestamp == "" {
		t.Error("加签时 timestamp 不应为空")
	}
	if req.Sign == "" {
		t.Error("加签时 sign 不应为空")
	}
	// 用相同算法重算 sign 对比
	stringToSign := req.Timestamp + "\n" + secret
	h := hmac.New(sha256.New, []byte(stringToSign))
	expected := base64.StdEncoding.EncodeToString(h.Sum(nil))
	if req.Sign != expected {
		t.Errorf("sign 不匹配: 期望 %q, 实际 %q", expected, req.Sign)
	}
}

func TestFeishuNotifier_TitleOrMessageEmpty(t *testing.T) {
	server, lastBody := feishuTestServer(t, http.StatusOK, `{"code":0,"msg":"success"}`)
	defer server.Close()

	n := NewFeishuNotifier(server.URL, "")

	if err := n.Send("只有标题", ""); err != nil {
		t.Fatalf("Send 错误: %v", err)
	}
	var req feishuRequest
	json.Unmarshal([]byte(*lastBody), &req)
	if req.Content.Text != "只有标题" {
		t.Errorf("message 空时 text 应为 title, 实际 %q", req.Content.Text)
	}

	if err := n.Send("", "只有内容"); err != nil {
		t.Fatalf("Send 错误: %v", err)
	}
	json.Unmarshal([]byte(*lastBody), &req)
	if req.Content.Text != "只有内容" {
		t.Errorf("title 空时 text 应为 message, 实际 %q", req.Content.Text)
	}
}

func TestFeishuNotifier_NonZeroCode(t *testing.T) {
	server, _ := feishuTestServer(t, http.StatusOK, `{"code":19021,"msg":"sign match fail"}`)
	defer server.Close()

	n := NewFeishuNotifier(server.URL, "")
	err := n.Send("标题", "内容")
	if err == nil {
		t.Fatal("code!=0 期望返回错误")
	}
	if !strings.Contains(err.Error(), "19021") {
		t.Errorf("错误应包含 code 19021, 实际: %v", err)
	}
	if !strings.Contains(err.Error(), "sign match fail") {
		t.Errorf("错误应包含 msg, 实际: %v", err)
	}
}

func TestFeishuNotifier_NetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	addr := server.URL
	server.Close() // 关闭使 URL 不可达

	n := NewFeishuNotifier(addr, "")
	err := n.Send("标题", "内容")
	if err == nil {
		t.Fatal("网络不可达应返回错误")
	}
	if !strings.Contains(err.Error(), "feishu request") {
		t.Errorf("错误应包含 feishu request, 实际: %v", err)
	}
}
