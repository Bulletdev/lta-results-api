package api

import (
	"log" // Adicione este import
	"net/http"
	"os"
	"strings" // Adicione este import

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	// Log na inicialização do middleware
	adminAPIKey := os.Getenv("ADMIN_API_KEY")
	log.Printf("Middleware inicializado. ADMIN_API_KEY configurada: %v", adminAPIKey != "")

	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		// Remover espaços em branco que possam estar causando problemas
		apiKey = strings.TrimSpace(apiKey)

		log.Printf("Requisição recebida. API Key fornecida: '%s'", apiKey)

		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key ausente"})
			return
		}

		adminAPIKey := os.Getenv("ADMIN_API_KEY")
		log.Printf("Comparando chaves: Recebida='%s', Esperada='%s', Iguais=%v",
			apiKey, adminAPIKey, apiKey == adminAPIKey)

		if apiKey != adminAPIKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key inválida"})
			return
		}

		c.Next()
	}
}
