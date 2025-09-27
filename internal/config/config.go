package config

import (
	"fmt"
	"github.com/go-ini/ini"
)

type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Driver   string
}

type ReplicationConfig struct {
	TableName        string
	IntervalSeconds  int
	SiteIdentifier   string
	EnableLogging    bool
	BatchSize        int
}

type ServerConfig struct {
	Port string
	Host string
}

type MetricsConfig struct {
	EnablePrometheus bool
	MetricsPath      string
}

type Config struct {
	DatabaseSource DatabaseConfig
	DatabaseTarget DatabaseConfig
	Replication    ReplicationConfig
	Server         ServerConfig
	Metrics        MetricsConfig
}

func LoadConfig(configPath string) (*Config, error) {
	cfg, err := ini.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	config := &Config{}

	dbSourceSection := cfg.Section("database_source")
	config.DatabaseSource = DatabaseConfig{
		Host:     dbSourceSection.Key("host").String(),
		Port:     dbSourceSection.Key("port").MustInt(3306),
		Username: dbSourceSection.Key("username").String(),
		Password: dbSourceSection.Key("password").String(),
		Database: dbSourceSection.Key("database").String(),
		Driver:   dbSourceSection.Key("driver").MustString("mysql"),
	}

	dbTargetSection := cfg.Section("database_target")
	config.DatabaseTarget = DatabaseConfig{
		Host:     dbTargetSection.Key("host").String(),
		Port:     dbTargetSection.Key("port").MustInt(3306),
		Username: dbTargetSection.Key("username").String(),
		Password: dbTargetSection.Key("password").String(),
		Database: dbTargetSection.Key("database").String(),
		Driver:   dbTargetSection.Key("driver").MustString("mysql"),
	}

	replicationSection := cfg.Section("replication")
	config.Replication = ReplicationConfig{
		TableName:       replicationSection.Key("table_name").String(),
		IntervalSeconds: replicationSection.Key("interval_seconds").MustInt(10),
		SiteIdentifier:  replicationSection.Key("site_identifier").String(),
		EnableLogging:   replicationSection.Key("enable_logging").MustBool(true),
		BatchSize:       replicationSection.Key("batch_size").MustInt(100),
	}

	serverSection := cfg.Section("server")
	config.Server = ServerConfig{
		Port: serverSection.Key("port").MustString("8080"),
		Host: serverSection.Key("host").MustString("0.0.0.0"),
	}

	metricsSection := cfg.Section("metrics")
	config.Metrics = MetricsConfig{
		EnablePrometheus: metricsSection.Key("enable_prometheus").MustBool(true),
		MetricsPath:      metricsSection.Key("metrics_path").MustString("/metrics"),
	}

	return config, nil
}

func (dc *DatabaseConfig) GetDSN() string {
	switch dc.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dc.Username, dc.Password, dc.Host, dc.Port, dc.Database)
	case "postgres":
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
			dc.Host, dc.Username, dc.Password, dc.Database, dc.Port)
	default:
		return ""
	}
}
