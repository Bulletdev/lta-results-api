package models

import (
	"context"
	"sort"
	"strconv"
	"time"

	"github.com/bulletdev/lta-results-api/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MatchResult representa o resultado de uma partida
type MatchResult struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	MatchID         string             `bson:"matchId" json:"matchId"`
	Date            time.Time          `bson:"date" json:"date"`
	TeamA           string             `bson:"teamA" json:"teamA"`
	TeamB           string             `bson:"teamB" json:"teamB"`
	ScoreA          int                `bson:"scoreA" json:"scoreA"`
	ScoreB          int                `bson:"scoreB" json:"scoreB"`
	Region          string             `bson:"region" json:"region"`
	Players         []Player           `bson:"players" json:"players"`
	Duration        string             `bson:"duration" json:"duration"`
	Winner          string             `bson:"winner" json:"winner"`
	MVP             string             `bson:"mvp,omitempty" json:"mvp,omitempty"`
	TournamentStage string             `bson:"tournamentStage,omitempty" json:"tournamentStage,omitempty"`
	VOD             string             `bson:"vod,omitempty" json:"vod,omitempty"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Player representa um jogador em uma partida
type Player struct {
	Name        string `bson:"name" json:"name"`
	Team        string `bson:"team" json:"team"`
	Position    string `bson:"position" json:"position"`
	Champion    string `bson:"champion" json:"champion"`
	Kills       int    `bson:"kills" json:"kills"`
	Deaths      int    `bson:"deaths" json:"deaths"`
	Assists     int    `bson:"assists" json:"assists"`
	CS          int    `bson:"cs" json:"cs"`
	Gold        int    `bson:"gold" json:"gold"`
	DamageDealt int    `bson:"damageDealt" json:"damageDealt"`
	VisionScore int    `bson:"visionScore" json:"visionScore"`
}

// PlayerStats representa estatísticas agregadas de um jogador
type PlayerStats struct {
	PlayerName     string  `json:"playerName"`
	TotalGames     int     `json:"totalGames"`
	Wins           int     `json:"wins"`
	Losses         int     `json:"losses"`
	WinRate        float64 `json:"winRate"`
	AverageKills   float64 `json:"averageKills"`
	AverageDeaths  float64 `json:"averageDeaths"`
	AverageAssists float64 `json:"averageAssists"`
	AverageCS      float64 `json:"averageCS"`
	KDA            string  `json:"kda"`
}

// TeamStats representa estatísticas agregadas de um time
type TeamStats struct {
	TeamName            string          `json:"teamName"`
	TotalGames          int             `json:"totalGames"`
	Wins                int             `json:"wins"`
	Losses              int             `json:"losses"`
	WinRate             float64         `json:"winRate"`
	AverageGameDuration float64         `json:"averageGameDuration"`
	MostPlayedChampions []ChampionStats `json:"mostPlayedChampions"`
}

// ChampionStats representa estatísticas de um campeão
type ChampionStats struct {
	Champion string  `json:"champion"`
	Games    int     `json:"games"`
	Wins     int     `json:"wins"`
	WinRate  float64 `json:"winRate"`
}

// GetMatchResults obtém resultados de partidas com base em um filtro
func GetMatchResults(filter bson.M, options *options.FindOptions) ([]MatchResult, int64, error) {
	collection := database.GetCollection("match_results")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Contar total de documentos
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Executar consulta
	cursor, err := collection.Find(ctx, filter, options)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Decodificar resultados
	var results []MatchResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// GetMatchResultByID obtém um resultado específico por ID
func GetMatchResultByID(matchID string) (*MatchResult, error) {
	collection := database.GetCollection("match_results")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"matchId": matchID}
	var result MatchResult

	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetPlayerStats calcula estatísticas agregadas para um jogador
func GetPlayerStats(playerName string) (*PlayerStats, error) {
	collection := database.GetCollection("match_results")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Filtrar partidas onde o jogador participou
	filter := bson.M{"players.name": playerName}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var matches []MatchResult
	if err := cursor.All(ctx, &matches); err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, nil
	}

	// Calcular estatísticas
	stats := &PlayerStats{
		PlayerName: playerName,
		TotalGames: len(matches),
	}

	var totalKills, totalDeaths, totalAssists, totalCS int

	for _, match := range matches {
		for _, player := range match.Players {
			if player.Name == playerName {
				// Incrementar contadores
				totalKills += player.Kills
				totalDeaths += player.Deaths
				totalAssists += player.Assists
				totalCS += player.CS

				// Verificar se o jogador ganhou
				if match.Winner == player.Team {
					stats.Wins++
				}
			}
		}
	}

	stats.Losses = stats.TotalGames - stats.Wins
	stats.WinRate = float64(stats.Wins) / float64(stats.TotalGames) * 100

	if stats.TotalGames > 0 {
		stats.AverageKills = float64(totalKills) / float64(stats.TotalGames)
		stats.AverageDeaths = float64(totalDeaths) / float64(stats.TotalGames)
		stats.AverageAssists = float64(totalAssists) / float64(stats.TotalGames)
		stats.AverageCS = float64(totalCS) / float64(stats.TotalGames)
	}

	if totalDeaths > 0 {
		kda := float64(totalKills+totalAssists) / float64(totalDeaths)
		stats.KDA = strconv.FormatFloat(kda, 'f', 2, 64)
	} else {
		stats.KDA = "Perfect"
	}

	return stats, nil
}

// GetTeamStats calcula estatísticas agregadas para um time
func GetTeamStats(teamName string) (*TeamStats, error) {
	collection := database.GetCollection("match_results")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Filtrar partidas onde o time participou
	filter := bson.M{
		"$or": []bson.M{
			{"teamA": teamName},
			{"teamB": teamName},
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var matches []MatchResult
	if err := cursor.All(ctx, &matches); err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, nil
	}

	// Calcular estatísticas
	stats := &TeamStats{
		TeamName:   teamName,
		TotalGames: len(matches),
	}

	// Mapa para rastrear campeões
	champStats := make(map[string]*ChampionStats)

	for _, match := range matches {
		// Verificar se o time ganhou
		if match.Winner == teamName {
			stats.Wins++
		}

		// Rastrear campeões usados
		for _, player := range match.Players {
			if player.Team == teamName {
				if _, exists := champStats[player.Champion]; !exists {
					champStats[player.Champion] = &ChampionStats{
						Champion: player.Champion,
					}
				}

				cs := champStats[player.Champion]
				cs.Games++
				if match.Winner == teamName {
					cs.Wins++
				}
			}
		}
	}

	stats.Losses = stats.TotalGames - stats.Wins
	stats.WinRate = float64(stats.Wins) / float64(stats.TotalGames) * 100

	// Converter mapa de campeões em slice
	for _, cs := range champStats {
		cs.WinRate = float64(cs.Wins) / float64(cs.Games) * 100
		stats.MostPlayedChampions = append(stats.MostPlayedChampions, *cs)
	}

	// Ordenar campeões por número de jogos
	sort.Slice(stats.MostPlayedChampions, func(i, j int) bool {
		return stats.MostPlayedChampions[i].Games > stats.MostPlayedChampions[j].Games
	})

	// Limitar a 5 campeões mais jogados
	if len(stats.MostPlayedChampions) > 5 {
		stats.MostPlayedChampions = stats.MostPlayedChampions[:5]
	}

	return stats, nil
}

// CreateMatchResult insere um novo resultado de partida
func CreateMatchResult(result *MatchResult) error {
	collection := database.GetCollection("match_results")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, result)
	return err
}

// UpdateMatchResult atualiza um resultado existente
func UpdateMatchResult(matchID string, result *MatchResult) error {
	collection := database.GetCollection("match_results")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"matchId": matchID}
	update := bson.M{"$set": result}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// DeleteMatchResult exclui um resultado
func DeleteMatchResult(matchID string) error {
	collection := database.GetCollection("match_results")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"matchId": matchID}

	_, err := collection.DeleteOne(ctx, filter)
	return err
}
