package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bulletdev/lta-results-api/models"
	"github.com/bulletdev/lta-results-api/scraper"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetMatchResults(c *gin.Context) {
	// Parâmetros de consulta
	region := c.Query("region")
	team := c.Query("team")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	skip := (page - 1) * limit

	// Construir filtro
	filter := bson.M{}
	if region != "" {
		filter["region"] = region
	}
	if team != "" {
		filter["$or"] = []bson.M{
			{"teama": team},
			{"teamb": team},
		}
	}

	// Opções de consulta
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"date": -1})
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	// Executar consulta
	results, total, err := models.GetMatchResults(filter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar resultados"})
		return
	}

	// Construir resposta com paginação
	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func GetMatchResultByID(c *gin.Context) {
	matchID := c.Param("matchId")

	// Buscar resultado por ID
	result, err := models.GetMatchResultByID(matchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resultado não encontrado"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func GetPlayerStats(c *gin.Context) {
	playerName := c.Param("playerName")

	// Buscar estatísticas do jogador
	stats, err := models.GetPlayerStats(playerName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar estatísticas"})
		return
	}

	if stats == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Jogador não encontrado"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func GetTeamStats(c *gin.Context) {
	teamName := c.Param("teamName")

	// Buscar estatísticas do time
	stats, err := models.GetTeamStats(teamName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar estatísticas"})
		return
	}

	if stats == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Time não encontrado"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func TriggerScraping(c *gin.Context) {
	go scraper.ScrapeMatchResults()
	c.JSON(http.StatusOK, gin.H{"message": "Scraping iniciado com sucesso"})
}

func CreateMatchResult(c *gin.Context) {
	var matchResult models.MatchResult

	if err := c.ShouldBindJSON(&matchResult); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Gerar novo ID se não fornecido
	if matchResult.ID.IsZero() {
		matchResult.ID = primitive.NewObjectID()
	}

	// Definir datas de criação e atualização
	now := time.Now()
	matchResult.CreatedAt = now
	matchResult.UpdatedAt = now

	// Inserir no banco de dados
	if err := models.CreateMatchResult(&matchResult); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar resultado"})
		return
	}

	c.JSON(http.StatusCreated, matchResult)
}

func UpdateMatchResult(c *gin.Context) {
	matchID := c.Param("matchId")

	var matchResult models.MatchResult
	if err := c.ShouldBindJSON(&matchResult); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Atualizar data de atualização
	matchResult.UpdatedAt = time.Now()

	// Atualizar no banco de dados
	if err := models.UpdateMatchResult(matchID, &matchResult); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar resultado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Resultado atualizado com sucesso"})
}

func DeleteMatchResult(c *gin.Context) {
	matchID := c.Param("matchId")

	// Excluir do banco de dados
	if err := models.DeleteMatchResult(matchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir resultado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Resultado excluído com sucesso"})
}
