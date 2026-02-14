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
	TopicPaymentRequested  = "payment.requested"
	TopicPaymentCompleted  = "payment.completed"
	TopicPaymentFailed     = "payment.failed"
)

type Consumer struct {
	transferReader *kafka.Reader
	paymentReader  *kafka.Reader
	repo           repository.AccountRepo
	producer       *Producer
}

func NewConsumer(brokers []string, groupID string, repo repository.AccountRepo, producer *Producer) *Consumer {
	transferReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       TopicTransferRequested,
		GroupID:     groupID,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		StartOffset: kafka.FirstOffset,
	})

	paymentReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       TopicPaymentRequested,
		GroupID:     groupID + "-payments",
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		StartOffset: kafka.FirstOffset,
	})

	return &Consumer{
		transferReader: transferReader,
		paymentReader:  paymentReader,
		repo:           repo,
		producer:       producer,
	}
}

// Start starts consuming transfer and payment requested events
func (c *Consumer) Start(ctx context.Context) {
	go c.consumeTransfers(ctx)
	go c.consumePayments(ctx)
}

func (c *Consumer) consumeTransfers(ctx context.Context) {
	log.Println("Starting transfer.requested consumer")
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping transfer.requested consumer")
			return
		default:
			msg, err := c.transferReader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Error fetching transfer message: %v", err)
				continue
			}

			c.processTransferMessage(ctx, msg)
			c.transferReader.CommitMessages(ctx, msg)
		}
	}
}

func (c *Consumer) consumePayments(ctx context.Context) {
	log.Println("Starting payment.requested consumer")
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping payment.requested consumer")
			return
		default:
			msg, err := c.paymentReader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Error fetching payment message: %v", err)
				continue
			}

			c.processPaymentMessage(ctx, msg)
			c.paymentReader.CommitMessages(ctx, msg)
		}
	}
}

func (c *Consumer) processTransferMessage(ctx context.Context, msg kafka.Message) {
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

func (c *Consumer) processPaymentMessage(ctx context.Context, msg kafka.Message) {
	var event models.PaymentEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("Error unmarshaling payment event: %v", err)
		return
	}

	log.Printf("Processing payment %d (ref: %s): account %d, amount: %s, type: %s",
		event.PaymentID, event.ReferenceID,
		event.AccountID, event.Amount.String(), event.PaymentType)

	// Get account info
	account, err := c.repo.GetByID(ctx, event.AccountID)
	if err != nil {
		log.Printf("Payment %d failed: account not found: %v", event.PaymentID, err)
		result := models.PaymentResultEvent{
			PaymentID:     event.PaymentID,
			ReferenceID:   event.ReferenceID,
			Status:        "failed",
			FailureReason: "account not found",
			AccountID:     event.AccountID,
			UserID:        event.UserID,
		}
		c.producer.PublishPaymentFailed(ctx, result)
		return
	}

	// Perform the withdrawal (debit from account)
	_, err = c.repo.Withdraw(ctx, event.AccountID, event.Amount)

	var result models.PaymentResultEvent
	result.PaymentID = event.PaymentID
	result.ReferenceID = event.ReferenceID
	result.AccountID = event.AccountID
	result.UserID = account.UserID

	if err != nil {
		result.Status = "failed"
		result.FailureReason = err.Error()
		log.Printf("Payment %d failed: %v", event.PaymentID, err)

		// Publish failure event
		if pubErr := c.producer.PublishPaymentFailed(ctx, result); pubErr != nil {
			log.Printf("Failed to publish payment.failed event: %v", pubErr)
		}
	} else {
		result.Status = "completed"
		log.Printf("Payment %d completed successfully (debited %s from account %d)",
			event.PaymentID, event.Amount.String(), event.AccountID)

		// Publish success event
		if pubErr := c.producer.PublishPaymentCompleted(ctx, result); pubErr != nil {
			log.Printf("Failed to publish payment.completed event: %v", pubErr)
		}
	}
}

func formatAccountID(id int64) string {
	return decimal.NewFromInt(id).String()
}

// Close closes both readers
func (c *Consumer) Close() error {
	if err := c.transferReader.Close(); err != nil {
		return err
	}
	return c.paymentReader.Close()
}
