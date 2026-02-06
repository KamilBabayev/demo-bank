package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"notification-service/models"
	"notification-service/repository"

	"github.com/segmentio/kafka-go"
)

const (
	TopicTransferCompleted = "transfer.completed"
	TopicTransferFailed    = "transfer.failed"
	TopicPaymentCompleted  = "payment.completed"
	TopicPaymentFailed     = "payment.failed"
)

type Consumer struct {
	transferCompletedReader *kafka.Reader
	transferFailedReader    *kafka.Reader
	paymentCompletedReader  *kafka.Reader
	paymentFailedReader     *kafka.Reader
	repo                    *repository.NotificationRepository
	// In a real system, we would have a user lookup service
	// For now, we'll simulate with placeholder user IDs
}

func NewConsumer(brokers []string, groupID string, repo *repository.NotificationRepository) *Consumer {
	transferCompletedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       TopicTransferCompleted,
		GroupID:     groupID,
		MinBytes:    10e3,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})

	transferFailedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       TopicTransferFailed,
		GroupID:     groupID,
		MinBytes:    10e3,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})

	paymentCompletedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       TopicPaymentCompleted,
		GroupID:     groupID,
		MinBytes:    10e3,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})

	paymentFailedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       TopicPaymentFailed,
		GroupID:     groupID,
		MinBytes:    10e3,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})

	return &Consumer{
		transferCompletedReader: transferCompletedReader,
		transferFailedReader:    transferFailedReader,
		paymentCompletedReader:  paymentCompletedReader,
		paymentFailedReader:     paymentFailedReader,
		repo:                    repo,
	}
}

// Start starts consuming messages from all topics
func (c *Consumer) Start(ctx context.Context) {
	go c.consumeTransferCompleted(ctx)
	go c.consumeTransferFailed(ctx)
	go c.consumePaymentCompleted(ctx)
	go c.consumePaymentFailed(ctx)
}

func (c *Consumer) consumeTransferCompleted(ctx context.Context) {
	log.Println("Starting transfer.completed consumer for notifications")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.transferCompletedReader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Error fetching transfer completed message: %v", err)
				continue
			}

			var event models.TransferResultEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Error unmarshaling transfer completed event: %v", err)
				c.transferCompletedReader.CommitMessages(ctx, msg)
				continue
			}

			log.Printf("Creating notifications for transfer %d completed (from user %d to user %d)",
				event.TransferID, event.FromUserID, event.ToUserID)

			metadata := map[string]interface{}{
				"transfer_id":  event.TransferID,
				"reference_id": event.ReferenceID,
			}

			// Create notification for sender
			if event.FromUserID > 0 {
				_, err = c.repo.CreateFromEvent(ctx,
					event.FromUserID,
					models.NotificationTypeTransferSent,
					models.ChannelEmail,
					"Transfer Completed",
					fmt.Sprintf("Your transfer (ref: %s) has been completed successfully.", event.ReferenceID),
					metadata,
				)
				if err != nil {
					log.Printf("Error creating sender notification: %v", err)
				}
				c.simulateSendNotification("email", fmt.Sprintf("Transfer completed notification for user %d", event.FromUserID))
			}

			// Create notification for receiver
			if event.ToUserID > 0 {
				_, err = c.repo.CreateFromEvent(ctx,
					event.ToUserID,
					models.NotificationTypeTransferReceived,
					models.ChannelEmail,
					"Transfer Received",
					fmt.Sprintf("You have received a transfer (ref: %s).", event.ReferenceID),
					metadata,
				)
				if err != nil {
					log.Printf("Error creating receiver notification: %v", err)
				}
				c.simulateSendNotification("email", fmt.Sprintf("Transfer received notification for user %d", event.ToUserID))
			}

			c.transferCompletedReader.CommitMessages(ctx, msg)
		}
	}
}

