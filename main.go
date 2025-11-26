package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	defaultTopic      = "yaf_csv"
	defaultTransport  = "kafka"
	defaultZMQAddress = "tcp://127.0.0.1:5555"
)

var KafkaBrokers = []string{"localhost:9092"}

// LineSender 描述统一的发送接口，便于扩展不同后端。
type LineSender interface {
	SendLine(line string)
}

func main() {
	topic := flag.String("topic", defaultTopic, "Kafka topic to send CSV lines")
	transport := flag.String("transport", defaultTransport, "output transport: kafka or zmq")
	zmqEndpoint := flag.String("zmq-endpoint", defaultZMQAddress, "ZMQ endpoint when transport=zmq")
	skipHeader := flag.Bool("skip-header", true, "skip first line as CSV header")
	flag.Parse()

	log.Printf("csv2kafka starting | transport=%s", *transport)

	// 初始化发送器
	sender, err := buildSender(*transport, *topic, *zmqEndpoint)
	if err != nil {
		log.Fatalf("sender init failed: %v", err)
	}

	// 从标准输入读取并发送
	processStdin(sender, *skipHeader)
}

// processStdin 从标准输入逐行读取 CSV 并发送到目标后端。
func processStdin(sender LineSender, skipHeader bool) {
	scanner := bufio.NewScanner(os.Stdin)
	lineNo := 0
	sentCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNo++

		// 跳过 header
		if skipHeader && lineNo == 1 {
			log.Printf("Skipping header: %s", line)
			continue
		}

		// 跳过空行
		if len(line) == 0 {
			continue
		}

		sender.SendLine(line)
		sentCount++
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("read from stdin error: %v", err)
	}

	log.Printf("Processed %d lines, sent %d lines to %s", lineNo, sentCount, "backend")
}

func buildSender(transport, topic, zmqEndpoint string) (LineSender, error) {
	switch transport {
	case "kafka":
		log.Printf("Using Kafka transport | brokers=%v topic=%s", KafkaBrokers, topic)
		return NewKafkaSender(KafkaBrokers, topic)
	case "zmq":
		log.Printf("Using ZMQ transport | endpoint=%s", zmqEndpoint)
		return NewZMQSender(zmqEndpoint)
	default:
		return nil, fmt.Errorf("unsupported transport %s", transport)
	}
}
