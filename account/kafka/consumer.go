package kafka

import (
	"context"
	"encoding/json"
	"log"

	"account/models"
	"account/repository"

	"github.com/segmentio/kafka-go"
	"github.com/shopspring/decimal"
)

const (
	TopicTransferRequested = "transfer.requested"
	TopicTransferCompleted = "transfer.completed"
	TopicTransferFailed    = "transfer.failed"
)

type Consumer struct {
	reader   *kafka.Reader
	repo     *repository.AccountRepository
	producer *Producer
}

func NewConsumer(brokers []string, groupID string, repo *repository.AccountRepository, producer *Producer) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       TopicTransferRequested,
		GroupID:     groupID,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		StartOffset: kafka.FirstOffset,
	})

	return &Consumer{
		reader:   reader,
		repo:     repo,
		producer: producer,
	}
}

// Start starts consuming transfer requested events
func (c *Consumer) Start(ctx context.Context) {
	log.Println("Starting transfer.requested consumer")
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping transfer.requested consumer")
			return
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Error fetching message: %v", err)
				continue
			}

			c.processMessage(ctx, msg)
			c.reader.CommitMessages(ctx, msg)
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) {
	var event models.TransferEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("Error unmarshaling transfer event: %v", err)
		return
	}

	log.Printf("Processing transfer %d (ref: %s): %s -> %s, amount: %s",
		event.TransferID, event.ReferenceID,
		formatAccountID(event.FromAccountID), formatAccountID(event.ToAccountID),
		event.Amount.String())

	// Get account info to include user IDs in the result event
	fromAccount, _ := c.repo.GetByID(ctx, event.FromAccountID)
	toAccount, _ := c.repo.GetByID(ctx, event.ToAccountID)

	// Perform the transfer
	err := c.repo.Transfer(ctx, event.FromAccountID, event.ToAccountID, event.Amount)

	var result models.TransferResultEvent
	result.TransferID = event.TransferID
	result.ReferenceID = event.ReferenceID
	result.FromAccountID = event.FromAccountID
	result.ToAccountID = event.ToAccountID
	if fromAccount != nil {
		result.FromUserID = fromAccount.UserID
	}
	if toAccount != nil {
		result.ToUserID = toAccount.UserID
	}

	if err != nil {
		result.Status = "failed"
		result.FailureReason = err.Error()
		log.Printf("Transfer %d failed: %v", event.TransferID, err)

		// Publish failure event
		if pubErr := c.producer.PublishTransferFailed(ctx, result); pubErr != nil {
			log.Printf("Failed to publish transfer.failed event: %v", pubErr)
		}
	} else {
		result.Status = "completed"
		log.Printf("Transfer %d completed successfully", event.TransferID)

		// Publish success event
		if pubErr := c.producer.PublishTransferCompleted(ctx, result); pubErr != nil {
			log.Printf("Failed to publish transfer.completed event: %v", pubErr)
		}
	}
}

func formatAccountID(id int64) string {
	return decimal.NewFromInt(id).String()
}

// Close closes the reader
func (c *Consumer) Close() error {
	return c.reader.Close()
}
