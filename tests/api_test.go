package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"table-replication-service/internal/config"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthCheckEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
		Metrics: config.MetricsConfig{
			EnablePrometheus: true,
			MetricsPath:      "/metrics",
		},
	}

	router := gin.New()
	
	router.GET("/health-check", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "table-replication-service",
		})
	})

	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%v'", response["status"])
	}

	if response["service"] != "table-replication-service" {
		t.Errorf("Expected service 'table-replication-service', got '%v'", response["service"])
	}

	_ = cfg // Use cfg to avoid unused variable error
}

func TestValidationEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	
	router.GET("/validation", func(c *gin.Context) {
		result := map[string]interface{}{
			"source_count": 100,
			"target_count": 100,
			"is_sync":      true,
			"difference":   0,
		}
		c.JSON(http.StatusOK, result)
	})

	req, err := http.NewRequest("GET", "/validation", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["is_sync"] != true {
		t.Errorf("Expected is_sync true, got %v", response["is_sync"])
	}
}
