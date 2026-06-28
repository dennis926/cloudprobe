package reporter

import (
	"context"
	"fmt"
	"time"

	"cloudprobe/pkg/agent/collector"
	"cloudprobe/pkg/agent/config"

	"github.com/gorilla/websocket"
)

// Reporter 数据上报器
type Reporter struct {
	cfg           *config.Config
	collector     *collector.Collector
	ws            *websocket.Conn
	done          chan struct{}
	reportedSystem bool
}

// New 创建上报器
func New(cfg *config.Config) (*Reporter, error) {
	c, err := collector.NewCollector()
	if err != nil {
		return nil, err
	}

	return &Reporter{
		cfg:       cfg,
		collector: c,
		done:      make(chan struct{}),
	}, nil
}

// Start 启动上报循环
func (r *Reporter) Start(ctx context.Context) {
	if err := r.connect(); err != nil {
		fmt.Printf("Initial connection failed: %v, will retry...\n", err)
	}

	ticker := time.NewTicker(time.Duration(r.cfg.Interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if r.ws == nil {
				if err := r.connect(); err != nil {
					fmt.Printf("Reconnect failed: %v\n", err)
					continue
				}
			}
			r.report()

		case <-ctx.Done():
			return
		case <-r.done:
			return
		}
	}
}

// Stop 停止上报
func (r *Reporter) Stop() {
	close(r.done)
	if r.ws != nil {
		r.ws.Close()
	}
}

// connect 建立WebSocket连接
func (r *Reporter) connect() error {
	if r.ws != nil {
		r.ws.Close()
	}

	ws, _, err := websocket.DefaultDialer.Dial(r.cfg.ServerURL, nil)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	// 发送认证消息
	auth := map[string]interface{}{
		"type":  "auth",
		"token": r.cfg.Token,
		"time":  time.Now().Unix(),
	}
	if err := ws.WriteJSON(auth); err != nil {
		ws.Close()
		return fmt.Errorf("auth failed: %w", err)
	}

	// 等待认证响应
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	var resp map[string]interface{}
	if err := ws.ReadJSON(&resp); err != nil {
		ws.Close()
		return fmt.Errorf("auth response failed: %w", err)
	}
	ws.SetReadDeadline(time.Time{})

	if resp["type"] != "auth_success" {
		ws.Close()
		return fmt.Errorf("auth rejected: %v", resp)
	}

	r.ws = ws
	r.reportedSystem = false
	fmt.Println("Connected to dashboard")

	// 启动消息读取协程
	go r.readLoop()

	return nil
}

// readLoop 读取服务端消息
func (r *Reporter) readLoop() {
	for {
		var msg map[string]interface{}
		if err := r.ws.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("Read error: %v\n", err)
			}
			r.ws = nil
			return
		}

		msgType, _ := msg["type"].(string)
		switch msgType {
		case "ping":
			r.ws.WriteJSON(map[string]string{"type": "pong"})
		}
	}
}

// report 上报指标数据
func (r *Reporter) report() {
	metrics, err := r.collector.Collect()
	if err != nil {
		fmt.Printf("Collect failed: %v\n", err)
		return
	}

	data := map[string]interface{}{
		"type": "metrics",
		"data": metrics,
		"time": time.Now().Unix(),
	}

	if err := r.ws.WriteJSON(data); err != nil {
		fmt.Printf("Report failed: %v\n", err)
		r.ws.Close()
		r.ws = nil
	}

	// 上报系统信息（仅在首次连接时）
	if !r.reportedSystem {
		r.reportedSystem = true
		sysInfo := map[string]interface{}{
			"type":  "system",
			"token": r.cfg.Token,
			"data": map[string]interface{}{
				"hostname": metrics.Hostname,
				"os":       metrics.OS,
				"platform": metrics.Platform,
				"cpu": map[string]interface{}{
					"model":         metrics.CPU.ModelName,
					"logical_count": metrics.CPU.LogicalCnt,
				},
				"memory": map[string]interface{}{
					"total": metrics.Memory.Total,
				},
				"disk": []map[string]interface{}{
					{
						"total": metrics.Disk[0].Total,
					},
				},
				"ip": map[string]interface{}{
					// IP 将由 Dashboard 端从连接中获取
				},
			},
			"time": time.Now().Unix(),
		}
		r.ws.WriteJSON(sysInfo)
	}
}
