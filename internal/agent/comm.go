package agent

import (
	"context"
)

// CommChannel 通信通道
type CommChannel struct {
	ch chan Message
}

// NewCommChannel 创建通信通道
func NewCommChannel(buffer int) *CommChannel {
	return &CommChannel{
		ch: make(chan Message, buffer),
	}
}

// Send 发送消息
func (c *CommChannel) Send(msg Message) {
	c.ch <- msg
}

// Receive 接收消息（阻塞）
func (c *CommChannel) Receive(ctx context.Context) (Message, bool) {
	select {
	case msg := <-c.ch:
		return msg, true
	case <-ctx.Done():
		return Message{}, false
	}
}

// Close 关闭通道
func (c *CommChannel) Close() {
	close(c.ch)
}
