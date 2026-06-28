#!/bin/bash
#
# CloudProbe Agent 一键安装脚本
# 支持国内/海外双源下载
# Usage: curl -fsSL http://your-dashboard:port/install.sh | bash -s -- [token]
#        curl -fsSL http://your-dashboard:port/install.sh | bash -s -- [region] [dashboard_url] [token]

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 参数解析 - 支持两种格式：
# 格式1（推荐）: 只传 token，自动从下载 URL 推断 Dashboard 地址
# 格式2（完整）: region dashboard_url token
if [ $# -eq 1 ]; then
    AGENT_TOKEN="$1"
    REGION="auto"
    # 从脚本下载 URL 自动推断 Dashboard 地址
    # 用户执行: curl http://host:port/install.sh | bash -s -- TOKEN
    # 需要设置 DASHBOARD_URL 环境变量，或通过交互式输入
    DASHBOARD_URL="${CP_DASHBOARD_URL:-}"
else
    REGION="${1:-auto}"
    DASHBOARD_URL="${2:-}"
    AGENT_TOKEN="${3:-}"
fi

# 下载源配置
GITHUB_RELEASE="https://github.com/dennis926/cloudprobe/releases/latest/download"
GITEE_RELEASE="https://gitee.com/den7hon/cloudprobe/releases/latest/download"

# 检测系统
ARCH=$(uname -m)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# 架构映射
case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    armv7l)  ARCH="armv7" ;;
    *) echo -e "${RED}不支持的架构: $ARCH${NC}"; exit 1 ;;
esac

case "$OS" in
    linux) OS="linux" ;;
    darwin) OS="darwin" ;;
    *) echo -e "${RED}不支持的操作系统: $OS${NC}"; exit 1 ;;
esac

# 自动检测国内/海外
if [ "$REGION" = "auto" ]; then
    echo -e "${BLUE}正在检测网络环境...${NC}"
    if curl -fsSL -m 5 https://gitee.com >/dev/null 2>&1; then
        if curl -fsSL -m 5 https://github.com >/dev/null 2>&1; then
            echo -e "${GREEN}检测到海外网络，使用 GitHub 源${NC}"
            REGION="海外"
        else
            echo -e "${GREEN}检测到国内网络，使用 Gitee 源${NC}"
            REGION="国内"
        fi
    else
        echo -e "${GREEN}检测到海外网络，使用 GitHub 源${NC}"
        REGION="海外"
    fi
fi

# 选择下载源
if [ "$REGION" = "国内" ] || [ "$REGION" = "cn" ] || [ "$REGION" = "CN" ]; then
    DOWNLOAD_URL="$GITEE_RELEASE/cloudprobe-agent-${OS}-${ARCH}"
    SOURCE="Gitee"
else
    DOWNLOAD_URL="$GITHUB_RELEASE/cloudprobe-agent-${OS}-${ARCH}"
    SOURCE="GitHub"
fi

BINARY_NAME="cloudprobe-agent"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/cloudprobe"
SERVICE_NAME="cloudprobe-agent"

print_banner() {
    echo -e "${BLUE}"
    echo "  ____ _                 _      ____            _            "
    echo " / ___| | ___  _   _  __| |    |  _ \ _ __ ___ | | ___  ___  "
    echo "| |   | |/ _ \| | | |/ _\` |    | |_) | '__/ _ \| |/ _ \/ __|"
    echo "| |___| | (_) | |_| | (_| |    |  __/| | | (_) | |  __/\__ \\"
    echo " \____|_|\___/ \__,_|\__,_|    |_|   |_|  \___/|_|\___||___/"
    echo -e "${NC}"
    echo -e "${GREEN}Agent 一键安装脚本${NC}"
    echo -e "系统: ${YELLOW}$OS/$ARCH${NC} | 源: ${YELLOW}$SOURCE${NC}"
    echo ""
}

# 检查 root 权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        echo -e "${RED}请使用 root 权限运行此脚本${NC}"
        echo -e "${YELLOW}sudo bash $0${NC}"
        exit 1
    fi
}

# 安装依赖
install_deps() {
    echo -e "${BLUE}检查依赖...${NC}"
    if ! command -v curl &>/dev/null && ! command -v wget &>/dev/null; then
        if command -v apt &>/dev/null; then
            apt update && apt install -y curl
        elif command -v yum &>/dev/null; then
            yum install -y curl
        elif command -v apk &>/dev/null; then
            apk add curl
        fi
    fi
}

