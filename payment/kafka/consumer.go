package kafka

import (
	"context"
	"encoding/json"
	"log"

	"payment/models"
	"payment/repository"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	completedReader *kafka.Reader
	failedReader    *kafka.Reader
	repo            *repository.PaymentRepository
}

func NewConsumer(brokers []string, groupID string, repo *repository.PaymentRepository) *Consumer {
	completedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       TopicPaymentCompleted,
		GroupID:     groupID,
		MinBytes:    10e3,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})

	failedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       TopicPaymentFailed,
		GroupID:     groupID,
		MinBytes:    10e3,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})

	return &Consumer{
		completedReader: completedReader,
		failedReader:    failedReader,
		repo:            repo,
	}
}

// Start starts consuming messages from both topics
func (c *Consumer) Start(ctx context.Context) {
	go c.consumeCompleted(ctx)
	go c.consumeFailed(ctx)
}

func (c *Consumer) consumeCompleted(ctx context.Context) {
	log.Println("Starting payment.completed consumer")
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping payment.completed consumer")
			return
		default:
			msg, err := c.completedReader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Error fetching completed message: %v", err)
				continue
			}

			var event models.PaymentResultEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Error unmarshaling completed event: %v", err)
				c.completedReader.CommitMessages(ctx, msg)
				continue
			}

			log.Printf("Received payment.completed event for payment %d", event.PaymentID)

			_, err = c.repo.MarkAsCompleted(ctx, event.PaymentID)
			if err != nil {
				log.Printf("Error marking payment %d as completed: %v", event.PaymentID, err)
			} else {
				log.Printf("Payment %d marked as completed", event.PaymentID)
			}

			c.completedReader.CommitMessages(ctx, msg)
		}
	}
}

func (c *Consumer) consumeFailed(ctx context.Context) {
	log.Println("Starting payment.failed consumer")
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping payment.failed consumer")
			return
		default:
			msg, err := c.failedReader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Error fetching failed message: %v", err)
				continue
			}

			var event models.PaymentResultEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Error unmarshaling failed event: %v", err)
				c.failedReader.CommitMessages(ctx, msg)
				continue
			}

			log.Printf("Received payment.failed event for payment %d: %s", event.PaymentID, event.FailureReason)

			_, err = c.repo.MarkAsFailed(ctx, event.PaymentID, event.FailureReason)
			if err != nil {
				log.Printf("Error marking payment %d as failed: %v", event.PaymentID, err)
			} else {
				log.Printf("Payment %d marked as failed", event.PaymentID)
			}

			c.failedReader.CommitMessages(ctx, msg)
		}
	}
}

// Close closes both readers
func (c *Consumer) Close() error {
	if err := c.completedReader.Close(); err != nil {
		return err
	}
	return c.failedReader.Close()
}
