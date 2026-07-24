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
// title 和 message 同时非空时合并为 "title\nmessage"，否则发送非空的那个。
func (n *FeishuNotifier) Send(title, message string) error {
	text := message
	if title != "" && message != "" {
		text = title + "\n" + message
	} else if title != "" {
		text = title
	}

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

// sign 计算飞书加签：base64(hmac_sha256(key=timestamp+"\n"+secret, message=""))。
func (n *FeishuNotifier) sign(timestamp string) (string, error) {
	stringToSign := timestamp + "\n" + n.secret
	h := hmac.New(sha256.New, []byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
