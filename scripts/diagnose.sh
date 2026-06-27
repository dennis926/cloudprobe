#!/bin/bash
# CloudProbe 网络诊断脚本
# 运行后把完整输出复制给我

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  CloudProbe 网络诊断报告${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

# 1. 服务器基本信息
echo -e "${YELLOW}[1] 服务器基本信息${NC}"
echo "主机名: $(hostname)"
echo "IP地址: $(curl -s ifconfig.me 2>/dev/null || echo '无法获取')"
echo ""

# 2. Docker 容器状态
echo -e "${YELLOW}[2] Docker 容器状态${NC}"
cd /www/wwwroot/cloudprobe 2>/dev/null || true
docker-compose -f docker-compose.bt.yml ps 2>/dev/null || docker compose -f docker-compose.bt.yml ps 2>/dev/null || echo "docker-compose 命令失败"
echo ""

# 3. 端口监听情况
echo -e "${YELLOW}[3] 端口监听情况 (8090 相关)${NC}"
ss -tlnp | grep -E ':809[0-9]' || echo "无 809x 端口监听"
echo ""

echo -e "${YELLOW}[3.1] 全部监听端口${NC}"
ss -tlnp | head -20
echo ""

# 4. 检查 8090 被谁占用
echo -e "${YELLOW}[4] 8090 端口占用详情${NC}"
ss -tlnp | grep ':8090 ' || echo "8090 未被占用"
echo ""

# 5. Docker 端口映射
echo -e "${YELLOW}[5] Docker 端口映射${NC}"
docker port cloudprobe-dashboard 2>/dev/null || echo "cloudprobe-dashboard 容器不存在或未运行"
echo ""

# 6. 本地访问测试
echo -e "${YELLOW}[6] 本地访问测试${NC}"
curl -s -o /dev/null -w "HTTP状态: %{http_code}, 耗时: %{time_total}s\n" http://127.0.0.1:8090/ 2>/dev/null || echo "本地访问 127.0.0.1:8090 失败"
curl -s -o /dev/null -w "HTTP状态: %{http_code}, 耗时: %{time_total}s\n" http://0.0.0.0:8090/ 2>/dev/null || echo "本地访问 0.0.0.0:8090 失败"
echo ""

# 7. 防火墙状态
echo -e "${YELLOW}[7] 防火墙状态${NC}"
if command -v ufw &> /dev/null; then
    ufw status verbose 2>/dev/null || echo "ufw 未启用"
elif command -v firewall-cmd &> /dev/null; then
    firewall-cmd --list-all 2>/dev/null || echo "firewalld 信息获取失败"
else
    echo "未检测到 ufw/firewalld"
fi
echo ""

# 8. Nginx 配置检查
echo -e "${YELLOW}[8] Nginx 配置 (8090 相关)${NC}"
if [ -d /www/server/panel/vhost/nginx ]; then
    grep -r "8090" /www/server/panel/vhost/nginx/ 2>/dev/null || echo "Nginx 配置中无 8090"
else
    echo "未检测到宝塔 Nginx 配置目录"
fi
echo ""

# 9. docker-compose 配置检查
echo -e "${YELLOW}[9] docker-compose.bt.yml 端口配置${NC}"
grep -A2 -B2 "ports:" /www/wwwroot/cloudprobe/docker-compose.bt.yml 2>/dev/null || echo "文件不存在"
echo ""

# 10. Docker 日志
echo -e "${YELLOW}[10] CloudProbe 最近日志${NC}"
cd /www/wwwroot/cloudprobe 2>/dev/null || true
docker-compose -f docker-compose.bt.yml logs --tail=20 dashboard 2>/dev/null || docker compose -f docker-compose.bt.yml logs --tail=20 dashboard 2>/dev/null || echo "无法获取日志"
echo ""

echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  诊断完成，请把上方全部输出复制给我${NC}"
echo -e "${CYAN}========================================${NC}"
