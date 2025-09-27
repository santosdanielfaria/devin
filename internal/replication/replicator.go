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
	
	lastID, err := r.dbManager.GetLastOffset(r.config.Replication.TableName, r.config.Replication.SiteIdentifier)
	if err != nil {
		r.metrics.IncrementReplicationErrors()
		if r.config.Replication.EnableLogging {
			log.Printf("Error getting last offset: %v", err)
		}
		return
	}

	records, err := r.dbManager.GetNewRecords(lastID, r.config.Replication.BatchSize)
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

	if err := r.dbManager.InsertRecords(records); err != nil {
		r.metrics.IncrementReplicationErrors()
		if r.config.Replication.EnableLogging {
			log.Printf("Error inserting records: %v", err)
		}
		return
	}

	newLastID := records[len(records)-1].ID
	if err := r.dbManager.UpdateOffset(r.config.Replication.TableName, r.config.Replication.SiteIdentifier, newLastID); err != nil {
		r.metrics.IncrementReplicationErrors()
		if r.config.Replication.EnableLogging {
			log.Printf("Error updating offset: %v", err)
		}
		return
	}

	r.metrics.AddReplicatedRecords(float64(len(records)))
	
	isSync, _, err := r.dbManager.ValidateSync()
	if err != nil {
		if r.config.Replication.EnableLogging {
			log.Printf("Error validating sync: %v", err)
		}
	} else {
		r.metrics.SetSyncStatus(isSync)
	}

	duration := time.Since(startTime)
	
	if r.config.Replication.EnableLogging {
		log.Printf("Replicated %d records in %v (last ID: %d)", len(records), duration, newLastID)
	}
}

func (r *Replicator) GetStatus() map[string]interface{} {
	lastID, err := r.dbManager.GetLastOffset(r.config.Replication.TableName, r.config.Replication.SiteIdentifier)
	if err != nil {
		lastID = 0
	}

	isSync, syncInfo, err := r.dbManager.ValidateSync()
	if err != nil {
		isSync = false
		syncInfo = map[string]interface{}{
			"error": err.Error(),
		}
	}

	return map[string]interface{}{
		"table_name":      r.config.Replication.TableName,
		"site_identifier": r.config.Replication.SiteIdentifier,
		"last_replicated_id": lastID,
		"is_sync":         isSync,
		"sync_info":       syncInfo,
		"status":          "running",
	}
}