func (c *Consumer) consumeTransferFailed(ctx context.Context) {
	log.Println("Starting transfer.failed consumer for notifications")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.transferFailedReader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Error fetching transfer failed message: %v", err)
				continue
			}

			var event models.TransferResultEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Error unmarshaling transfer failed event: %v", err)
				c.transferFailedReader.CommitMessages(ctx, msg)
				continue
			}

			log.Printf("Creating notification for transfer %d failed (user %d): %s",
				event.TransferID, event.FromUserID, event.FailureReason)

			metadata := map[string]interface{}{
				"transfer_id":    event.TransferID,
				"reference_id":   event.ReferenceID,
				"failure_reason": event.FailureReason,
			}

			// Create notification for sender about the failure
			if event.FromUserID > 0 {
				_, err = c.repo.CreateFromEvent(ctx,
					event.FromUserID,
					models.NotificationTypeTransferFailed,
					models.ChannelEmail,
					"Transfer Failed",
					fmt.Sprintf("Your transfer (ref: %s) has failed: %s", event.ReferenceID, event.FailureReason),
					metadata,
				)
				if err != nil {
					log.Printf("Error creating failed transfer notification: %v", err)
				}
				c.simulateSendNotification("email", fmt.Sprintf("Transfer failed notification for user %d", event.FromUserID))
			}

			c.transferFailedReader.CommitMessages(ctx, msg)
		}
	}
}

func (c *Consumer) consumePaymentCompleted(ctx context.Context) {
	log.Println("Starting payment.completed consumer for notifications")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.paymentCompletedReader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Error fetching payment completed message: %v", err)
				continue
			}

			var event models.PaymentResultEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Error unmarshaling payment completed event: %v", err)
				c.paymentCompletedReader.CommitMessages(ctx, msg)
				continue
			}

			log.Printf("Creating notification for payment %d completed", event.PaymentID)

			metadata := map[string]interface{}{
				"payment_id":   event.PaymentID,
				"reference_id": event.ReferenceID,
			}

			_, err = c.repo.CreateFromEvent(ctx,
				1, // Placeholder: would be payer's user_id
				models.NotificationTypePaymentProcessed,
				models.ChannelEmail,
				"Payment Processed",
				fmt.Sprintf("Your payment (ref: %s) has been processed successfully.", event.ReferenceID),
				metadata,
			)
			if err != nil {
				log.Printf("Error creating payment notification: %v", err)
			}

			c.simulateSendNotification("email", "Payment processed notification")

			c.paymentCompletedReader.CommitMessages(ctx, msg)
		}
	}
}

func (c *Consumer) consumePaymentFailed(ctx context.Context) {
	log.Println("Starting payment.failed consumer for notifications")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.paymentFailedReader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Error fetching payment failed message: %v", err)
				continue
			}

			var event models.PaymentResultEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Error unmarshaling payment failed event: %v", err)
				c.paymentFailedReader.CommitMessages(ctx, msg)
				continue
			}

			log.Printf("Creating notification for payment %d failed: %s", event.PaymentID, event.FailureReason)

			metadata := map[string]interface{}{
				"payment_id":     event.PaymentID,
				"reference_id":   event.ReferenceID,
				"failure_reason": event.FailureReason,
			}

			_, err = c.repo.CreateFromEvent(ctx,
				1, // Placeholder: would be payer's user_id
				models.NotificationTypePaymentFailed,
				models.ChannelEmail,
				"Payment Failed",
				fmt.Sprintf("Your payment (ref: %s) has failed: %s", event.ReferenceID, event.FailureReason),
				metadata,
			)
			if err != nil {
				log.Printf("Error creating failed payment notification: %v", err)
			}

			c.simulateSendNotification("email", "Payment failed notification")

			c.paymentFailedReader.CommitMessages(ctx, msg)
		}
	}
}

// simulateSendNotification simulates sending a notification via a channel
func (c *Consumer) simulateSendNotification(channel, message string) {
	log.Printf("[SIMULATED %s] Sending: %s", channel, message)
	// In a real system, this would call an email/SMS service
}

// Close closes all readers
func (c *Consumer) Close() error {
	if err := c.transferCompletedReader.Close(); err != nil {
		return err
	}
	if err := c.transferFailedReader.Close(); err != nil {
		return err
	}
	if err := c.paymentCompletedReader.Close(); err != nil {
		return err
	}
	return c.paymentFailedReader.Close()
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
