package notification

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// FeishuNotifier 通过飞书自定义机器人 webhook 发送通知。
// secret 非空时启用加签校验。
type FeishuNotifier struct {
	webhook string
	secret  string
	client  *http.Client
}

// NewFeishuNotifier 创建飞书通知器。secret 可为空（不加签）。
func NewFeishuNotifier(webhook, secret string) *FeishuNotifier {
	return &FeishuNotifier{
		webhook: webhook,
		secret:  secret,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

type feishuRequest struct {
	MsgType   string        `json:"msg_type"`
	Content   feishuContent `json:"content"`
	Timestamp string        `json:"timestamp,omitempty"`
	Sign      string        `json:"sign,omitempty"`
}

type feishuContent struct {
	Text string `json:"text"`
}

type feishuResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// Send 向飞书 webhook 发送 text 消息。
// 消息文本由 buildFeishuText 构造，会附带会话 ID（前 8 位）
// 和会话标题（前 50 字符）- 仅当 SessionID/SessionTitle 非空时追加。
func (n *FeishuNotifier) Send(noti Notification) error {
	text := buildFeishuText(noti)

	req := feishuRequest{
		MsgType: "text",
		Content: feishuContent{Text: text},
	}

	if n.secret != "" {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		sign, err := n.sign(timestamp)
		if err != nil {
			return fmt.Errorf("feishu sign: %w", err)
		}
		req.Timestamp = timestamp
		req.Sign = sign
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("feishu marshal: %w", err)
	}

	resp, err := n.client.Post(n.webhook, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("feishu request: %w", err)
	}
	defer resp.Body.Close()

	var fr feishuResponse
	if err := json.NewDecoder(resp.Body).Decode(&fr); err != nil {
		return fmt.Errorf("feishu decode: %w", err)
	}

	if fr.Code != 0 {
		return fmt.Errorf("feishu send failed: code=%d msg=%s", fr.Code, fr.Msg)
	}
	return nil
}

// buildFeishuText 构造飞书消息文本。
// 格式:
//
//	<title>
//	<message>
//
//	会话: <标题前 50 字符>...
//	ID: <session_id 前 8 位>
//
// SessionID 和 SessionTitle 都为空时不追加会话信息块，
// test 命令发的测试通知不会带会话信息。
func buildFeishuText(n Notification) string {
	var sb strings.Builder
	if n.Title != "" {
		sb.WriteString(n.Title)
	}
	if n.Message != "" {
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(n.Message)
	}
	// 仅当存在会话上下文时追加
	if n.SessionID != "" || n.SessionTitle != "" {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		if n.SessionTitle != "" {
			sb.WriteString("会话: " + n.SessionTitle + "\n")
		}
		if n.SessionID != "" {
			sid := n.SessionID
			// UUID 形如 230c9a65-b2f7-4848-bc52-464070178d6e，取前 8 位足够区分
			if len(sid) > 8 {
				sid = sid[:8]
			}
			sb.WriteString("ID: " + sid)
		}
	}
	return sb.String()
}

// sign 计算飞书加签：base64(hmac_sha256(key=timestamp+"\n"+secret, message=""))。
func (n *FeishuNotifier) sign(timestamp string) (string, error) {
	stringToSign := timestamp + "\n" + n.secret
	h := hmac.New(sha256.New, []byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
