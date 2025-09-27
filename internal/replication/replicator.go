package replication

import (
	"context"
	"log"
	"table-replication-service/internal/config"
	"table-replication-service/internal/database"
	"table-replication-service/internal/metrics"
	"time"
)

type Replicator struct {
	dbManager *database.DatabaseManager
	config    *config.Config
	metrics   *metrics.Metrics
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewReplicator(dbManager *database.DatabaseManager, cfg *config.Config, metrics *metrics.Metrics) *Replicator {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Replicator{
		dbManager: dbManager,
		config:    cfg,
		metrics:   metrics,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (r *Replicator) Start() {
	ticker := time.NewTicker(time.Duration(r.config.Replication.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	if r.config.Replication.EnableLogging {
		log.Printf("Starting replication service for table %s, site %s", 
			r.config.Replication.TableName, r.config.Replication.SiteIdentifier)
	}

	r.replicate()

	for {
		select {
		case <-ticker.C:
			r.replicate()
		case <-r.ctx.Done():
			if r.config.Replication.EnableLogging {
				log.Println("Replication service stopped")
			}
			return
		}
	}
}

func (r *Replicator) Stop() {
	r.cancel()
}

func (r *Replicator) replicate() {
	startTime := time.Now()
	
	lastTimestamp, lastID, err := r.dbManager.GetLastOffset(r.config.Replication.TableName, r.config.Replication.SiteIdentifier, r.config.Replication.SourceAZ)
	if err != nil {
		r.metrics.IncrementReplicationErrors()
		if r.config.Replication.EnableLogging {
			log.Printf("Error getting last offset: %v", err)
		}
		return
	}

	records, err := r.dbManager.GetNewRecords(lastTimestamp, lastID, r.config.Replication.BatchSize, r.config.Replication.SourceAZ)
	if err != nil {
		r.metrics.IncrementReplicationErrors()
		if r.config.Replication.EnableLogging {
			log.Printf("Error getting new records: %v", err)
		}
		return
	}

	if len(records) == 0 {
		return
	}

	if err := r.dbManager.InsertRecords(records, r.config.Replication.TargetAZ); err != nil {
		r.metrics.IncrementReplicationErrors()
		if r.config.Replication.EnableLogging {
			log.Printf("Error inserting records: %v", err)
		}
		return
	}

	newLastTimestamp := records[len(records)-1].LastUpdateTime
	newLastID := records[len(records)-1].ID
	if err := r.dbManager.UpdateOffset(r.config.Replication.TableName, r.config.Replication.SiteIdentifier, r.config.Replication.SourceAZ, newLastTimestamp, newLastID); err != nil {
		r.metrics.IncrementReplicationErrors()
		if r.config.Replication.EnableLogging {
			log.Printf("Error updating offset: %v", err)
		}
		return
	}

	r.metrics.AddReplicatedRecords(float64(len(records)))
	
	isSync, _, err := r.dbManager.ValidateSync(r.config.Replication.SourceAZ, r.config.Replication.TargetAZ)
	if err != nil {
		if r.config.Replication.EnableLogging {
			log.Printf("Error validating sync: %v", err)
		}
	} else {
		r.metrics.SetSyncStatus(isSync)
	}

	duration := time.Since(startTime)
	
	if r.config.Replication.EnableLogging {
		log.Printf("Replicated %d records in %v (last timestamp: %v)", len(records), duration, newLastTimestamp)
	}
}

func (r *Replicator) GetStatus() map[string]interface{} {
	lastTimestamp, lastID, err := r.dbManager.GetLastOffset(r.config.Replication.TableName, r.config.Replication.SiteIdentifier, r.config.Replication.SourceAZ)
	if err != nil {
		lastTimestamp = time.Time{}
		lastID = 0
	}

	isSync, syncInfo, err := r.dbManager.ValidateSync(r.config.Replication.SourceAZ, r.config.Replication.TargetAZ)
	if err != nil {
		isSync = false
		syncInfo = map[string]interface{}{
			"error": err.Error(),
		}
	}

	return map[string]interface{}{
		"table_name":              r.config.Replication.TableName,
		"site_identifier":         r.config.Replication.SiteIdentifier,
		"source_az":               r.config.Replication.SourceAZ,
		"target_az":               r.config.Replication.TargetAZ,
		"last_replicated_timestamp": lastTimestamp.Format(time.RFC3339),
		"last_replicated_id":      lastID,
		"is_sync":                 isSync,
		"sync_info":               syncInfo,
		"status":                  "running",
	}
}
