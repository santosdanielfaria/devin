package models

import (
	"time"
)

type SimImeiBinding struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	MSISDN         string    `gorm:"column:msisdn;type:varchar(20);not null;index:idx_sim_imei_msisdn" json:"msisdn"`
	IMEI           *string   `gorm:"column:imei;type:varchar(15)" json:"imei"`
	AZ             string    `gorm:"column:az;type:varchar(10);default:'sa-east-1a'" json:"az"`
	Locked         bool      `gorm:"column:locked;type:tinyint(1);not null;default:0" json:"locked"`
	LastUpdateTime time.Time `gorm:"column:lastupdatetime;type:timestamp;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"lastupdatetime"`
}

func (SimImeiBinding) TableName() string {
	return "sim_imei_binding"
}

type ReplicationOffset struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Table        string    `gorm:"column:table_name;type:varchar(100);not null" json:"table_name"`
	LastID       uint64    `gorm:"column:last_id;not null;default:0" json:"last_id"`
	SiteID       string    `gorm:"column:site_id;type:varchar(20);not null" json:"site_id"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ReplicationOffset) TableName() string {
	return "replication_offset"
}
