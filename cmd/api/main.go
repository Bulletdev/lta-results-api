package main

import (
	"context"
	_ "fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bulletdev/lta-results-api/api"
	"github.com/bulletdev/lta-results-api/database"
	"github.com/bulletdev/lta-results-api/scraper"
)

func main() {
	// Conectar ao banco de dados
	if err := database.Connect(); err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer database.Close()

	// Configurar API
	router := api.SetupRouter()
	server := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: router,
	}

	// Configurar cron job para scraping
	go scraper.ScheduleScraping()

	// Iniciar servidor em uma goroutine
	go func() {
		log.Printf("Servidor API iniciado na porta %s", os.Getenv("PORT"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar servidor: %v", err)
		}
	}()

	// Configurar graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Desligando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Erro ao desligar servidor: %v", err)
	}

	log.Println("Servidor encerrado com sucesso")
}
