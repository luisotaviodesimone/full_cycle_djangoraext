package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"videoconverter/internal/converter"

  _ "github.com/lib/pq"
)

func main() {
	db, err := connectPostgres()

	if err != nil {
    panic(err)
	}

	vc := converter.New(db)

	vc.Handle([]byte(`{"video_id": 1, "path": "/media/uploads/1"}`))
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}

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
