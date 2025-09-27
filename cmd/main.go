package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"table-replication-service/internal/api"
	"table-replication-service/internal/config"
	"table-replication-service/internal/database"
	"table-replication-service/internal/metrics"
	"table-replication-service/internal/replication"
	"table-replication-service/pkg/utils"
)

func main() {
	configPath := "config.ini"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbManager, err := database.NewDatabaseManager(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database manager: %v", err)
	}
	defer dbManager.Close()

	metricsInstance := metrics.NewMetrics()

	replicator := replication.NewReplicator(dbManager, cfg, metricsInstance)

	go replicator.Start()

	apiServer := api.NewAPIServer(dbManager, replicator, cfg)

	server := &http.Server{
		Addr:    cfg.Server.Host + ":" + cfg.Server.Port,
		Handler: apiServer.GetRouter(),
	}

	go func() {
		log.Printf("Starting server on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	sigChan := utils.WaitForShutdownSignal()
	<-sigChan

	log.Println("Shutting down gracefully...")

	replicator.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
