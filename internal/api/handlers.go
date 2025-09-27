package api

import (
	"net/http"
	"table-replication-service/internal/config"
	"table-replication-service/internal/database"
	"table-replication-service/internal/replication"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type APIServer struct {
	router     *gin.Engine
	dbManager  *database.DatabaseManager
	replicator *replication.Replicator
	config     *config.Config
}

func NewAPIServer(dbManager *database.DatabaseManager, replicator *replication.Replicator, cfg *config.Config) *APIServer {
	router := gin.Default()
	
	server := &APIServer{
		router:     router,
		dbManager:  dbManager,
		replicator: replicator,
		config:     cfg,
	}
	
	server.setupRoutes()
	return server
}

func (s *APIServer) setupRoutes() {
	s.router.GET("/health-check", s.healthCheck)
	
	s.router.GET("/validation", s.validation)
	
	s.router.GET("/status", s.status)
	
	if s.config.Metrics.EnablePrometheus {
		s.router.GET(s.config.Metrics.MetricsPath, gin.WrapH(promhttp.Handler()))
	}
	
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (s *APIServer) healthCheck(c *gin.Context) {
	sourceDB, err := s.dbManager.SourceDB.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "source database connection failed",
		})
		return
	}
	
	if err := sourceDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "source database ping failed",
		})
		return
	}
	
	targetDB, err := s.dbManager.TargetDB.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "target database connection failed",
		})
		return
	}
	
	if err := targetDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "target database ping failed",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "table-replication-service",
	})
}

func (s *APIServer) validation(c *gin.Context) {
	isSync, result, err := s.dbManager.ValidateSync(s.config.Replication.SourceAZ, s.config.Replication.TargetAZ)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	status := http.StatusOK
	if !isSync {
		status = http.StatusConflict
	}
	
	c.JSON(status, result)
}

func (s *APIServer) status(c *gin.Context) {
	status := s.replicator.GetStatus()
	c.JSON(http.StatusOK, status)
}

func (s *APIServer) Start() error {
	address := s.config.Server.Host + ":" + s.config.Server.Port
	return s.router.Run(address)
}

func (s *APIServer) GetRouter() *gin.Engine {
	return s.router
}
