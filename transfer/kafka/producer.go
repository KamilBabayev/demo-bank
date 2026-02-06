package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"transfer/models"

	"github.com/segmentio/kafka-go"
)

const (
	TopicTransferRequested = "transfer.requested"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        TopicTransferRequested,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		Async:        false, // Synchronous writes for reliability
	}

	return &Producer{writer: writer}
}

// PublishTransferRequested publishes a transfer requested event
func (p *Producer) PublishTransferRequested(ctx context.Context, transfer *models.Transfer) error {
	event := models.TransferRequestedEvent{
		TransferID:    transfer.ID,
		ReferenceID:   transfer.ReferenceID.String(),
		FromAccountID: transfer.FromAccountID,
		ToAccountID:   transfer.ToAccountID,
		Amount:        transfer.Amount,
		Currency:      transfer.Currency,
	}

	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(transfer.ReferenceID.String()),
		Value: value,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte("transfer.requested")},
			{Key: "transfer_id", Value: []byte(fmt.Sprintf("%d", transfer.ID))},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Published transfer.requested event for transfer %d (ref: %s)", transfer.ID, transfer.ReferenceID)
	return nil
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.writer.Close()
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
		// Topic might already exist, that's okay
		log.Printf("Topic creation result for %s: %v", topic, err)
	}

	return nil
}
