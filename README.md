# csv2kafka

使用 Go 从标准输入（管道）读取 CSV 数据，逐行发送到 Kafka 或 ZeroMQ。适用于与 super_mediator、YAF 等工具通过管道集成。

## 构建
- 需要 Go 1.21+
- 拉取依赖并编译：
  ```bash
  go mod tidy
  go build -o csv2kafka .
  
  set GOOS=linux
  set GOARCH=amd64
  go build -o csv2kafka .

  ```

## 使用方法

### 主要参数
- `-transport`：发送后端，`kafka` 或 `zmq`，默认 `kafka`  
- `-topic`：Kafka topic，`transport=kafka` 时生效，默认 `yaf_csv`  
- `-zmq-endpoint`：ZeroMQ PUSH 端点，`transport=zmq` 时生效，默认 `tcp://127.0.0.1:5555`
- `-skip-header`：是否跳过第一行作为 CSV 表头，默认 `true`

### 运行方式
程序从**标准输入（stdin）**读取 CSV 数据，支持以下使用场景：

1. **管道输入**（推荐）：与其他工具通过管道连接
2. **重定向输入**：从文件重定向到标准输入
3. **Docker 容器内**：作为 super_mediator 的下游处理程序

## 运行示例

### 1. 管道模式（与 super_mediator 集成）
```bash
# Kafka 模式
super_mediator --output-mode=TEXT --out=- | \
  ./csv2kafka -transport kafka -topic yaf_csv

# ZeroMQ 模式
super_mediator --output-mode=TEXT --out=- | \
  ./csv2kafka -transport zmq -zmq-endpoint tcp://10.0.0.5:5555
```

### 2. 文件重定向模式
```bash
# 从 CSV 文件读取
./csv2kafka -transport kafka -topic yaf_csv < data.csv

# 或使用 cat
cat data.csv | ./csv2kafka -transport kafka -topic yaf_csv
```

### 3. Docker 容器内使用
参考 Dockerfile 中的集成方式：
```bash
super_mediator \
  --ipfix-input=tcp \
  --ipfix-port=18000 \
  --output-mode=TEXT \
  --out=- \
  | /usr/local/bin/csv2kafka -transport kafka -topic yaf_csv
```

### 4. 后台运行（配合 nohup）
```bash
nohup super_mediator --output-mode=TEXT --out=- | \
  ./csv2kafka -transport kafka -topic yaf_csv > run.log 2>&1 &
```

## CSV 数据格式要求
- 数据通过标准输入逐行传入
- 默认跳过第一行作为表头（可通过 `-skip-header=false` 禁用）
- 空行会被自动跳过
- 每行数据会原样发送到 Kafka 或 ZeroMQ

## Kafka & ZMQ 说明
- Kafka：生产者使用 sarama 异步发送，等待 leader 确认（`WaitForLocal`），错误会打印日志，成功不回传。如需修改 broker 列表或更多配置，可在 `main.go` 与 `kafka.go` 中调整。
- ZMQ：当前实现创建 PUSH socket 并连接到 `-zmq-endpoint`。接收端可使用 PULL/SUB 等模式配合，ZeroMQ C 库需预先安装。若需更复杂的套接字模式，可参考 `zmq.go`。

## ZeroMQ 环境准备
运行 `-transport=zmq` 前，必须在运行主程序的服务器上安装 ZeroMQ C 库（`libzmq`）以及开发头文件。常用系统的安装方式：

- **Debian / Ubuntu**
  ```bash
  sudo apt update
  sudo apt install -y libzmq3-dev
  ```
- **CentOS / RHEL / Rocky**
  ```bash
  sudo yum install -y epel-release
  sudo yum install -y zeromq zeromq-devel
  ```
- **macOS (Homebrew)**
  ```bash
  brew install zeromq
  ```
- **Windows**
  - 安装 [ZeroMQ 官方 prebuilt binaries](https://github.com/zeromq/libzmq/releases) 或使用 `choco install zeromq`。
  - 将 `libzmq.dll` 所在目录加入 `PATH`，并确保 Go 编译/运行环境可以找到相应头文件（可通过安装 [libsodium+libzmq 包](https://github.com/zeromq/libzmq) 或使用 MSYS2 `pacman -S mingw-w64-x86_64-zeromq`）。

安装完成后，可运行 `pkg-config --modversion libzmq`（Linux/macOS）或使用 `where libzmq.dll`（Windows）验证库是否就绪，再编译/运行本项目即可。

## 常见问题
- 无法拉取依赖：设置代理 `go env -w GOPROXY=https://goproxy.cn,direct`，必要时 `go env -w GOSUMDB=off`
- topic 不存在：确保 Kafka 允许自动创建或提前创建目标 topic
