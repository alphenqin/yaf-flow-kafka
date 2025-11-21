package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	defaultWatchDir = "/host/output"
	defaultTopic    = "yaf_csv"
)

var KafkaBrokers = []string{"localhost:9092"}

func main() {
	watchDir := flag.String("watchdir", defaultWatchDir, "directory to watch for CSV files")
	topic := flag.String("topic", defaultTopic, "Kafka topic to send CSV lines")
	flag.Parse()

	log.Printf("csv2kafka starting | watchdir=%s topic=%s brokers=%v", *watchDir, *topic, KafkaBrokers)

	// 初始化 Kafka 生产者
	sender, err := NewKafkaSender(KafkaBrokers, *topic)
	if err != nil {
		log.Fatalf("Kafka init failed: %v", err)
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

// sendCSV 逐行读取 CSV 并发送到 Kafka。
func sendCSV(path string, sender *KafkaSender) {
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
	fmt.Printf("Sent %d lines from %s to Kafka\n", lineNo-1, path)
}
