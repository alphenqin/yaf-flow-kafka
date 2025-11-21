package main

import (
	"log"

	"github.com/Shopify/sarama"
)

// KafkaSender 封装异步 Kafka 生产者。
type KafkaSender struct {
	producer sarama.AsyncProducer
	topic    string
}

func NewKafkaSender(brokers []string, topic string) (*KafkaSender, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = false
	config.Producer.RequiredAcks = sarama.WaitForLocal

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	// 持续读取错误通道，避免阻塞。
	go func() {
		for err := range producer.Errors() {
			log.Printf("Kafka producer error: %v", err)
		}
	}()

	return &KafkaSender{
		producer: producer,
		topic:    topic,
	}, nil
}

// SendLine 将单行数据发送到 Kafka。
func (ks *KafkaSender) SendLine(line string) {
	msg := &sarama.ProducerMessage{
		Topic: ks.topic,
		Value: sarama.StringEncoder(line),
	}
	ks.producer.Input() <- msg
}
