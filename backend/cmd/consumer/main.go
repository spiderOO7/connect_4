package main

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	brokers := []string{"localhost:9092"}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   "game-analytics",
		GroupID: "analytics-consumer",
	})
	defer reader.Close()

	log.Println("analytics consumer listening on topic game-analytics")
	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("read error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		log.Printf("event: %s", string(m.Value))
	}
}
