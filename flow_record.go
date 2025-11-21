package main

// FlowRecord 可选的字段结构，当前示例未使用，但保留以便未来扩展。
type FlowRecord struct {
	StartTime string
	EndTime   string
	SrcIP     string
	SrcPort   string
	DstIP     string
	DstPort   string
	Protocol  string
	Packets   string
	Bytes     string
	TCPFlags  string
	AppLabel  string
	Duration  string
}
