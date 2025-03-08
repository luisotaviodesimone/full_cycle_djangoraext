package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"videoconverter/internal/utils"
)

func main() {
	dbHost := utils.GetEnvOrDefault("DB_HOST", "localhost")
	dbPort := utils.GetEnvOrDefault("DB_PORT", "5432")
	dbUser := utils.GetEnvOrDefault("DB_USER", "postgres")
	dbPassword := utils.GetEnvOrDefault("DB_PASSWORD", "root")

	os.Setenv("PGPASSWORD", dbPassword)
	dbName := utils.GetEnvOrDefault("DB_NAME", "postgres")

	currentDir, _ := os.Getwd()
	migrationScript := utils.GetEnvOrDefault("MIGRATION_SCRIPT", filepath.Join(currentDir, "db.sql"))

	cmd := exec.Command("psql", "-h", dbHost, "-p", dbPort, "-U", dbUser, dbName, "-f", migrationScript)

	if err := cmd.Run(); err != nil {
		panic(err)
	}

}
