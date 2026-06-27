package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"cloudprobe/internal/config"
)

// Client 3x-ui API 代理客户端
type Client struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewClient 创建3x-ui代理客户端
func NewClient(cfg *config.XUIConfig) *Client {
	return &Client{
		baseURL: cfg.PanelURL,
		token:   cfg.APIToken,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// request 发送HTTP请求到3x-ui面板
func (c *Client) request(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("3x-ui request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("3x-ui returned %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetStatus 获取3x-ui面板状态
func (c *Client) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	data, err := c.request(ctx, http.MethodPost, "/login", map[string]string{
		"username": "admin",
		"password": c.token,
	})
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetInbounds 获取入站列表
func (c *Client) GetInbounds(ctx context.Context) ([]map[string]interface{}, error) {
	// 3x-ui 的 API 路径可能因版本不同而变化
	// 这里使用常见的 API 路径
	data, err := c.request(ctx, http.MethodGet, "/xui/API/inbounds", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Success bool                     `json:"success"`
		Obj     []map[string]interface{} `json:"obj"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		// 尝试直接解析为数组
		var list []map[string]interface{}
		if err := json.Unmarshal(data, &list); err != nil {
			return nil, err
		}
		return list, nil
	}

	if !result.Success {
		return nil, fmt.Errorf("3x-ui API returned success=false")
	}
	return result.Obj, nil
}

// GetClients 获取客户端列表（按入站ID）
func (c *Client) GetClients(ctx context.Context, inboundID int) ([]map[string]interface{}, error) {
	data, err := c.request(ctx, http.MethodGet, fmt.Sprintf("/xui/API/inbounds/%d", inboundID), nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Success bool                   `json:"success"`
		Obj     map[string]interface{} `json:"obj"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// 从入站配置中提取客户端信息
	if settings, ok := result.Obj["settings"].(string); ok {
		var settingsMap map[string]interface{}
		if err := json.Unmarshal([]byte(settings), &settingsMap); err == nil {
			if clients, ok := settingsMap["clients"].([]interface{}); ok {
				result := make([]map[string]interface{}, 0, len(clients))
				for _, c := range clients {
					if cm, ok := c.(map[string]interface{}); ok {
						result = append(result, cm)
					}
				}
				return result, nil
			}
		}
	}

	return []map[string]interface{}{}, nil
}

// GetXrayStatus 获取Xray运行状态
func (c *Client) GetXrayStatus(ctx context.Context) (map[string]interface{}, error) {
	data, err := c.request(ctx, http.MethodGet, "/xui/API/xray/status", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ToggleXray 启动/停止Xray
func (c *Client) ToggleXray(ctx context.Context) error {
	_, err := c.request(ctx, http.MethodPost, "/xui/API/xray/restart", nil)
	return err
}

// GetNodes 获取节点流量统计
func (c *Client) GetNodes(ctx context.Context) ([]map[string]interface{}, error) {
	inbounds, err := c.GetInbounds(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为节点视图
	nodes := make([]map[string]interface{}, 0, len(inbounds))
	for _, in := range inbounds {
		node := map[string]interface{}{
			"id":         in["id"],
			"protocol":   in["protocol"],
			"port":       in["port"],
			"remark":     in["remark"],
			"enable":     in["enable"],
			"up":         in["up"],
			"down":       in["down"],
			"total":      in["total"],
			"expiryTime": in["expiryTime"],
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}
