package tests

import (
	"fmt"
	"testing"
	"time"
	"table-replication-service/internal/config"
	"table-replication-service/internal/database"
	"table-replication-service/internal/models"
)

func TestAlternatingInsertionFix(t *testing.T) {
	cfg, err := config.LoadConfig("../config.test.ini")
	if err != nil {
		t.Skipf("Skipping test - config file not found: %v", err)
	}

	dbManager, err := database.NewDatabaseManager(cfg)
	if err != nil {
		t.Skipf("Skipping test - database connection failed: %v", err)
	}
	defer dbManager.Close()

	testPattern := "5511888888%"
	dbManager.SourceDB.Where("msisdn LIKE ?", testPattern).Delete(&models.SimImeiBinding{})
	dbManager.TargetDB.Where("msisdn LIKE ?", testPattern).Delete(&models.SimImeiBinding{})

	for i := 1; i <= 10; i++ {
		sourceRecord := models.SimImeiBinding{
			MSISDN: fmt.Sprintf("551188888800%02d", i),
			AZ:     "sa-east-1a",
			Locked: false,
		}
		dbManager.SourceDB.Create(&sourceRecord)
		
		time.Sleep(50 * time.Millisecond) // Ensure different timestamps
		
		targetRecord := models.SimImeiBinding{
			MSISDN: fmt.Sprintf("551188888801%02d", i),
			AZ:     "sa-east-1c", 
			Locked: false,
		}
		dbManager.TargetDB.Create(&targetRecord)
		
		time.Sleep(50 * time.Millisecond) // Ensure different timestamps
	}

	lastTimestamp := time.Time{}
	lastID := uint64(0)
	totalReplicated := 0
	maxIterations := 20

	for i := 0; i < maxIterations; i++ {
		records, err := dbManager.GetNewRecords(lastTimestamp, lastID, 2, "sa-east-1a")
		if err != nil {
			t.Fatalf("Failed to get new records: %v", err)
		}

		if len(records) == 0 {
			break
		}

		if err := dbManager.InsertRecords(records, "sa-east-1c"); err != nil {
			t.Fatalf("Failed to insert records: %v", err)
		}

		lastTimestamp = records[len(records)-1].LastUpdateTime
		lastID = records[len(records)-1].ID
		totalReplicated += len(records)
		t.Logf("Iteration %d: replicated %d records, new timestamp: %v, new ID: %d", i+1, len(records), lastTimestamp, lastID)
	}

	var sourceCount int64
	dbManager.SourceDB.Model(&models.SimImeiBinding{}).Where("msisdn LIKE ? AND original_id IS NULL", testPattern).Count(&sourceCount)
	
	var replicatedCount int64
	dbManager.TargetDB.Model(&models.SimImeiBinding{}).Where("msisdn LIKE ? AND original_id IS NOT NULL", testPattern).Count(&replicatedCount)

	if replicatedCount != sourceCount {
		t.Errorf("Expected %d replicated records, got %d. Records were still skipped!", sourceCount, replicatedCount)
	} else {
		t.Logf("SUCCESS: All %d source records were replicated correctly", sourceCount)
	}

	dbManager.SourceDB.Where("msisdn LIKE ?", testPattern).Delete(&models.SimImeiBinding{})
	dbManager.TargetDB.Where("msisdn LIKE ?", testPattern).Delete(&models.SimImeiBinding{})
}
