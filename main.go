package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	defaultWatchDir   = "/host/output"
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
	watchDir := flag.String("watchdir", defaultWatchDir, "directory to watch for CSV files")
	topic := flag.String("topic", defaultTopic, "Kafka topic to send CSV lines")
	transport := flag.String("transport", defaultTransport, "output transport: kafka or zmq")
	zmqEndpoint := flag.String("zmq-endpoint", defaultZMQAddress, "ZMQ endpoint when transport=zmq")
	flag.Parse()

	log.Printf("csv2kafka starting | watchdir=%s transport=%s", *watchDir, *transport)

	// 初始化发送器
	sender, err := buildSender(*transport, *topic, *zmqEndpoint)
	if err != nil {
		log.Fatalf("sender init failed: %v", err)
	}

	// 启动目录监控
	err = WatchDir(*watchDir, func(path string) {
		if !isCSV(path) {
			return
		}
		log.Println("New CSV ready:", path)
		sendCSV(path, sender)
	})
	if err != nil {
		log.Fatal("watch dir error:", err)
	}
	log.Printf("watch established on %s", *watchDir)

	// 阻塞主协程
	select {}
}

func isCSV(path string) bool {
	return len(path) > 4 && path[len(path)-4:] == ".csv"
}

// sendCSV 逐行读取 CSV 并发送到目标后端。
func sendCSV(path string, sender LineSender) {
	f, err := os.Open(path)
	if err != nil {
		log.Println("open csv error:", err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		line := scanner.Text()
		if lineNo == 0 {
			lineNo++
			continue // 跳过 header
		}
		sender.SendLine(line)
		lineNo++
	}
	if err := scanner.Err(); err != nil {
		log.Println("read csv error:", err)
	}
	fmt.Printf("Sent %d lines from %s\n", lineNo-1, path)
}

func buildSender(transport, topic, zmqEndpoint string) (LineSender, error) {
	switch transport {
	case "kafka":
		return NewKafkaSender(KafkaBrokers, topic)
	case "zmq":
		return NewZMQSender(zmqEndpoint)
	default:
		return nil, fmt.Errorf("unsupported transport %s", transport)
	}
}
