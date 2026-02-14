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
	completedWriter        *kafka.Writer
	failedWriter           *kafka.Writer
	paymentCompletedWriter *kafka.Writer
	paymentFailedWriter    *kafka.Writer
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

	paymentCompletedWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        TopicPaymentCompleted,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	paymentFailedWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        TopicPaymentFailed,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	return &Producer{
		completedWriter:        completedWriter,
		failedWriter:           failedWriter,
		paymentCompletedWriter: paymentCompletedWriter,
		paymentFailedWriter:    paymentFailedWriter,
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

// PublishPaymentCompleted publishes a payment completed event
func (p *Producer) PublishPaymentCompleted(ctx context.Context, event models.PaymentResultEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.ReferenceID),
		Value: value,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte("payment.completed")},
			{Key: "payment_id", Value: []byte(fmt.Sprintf("%d", event.PaymentID))},
		},
	}

	if err := p.paymentCompletedWriter.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Published payment.completed event for payment %d", event.PaymentID)
	return nil
}

// PublishPaymentFailed publishes a payment failed event
func (p *Producer) PublishPaymentFailed(ctx context.Context, event models.PaymentResultEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.ReferenceID),
		Value: value,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte("payment.failed")},
			{Key: "payment_id", Value: []byte(fmt.Sprintf("%d", event.PaymentID))},
		},
	}

	if err := p.paymentFailedWriter.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Published payment.failed event for payment %d: %s", event.PaymentID, event.FailureReason)
	return nil
}

// Close closes all writers
func (p *Producer) Close() error {
	if err := p.completedWriter.Close(); err != nil {
		return err
	}
	if err := p.failedWriter.Close(); err != nil {
		return err
	}
	if err := p.paymentCompletedWriter.Close(); err != nil {
		return err
	}
	return p.paymentFailedWriter.Close()
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
