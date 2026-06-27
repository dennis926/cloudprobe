package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// FeishuChannel 飞书通知渠道
type FeishuChannel struct {
	WebhookURL string
}

func (f *FeishuChannel) Name() string {
	return "feishu"
}

func (f *FeishuChannel) Send(ctx context.Context, title, content string, recipient string) error {
	if f.WebhookURL == "" {
		return fmt.Errorf("feishu webhook url is empty")
	}

	// 飞书机器人消息格式 - 富文本卡片
	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": title,
				},
				"template": "red",
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"tag":     "lark_md",
						"content": content,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("feishu notification failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}
	if result.Code != 0 {
		return fmt.Errorf("feishu notification failed: %s", result.Msg)
	}
	return nil
}

func (f *FeishuChannel) Test(ctx context.Context) error {
	return f.Send(ctx, "CloudProbe 飞书测试", "这是一则测试通知，如果收到说明配置正确。", "")
}
