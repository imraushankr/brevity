package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	appName       = "Brevity"
	migrationsDir = "src/migrations"
	fileTemplate  = `-- %s Migration: %s
-- Generated: %s
-- Direction: %s

-- Add your SQL below this line
`
	helpMessage = `Brevity Migration Creator

Usage:
  go run src/cmd/migrate/create/main.go <migration_name>

Example:
  go run src/cmd/migrate/create/main.go create_users_table

Rules:
  - Migration names should be descriptive and use underscores
  - Only lowercase letters, numbers and underscores allowed
  - Must be run from project root or a subdirectory`
)

func main() {
	if err := run(); err != nil {
		if errors.Is(err, ErrHelpRequested) {
			fmt.Print(helpMessage) // Changed from Println to Print
			os.Exit(0)
		}
		fmt.Printf("%s Migration Error: %v\n", appName, err)
		os.Exit(1)
	}
}

var (
	ErrHelpRequested = errors.New("help requested")
)

func run() error {
	if len(os.Args) < 2 {
		return ErrHelpRequested
	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		return ErrHelpRequested
	}

	migrationName := os.Args[1]
	if err := validateMigrationName(migrationName); err != nil {
		return fmt.Errorf("invalid migration name: %w", err)
	}

	migrationsPath, err := findOrCreateMigrationsDir()
	if err != nil {
		return fmt.Errorf("directory setup failed: %w", err)
	}

	timestamp := time.Now().UTC().Format("20060102150405")
	baseName := fmt.Sprintf("%s_%s", timestamp, migrationName)

	for _, direction := range []string{"up", "down"} {
		filename := filepath.Join(migrationsPath, fmt.Sprintf("%s.%s.sql", baseName, direction))
		if err := createMigrationFile(filename, direction, migrationName); err != nil {
			return fmt.Errorf("file creation failed: %w", err)
		}
		fmt.Printf("%s: Created migration %s\n", appName, filepath.Base(filename))
	}

	return nil
}

func validateMigrationName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	if strings.HasPrefix(name, "-") {
		return errors.New("name cannot start with a hyphen")
	}

	validName := regexp.MustCompile(`^[a-z0-9_]+$`)
	if !validName.MatchString(name) {
		return errors.New("name must contain only lowercase letters, numbers and underscores")
	}

	return nil
}

func findOrCreateMigrationsDir() (string, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return "", err
	}

	migrationsPath := filepath.Join(projectRoot, migrationsDir)
	if err := os.MkdirAll(migrationsPath, 0755); err != nil {
		return "", fmt.Errorf("cannot create migrations directory: %w", err)
	}

	return migrationsPath, nil
}

func findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot determine working directory: %w", err)
	}

	current := wd
	for {
		if isProjectRoot(current) {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", fmt.Errorf("could not locate %s project root (go.mod not found)", appName)
}

func isProjectRoot(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	return err == nil
}

func createMigrationFile(filename, direction, migrationName string) error {
	content := fmt.Sprintf(fileTemplate,
		appName,
		migrationName,
		time.Now().UTC().Format(time.RFC3339),
		strings.ToUpper(direction),
	)

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("cannot write to file: %w", err)
	}

	return nil
}