# 交互式配置
interactive_config() {
    if [ -z "$DASHBOARD_URL" ]; then
        echo -e "${YELLOW}请输入 Dashboard WebSocket 地址${NC}"
        echo -e "示例: wss://your-domain.com/ws/agent"
        read -rp "Dashboard URL: " DASHBOARD_URL
    fi

    if [ -z "$AGENT_TOKEN" ]; then
        echo -e "${YELLOW}请输入 Agent Token${NC}"
        echo -e "（在 Dashboard 的 服务器管理 -> 添加服务器 中获取）"
        read -rp "Agent Token: " AGENT_TOKEN
    fi

    if [ -z "$DASHBOARD_URL" ] || [ -z "$AGENT_TOKEN" ]; then
        echo -e "${RED}Dashboard URL 和 Agent Token 不能为空${NC}"
        exit 1
    fi
}

# 下载 Agent
download_agent() {
    echo -e "${BLUE}正在从 $SOURCE 下载 Agent...${NC}"
    echo -e "${BLUE}下载地址: $DOWNLOAD_URL${NC}"

    TMP_DIR=$(mktemp -d)
    TMP_FILE="$TMP_DIR/$BINARY_NAME"

    if command -v curl &>/dev/null; then
        curl -fsSL -o "$TMP_FILE" "$DOWNLOAD_URL" --progress-bar
    else
        wget -q --show-progress -O "$TMP_FILE" "$DOWNLOAD_URL"
    fi

    if [ ! -f "$TMP_FILE" ]; then
        echo -e "${RED}下载失败，请检查网络连接${NC}"
        exit 1
    fi

    chmod +x "$TMP_FILE"
    echo -e "${GREEN}下载完成${NC}"
}

# 安装 Agent
install_binary() {
    echo -e "${BLUE}安装 Agent...${NC}"
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"

    cp "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"

    # 创建配置文件
    cat > "$CONFIG_DIR/agent.yml" <<EOF
# CloudProbe Agent 配置文件
server_url: "$DASHBOARD_URL"
token: "$AGENT_TOKEN"
interval: 30
heartbeat: 30
EOF

    echo -e "${GREEN}Agent 安装完成: $INSTALL_DIR/$BINARY_NAME${NC}"
    echo -e "${GREEN}配置文件: $CONFIG_DIR/agent.yml${NC}"
}

# 创建 systemd 服务
install_systemd() {
    echo -e "${BLUE}创建 systemd 服务...${NC}"

    cat > "/etc/systemd/system/${SERVICE_NAME}.service" <<EOF
[Unit]
Description=CloudProbe Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/$BINARY_NAME -c $CONFIG_DIR/agent.yml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable "$SERVICE_NAME"
    systemctl start "$SERVICE_NAME"

    echo -e "${GREEN}systemd 服务已创建并启动${NC}"
}

# 验证安装
verify_install() {
    echo -e "${BLUE}验证安装...${NC}"
    sleep 2

    if systemctl is-active --quiet "$SERVICE_NAME"; then
        echo -e "${GREEN}Agent 运行正常${NC}"
        echo -e "${GREEN}查看日志: journalctl -u $SERVICE_NAME -f${NC}"
    else
        echo -e "${RED}Agent 启动失败，查看日志排查:${NC}"
        echo -e "${YELLOW}journalctl -u $SERVICE_NAME --no-pager -n 50${NC}"
        exit 1
    fi
}

# 卸载
uninstall() {
    echo -e "${YELLOW}正在卸载 CloudProbe Agent...${NC}"
    systemctl stop "$SERVICE_NAME" 2>/dev/null || true
    systemctl disable "$SERVICE_NAME" 2>/dev/null || true
    rm -f "/etc/systemd/system/${SERVICE_NAME}.service"
    rm -f "$INSTALL_DIR/$BINARY_NAME"
    rm -rf "$CONFIG_DIR"
    systemctl daemon-reload
    echo -e "${GREEN}卸载完成${NC}"
    exit 0
}

# 主流程
main() {
    print_banner

    # 检查卸载参数
    if [ "${1:-}" = "uninstall" ] || [ "${1:-}" = "remove" ]; then
        uninstall
    fi

    check_root
    install_deps
    interactive_config
    download_agent
    install_binary
    install_systemd
    verify_install

    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  CloudProbe Agent 安装成功!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo -e "${BLUE}常用命令:${NC}"
    echo -e "  查看状态: ${YELLOW}systemctl status $SERVICE_NAME${NC}"
    echo -e "  查看日志: ${YELLOW}journalctl -u $SERVICE_NAME -f${NC}"
    echo -e "  重启服务: ${YELLOW}systemctl restart $SERVICE_NAME${NC}"
    echo -e "  停止服务: ${YELLOW}systemctl stop $SERVICE_NAME${NC}"
    echo -e "  卸载: ${YELLOW}bash $0 uninstall${NC}"
    echo ""
}

main "$@"
