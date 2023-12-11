package pulsar

import (
	"context"
	"os"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/pkg/errors"
)

// Consumer pulsar消费者客户端
type Consumer struct {
	ctx            context.Context             // 上下文
	cancel         func()                      // 取消函数
	baseClient     pulsar.Client               // 原始client
	consumer       pulsar.Consumer             // 消费者
	messageChannel chan pulsar.ConsumerMessage // 存放消息通道
}

// Close .
func (c *Consumer) Close() error {
	c.cancel()
	c.consumer.Close()
	c.baseClient.Close()
	return nil
}

// ReceiverChannel 接收通道内数据
func (c *Consumer) ReceiverChannel(out chan<- []byte) {
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return
			case msg := <-c.messageChannel:
				out <- msg.Payload()
			}
		}
	}()
}

// NewConsumer .
// url pulsar://localhost:6600,localhost:6650
// topic 主题
func NewConsumer(url string, topic string) (*Consumer, error) {
	c, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               url,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		return nil, errors.Wrap(err, "创建pulsar连接失败")
	}
	hostname, err := os.Hostname()
	if err != nil {
		return nil, errors.Wrap(err, "获取主机名失败")
	}
	mc := make(chan pulsar.ConsumerMessage, 1)
	cs, err := c.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: hostname,
		Type:             pulsar.Shared,
		MessageChannel:   mc,
	})
	if err != nil {
		return nil, errors.Wrap(err, "创建消费者失败")
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Consumer{
		ctx:            ctx,
		cancel:         cancel,
		baseClient:     c,
		consumer:       cs,
		messageChannel: mc,
	}, nil
}
