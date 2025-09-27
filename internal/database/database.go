package database

import (
	"fmt"
	"table-replication-service/internal/config"
	"table-replication-service/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseManager struct {
	SourceDB *gorm.DB
	TargetDB *gorm.DB
	config   *config.Config
}

func NewDatabaseManager(cfg *config.Config) (*DatabaseManager, error) {
	dm := &DatabaseManager{
		config: cfg,
	}

	sourceDB, err := dm.initConnection(&cfg.DatabaseSource)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to source database: %w", err)
	}
	dm.SourceDB = sourceDB

	targetDB, err := dm.initConnection(&cfg.DatabaseTarget)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to target database: %w", err)
	}
	dm.TargetDB = targetDB

	if err := dm.autoMigrate(); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate tables: %w", err)
	}

	return dm, nil
}

func (dm *DatabaseManager) initConnection(dbConfig *config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch dbConfig.Driver {
	case "mysql":
		dialector = mysql.Open(dbConfig.GetDSN())
	case "postgres":
		dialector = postgres.Open(dbConfig.GetDSN())
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", dbConfig.Driver)
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	if dm.config.Replication.EnableLogging {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (dm *DatabaseManager) autoMigrate() error {
	if err := dm.SourceDB.AutoMigrate(&models.ReplicationOffset{}); err != nil {
		return fmt.Errorf("failed to migrate source replication_offset table: %w", err)
	}

	if err := dm.TargetDB.AutoMigrate(&models.ReplicationOffset{}); err != nil {
		return fmt.Errorf("failed to migrate target replication_offset table: %w", err)
	}

	if err := dm.SourceDB.AutoMigrate(&models.SimImeiBinding{}); err != nil {
		return fmt.Errorf("failed to verify source sim_imei_binding table: %w", err)
	}

	if err := dm.TargetDB.AutoMigrate(&models.SimImeiBinding{}); err != nil {
		return fmt.Errorf("failed to verify target sim_imei_binding table: %w", err)
	}

	return nil
}

func (dm *DatabaseManager) Close() error {
	if dm.SourceDB != nil {
		if sqlDB, err := dm.SourceDB.DB(); err == nil {
			sqlDB.Close()
		}
	}

	if dm.TargetDB != nil {
		if sqlDB, err := dm.TargetDB.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return nil
}

func (dm *DatabaseManager) GetNewRecords(lastID uint64, limit int) ([]models.SimImeiBinding, error) {
	var records []models.SimImeiBinding
	
	err := dm.SourceDB.Where("id > ?", lastID).
		Order("id ASC").
		Limit(limit).
		Find(&records).Error
	
	return records, err
}

func (dm *DatabaseManager) InsertRecords(records []models.SimImeiBinding) error {
	if len(records) == 0 {
		return nil
	}

	return dm.TargetDB.CreateInBatches(records, len(records)).Error
}

func (dm *DatabaseManager) GetLastOffset(tableName, siteID string) (uint64, error) {
	var offset models.ReplicationOffset
	
	err := dm.SourceDB.Where("table_name = ? AND site_id = ?", tableName, siteID).
		First(&offset).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	
	return offset.LastID, nil
}

func (dm *DatabaseManager) UpdateOffset(tableName, siteID string, lastID uint64) error {
	offset := models.ReplicationOffset{
		Table:  tableName,
		LastID: lastID,
		SiteID: siteID,
	}

	return dm.SourceDB.Where("table_name = ? AND site_id = ?", tableName, siteID).
		Assign(models.ReplicationOffset{LastID: lastID}).
		FirstOrCreate(&offset).Error
}

func (dm *DatabaseManager) ValidateSync() (bool, map[string]interface{}, error) {
	var sourceCount, targetCount int64
	
	if err := dm.SourceDB.Model(&models.SimImeiBinding{}).Count(&sourceCount).Error; err != nil {
		return false, nil, fmt.Errorf("failed to count source records: %w", err)
	}
	
	if err := dm.TargetDB.Model(&models.SimImeiBinding{}).Count(&targetCount).Error; err != nil {
		return false, nil, fmt.Errorf("failed to count target records: %w", err)
	}
	
	isSync := sourceCount == targetCount
	
	result := map[string]interface{}{
		"source_count": sourceCount,
		"target_count": targetCount,
		"is_sync":      isSync,
		"difference":   sourceCount - targetCount,
	}
	
	return isSync, result, nil
}
