package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	appName       = "Brevity"
	defaultDBFile = "/data/brevity.db" // Updated to use /data directory
	migrationsDir = "src/migrations"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("%s Migration Error: %v", appName, err)
	}
}

func run() error {
	// Set up flags
	dbFile := flag.String("db", defaultDBFile, "database file path")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		return fmt.Errorf("usage: go run cmd/migrate/run/main.go [up|down|steps|force VERSION|status|version]\n" +
			"Commands:\n" +
			"  up         - Apply all available migrations\n" +
			"  down       - Roll back one migration\n" +
			"  steps N    - Apply N migrations (negative for rollback)\n" +
			"  force V    - Set migration version V\n" +
			"  status     - Show current migration status\n" +
			"  version    - Show current migration version")
	}

	// Get project root for migrations
	projectRoot, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("project root: %w", err)
	}

	migrationsPath := filepath.Join(projectRoot, migrationsDir)

	// Validate paths
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory not found: %s", migrationsPath)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(filepath.Dir(*dbFile), 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	log.Printf("%s: Using migrations from: %s", appName, migrationsPath)
	log.Printf("%s: Using database: %s", appName, *dbFile)

	// Initialize migrator with SQLite configuration
	dbURL := fmt.Sprintf("sqlite3://%s?_foreign_keys=on&_journal_mode=WAL", *dbFile)
	m, err := migrate.New(
		fmt.Sprintf("file://%s", filepath.ToSlash(migrationsPath)),
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("migrate initialization: %w", err)
	}
	defer func() {
		if _, err := m.Close(); err != nil {
			log.Printf("%s: Warning - migration cleanup error: %v", appName, err)
		}
	}()

	// Handle commands
	switch cmd := args[0]; cmd {
	case "up":
		if err := m.Up(); err != nil {
			if err == migrate.ErrNoChange {
				log.Printf("%s: Database already up to date", appName)
				return nil
			}
			return fmt.Errorf("migrate up: %w", err)
		}
		log.Printf("%s: Migrations applied successfully", appName)

	case "down":
		if err := m.Down(); err != nil {
			if err == migrate.ErrNoChange {
				log.Printf("%s: Database already at base version", appName)
				return nil
			}
			return fmt.Errorf("migrate down: %w", err)
		}
		log.Printf("%s: Migration rolled back successfully", appName)

	case "steps":
		if len(args) < 2 {
			return fmt.Errorf("steps requires a number argument")
		}
		steps, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid steps value: %w", err)
		}
		if err := m.Steps(steps); err != nil {
			if err == migrate.ErrNoChange {
				log.Printf("%s: No migrations to apply", appName)
				return nil
			}
			return fmt.Errorf("migrate steps %d: %w", steps, err)
		}
		log.Printf("%s: Applied %d migration steps successfully", appName, steps)

	case "force":
		if len(args) < 2 {
			return fmt.Errorf("force requires a version number")
		}
		version, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid version number: %w", err)
		}
		if err := m.Force(version); err != nil {
			return fmt.Errorf("force version %d: %w", version, err)
		}
		log.Printf("%s: Force set version to %d", appName, version)

	case "status":
		version, dirty, err := m.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				log.Printf("%s: No migrations applied", appName)
				return nil
			}
			return fmt.Errorf("status check: %w", err)
		}
		log.Printf("%s: Current version: %d (dirty: %v)", appName, version, dirty)

	case "version":
		version, _, err := m.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				log.Printf("%s: No migrations applied", appName)
				return nil
			}
			return fmt.Errorf("version check: %w", err)
		}
		log.Printf("%s: Current migration version: %d", appName, version)

	default:
		return fmt.Errorf("invalid command '%s'", cmd)
	}

	return nil
}

func getProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("working directory: %w", err)
	}

	current := wd
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", fmt.Errorf("project root not found (go.mod missing)")
}