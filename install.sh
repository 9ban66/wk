#!/usr/bin/env bash
set -euo pipefail

# 一键安装脚本
# 使用方式：
#   curl -fsSL https://raw.githubusercontent.com/yatori-dev/yatori-go-core/main/install.sh | bash
#   或者：
#   bash install.sh

REPO_URL="${REPO_URL:-https://github.com/yatori-dev/yatori-go-core.git}"
REPO_DIR="${REPO_DIR:-$HOME/yatori-go-core}"
BRANCH="${BRANCH:-main}"
PORT="${PORT:-8080}"
APP_NAME="yatori-web"

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

fail() {
  echo "错误: $*" >&2
  exit 1
}

install_base_packages() {
  if command_exists apt-get; then
    export DEBIAN_FRONTEND=noninteractive
    sudo apt-get update
    sudo apt-get install -y curl wget git ca-certificates
  elif command_exists yum; then
    sudo yum install -y curl wget git ca-certificates
  elif command_exists apk; then
    sudo apk add --no-cache curl wget git ca-certificates
  else
    fail "当前系统不支持自动安装依赖，请手动安装 curl wget git ca-certificates"
  fi
}

install_go() {
  if command_exists go; then
    return 0
  fi

  if command_exists apt-get; then
    export DEBIAN_FRONTEND=noninteractive
    sudo apt-get install -y golang-go
  elif command_exists yum; then
    sudo yum install -y golang
  elif command_exists apk; then
    sudo apk add --no-cache go
  else
    fail "无法安装 Go，请手动安装 Go 1.24+"
  fi

  command_exists go || fail "Go 安装失败"
}

ensure_repo() {
  mkdir -p "$(dirname "$REPO_DIR")"

  if [ -d "$REPO_DIR/.git" ]; then
    echo "更新仓库: $REPO_DIR"
    cd "$REPO_DIR"
    git remote set-url origin "$REPO_URL" 2>/dev/null || true
    git fetch --depth 1 origin "$BRANCH" || true
    git checkout "$BRANCH" || true
    git pull --ff-only origin "$BRANCH" || true
  else
    echo "克隆仓库到: $REPO_DIR"
    git clone --depth 1 --branch "$BRANCH" "$REPO_URL" "$REPO_DIR"
  fi
}

build_binary() {
  cd "$REPO_DIR"
  if [ ! -f go.mod ]; then
    fail "未找到 go.mod，请检查仓库路径"
  fi

  echo "正在构建 $APP_NAME..."
  GOOS=linux GOARCH=amd64 go build -o "$REPO_DIR/$APP_NAME" .
}

start_service() {
  cd "$REPO_DIR"
  if pgrep -f "$APP_NAME -addr :$PORT" >/dev/null 2>&1; then
    pkill -f "$APP_NAME -addr :$PORT" || true
  fi

  echo "启动服务，监听端口 $PORT"
  nohup "$REPO_DIR/$APP_NAME" -addr ":$PORT" > "$REPO_DIR/${APP_NAME}.log" 2>&1 &
  echo "服务已启动"
  echo "访问地址: http://127.0.0.1:$PORT/"
  echo "日志文件: $REPO_DIR/${APP_NAME}.log"
}

main() {
  install_base_packages
  install_go
  ensure_repo
  build_binary
  start_service
}

case "${1:-all}" in
  all)
    main
    ;;
  build)
    if [ -d "$REPO_DIR/.git" ]; then
      ensure_repo
    fi
    build_binary
    ;;
  start)
    start_service
    ;;
  *)
    echo "Usage: bash install.sh [all|build|start]" >&2
    exit 1
    ;;
esac
