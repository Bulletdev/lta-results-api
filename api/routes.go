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
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	// Middleware de autenticação para rotas protegidas
	authMiddleware := AuthMiddleware()

	// Health Check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
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
