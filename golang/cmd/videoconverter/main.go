package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"videoconverter/internal/converter"
	"videoconverter/internal/rabbitmq"
	"videoconverter/pkg/log"

	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
)

func connectPostgres() (db *sql.DB, err error) {
	postgresHost := getEnvOrDefault("POSTGRES_HOST", "localhost")
	postgresPort := getEnvOrDefault("POSTGRES_PORT", "5432")
	postgresUser := getEnvOrDefault("POSTGRES_USER", "postgres")
	postgresPassword := getEnvOrDefault("POSTGRES_PASSWORD", "root")
	postgresDB := getEnvOrDefault("POSTGRES_DB", "converter")
	postgresSSL := getEnvOrDefault("POSTGRES_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", postgresHost, postgresPort, postgresUser, postgresPassword, postgresDB, postgresSSL)

	db, err = sql.Open("postgres", connStr)

	if err != nil {
		slog.Error("failed to connect to postgres", slog.String("error", err.Error()))
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		slog.Error("failed to ping postgres", slog.String("error", err.Error()))
		return nil, err
	}

	slog.Info("Connected to postgres successfully!")
	return db, nil
}

func main() {

	isDebug := getEnvOrDefault("DEBUG", "true") == "true"
	logger := log.NewLogger(isDebug)
	slog.SetDefault(logger)

	db, err := connectPostgres()

	if err != nil {
		panic(err)
	}

	rabbitMQURL := getEnvOrDefault("RABBITMQ_URL", "amqp://admin:admin@localhost:5672/")
	rabbitClient, err := rabbitmq.NewRabbitClient(rabbitMQURL)
	slog.Info("Connected to rabbit successfully!")
	defer rabbitClient.Close()

	if err != nil {
		panic(err)
	}

	convertionExchange := getEnvOrDefault("CONVERSION_EXCHANGE", "conversion_exchange")
	queueName := getEnvOrDefault("CONVERSION_QUEUE", "video_conversion_queue")
	conversionRoutingKey := getEnvOrDefault("CONVERSION_KEY", "conversion")
	confirmationKey := getEnvOrDefault("CONFIRMATION_KEY", "finish-conversion")
	confirmationQueue := getEnvOrDefault("CONFIRMATION_QUEUE", "video_confirmation_queue")

	vc := converter.NewVideoConverter(db, rabbitClient)
	// vc.Handle([]byte(`{"video_id": 1, "path": "/media/uploads/1"}`))

	msgs, err := rabbitClient.ConsumeMessages(convertionExchange, conversionRoutingKey, queueName)
	if err != nil {
		slog.Error("Failed to consume messages", slog.String("error", err.Error()))
	}

	for d := range msgs {
		go func(delivery amqp.Delivery) {
			vc.Handle(delivery, convertionExchange, confirmationQueue, confirmationKey)
		}(d)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}
