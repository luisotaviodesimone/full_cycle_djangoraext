package converter

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

func IsProcessed(db *sql.DB, videoID int) bool {
	var isProcessed bool
	query := `SELECT EXISTS(SELECT 1 FROM processed_videos WHERE video_id = $1 AND status = 'success')`
	err := db.QueryRow(query, videoID).Scan(&isProcessed)

	if err != nil {
		slog.Error("failed to check if video is processed", slog.Int("video_id", videoID), slog.String("error", err.Error()))
		return false
	}

	return isProcessed
}

func MarkAsProcessed(db *sql.DB, videoID int) error {
	query := `INSERT INTO processed_videos (video_id, status, processed_at) VALUES ($1, $2, $3)`
	_, err := db.Exec(query, videoID, "success", time.Now())

	if err != nil {
		slog.Error("failed to mark video as processed", slog.Int("video_id", videoID), slog.String("error", err.Error()))
		return err
	}

	return nil
}

func RegisterError(db *sql.DB, errorData map[string]interface{}, err error) {
	serializedError, err := json.Marshal(errorData)

	if err != nil {
		slog.Error("failed to unmarshal error", slog.String("error", err.Error()))
	}

	query := `INSERT INTO process_errors_log (error_details, created_at) VALUES ($1, $2)`

	fmt.Println("serializedError", string(serializedError))
	_, dbErr := db.Exec(query, serializedError, time.Now())

	if dbErr != nil {
		slog.Error("error registering error", slog.String("errors_details", string(serializedError)), slog.String("error", err.Error()))
	}
}
