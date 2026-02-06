package kafka

import (
	"context"
	"encoding/json"
	"log"

	"transfer/models"
	"transfer/repository"

	"github.com/segmentio/kafka-go"
)

const (
	TopicTransferCompleted = "transfer.completed"
	TopicTransferFailed    = "transfer.failed"
)

type Consumer struct {
	completedReader *kafka.Reader
	failedReader    *kafka.Reader
	repo            *repository.TransferRepository
}

func NewConsumer(brokers []string, groupID string, repo *repository.TransferRepository) *Consumer {
	completedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       TopicTransferCompleted,
		GroupID:     groupID,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		StartOffset: kafka.FirstOffset,
	})

	failedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       TopicTransferFailed,
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
	log.Println("Starting transfer.completed consumer")
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping transfer.completed consumer")
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

			var event models.TransferResultEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Error unmarshaling completed event: %v", err)
				c.completedReader.CommitMessages(ctx, msg)
				continue
			}

			log.Printf("Received transfer.completed event for transfer %d", event.TransferID)

			_, err = c.repo.MarkAsCompleted(ctx, event.TransferID)
			if err != nil {
				log.Printf("Error marking transfer %d as completed: %v", event.TransferID, err)
			} else {
				log.Printf("Transfer %d marked as completed", event.TransferID)
			}

			c.completedReader.CommitMessages(ctx, msg)
		}
	}
}

func (c *Consumer) consumeFailed(ctx context.Context) {
	log.Println("Starting transfer.failed consumer")
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping transfer.failed consumer")
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

			var event models.TransferResultEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Error unmarshaling failed event: %v", err)
				c.failedReader.CommitMessages(ctx, msg)
				continue
			}

			log.Printf("Received transfer.failed event for transfer %d: %s", event.TransferID, event.FailureReason)

			_, err = c.repo.MarkAsFailed(ctx, event.TransferID, event.FailureReason)
			if err != nil {
				log.Printf("Error marking transfer %d as failed: %v", event.TransferID, err)
			} else {
				log.Printf("Transfer %d marked as failed", event.TransferID)
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
