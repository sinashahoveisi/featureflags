package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"featureflags/entity"
	"featureflags/migrations"
	"featureflags/pkg/logger"
	"featureflags/repository"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestDB wraps a test database connection
type TestDB struct {
	DB *sqlx.DB
}

// SetupTestDB creates a test database and runs migrations
func SetupTestDB(t *testing.T) *TestDB {
	// Use environment variables or defaults for test database
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5432")
	user := getEnvOrDefault("TEST_DB_USER", "featureflags")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "featureflags")
	
	// Get base database name and add _test suffix
	baseDBName := getEnvOrDefault("POSTGRES_DB", "featureflags")
	dbName := getEnvOrDefault("TEST_DB_NAME", baseDBName+"_test")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)

	db, err := sqlx.Connect("postgres", connStr)
	require.NoError(t, err, "Failed to connect to test database")

	// Run migrations - check multiple possible paths
	migrationPaths := []string{"./migrations", "../migrations", "/app/migrations"}
	for _, path := range migrationPaths {
		err = migrations.RunMigrations(db.DB, path)
		if err == nil {
			break
		}
	}
	require.NoError(t, err, "Failed to run test migrations")

	return &TestDB{DB: db}
}

// Close closes the test database connection
func (tdb *TestDB) Close() {
	if tdb.DB != nil {
		tdb.DB.Close()
	}
}

// CleanTables removes all data from tables (for test isolation)
func (tdb *TestDB) CleanTables(t *testing.T) {
	_, err := tdb.DB.Exec("TRUNCATE TABLE audit_logs, flag_dependencies, flags RESTART IDENTITY CASCADE")
	require.NoError(t, err, "Failed to clean test tables")
}

// CreateTestFlag creates a test flag in the database
func (tdb *TestDB) CreateTestFlag(t *testing.T, name string, status entity.FlagStatus) *entity.Flag {
	flag := &entity.Flag{
		Name:   name,
		Status: status,
	}

	flagRepo := repository.NewFlagRepository(tdb.DB)
	flagID, err := flagRepo.CreateFlag(context.Background(), flag)
	require.NoError(t, err, "Failed to create test flag")

	flag.ID = flagID
	return flag
}

// CreateTestFlagWithDependencies creates a test flag with dependencies
func (tdb *TestDB) CreateTestFlagWithDependencies(t *testing.T, name string, status entity.FlagStatus, deps []int64) *entity.Flag {
	flag := tdb.CreateTestFlag(t, name, status)
	
	if len(deps) > 0 {
		flagRepo := repository.NewFlagRepository(tdb.DB)
		for _, depID := range deps {
			err := flagRepo.AddDependency(context.Background(), flag.ID, depID)
			require.NoError(t, err, "Failed to add test dependency")
		}
		flag.Dependencies = deps
	}

	return flag
}

// GetTestLogger creates a test logger
func GetTestLogger() *logger.Logger {
	log, err := logger.New("debug", "development")
	if err != nil {
		panic(fmt.Sprintf("Failed to create test logger: %v", err))
	}
	return log
}

// AssertFlagStatus asserts that a flag has the expected status
func (tdb *TestDB) AssertFlagStatus(t *testing.T, flagID int64, expectedStatus entity.FlagStatus) {
	flagRepo := repository.NewFlagRepository(tdb.DB)
	flag, err := flagRepo.GetFlagByID(context.Background(), flagID)
	require.NoError(t, err, "Failed to get flag for status assertion")
	require.Equal(t, expectedStatus, flag.Status, "Flag status mismatch")
}

// AssertAuditLogExists asserts that an audit log entry exists for a flag
func (tdb *TestDB) AssertAuditLogExists(t *testing.T, flagID int64, action entity.AuditAction, actor string) {
	auditRepo := repository.NewAuditRepository(tdb.DB)
	logs, err := auditRepo.ListAuditLogsByFlagID(context.Background(), flagID)
	require.NoError(t, err, "Failed to get audit logs")
	
	found := false
	for _, log := range logs {
		if log.Action == action && log.Actor == actor {
			found = true
			break
		}
	}
	require.True(t, found, "Expected audit log not found: action=%s, actor=%s", action, actor)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 