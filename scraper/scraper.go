package scraper

import (
	"context"
	"fmt"
	"log"
	_ "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bulletdev/lta-results-api/models"
	"github.com/chromedp/chromedp"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// URLs para scraping
var LTA_URLS = map[string]string{
	"sul":   "https://maisesports.com.br/campeonatos/league-of-legends-lta-sul-split-2-2025/",
	"norte": "https://maisesports.com.br/campeonatos/league-of-legends-lta-norte-split-2-2025/",
}

// Configurações do scraper
const (
	maxRetries = 3
	retryDelay = 5 * time.Second
	timeout    = 180 * time.Second
)

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

	// --- Início da Criação Única do Contexto Chrome ---
	// Configurar contexto para o Chrome headless
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-zygote", true),
		chromedp.Flag("user-data-dir", "/app/chrome-data"),
		chromedp.UserDataDir("/app/chrome-data"),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	// Criar contexto principal do chromedp
	mainCtx, cancelCtx := chromedp.NewContext(
		allocCtx,
		chromedp.WithErrorf(log.Printf),
	)
	defer cancelCtx()

	// Criar um contexto com timeout a partir do contexto principal do chromedp
	timeoutCtx, cancelTimeout := context.WithTimeout(mainCtx, timeout)
	defer cancelTimeout()
	// --- Fim da Criação Única do Contexto Chrome ---

	// Para cada região, extrair os resultados usando o mesmo contexto
	for region, url := range LTA_URLS {
		log.Printf("Extraindo resultados da região %s...", region)
		log.Printf("URL: %s", url)

		// Tentar extrair os dados com retry
		var html string
		var err error
		for i := 0; i < maxRetries; i++ {
			// Passar o contexto com timeout para extractHTML
			html, err = extractHTML(timeoutCtx, url) // Passa o contexto
			if err == nil {
				break
			}
			// Se o erro for context deadline exceeded, não adianta tentar de novo com o mesmo contexto
			if err == context.DeadlineExceeded {
				log.Printf("Timeout atingido na tentativa %d para %s. Abortando retries.", i+1, region)
				break
			}
			log.Printf("Tentativa %d falhou para %s: %v", i+1, region, err)
			time.Sleep(retryDelay)
		}

		if err != nil {
			log.Printf("Erro ao extrair dados da região %s após tentativas: %v", region, err)
			continue
		}

		// Processar o HTML
		matchResults, err := parseHTML(html, region)
		if err != nil {
			log.Printf("Erro ao processar HTML da região %s: %v", region, err)
			continue
		}

		log.Printf("Processados %d resultados da região %s", len(matchResults), region)

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

// extractHTML agora recebe um contexto chromedp existente
func extractHTML(ctx context.Context, url string) (string, error) {
	// Variável para armazenar o HTML extraído
	var html string

	// Navegar para a URL e extrair o HTML usando o contexto fornecido
	err := chromedp.Run(ctx, // Usa o contexto passado como argumento
		chromedp.Navigate(url),
		chromedp.WaitVisible(".recent-matches", chromedp.ByQuery), // Ajuste o seletor se necessário
		// Extrair o HTML do container específico em vez do body inteiro
		chromedp.OuterHTML(".recent-matches", &html, chromedp.ByQuery),
	)

	if err != nil {
		// Retornar o erro original, incluindo context deadline exceeded
		return "", fmt.Errorf("erro durante execução do chromedp: %w", err)
	}

	return html, nil
}

// parseHTML processa o HTML extraído para obter resultados de partidas
func parseHTML(html, region string) ([]*models.MatchResult, error) {
	// Criar um novo documento goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar documento: %v", err)
	}

	var results []*models.MatchResult

	// Encontrar todos os cards de partida
	doc.Find(".match-card").Each(func(i int, s *goquery.Selection) {
		// Extrair dados básicos da partida
		matchID, _ := s.Attr("data-match-id")
		if matchID == "" {
			matchID = fmt.Sprintf("%s-%d", region, i)
		}

		dateStr := s.Find(".match-date").Text()
		teamA := s.Find(".team-a .team-name").Text()
		teamB := s.Find(".team-b .team-name").Text()
		scoreAStr := s.Find(".team-a .score").Text()
		scoreBStr := s.Find(".team-b .score").Text()

		// Converter valores
		date, err := time.Parse("02 Jan 2006", strings.TrimSpace(dateStr))
		if err != nil {
			log.Printf("Erro ao converter data: %v", err)
			return
		}

		scoreA, err := strconv.Atoi(strings.TrimSpace(scoreAStr))
		if err != nil {
			log.Printf("Erro ao converter score A: %v", err)
			return
		}

		scoreB, err := strconv.Atoi(strings.TrimSpace(scoreBStr))
		if err != nil {
			log.Printf("Erro ao converter score B: %v", err)
			return
		}

		// Determinar vencedor
		var winner string
		if scoreA > scoreB {
			winner = teamA
		} else if scoreB > scoreA {
			winner = teamB
		} else {
			winner = "Empate"
		}

		// Extrair informações dos jogadores
		var players []models.Player
		s.Find(".player-stats").Each(func(_ int, playerSel *goquery.Selection) {
			player := models.Player{
				Name:        strings.TrimSpace(playerSel.Find(".player-name").Text()),
				Team:        strings.TrimSpace(playerSel.Find(".team-name").Text()),
				Position:    strings.TrimSpace(playerSel.Find(".position").Text()),
				Champion:    strings.TrimSpace(playerSel.Find(".champion").Text()),
				Kills:       parseInt(playerSel.Find(".kills").Text()),
				Deaths:      parseInt(playerSel.Find(".deaths").Text()),
				Assists:     parseInt(playerSel.Find(".assists").Text()),
				CS:          parseInt(playerSel.Find(".cs").Text()),
				Gold:        parseInt(playerSel.Find(".gold").Text()),
				DamageDealt: parseInt(playerSel.Find(".damage-dealt").Text()),
				VisionScore: parseInt(playerSel.Find(".vision-score").Text()),
			}
			players = append(players, player)
		})

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
			Players:   players,
			Winner:    winner,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Adicionar à lista de resultados
		results = append(results, result)
	})

	return results, nil
}

// parseInt converte string para int com tratamento de erro
func parseInt(s string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return i
}
