package tests

import (
	"testing"
	"time"
	"table-replication-service/internal/config"
	"table-replication-service/internal/database"
	"table-replication-service/internal/models"
)

func TestDebugOffsetTracking(t *testing.T) {
	cfg, err := config.LoadConfig("../config.test.ini")
	if err != nil {
		t.Skipf("Skipping test - config file not found: %v", err)
	}

	dbManager, err := database.NewDatabaseManager(cfg)
	if err != nil {
		t.Skipf("Skipping test - database connection failed: %v", err)
	}
	defer dbManager.Close()

	dbManager.SourceDB.Where("msisdn LIKE ?", "5511888%").Delete(&models.SimImeiBinding{})
	dbManager.TargetDB.Where("msisdn LIKE ?", "5511888%").Delete(&models.SimImeiBinding{})

	record1 := models.SimImeiBinding{MSISDN: "55118880001", AZ: "sa-east-1a", Locked: false}
	record2 := models.SimImeiBinding{MSISDN: "55118880002", AZ: "sa-east-1a", Locked: false}
	record3 := models.SimImeiBinding{MSISDN: "55118880003", AZ: "sa-east-1a", Locked: false}

	dbManager.SourceDB.Create(&record1)
	time.Sleep(10 * time.Millisecond)
	dbManager.SourceDB.Create(&record2)
	time.Sleep(10 * time.Millisecond)
	dbManager.SourceDB.Create(&record3)

	var allRecords []models.SimImeiBinding
	dbManager.SourceDB.Where("msisdn LIKE ? AND original_id IS NULL", "5511888%").Order("id ASC").Find(&allRecords)
	
	t.Logf("Created %d records:", len(allRecords))
	for _, r := range allRecords {
		t.Logf("  ID: %d, MSISDN: %s, Timestamp: %v", r.ID, r.MSISDN, r.LastUpdateTime)
	}

	lastTimestamp := time.Time{}
	lastID := uint64(0)

	records, err := dbManager.GetNewRecords(lastTimestamp, lastID, 2, "sa-east-1a")
	if err != nil {
		t.Fatalf("Failed to get new records: %v", err)
	}

	t.Logf("First batch: got %d records", len(records))
	for _, r := range records {
		t.Logf("  ID: %d, MSISDN: %s, Timestamp: %v", r.ID, r.MSISDN, r.LastUpdateTime)
	}

	if len(records) > 0 {
		lastTimestamp = records[len(records)-1].LastUpdateTime
		lastID = records[len(records)-1].ID
		t.Logf("Updated offset to timestamp: %v, ID: %d", lastTimestamp, lastID)
	}

	records, err = dbManager.GetNewRecords(lastTimestamp, lastID, 2, "sa-east-1a")
	if err != nil {
		t.Fatalf("Failed to get new records: %v", err)
	}

	t.Logf("Second batch: got %d records", len(records))
	for _, r := range records {
		t.Logf("  ID: %d, MSISDN: %s, Timestamp: %v", r.ID, r.MSISDN, r.LastUpdateTime)
	}

	dbManager.SourceDB.Where("msisdn LIKE ?", "5511888%").Delete(&models.SimImeiBinding{})
	dbManager.TargetDB.Where("msisdn LIKE ?", "5511888%").Delete(&models.SimImeiBinding{})
}
