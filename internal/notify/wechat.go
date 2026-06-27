package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WechatChannel 微信通知渠道（ServerChan / PushPlus）
type WechatChannel struct {
	Provider string // serverchan | pushplus
	Token    string
}

func (w *WechatChannel) Name() string {
	return "wechat"
}

func (w *WechatChannel) Send(ctx context.Context, title, content string, recipient string) error {
	switch w.Provider {
	case "serverchan":
		return w.sendServerChan(title, content)
	case "pushplus":
		return w.sendPushPlus(title, content)
	default:
		return fmt.Errorf("unsupported wechat provider: %s", w.Provider)
	}
}

// ServerChan (Server酱) - https://sct.ftqq.com/
func (w *WechatChannel) sendServerChan(title, content string) error {
	url := fmt.Sprintf("https://sctapi.ftqq.com/%s.send", w.Token)
	payload := map[string]string{
		"title": title,
		"desp":  content,
	}
	return w.postJSON(url, payload)
}

// PushPlus - http://www.pushplus.plus/
func (w *WechatChannel) sendPushPlus(title, content string) error {
	url := "https://www.pushplus.plus/send"
	payload := map[string]interface{}{
		"token":    w.Token,
		"title":    title,
		"content":  content,
		"template": "html",
	}
	return w.postJSON(url, payload)
}

func (w *WechatChannel) postJSON(url string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wechat notification failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil // 忽略解析错误，HTTP 200 即视为成功
	}
	if result.Code != 0 {
		return fmt.Errorf("wechat notification failed: %s", result.Message)
	}
	return nil
}

func (w *WechatChannel) Test(ctx context.Context) error {
	return w.Send(ctx, "CloudProbe 微信测试", "这是一则测试通知，如果收到说明配置正确。", "")
}
