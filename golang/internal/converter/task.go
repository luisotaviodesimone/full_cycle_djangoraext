package converter

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"
	"videoconverter/internal/rabbitmq"

	"github.com/streadway/amqp"
)

type VideoConverter struct {
	db             *sql.DB
	rabbitmqClient *rabbitmq.RabbitClient
}

type VideoTask struct {
	VideoID int    `json:"video_id"`
	Path    string `json:"path"`
}

func NewVideoConverter(db *sql.DB, rabbitmqClient *rabbitmq.RabbitClient) *VideoConverter {
	return &VideoConverter{
		db:             db,
		rabbitmqClient: rabbitmqClient,
	}
}

func (vc *VideoConverter) Handle(delivery amqp.Delivery) {
  defer delivery.Ack(false)
	var task VideoTask

	if err := json.Unmarshal(delivery.Body, &task); err != nil {
		vc.logError(task, "failed to unmarshal task", err)
		return
	}

	if IsProcessed(vc.db, task.VideoID) {
		slog.Warn("Video already processed", slog.Int("video_id", task.VideoID))
		return
	}

	if err := vc.processVideo(&task); err != nil {
		return
	}

	slog.Info("Video processed successfully", slog.Int("video_id", task.VideoID))

	if err := MarkAsProcessed(vc.db, task.VideoID); err != nil {
		vc.logError(task, "failed to mark video as processed", err)
		return
	}

	slog.Info("Video marked as processed", slog.Int("video_id", task.VideoID))

	return
}

func (vc *VideoConverter) processVideo(task *VideoTask) error {
	mergedFile := filepath.Join(task.Path, "merged.mp4")
	mpegDashPath := filepath.Join(task.Path, "mpeg-dash")

	slog.Info("Merging chunks", slog.String("path", task.Path))

	if err := vc.mergeChunks(task.Path, mergedFile); err != nil {
		vc.logError(*task, "failed to merge chunks", err)
		return err
	}

	slog.Info("Creating mpeg-dash directory", slog.String("path", task.Path))

	if err := os.MkdirAll(mpegDashPath, os.ModePerm); err != nil {
		vc.logError(*task, "failed to create mpeg-dash directory", err)
		return err
	}

	slog.Info("Converting video to mpeg-dash", slog.String("path", task.Path))
	ffmpegCmd := exec.Command("ffmpeg",
		"-i",
		mergedFile,
		"-f",
		"dash",
		filepath.Join(mpegDashPath, "output.mpd"),
	)

	if output, err := ffmpegCmd.CombinedOutput(); err != nil {
		vc.logError(*task, "failed to convert video to mpeg-dash, output: "+string(output), err)
		return err
	}

	slog.Info("Video converted to mpeg-dash", slog.String("path", mpegDashPath))

	slog.Info("Video converted to mpeg-dash", slog.String("path", mergedFile))

	if err := os.RemoveAll(mergedFile); err != nil {
		vc.logError(*task, "failed to remove merged file", err)
		return err
	}

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

	RegisterError(vc.db, errorData, err)
}

func (vc *VideoConverter) extractNumber(fileName string) int {
	regex := regexp.MustCompile(`\d+`)
	numberStr := regex.FindString(filepath.Base(fileName))
	chunkNumber, error := strconv.Atoi(numberStr)

	if error != nil {
		log.Fatal(error)
		return -1
	}

	return chunkNumber
}

func (vc *VideoConverter) mergeChunks(inputDir, outputFile string) error {
	chunks, err := filepath.Glob(filepath.Join(inputDir, "*.chunk"))

	if err != nil {
		return fmt.Errorf("failed to find chunks: %v", err)
	}

	sort.Slice(chunks, func(i, j int) bool {
		return vc.extractNumber(chunks[i]) < vc.extractNumber(chunks[j])
	})

	output, err := os.Create(outputFile)

	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}

	defer output.Close()

	for _, chunk := range chunks {
		input, err := os.Open(chunk)

		if err != nil {
			return fmt.Errorf("failed to open chunk: %v", err)
		}

		if _, err = output.ReadFrom(input); err != nil {
			return fmt.Errorf("failed to write chunk %s to output: %v", chunk, err)
		}

		input.Close()
	}

	return nil
}
