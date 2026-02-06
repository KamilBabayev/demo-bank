package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"card/models"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	createdWriter   *kafka.Writer
	blockedWriter   *kafka.Writer
	activatedWriter *kafka.Writer
	cancelledWriter *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	createdWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        TopicCardCreated,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	blockedWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        TopicCardBlocked,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	activatedWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        TopicCardActivated,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	cancelledWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        TopicCardCancelled,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	return &Producer{
		createdWriter:   createdWriter,
		blockedWriter:   blockedWriter,
		activatedWriter: activatedWriter,
		cancelledWriter: cancelledWriter,
	}
}

// PublishCardCreated publishes a card created event
func (p *Producer) PublishCardCreated(ctx context.Context, event models.CardEvent) error {
	event.EventType = "created"
	return p.publishEvent(ctx, p.createdWriter, event)
}

// PublishCardBlocked publishes a card blocked event
func (p *Producer) PublishCardBlocked(ctx context.Context, event models.CardEvent) error {
	event.EventType = "blocked"
	return p.publishEvent(ctx, p.blockedWriter, event)
}

// PublishCardActivated publishes a card activated event
func (p *Producer) PublishCardActivated(ctx context.Context, event models.CardEvent) error {
	event.EventType = "activated"
	return p.publishEvent(ctx, p.activatedWriter, event)
}

// PublishCardCancelled publishes a card cancelled event
func (p *Producer) PublishCardCancelled(ctx context.Context, event models.CardEvent) error {
	event.EventType = "cancelled"
	return p.publishEvent(ctx, p.cancelledWriter, event)
}

func (p *Producer) publishEvent(ctx context.Context, writer *kafka.Writer, event models.CardEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("%d", event.CardID)),
		Value: value,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "card_id", Value: []byte(fmt.Sprintf("%d", event.CardID))},
		},
	}

	if err := writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Published %s event for card %d", event.EventType, event.CardID)
	return nil
}

// Close closes all writers
func (p *Producer) Close() error {
	if err := p.createdWriter.Close(); err != nil {
		return err
	}
	if err := p.blockedWriter.Close(); err != nil {
		return err
	}
	if err := p.activatedWriter.Close(); err != nil {
		return err
	}
	return p.cancelledWriter.Close()
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
