package pb

// 手写 protobuf 兼容消息结构体
// 实际项目应通过 `make proto` 由 protoc 自动生成

// ReportRequest 上报请求
type ReportRequest struct {
	Token     string `json:"token,omitempty"`
	Payload   []byte `json:"payload,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Hostname  string `json:"hostname,omitempty"`
}

func (x *ReportRequest) Reset()         { *x = ReportRequest{} }
func (x *ReportRequest) String() string { return string(x.Payload) }
func (x *ReportRequest) ProtoMessage()  {}

// ReportResponse 上报响应
type ReportResponse struct {
	Success bool   `json:"success,omitempty"`
	Message string `json:"message,omitempty"`
}

func (x *ReportResponse) Reset()         { *x = ReportResponse{} }
func (x *ReportResponse) String() string { return x.Message }
func (x *ReportResponse) ProtoMessage()  {}

// HeartbeatRequest 心跳请求
type HeartbeatRequest struct {
	Token     string `json:"token,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

func (x *HeartbeatRequest) Reset()         { *x = HeartbeatRequest{} }
func (x *HeartbeatRequest) String() string { return x.Token }
func (x *HeartbeatRequest) ProtoMessage()  {}

// HeartbeatResponse 心跳响应
type HeartbeatResponse struct {
	Success    bool   `json:"success,omitempty"`
	Message    string `json:"message,omitempty"`
	ServerTime int64  `json:"server_time,omitempty"`
}

func (x *HeartbeatResponse) Reset()         { *x = HeartbeatResponse{} }
func (x *HeartbeatResponse) String() string { return x.Message }
func (x *HeartbeatResponse) ProtoMessage()  {}

// CommandResponse 命令响应
type CommandResponse struct {
	Command string `json:"command,omitempty"`
	Payload []byte `json:"payload,omitempty"`
}

func (x *CommandResponse) Reset()         { *x = CommandResponse{} }
func (x *CommandResponse) String() string { return x.Command }
func (x *CommandResponse) ProtoMessage()  {}
