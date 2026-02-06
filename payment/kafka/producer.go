package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"payment/models"

	"github.com/segmentio/kafka-go"
)

const (
	TopicPaymentRequested = "payment.requested"
	TopicPaymentCompleted = "payment.completed"
	TopicPaymentFailed    = "payment.failed"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        TopicPaymentRequested,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	return &Producer{writer: writer}
}

// PublishPaymentRequested publishes a payment requested event
func (p *Producer) PublishPaymentRequested(ctx context.Context, payment *models.Payment) error {
	recipientName := ""
	if payment.RecipientName != nil {
		recipientName = *payment.RecipientName
	}
	recipientAccount := ""
	if payment.RecipientAccount != nil {
		recipientAccount = *payment.RecipientAccount
	}

	event := models.PaymentRequestedEvent{
		PaymentID:        payment.ID,
		ReferenceID:      payment.ReferenceID.String(),
		AccountID:        payment.AccountID,
		UserID:           payment.UserID,
		PaymentType:      payment.PaymentType,
		RecipientName:    recipientName,
		RecipientAccount: recipientAccount,
		Amount:           payment.Amount,
		Currency:         payment.Currency,
	}

	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(payment.ReferenceID.String()),
		Value: value,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte("payment.requested")},
			{Key: "payment_id", Value: []byte(fmt.Sprintf("%d", payment.ID))},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Published payment.requested event for payment %d (ref: %s)", payment.ID, payment.ReferenceID)
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
		log.Printf("Topic creation result for %s: %v", topic, err)
	}

	return nil
}
