# csv2kafka

使用 Go 监听目录中写完的 CSV 文件，逐行发送到 Kafka 或 ZeroMQ。

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
- 主要参数  
  - `-watchdir`：被监控的 CSV 输出目录，默认 `/host/output`  
  - `-transport`：发送后端，`kafka` 或 `zmq`，默认 `kafka`  
  - `-topic`：Kafka topic，`transport=kafka` 时生效，默认 `yaf_csv`  
  - `-zmq-endpoint`：ZeroMQ PUSH 端点，`transport=zmq` 时生效，默认 `tcp://127.0.0.1:5555`

## 运行示例
- Kafka 模式
  ```bash
  ./csv2kafka \
    -watchdir /data/out \
    -transport kafka \
    -topic yaf_csv
  ```
- ZeroMQ 模式
  ```bash
  ./csv2kafka \
    -watchdir /data/out \
    -transport zmq \
    -zmq-endpoint tcp://10.0.0.5:5555
  ```
- 后台运行（Kafka 示例）
  ```bash
  nohup ./csv2kafka -watchdir /data/out -transport kafka -topic yaf_csv > run.log 2>&1 &
  ```

## CSV 要求
- 文件扩展名 `.csv`
- 首行为表头，会被跳过
- 程序通过 fsnotify 捕获 `CloseWrite` 事件，并用多次检测文件大小判断“写完”后开始读取

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
