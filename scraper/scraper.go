package scraper

import (
	"context"
	"fmt"
	"log"
	_ "net/http"
	"os"
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
	timeout    = 60 * time.Second
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

	// Para cada região, extrair os resultados
	for region, url := range LTA_URLS {
		log.Printf("Extraindo resultados da região %s...", region)
		log.Printf("URL: %s", url)

		// Tentar extrair os dados com retry
		var html string
		var err error
		for i := 0; i < maxRetries; i++ {
			html, err = extractHTML(url)
			if err == nil {
				break
			}
			log.Printf("Tentativa %d falhou: %v", i+1, err)
			time.Sleep(retryDelay)
		}

		if err != nil {
			log.Printf("Erro ao extrair dados da região %s após %d tentativas: %v", region, maxRetries, err)
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

// extractHTML extrai o HTML da página usando Chrome headless
func extractHTML(url string) (string, error) {
	// Definir o caminho para o executável do Chrome (ajuste se necessário)
	// Tenta os locais mais comuns no Windows.
	chromePath := ""
	possiblePaths := []string{
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		// Adicione outros caminhos se o Chrome estiver em local diferente
	}
	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			chromePath = p
			log.Printf("Usando executável do Chrome encontrado em: %s", chromePath)
			break
		}
	}

	// Configurar contexto para o Chrome headless
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// Se encontrou um caminho, usa ele
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	// Adiciona o caminho do executável se ele foi encontrado
	if chromePath != "" {
		opts = append(opts, chromedp.ExecPath(chromePath))
	} else {
		log.Println("AVISO: Não foi possível encontrar o executável do Chrome nos caminhos padrão. Tentando usar o PATH.")
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// Adicionar timeout
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	// Variável para armazenar o HTML extraído
	var html string

	// Navegar para a URL e extrair o HTML
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(".recent-matches", chromedp.ByQuery),
		chromedp.OuterHTML("body", &html),
	)

	if err != nil {
		return "", fmt.Errorf("erro ao extrair HTML: %v", err)
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
