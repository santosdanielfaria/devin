# Table Replication Service

A Go-based microservice for bidirectional database table replication between MySQL/PostgreSQL databases. This service ensures data consistency across multiple database instances with offset-based tracking to prevent data loss or duplication.

## Features

- **Bidirectional Replication**: Replicates data between two database instances
- **Offset-based Consistency**: Tracks last replicated record ID to ensure no data loss
- **Multi-database Support**: Works with both MySQL and PostgreSQL via GORM
- **REST API**: Provides endpoints for validation, health checks, and status monitoring
- **Prometheus Metrics**: Built-in instrumentation for monitoring replication health
- **Graceful Shutdown**: Safe shutdown handling for Kubernetes deployments
- **Configurable**: All settings managed via INI configuration file
- **Package Structure**: Modular design for code reusability

## Architecture

```
table-replication-service/
├── cmd/                    # Application entry point
├── internal/
│   ├── api/               # REST API handlers
│   ├── config/            # Configuration management
│   ├── database/          # Database connection and operations
│   ├── metrics/           # Prometheus metrics
│   ├── models/            # Data models
│   └── replication/       # Core replication logic
├── pkg/
│   ├── logger/            # Logging utilities
│   └── utils/             # Shared utilities
├── tests/                 # Unit tests
├── docs/                  # Documentation
└── config.ini            # Configuration file
```

## Configuration

The service is configured via an INI file (`config.ini`):

```ini
[database_source]
host = ${DB_SOURCE_HOST:localhost}
port = ${DB_SOURCE_PORT:3306}
username = ${DB_SOURCE_USER:root}
password = ${DB_SOURCE_PASSWORD}
database = ${DB_SOURCE_DB:site_a}
driver = ${DB_SOURCE_DRIVER:mysql}

[database_target]
host = ${DB_TARGET_HOST:localhost}
port = ${DB_TARGET_PORT:3306}
username = ${DB_TARGET_USER:root}
password = ${DB_TARGET_PASSWORD}
database = ${DB_TARGET_DB:site_b}
driver = ${DB_TARGET_DRIVER:mysql}

[replication]
table_name = sim_imei_binding
interval_seconds = 10
site_identifier = sa-east-1a
enable_logging = true
batch_size = 100

[server]
port = 8080
host = 0.0.0.0

[metrics]
enable_prometheus = true
metrics_path = /metrics
```

## Database Schema

The service replicates the `sim_imei_binding` table with the following structure:

```sql
CREATE TABLE sim_imei_binding (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  msisdn VARCHAR(20) NOT NULL,
  imei VARCHAR(15) DEFAULT NULL,
  az VARCHAR(10) DEFAULT "sa-east-1a",
  locked TINYINT(1) NOT NULL DEFAULT 0,
  lastupdatetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_sim_imei_msisdn (msisdn)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
```

The service automatically creates a `replication_offset` table to track replication progress.

## Installation

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd table-replication-service
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Configure the service**:
   - Copy `config.ini` and update database credentials
   - Set appropriate `site_identifier` for each instance

4. **Build the application**:
   ```bash
   go build -o replication-service ./cmd
   ```

## Usage

### Running the Service

```bash
# Using default config.ini
./replication-service

# Using custom config file
./replication-service /path/to/custom-config.ini
```

### Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o replication-service ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/replication-service .
COPY --from=builder /app/config.ini .
CMD ["./replication-service"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: table-replication-service
spec:
  replicas: 1
  selector:
    matchLabels:
      app: table-replication-service
  template:
    metadata:
      labels:
        app: table-replication-service
    spec:
      containers:
      - name: replication-service
        image: table-replication-service:latest
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health-check
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health-check
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## API Endpoints

### Health Check
```
GET /health-check
```
Returns service health status for Kubernetes probes.

### Validation
```
GET /validation
```
Compares record counts between source and target databases.

### Status
```
GET /status
```
Returns current replication status and information.

### Metrics
```
GET /metrics
```
Prometheus metrics endpoint.

### Swagger Documentation
```
GET /swagger/index.html
```
Interactive API documentation.

## Monitoring

The service exposes Prometheus metrics:

- `table_replication_records_total`: Total number of replicated records
- `table_replication_errors_total`: Total number of replication errors
- `table_replication_sync_status`: Sync status (1 = synced, 0 = not synced)

## Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./tests -run TestLoadConfig
```

## Replication Logic

1. **Offset Tracking**: Each replication cycle starts by retrieving the last replicated record ID
2. **Batch Processing**: Fetches new records in configurable batches for performance
3. **Atomic Updates**: Updates offset only after successful replication
4. **Error Handling**: Comprehensive error handling with metrics and logging
5. **Consistency**: Ensures no data duplication or loss during service restarts

## Multi-table Support

To replicate additional tables:

1. Create new model structs in `internal/models/`
2. Update configuration to specify table name
3. Modify database operations to handle different table structures
4. The package structure allows easy extension for multiple table types

## Troubleshooting

### Common Issues

1. **Database Connection Errors**:
   - Verify database credentials in config.ini
   - Check network connectivity
   - Ensure database exists and user has proper permissions

2. **Replication Lag**:
   - Increase batch_size for better performance
   - Check database performance and indexing
   - Monitor Prometheus metrics for bottlenecks

3. **Sync Issues**:
   - Use `/validation` endpoint to check sync status
   - Check replication logs for errors
   - Verify offset table consistency

### Logging

Enable detailed logging by setting `enable_logging = true` in the configuration. Logs include:
- Replication cycle information
- Error details
- Performance metrics
- Database connection status

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License.
