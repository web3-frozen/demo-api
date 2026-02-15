package queue

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/web3-frozen/demo-api/internal/model"
)

type KafkaProducer struct {
	writer *kafka.Writer
	logger *slog.Logger
}

func NewKafkaProducer(brokers string, topic string, logger *slog.Logger) *KafkaProducer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		Async:        true,
	}
	return &KafkaProducer{writer: w, logger: logger}
}

func (p *KafkaProducer) PublishEvent(ctx context.Context, eventType string, taskID string, data any) {
	event := model.TaskEvent{
		Type:      eventType,
		TaskID:    taskID,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
	payload, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("failed to marshal event", "error", err)
		return
	}
	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(taskID),
		Value: payload,
	})
	if err != nil {
		p.logger.Warn("failed to publish event", "error", err, "type", eventType)
	}
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
