package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"account/models"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	completedWriter *kafka.Writer
	failedWriter    *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	completedWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        TopicTransferCompleted,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	failedWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        TopicTransferFailed,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	return &Producer{
		completedWriter: completedWriter,
		failedWriter:    failedWriter,
	}
}

// PublishTransferCompleted publishes a transfer completed event
func (p *Producer) PublishTransferCompleted(ctx context.Context, event models.TransferResultEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.ReferenceID),
		Value: value,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte("transfer.completed")},
			{Key: "transfer_id", Value: []byte(fmt.Sprintf("%d", event.TransferID))},
		},
	}

	if err := p.completedWriter.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Published transfer.completed event for transfer %d", event.TransferID)
	return nil
}

// PublishTransferFailed publishes a transfer failed event
func (p *Producer) PublishTransferFailed(ctx context.Context, event models.TransferResultEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.ReferenceID),
		Value: value,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte("transfer.failed")},
			{Key: "transfer_id", Value: []byte(fmt.Sprintf("%d", event.TransferID))},
		},
	}

	if err := p.failedWriter.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Published transfer.failed event for transfer %d: %s", event.TransferID, event.FailureReason)
	return nil
}

// Close closes both writers
func (p *Producer) Close() error {
	if err := p.completedWriter.Close(); err != nil {
		return err
	}
	return p.failedWriter.Close()
}

// EnsureTopicExists creates the topic if it doesn't exist
func EnsureTopicExists(brokers []string, topic string) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     3,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		log.Printf("Topic creation result for %s: %v", topic, err)
	}

	return nil
}
