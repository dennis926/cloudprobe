package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloudprobe/internal/grpc/pb"
	"cloudprobe/pkg/agent/collector"
	"cloudprobe/pkg/agent/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// GRPCReporter gRPC上报器
type GRPCReporter struct {
	cfg       *config.Config
	collector *collector.Collector
	conn      *grpc.ClientConn
	client    pb.AgentServiceClient
	done      chan struct{}
}

// NewGRPCReporter 创建gRPC上报器
func NewGRPCReporter(cfg *config.Config) (*GRPCReporter, error) {
	c, err := collector.NewCollector()
	if err != nil {
		return nil, err
	}

	return &GRPCReporter{
		cfg:       cfg,
		collector: c,
		done:      make(chan struct{}),
	}, nil
}

// Start 启动上报循环
func (r *GRPCReporter) Start(ctx context.Context) {
	if err := r.connect(); err != nil {
		fmt.Printf("Initial gRPC connection failed: %v, will retry...\n", err)
	}

	ticker := time.NewTicker(time.Duration(r.cfg.Interval) * time.Second)
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-ticker.C:
			if r.conn == nil || r.conn.GetState() == connectivity.Shutdown {
				if err := r.connect(); err != nil {
					fmt.Printf("gRPC reconnect failed: %v\n", err)
					continue
				}
			}
			r.report()

		case <-heartbeatTicker.C:
			if r.conn != nil && r.conn.GetState() == connectivity.Ready {
				r.heartbeat()
			}

		case <-ctx.Done():
			return
		case <-r.done:
			return
		}
	}
}

// Stop 停止上报
func (r *GRPCReporter) Stop() {
	close(r.done)
	if r.conn != nil {
		r.conn.Close()
	}
}

// connect 建立gRPC连接
func (r *GRPCReporter) connect() error {
	if r.conn != nil {
		r.conn.Close()
	}

	// 配置TLS（海外部署必须加密）
	cred, err := credentials.NewClientTLSFromFile("/etc/cloudprobe/ca.crt", "")
	if err != nil {
		// 如果证书不存在，使用insecure（开发环境）
		fmt.Printf("TLS cert not found, using insecure connection: %v\n", err)
		conn, err := grpc.NewClient(r.cfg.ServerURL, grpc.WithInsecure())
		if err != nil {
			return fmt.Errorf("grpc dial failed: %w", err)
		}
		r.conn = conn
	} else {
		conn, err := grpc.NewClient(r.cfg.ServerURL, grpc.WithTransportCredentials(cred))
		if err != nil {
			return fmt.Errorf("grpc tls dial failed: %w", err)
		}
		r.conn = conn
	}

	// 等待连接就绪
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r.conn.Connect()
	if !r.conn.WaitForStateChange(ctx, connectivity.Idle) {
		return fmt.Errorf("grpc connection timeout")
	}

	r.client = &agentServiceClient{cc: r.conn}
	fmt.Println("gRPC connected to dashboard")
	return nil
}

// report 上报指标
func (r *GRPCReporter) report() {
	metrics, err := r.collector.Collect()
	if err != nil {
		fmt.Printf("Collect failed: %v\n", err)
		return
	}

	payload, err := json.Marshal(metrics)
	if err != nil {
		fmt.Printf("Marshal metrics failed: %v\n", err)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "x-agent-token", r.cfg.Token)
	_, err = r.client.Report(ctx, &pb.ReportRequest{
		Token:     r.cfg.Token,
		Payload:   payload,
		Timestamp: time.Now().Unix(),
		Hostname:  metrics.Hostname,
	})
	if err != nil {
		fmt.Printf("gRPC report failed: %v\n", err)
	}
}

// heartbeat 发送心跳
func (r *GRPCReporter) heartbeat() {
	ctx := metadata.AppendToOutgoingContext(context.Background(), "x-agent-token", r.cfg.Token)
	_, err := r.client.Heartbeat(ctx, &pb.HeartbeatRequest{
		Token:     r.cfg.Token,
		Timestamp: time.Now().Unix(),
	})
	if err != nil {
		fmt.Printf("gRPC heartbeat failed: %v\n", err)
	}
}

// ==================== gRPC Client 手写实现 ====================

type agentServiceClient struct {
	cc *grpc.ClientConn
}

func (c *agentServiceClient) Report(ctx context.Context, in *pb.ReportRequest, opts ...grpc.CallOption) (*pb.ReportResponse, error) {
	out := new(pb.ReportResponse)
	err := c.cc.Invoke(ctx, "/agent.AgentService/Report", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentServiceClient) Heartbeat(ctx context.Context, in *pb.HeartbeatRequest, opts ...grpc.CallOption) (*pb.HeartbeatResponse, error) {
	out := new(pb.HeartbeatResponse)
	err := c.cc.Invoke(ctx, "/agent.AgentService/Heartbeat", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentServiceClient) StreamReport(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[pb.ReportRequest, pb.CommandResponse], error) {
	stream, err := c.cc.NewStream(ctx, &grpc.StreamDesc{
		StreamName:    "StreamReport",
		ClientStreams: true,
		ServerStreams: true,
	}, "/agent.AgentService/StreamReport", opts...)
	if err != nil {
		return nil, err
	}
	return &agentServiceStreamReportClient{stream}, nil
}

type agentServiceStreamReportClient struct {
	grpc.ClientStream
}

func (x *agentServiceStreamReportClient) Send(m *pb.ReportRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *agentServiceStreamReportClient) Recv() (*pb.CommandResponse, error) {
	m := new(pb.CommandResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}
