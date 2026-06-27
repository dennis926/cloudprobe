package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// QQChannel QQ通知渠道（通过 go-cqhttp / NapCatQQ HTTP API）
type QQChannel struct {
	APIBaseURL string // 如 http://127.0.0.1:3000
	AccessToken string
	DefaultQQ  string // 默认接收QQ号
}

func (q *QQChannel) Name() string {
	return "qq"
}

func (q *QQChannel) Send(ctx context.Context, title, content string, recipient string) error {
	qq := q.DefaultQQ
	if recipient != "" {
		qq = recipient
	}
	if qq == "" {
		return fmt.Errorf("qq recipient is empty")
	}

	// go-cqhttp 发送私聊消息 API
	url := fmt.Sprintf("%s/send_private_msg", q.APIBaseURL)
	payload := map[string]interface{}{
		"user_id": qq,
		"message": fmt.Sprintf("%s\n\n%s", title, content),
	}

	return q.postAPI(ctx, url, payload)
}

// SendGroup 发送群消息
func (q *QQChannel) SendGroup(ctx context.Context, groupID, title, content string) error {
	url := fmt.Sprintf("%s/send_group_msg", q.APIBaseURL)
	payload := map[string]interface{}{
		"group_id": groupID,
		"message":  fmt.Sprintf("%s\n\n%s", title, content),
	}
	return q.postAPI(ctx, url, payload)
}

func (q *QQChannel) postAPI(ctx context.Context, url string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if q.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+q.AccessToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("qq notification failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Status  string `json:"status"`
		RetCode int    `json:"retcode"`
		Msg     string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}
	if result.Status != "ok" && result.RetCode != 0 {
		return fmt.Errorf("qq notification failed: %s", result.Msg)
	}
	return nil
}

func (q *QQChannel) Test(ctx context.Context) error {
	return q.Send(ctx, "CloudProbe QQ测试", "这是一则测试通知，如果收到说明配置正确。", q.DefaultQQ)
}
