package setup

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"ccards/pkg/config"
	"ccards/pkg/database"
)

var (
	testConfig   *config.Config
	testDB       *sql.DB
	testGormDB   *gorm.DB
	testRedis    *redis.Client
	setupOnce    sync.Once
	cleanupFuncs []func()
	mu           sync.Mutex
)

type TestHelper struct {
	Config  *config.Config
	DB      *sql.DB
	GormDB  *gorm.DB
	Redis   *redis.Client
	cleanup []func()
}

func SetupTestEnvironment() (*TestHelper, error) {
	var setupErr error

	setupOnce.Do(func() {
		os.Setenv("APP_ENVIRONMENT", "test")

		cfg, err := config.LoadConfig()
		if err != nil {
			setupErr = fmt.Errorf("failed to load test config: %w", err)
			return
		}
		testConfig = cfg

		if err := setupTestDatabase(cfg); err != nil {
			setupErr = fmt.Errorf("failed to setup test database: %w", err)
			return
		}

		if err := setupTestRedis(cfg); err != nil {
			setupErr = fmt.Errorf("failed to setup test redis: %w", err)
			return
		}
	})

	if setupErr != nil {
		return nil, setupErr
	}

	// Create a new TestHelper instance using the global variables
	helper := &TestHelper{
		Config: testConfig,
		DB:     testDB,
		GormDB: testGormDB,
		Redis:  testRedis,
	}

	return helper, nil
}

func setupTestDatabase(cfg *config.Config) error {
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %w", err)
	}
	testDB = db

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to setup gorm: %w", err)
	}
	testGormDB = gormDB

	migrationsPath := findMigrationsPath()
	if migrationsPath == "" {
		return fmt.Errorf("could not find migrations directory")
	}

	if err := database.RunMigrations(db, migrationsPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	addCleanupFunc(func() {
		log.Println("Closing test database connection...")
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	})

	return nil
}

func findMigrationsPath() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			migrationsPath := filepath.Join(dir, "db", "migrations")
			if _, err := os.Stat(migrationsPath); err == nil {
				return migrationsPath
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}

func setupTestRedis(cfg *config.Config) error {
	ctx := context.Background()

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
		PoolSize:     cfg.Redis.PoolSize,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	testRedis = client

	addCleanupFunc(func() {
		log.Println("Cleaning Redis test database...")
		if err := client.FlushDB(ctx).Err(); err != nil {
			log.Printf("Error flushing Redis: %v", err)
		}
		if err := client.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		}
	})

	return nil
}

func CleanupTestEnvironment() {
	mu.Lock()
	defer mu.Unlock()

	for i := len(cleanupFuncs) - 1; i >= 0; i-- {
		cleanupFuncs[i]()
	}
	cleanupFuncs = nil
}

func addCleanupFunc(fn func()) {
	mu.Lock()
	defer mu.Unlock()
	cleanupFuncs = append(cleanupFuncs, fn)
}

func NewTestHelper(t *testing.T) *TestHelper {
	helper, err := SetupTestEnvironment()
	if err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}

	if err := helper.CleanDatabase(); err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}

	if err := helper.CleanRedis(); err != nil {
		t.Fatalf("Failed to clean Redis: %v", err)
	}

	return helper
}

func (h *TestHelper) CleanDatabase() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get all table names
	rows, err := h.DB.QueryContext(ctx, `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		AND tablename NOT IN ('goose_db_version', 'schema_migrations')
		ORDER BY tablename
	`)
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, table := range tables {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)); err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (h *TestHelper) CleanRedis() error {
	ctx := context.Background()
	return h.Redis.FlushDB(ctx).Err()
}

func (h *TestHelper) BeginTx() (*sql.Tx, error) {
	return h.DB.Begin()
}

func (h *TestHelper) MustExec(t *testing.T, query string, args ...interface{}) {
	_, err := h.DB.Exec(query, args...)
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
}