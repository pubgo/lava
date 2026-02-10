#!/bin/bash

# 测试 watch 命令
echo "Testing watch command..."

# 创建测试配置文件
cat > .lava.yaml << 'EOF'
watch:
  watchers:
    - name: "test"
      directory: "."
      patterns:
        - "*.test"
      commands:
        - "echo 'Test file changed'"
      ignore:
        - ".git"
      run_on_startup: false
      timeout: 10
EOF

# 在后台运行 watch 命令
go run main.go watch &
WATCH_PID=$!

# 等待 watch 命令启动
sleep 2

# 创建测试文件
echo "Creating test file..."
echo "test content" > test.test

# 等待 watch 命令检测到文件变更
sleep 2

# 终止 watch 命令
kill $WATCH_PID

# 清理测试文件
rm -f test.test .lava.yaml

echo "Test completed"
