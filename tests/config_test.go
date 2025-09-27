package tests

import (
	"os"
	"table-replication-service/internal/config"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	configContent := `[database_source]
host = localhost
port = 3306
username = ${DB_SOURCE_USER:testuser}
password = ${DB_SOURCE_PASSWORD:testpass}
database = testdb
driver = mysql

[database_target]
host = localhost
port = 3306
username = ${DB_TARGET_USER:testuser}
password = ${DB_TARGET_PASSWORD:testpass}
database = testdb2
driver = mysql

[replication]
table_name = test_table
interval_seconds = 5
site_identifier = test-site
enable_logging = true
batch_size = 50

[server]
port = 8080
host = 0.0.0.0

[metrics]
enable_prometheus = true
metrics_path = /metrics`

	tmpFile, err := os.CreateTemp("", "test_config_*.ini")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := config.LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.DatabaseSource.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", cfg.DatabaseSource.Host)
	}

	if cfg.DatabaseSource.Port != 3306 {
		t.Errorf("Expected port 3306, got %d", cfg.DatabaseSource.Port)
	}

	if cfg.Replication.TableName != "test_table" {
		t.Errorf("Expected table name 'test_table', got '%s'", cfg.Replication.TableName)
	}

	if cfg.Replication.IntervalSeconds != 5 {
		t.Errorf("Expected interval 5, got %d", cfg.Replication.IntervalSeconds)
	}
}

func TestGetDSN(t *testing.T) {
	dbConfig := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "testuser",
		Password: "testpass",
		Database: "testdb",
		Driver:   "mysql",
	}

	dsn := dbConfig.GetDSN()
	expected := "testuser:testpass@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	
	if dsn != expected {
		t.Errorf("Expected DSN '%s', got '%s'", expected, dsn)
	}

	dbConfig.Driver = "postgres"
	dsn = dbConfig.GetDSN()
	expected = "host=localhost user=testuser password=testpass dbname=testdb port=3306 sslmode=disable TimeZone=UTC"
	
	if dsn != expected {
		t.Errorf("Expected PostgreSQL DSN '%s', got '%s'", expected, dsn)
	}
}
