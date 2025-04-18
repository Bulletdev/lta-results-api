package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	// Configurar modo de execução
	if gin.Mode() == gin.DebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Configurar CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "X-API-Key"}
	router.Use(cors.New(config))

	// Middleware de autenticação para rotas protegidas
	authMiddleware := AuthMiddleware()

	// Health Check
	router.GET("/health", func(c *gin.Context) {
		// Adicionar headers importantes
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		// Responder com status 200 e dados
		c.JSON(200, gin.H{
			"status":  "healthy",
			"time":    time.Now().Format(time.RFC3339),
			"version": "1.0.0",
		})
	})

	// Adicionar rota OPTIONS para o health check
	router.OPTIONS("/health", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
		c.Status(204)
	})

	// Rotas públicas
	v1 := router.Group("/api/v1")
	{
		// Resultados de partidas
		v1.GET("/results", GetMatchResults)
		v1.GET("/results/:matchId", GetMatchResultByID)

		// Estatísticas de jogadores
		v1.GET("/players/:playerName/stats", GetPlayerStats)

		// Estatísticas de times
		v1.GET("/teams/:teamName/stats", GetTeamStats)

		// Rotas protegidas (admin)
		admin := v1.Group("/admin")
		admin.Use(authMiddleware)
		{
			admin.POST("/scrape", TriggerScraping)
			admin.POST("/results", CreateMatchResult)
			admin.PUT("/results/:matchId", UpdateMatchResult)
			admin.DELETE("/results/:matchId", DeleteMatchResult)
		}
	}

	return router
}
