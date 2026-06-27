#!/bin/bash
# 重置 CloudProbe 默认管理员密码为 admin

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}正在生成正确的 bcrypt 哈希...${NC}"

# 使用 Go 生成正确的 bcrypt 哈希
export PATH=$PATH:/usr/local/go/bin

cat > /tmp/gen_hash.go << 'EOF'
package main
import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)
func main() {
    h, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
    if err != nil {
        fmt.Println("ERROR:", err)
        return
    }
    fmt.Println(string(h))
}
EOF

HASH=$(go run /tmp/gen_hash.go)

if [ "$HASH" = "" ] || [ "${HASH:0:5}" != "$2a$" ]; then
    echo -e "${RED}哈希生成失败${NC}"
    exit 1
fi

echo -e "${GREEN}哈希生成成功${NC}"
echo -e "${YELLOW}正在更新数据库...${NC}"

# 更新数据库
docker exec -i cloudprobe-postgres psql -U cpuser -d cloudprobe -c "
UPDATE users SET password = '$HASH' WHERE username = 'admin';
"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}密码重置成功！${NC}"
    echo -e "${GREEN}请用 admin / admin 登录${NC}"
else
    echo -e "${RED}密码重置失败${NC}"
    exit 1
fi

rm -f /tmp/gen_hash.go
