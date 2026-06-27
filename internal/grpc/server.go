package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"cloudprobe/internal/database"
	"cloudprobe/internal/grpc/pb"
	"cloudprobe/internal/model"
	"cloudprobe/internal/service"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AgentServer gRPC Agent服务
type AgentServer struct {
	logger *zap.Logger
}

// NewAgentServer 创建gRPC Agent服务
func NewAgentServer(logger *zap.Logger) *AgentServer {
	return &AgentServer{logger: logger}
}

// authInterceptor Token认证拦截器
func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	tokens := md.Get("x-agent-token")
	if len(tokens) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	var server model.Server
	if err := database.GetDB().Where("agent_token = ?", tokens[0]).First(&server).Error; err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// 将serverID写入context
	ctx = context.WithValue(ctx, "server_id", server.ID)
	return handler(ctx, req)
}

// Report 处理Agent上报
func (s *AgentServer) Report(ctx context.Context, req *pb.ReportRequest) (*pb.ReportResponse, error) {
	serverID, ok := ctx.Value("server_id").(uint)
	if !ok {
		return nil, status.Error(codes.Internal, "server id not found")
	}

	if len(req.Payload) > 0 {
		var data map[string]interface{}
		if err := json.Unmarshal(req.Payload, &data); err == nil {
			if err := service.HandleMetricsFromAgent(serverID, data); err != nil {
				s.logger.Error("handle metrics failed", zap.Error(err))
			}
		}
	}

	// 更新在线状态
	svc := service.NewServerService()
	if err := svc.UpdateServerStatus(serverID, "online"); err != nil {
		s.logger.Error("update server status failed", zap.Error(err))
	}

	return &pb.ReportResponse{
		Success: true,
		Message: "report received",
	}, nil
}

// Heartbeat 处理Agent心跳
func (s *AgentServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	serverID, ok := ctx.Value("server_id").(uint)
	if !ok {
		return nil, status.Error(codes.Internal, "server id not found")
	}

	svc := service.NewServerService()
	if err := svc.UpdateServerStatus(serverID, "online"); err != nil {
		s.logger.Error("heartbeat update failed", zap.Error(err))
	}

	return &pb.HeartbeatResponse{
		Success:    true,
		Message:    "pong",
		ServerTime: time.Now().Unix(),
	}, nil
}

// StreamReport 流式上报
func (s *AgentServer) StreamReport(stream grpc.BidiStreamingServer[pb.ReportRequest, pb.CommandResponse]) error {
	// 从第一条消息获取token进行认证
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	var server model.Server
	if err := database.GetDB().Where("agent_token = ?", req.Token).First(&server).Error; err != nil {
		return status.Error(codes.Unauthenticated, "invalid token")
	}

	serverID := server.ID
	s.logger.Info("agent stream connected", zap.Uint("server_id", serverID))

	// 发送欢迎命令
	if err := stream.Send(&pb.CommandResponse{Command: "connected"}); err != nil {
		return err
	}

	// 循环接收上报数据
	for {
		req, err := stream.Recv()
		if err != nil {
			s.logger.Info("agent stream disconnected", zap.Uint("server_id", serverID))
			break
		}

		if len(req.Payload) > 0 {
			var data map[string]interface{}
			if err := json.Unmarshal(req.Payload, &data); err == nil {
				if err := service.HandleMetricsFromAgent(serverID, data); err != nil {
					s.logger.Error("stream handle metrics failed", zap.Error(err))
				}
			}
		}
	}

	return nil
}

// StartGRPCServer 启动gRPC服务
func StartGRPCServer(addr string, logger *zap.Logger) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	server := NewAgentServer(logger)

	// 注册服务描述符
	s := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor),
	)

	// 使用反射注册服务（因为手写protobuf没有生成的Register函数）
	// 这里使用grpc的通用服务注册
	RegisterAgentServiceServer(s, server)

	go func() {
		logger.Info("gRPC server started", zap.String("addr", addr))
		if err := s.Serve(lis); err != nil {
			logger.Error("gRPC server failed", zap.Error(err))
		}
	}()

	return s, nil
}

// AgentServiceServer 接口
type AgentServiceServer interface {
	Report(context.Context, *pb.ReportRequest) (*pb.ReportResponse, error)
	Heartbeat(context.Context, *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error)
	StreamReport(grpc.BidiStreamingServer[pb.ReportRequest, pb.CommandResponse]) error
}

// RegisterAgentServiceServer 注册服务
func RegisterAgentServiceServer(s *grpc.Server, srv AgentServiceServer) {
	desc := &grpc.ServiceDesc{
		ServiceName: "agent.AgentService",
		HandlerType: (*AgentServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "Report",
				Handler:    _AgentService_Report_Handler,
			},
			{
				MethodName: "Heartbeat",
				Handler:    _AgentService_Heartbeat_Handler,
			},
		},
		Streams: []grpc.StreamDesc{
			{
				StreamName:    "StreamReport",
				Handler:       _AgentService_StreamReport_Handler,
				ServerStreams: true,
				ClientStreams: true,
			},
		},
		Metadata: "agent.proto",
	}
	s.RegisterService(desc, srv)
}

func _AgentService_Report_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(pb.ReportRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).Report(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/agent.AgentService/Report",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).Report(ctx, req.(*pb.ReportRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_Heartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(pb.HeartbeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).Heartbeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/agent.AgentService/Heartbeat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).Heartbeat(ctx, req.(*pb.HeartbeatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_StreamReport_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(AgentServiceServer).StreamReport(&agentServiceStreamReportServer{stream})
}

type agentServiceStreamReportServer struct {
	grpc.ServerStream
}

func (x *agentServiceStreamReportServer) Send(m *pb.CommandResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *agentServiceStreamReportServer) Recv() (*pb.ReportRequest, error) {
	m := new(pb.ReportRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}
