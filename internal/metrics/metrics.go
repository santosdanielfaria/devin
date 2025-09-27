package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	ReplicatedRecords prometheus.Counter
	ReplicationErrors prometheus.Counter
	SyncStatus        prometheus.Gauge
}

func NewMetrics() *Metrics {
	return &Metrics{
		ReplicatedRecords: promauto.NewCounter(prometheus.CounterOpts{
			Name: "table_replication_records_total",
			Help: "The total number of replicated records",
		}),
		ReplicationErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "table_replication_errors_total",
			Help: "The total number of replication errors",
		}),
		SyncStatus: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "table_replication_sync_status",
			Help: "The sync status between source and target databases (1 = synced, 0 = not synced)",
		}),
	}
}

func (m *Metrics) AddReplicatedRecords(count float64) {
	m.ReplicatedRecords.Add(count)
}

func (m *Metrics) IncrementReplicationErrors() {
	m.ReplicationErrors.Inc()
}

func (m *Metrics) SetSyncStatus(isSync bool) {
	if isSync {
		m.SyncStatus.Set(1)
	} else {
		m.SyncStatus.Set(0)
	}
}
