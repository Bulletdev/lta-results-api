package scraper

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bulletdev/lta-results-api/models"
	"github.com/chromedp/chromedp"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// URLs para scraping
var LTA_URLS = map[string]string{
	"sul":   "https://lolesports.com/standings/lta-sul/lta_2025_split1/regular_season",
	"norte": "https://lolesports.com/standings/lta-norte/lta_2025_split1/regular_season",
}

// ScheduleScraping configura o agendamento do scraping
func ScheduleScraping() {
	c := cron.New()

	// Agendar scraping diário às 2h da manhã
	_, err := c.AddFunc("0 2 * * *", func() {
		log.Println("Executando scraping agendado...")
		if err := ScrapeMatchResults(); err != nil {
			log.Printf("Erro no scraping agendado: %v", err)
		}
	})

	if err != nil {
		log.Printf("Erro ao configurar agendamento: %v", err)
		return
	}

	c.Start()
	log.Println("Scraping agendado configurado com sucesso")
}

// ScrapeMatchResults extrai os resultados de partidas
func ScrapeMatchResults() error {
	log.Println("Iniciando extração de resultados de partidas...")

	// Para cada região, extrair os resultados
	for region, url := range LTA_URLS {
		log.Printf("Extraindo resultados da região %s...", region)

		// Configurar contexto para o Chrome headless
		ctx, cancel := chromedp.NewContext(
			context.Background(),
			chromedp.WithLogf(log.Printf),
		)
		defer cancel()

		// Adicionar timeout
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		// Variáveis para armazenar o HTML extraído
		var html string

		// Navegar para a URL e extrair o HTML
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.WaitVisible(".recent-matches", chromedp.ByQuery),
			chromedp.OuterHTML("body", &html),
		)

		if err != nil {
			log.Printf("Erro ao extrair dados da região %s: %v", region, err)
			continue
		}

		// Processar o HTML
		matchResults, err := parseHTML(html, region)
		if err != nil {
			log.Printf("Erro ao processar HTML da região %s: %v", region, err)
			continue
		}

		// Salvar os resultados
		for _, result := range matchResults {
			if err := models.CreateMatchResult(result); err != nil {
				log.Printf("Erro ao salvar resultado %s: %v", result.MatchID, err)
			} else {
				log.Printf("Resultado %s salvo com sucesso", result.MatchID)
			}
		}

		log.Printf("Extração da região %s concluída!", region)
	}

	log.Println("Extração de resultados concluída!")
	return nil
}

// parseHTML processa o HTML extraído para obter resultados de partidas
func parseHTML(html, region string) ([]*models.MatchResult, error) {
	// Implementação simplificada para o exemplo
	// Em uma implementação real, você usaria um parser HTML como goquery

	var results []*models.MatchResult

	// Extrair blocos de partidas
	matchBlocks := strings.Split(html, "class=\"match-card\"")

	for i, block := range matchBlocks {
		if i == 0 {
			continue // Pular o primeiro split que não contém partida
		}

		// Extrair dados básicos da partida
		matchID := extractDataAttribute(block, "data-match-id")
		if matchID == "" {
			matchID = fmt.Sprintf("%s-%d", region, i)
		}

		dateStr := extractText(block, "class=\"match-date\"")
		teamA := extractText(block, "class=\"team-a team-name\"")
		teamB := extractText(block, "class=\"team-b team-name\"")
		scoreAStr := extractText(block, "class=\"team-a score\"")
		scoreBStr := extractText(block, "class=\"team-b score\"")

		// Converter valores
		date, _ := time.Parse("02 Jan 2006", dateStr)
		scoreA, _ := strconv.Atoi(scoreAStr)
		scoreB, _ := strconv.Atoi(scoreBStr)

		// Determinar vencedor
		var winner string
		if scoreA > scoreB {
			winner = teamA
		} else if scoreB > scoreA {
			winner = teamB
		} else {
			winner = "Empate"
		}

		// Criar objeto de resultado
		result := &models.MatchResult{
			ID:        primitive.NewObjectID(),
			MatchID:   matchID,
			Date:      date,
			TeamA:     teamA,
			TeamB:     teamB,
			ScoreA:    scoreA,
			ScoreB:    scoreB,
			Region:    region,
			Winner:    winner,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Adicionar à lista de resultados
		results = append(results, result)
	}

	return results, nil
}

// extractText extrai texto de um elemento HTML
func extractText(html, selector string) string {
	start := strings.Index(html, selector)
	if start == -1 {
		return ""
	}

	// Avançar para o conteúdo
	contentStartOffset := strings.Index(html[start:], ">")
	if contentStartOffset == -1 {
		return ""
	}
	contentStart := start + contentStartOffset + 1

	// Encontrar o fim do elemento
	contentEndOffset := strings.Index(html[contentStart:], "<")
	if contentEndOffset == -1 {
		return ""
	}
	contentEnd := contentStart + contentEndOffset

	// Extrair e limpar o texto
	text := html[contentStart:contentEnd]
	return strings.TrimSpace(text)
}

// extractDataAttribute extrai valor de um atributo data-*
func extractDataAttribute(html, attrName string) string {
	attrStart := strings.Index(html, attrName+"=\"")
	if attrStart == -1 {
		return ""
	}

	valueStart := attrStart + len(attrName) + 2 // +2 para pular ="
	valueEndOffset := strings.Index(html[valueStart:], "\"")
	if valueEndOffset == -1 {
		return ""
	}
	valueEnd := valueStart + valueEndOffset

	return html[valueStart:valueEnd]
}
