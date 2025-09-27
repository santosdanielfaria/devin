package tests

import (
	"testing"
	"table-replication-service/internal/config"
	"table-replication-service/internal/database"
	"table-replication-service/internal/models"
)

func TestAZFieldPreservation(t *testing.T) {
	cfg, err := config.LoadConfig("../config.test.ini")
	if err != nil {
		t.Skipf("Skipping test - config file not found or database not available: %v", err)
	}

	dbManager, err := database.NewDatabaseManager(cfg)
	if err != nil {
		t.Skipf("Skipping test - database connection failed: %v", err)
	}
	defer dbManager.Close()

	testRecord := models.SimImeiBinding{
		MSISDN: "5511999999999",
		IMEI:   stringPtr("999999999999999"),
		AZ:     "sa-east-1a",
		Locked: false,
	}

	if err := dbManager.SourceDB.Create(&testRecord).Error; err != nil {
		t.Fatalf("Failed to create test record: %v", err)
	}

	records, err := dbManager.GetNewRecords(0, 10, "sa-east-1a")
	if err != nil {
		t.Fatalf("Failed to get new records: %v", err)
	}

	if len(records) == 0 {
		t.Fatal("No records found for replication")
	}

	originalAZ := records[0].AZ
	if originalAZ != "sa-east-1a" {
		t.Errorf("Expected original AZ to be 'sa-east-1a', got '%s'", originalAZ)
	}

	if err := dbManager.InsertRecords(records, "sa-east-1c"); err != nil {
		t.Fatalf("Failed to insert records: %v", err)
	}

	var replicatedRecord models.SimImeiBinding
	if err := dbManager.TargetDB.Where("msisdn = ?", testRecord.MSISDN).First(&replicatedRecord).Error; err != nil {
		t.Fatalf("Failed to find replicated record: %v", err)
	}

	if replicatedRecord.AZ != "sa-east-1a" {
		t.Errorf("Expected replicated AZ to be 'sa-east-1a', got '%s'", replicatedRecord.AZ)
	}

	if replicatedRecord.OriginalID == nil {
		t.Error("Expected OriginalID to be set")
	}

	dbManager.SourceDB.Where("msisdn = ?", testRecord.MSISDN).Delete(&models.SimImeiBinding{})
	dbManager.TargetDB.Where("msisdn = ?", testRecord.MSISDN).Delete(&models.SimImeiBinding{})
}

func stringPtr(s string) *string {
	return &s
}
