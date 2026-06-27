package ssh

import (
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client SSH客户端封装
type Client struct {
	conn       *ssh.Client
	session    *ssh.Session
	stdin      io.WriteCloser
	stdout     io.Reader
	stderr     io.Reader
	closeCh    chan struct{}
	closeOnce  chan struct{}
	isClosed   bool
}

// Config SSH连接配置
type Config struct {
	Host       string
	Port       int
	User       string
	Password   string
	PrivateKey string
	Timeout    time.Duration
}

// NewClient 创建SSH客户端
func NewClient(cfg *Config) (*Client, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	authMethods := []ssh.AuthMethod{}

	// 优先使用私钥认证
	if cfg.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(cfg.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("parse private key failed: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// 密码认证
	if cfg.Password != "" {
		authMethods = append(authMethods, ssh.Password(cfg.Password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no auth method provided")
	}

	sshConfig := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境应使用 known_hosts
		Timeout:         cfg.Timeout,
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	conn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("ssh dial failed: %w", err)
	}

	client := &Client{
		conn:      conn,
		closeCh:   make(chan struct{}),
		closeOnce: make(chan struct{}),
	}

	return client, nil
}

// OpenShell 打开交互式Shell
func (c *Client) OpenShell(rows, cols int) error {
	session, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("new session failed: %w", err)
	}
	c.session = session

	// 请求伪终端
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	term := "xterm-256color"
	if err := session.RequestPty(term, rows, cols, modes); err != nil {
		return fmt.Errorf("request pty failed: %w", err)
	}

	// 获取管道
	c.stdin, err = session.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe failed: %w", err)
	}
	c.stdout, err = session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe failed: %w", err)
	}
	c.stderr, err = session.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe failed: %w", err)
	}

	// 启动shell
	if err := session.Shell(); err != nil {
		return fmt.Errorf("start shell failed: %w", err)
	}

	return nil
}

// Resize 调整终端大小
func (c *Client) Resize(rows, cols int) error {
	if c.session == nil {
		return fmt.Errorf("session not opened")
	}
	return c.session.WindowChange(rows, cols)
}

// Write 向SSH会话写入数据
func (c *Client) Write(data []byte) (int, error) {
	if c.stdin == nil {
		return 0, fmt.Errorf("stdin not ready")
	}
	return c.stdin.Write(data)
}

// ReadStdout 读取SSH stdout
func (c *Client) ReadStdout() io.Reader {
	return c.stdout
}

// ReadStderr 读取SSH stderr
func (c *Client) ReadStderr() io.Reader {
	return c.stderr
}

// Close 关闭连接
func (c *Client) Close() error {
	select {
	case <-c.closeOnce:
		return nil
	default:
		close(c.closeOnce)
	}

	c.isClosed = true
	close(c.closeCh)

	if c.session != nil {
		c.session.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsClosed 检查是否已关闭
func (c *Client) IsClosed() bool {
	return c.isClosed
}

// Done 返回关闭信号通道
func (c *Client) Done() <-chan struct{} {
	return c.closeCh
}

// Wait 等待会话结束
func (c *Client) Wait() error {
	if c.session == nil {
		return fmt.Errorf("session not opened")
	}
	return c.session.Wait()
}
