package kafka

import (
	"context"
	"log"
	"time"

	"cat-uploader-go/internal/storage/postgres"
	kafkago "github.com/segmentio/kafka-go"
)

// PublishOutbox retries unsent database events until Kafka acknowledges them.
func PublishOutbox(ctx context.Context, repo *postgres.PhotoRepository, brokers []string, topic string) {
	writer := &kafkago.Writer{Addr: kafkago.TCP(brokers...), Topic: topic, Balancer: &kafkago.Hash{}, RequiredAcks: kafkago.RequireAll}
	defer writer.Close()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		publishBatch(ctx, repo, writer)
		select { case <-ctx.Done(): return; case <-ticker.C: }
	}
}

func publishBatch(ctx context.Context, repo *postgres.PhotoRepository, writer *kafkago.Writer) {
	events, err := repo.PendingOutbox(ctx, 100)
	if err != nil { log.Printf("read Kafka outbox: %v", err); return }
	for _, event := range events {
		if err := writer.WriteMessages(ctx, kafkago.Message{Key: []byte(event.JobID), Value: []byte(event.JobID)}); err != nil {
			log.Printf("publish upload job %s: %v", event.JobID, err); return
		}
		if err := repo.MarkOutboxPublished(ctx, event.ID); err != nil { log.Printf("mark outbox event %d published: %v", event.ID, err); return }
	}
}
