package tests

import (
	"testing"
	"table-replication-service/internal/config"
	"table-replication-service/internal/database"
	"table-replication-service/internal/models"
	"time"
)

func TestAlternateInsertionReplication(t *testing.T) {
	cfg, err := config.LoadConfig("../config.test.ini")
	if err != nil {
		t.Skipf("Skipping test - config file not found or database not available: %v", err)
	}

	dbManager, err := database.NewDatabaseManager(cfg)
	if err != nil {
		t.Skipf("Skipping test - database connection failed: %v", err)
	}
	defer dbManager.Close()

	dbManager.SourceDB.Where("msisdn LIKE ?", "551199999%").Delete(&models.SimImeiBinding{})
	dbManager.TargetDB.Where("msisdn LIKE ?", "551199999%").Delete(&models.SimImeiBinding{})

	testRecords := []models.SimImeiBinding{
		{MSISDN: "5511999990001", IMEI: stringPtr("100000000000001"), AZ: "sa-east-1a", Locked: false},
		{MSISDN: "5511999990002", IMEI: stringPtr("100000000000002"), AZ: "sa-east-1a", Locked: false},
		{MSISDN: "5511999990003", IMEI: stringPtr("100000000000003"), AZ: "sa-east-1a", Locked: false},
		{MSISDN: "5511999990004", IMEI: stringPtr("100000000000004"), AZ: "sa-east-1a", Locked: false},
		{MSISDN: "5511999990005", IMEI: stringPtr("100000000000005"), AZ: "sa-east-1a", Locked: false},
	}

	for i, record := range testRecords {
		if err := dbManager.SourceDB.Create(&record).Error; err != nil {
			t.Fatalf("Failed to create test record %d: %v", i+1, err)
		}
		
		if i%2 == 0 {
			targetRecord := models.SimImeiBinding{
				MSISDN: "5511999990" + string(rune('6'+i)),
				IMEI:   stringPtr("200000000000" + string(rune('1'+i))),
				AZ:     "sa-east-1c",
				Locked: false,
			}
			dbManager.TargetDB.Create(&targetRecord)
		}
		
		time.Sleep(10 * time.Millisecond)
	}

	lastTimestamp := time.Time{}
	lastID := uint64(0)
	totalReplicated := 0
	maxIterations := 10

	for i := 0; i < maxIterations; i++ {
		records, err := dbManager.GetNewRecords(lastTimestamp, lastID, 2, "sa-east-1a") // Small batch size
		if err != nil {
			t.Fatalf("Failed to get new records: %v", err)
		}

		if len(records) == 0 {
			break
		}

		if err := dbManager.InsertRecords(records, "sa-east-1c"); err != nil {
			t.Fatalf("Failed to insert records: %v", err)
		}

		if len(records) > 0 {
			lastTimestamp = records[len(records)-1].LastUpdateTime
			lastID = records[len(records)-1].ID
		}
		
		totalReplicated += len(records)
		t.Logf("Iteration %d: replicated %d records, new timestamp: %v, new ID: %d", i+1, len(records), lastTimestamp, lastID)
	}

	var sourceCount int64
	dbManager.SourceDB.Model(&models.SimImeiBinding{}).Where("msisdn LIKE ? AND original_id IS NULL", "551199999%").Count(&sourceCount)
	
	var replicatedCount int64
	dbManager.TargetDB.Model(&models.SimImeiBinding{}).Where("msisdn LIKE ? AND original_id IS NOT NULL", "551199999%").Count(&replicatedCount)

	t.Logf("Source records: %d, Replicated records: %d, Total replicated: %d", sourceCount, replicatedCount, totalReplicated)

	if replicatedCount != sourceCount {
		t.Errorf("Expected %d replicated records, got %d. Some records were skipped!", sourceCount, replicatedCount)
		
		var unreplicatedRecords []models.SimImeiBinding
		dbManager.SourceDB.Where("msisdn LIKE ? AND original_id IS NULL", "551199999%").Find(&unreplicatedRecords)
		
		for _, record := range unreplicatedRecords {
			var exists int64
			dbManager.TargetDB.Model(&models.SimImeiBinding{}).Where("original_id = ?", record.ID).Count(&exists)
			if exists == 0 {
				t.Logf("Record not replicated: ID=%d, MSISDN=%s", record.ID, record.MSISDN)
			}
		}
	}

	dbManager.SourceDB.Where("msisdn LIKE ?", "551199999%").Delete(&models.SimImeiBinding{})
	dbManager.TargetDB.Where("msisdn LIKE ?", "551199999%").Delete(&models.SimImeiBinding{})
}
