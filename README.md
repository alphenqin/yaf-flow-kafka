# csv2kafka

使用 Go 监听目录中写完的 CSV 文件，逐行发送到 Kafka。

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

## 运行
- 默认参数：监听 `/host/output`，发送到 topic `yaf_csv`，Kafka broker `localhost:9092`（可在代码中修改）。
- 运行命令：
  ```bash
  ./csv2kafka -watchdir <监听目录> -topic <目标topic>
  ```
  例如：
  ```bash
  ./csv2kafka -watchdir /data/out -topic my_topic
  ```
- 后台运行示例：
  ```bash
  nohup ./csv2kafka -watchdir /data/out -topic my_topic > run.log 2>&1 &
  ```

## CSV 要求
- 文件扩展名 `.csv`
- 首行为表头，会被跳过
- 程序通过 fsnotify 捕获 `CloseWrite` 事件，并用多次检测文件大小判断“写完”后开始读取

## Kafka 说明
- 生产者使用 sarama 异步发送，等待 leader 确认（`WaitForLocal`）
- 错误会打印日志，成功不回传
- 如需修改 broker 列表或更多 Kafka 配置，可在 `main.go` 与 `kafka.go` 中调整

## 常见问题
- 无法拉取依赖：设置代理 `go env -w GOPROXY=https://goproxy.cn,direct`，必要时 `go env -w GOSUMDB=off`
- topic 不存在：确保 Kafka 允许自动创建或提前创建目标 topic
