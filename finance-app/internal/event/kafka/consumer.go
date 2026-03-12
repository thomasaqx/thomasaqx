package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

// MessageHandler is a callback invoked for each consumed message.
type MessageHandler func(ctx context.Context, topic string, payload json.RawMessage) error

// Consumer reads messages from Kafka topics using a consumer group.
type Consumer struct {
	group   sarama.ConsumerGroup
	topics  []string
	handler MessageHandler
}

func NewConsumer(brokers []string, groupID string, topics []string, handler MessageHandler) (*Consumer, error) {
	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	group, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, fmt.Errorf("kafka consumer group: %w", err)
	}
	return &Consumer{
		group:   group,
		topics:  topics,
		handler: handler,
	}, nil
}

// Start begins consuming in a goroutine; cancel ctx to stop.
func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for {
			if err := c.group.Consume(ctx, c.topics, c); err != nil {
				log.Printf("consumer error: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()
}

// Close releases the consumer group.
func (c *Consumer) Close() error {
	return c.group.Close()
}

// --- sarama.ConsumerGroupHandler interface ---

func (c *Consumer) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (c *Consumer) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if err := c.handler(session.Context(), msg.Topic, msg.Value); err != nil {
			log.Printf("handle message [%s]: %v", msg.Topic, err)
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
