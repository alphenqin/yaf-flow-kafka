package main

import (
	"log"

	"github.com/pebbe/zmq4"
)

// ZMQSender 使用 PUSH 模式向指定 endpoint 发送文本行。
type ZMQSender struct {
	socket *zmq4.Socket
}

func NewZMQSender(endpoint string) (*ZMQSender, error) {
	socket, err := zmq4.NewSocket(zmq4.PUSH)
	if err != nil {
		return nil, err
	}
	if err := socket.Connect(endpoint); err != nil {
		socket.Close()
		return nil, err
	}
	return &ZMQSender{socket: socket}, nil
}

func (zs *ZMQSender) SendLine(line string) {
	if _, err := zs.socket.Send(line, 0); err != nil {
		log.Printf("ZMQ send error: %v", err)
	}
}

