package converter

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"
)

type VideoConverter struct {
}

type VideoTask struct {
	VideoID int    `json:"video_id"`
	Path    string `json:"path"`
}

func (vc *VideoConverter) Handle(msg []byte) error {
	var task VideoTask
	err := json.Unmarshal(msg, &task)

	if err != nil {
		vc.logError(task, "failed to unmarshal task", err)
	}

	return nil
}

func (vc *VideoConverter) processVideo(task *VideoTask) error {
  mergedFile := filepath.Join(task.Path, "merged.mp4")
  fmt.Println(mergedFile)
  return nil
}

func (vc *VideoConverter) logError(task VideoTask, message string, err error) {
	errorData := map[string]any{
		"video_id": task.VideoID,
		"error":    message,
		"details":  err.Error(),
		"time":     time.Now(),
	}

	serializedError, err := json.Marshal(errorData)

	if err != nil {
		slog.Error(fmt.Sprintf("Failed to unmarshal error: %v", err))
	}

	slog.Error("Processing error", slog.String("error_details", string(serializedError)))

}
