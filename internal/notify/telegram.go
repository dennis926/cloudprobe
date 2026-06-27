package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TelegramChannel Telegram通知渠道
type TelegramChannel struct {
	BotToken string
	ChatID   string // 默认接收人，可以是频道ID或用户ID
}

func (t *TelegramChannel) Name() string {
	return "telegram"
}

func (t *TelegramChannel) Send(ctx context.Context, title, content string, recipient string) error {
	chatID := t.ChatID
	if recipient != "" {
		chatID = recipient
	}
	if chatID == "" {
		return fmt.Errorf("telegram chat_id is empty")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.BotToken)

	// 合并标题和内容
	text := fmt.Sprintf("*%s*\n\n%s", escapeMarkdown(title), content)

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "MarkdownV2",
		"disable_web_page_preview": true,
	}

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

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram notification failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}
	if !result.Ok {
		return fmt.Errorf("telegram notification failed: %s", result.Description)
	}
	return nil
}

// escapeMarkdown 转义MarkdownV2特殊字符
func escapeMarkdown(text string) string {
	chars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, ch := range chars {
		text = replaceAll(text, ch, "\\"+ch)
	}
	return text
}

func replaceAll(s, old, new string) string {
	// 简单替换
	result := ""
	for _, r := range s {
		if string(r) == old {
			result += new
		} else {
			result += string(r)
		}
	}
	return result
}

func (t *TelegramChannel) Test(ctx context.Context) error {
	return t.Send(ctx, "CloudProbe Telegram测试", "这是一则测试通知，如果收到说明配置正确。", "")
}
