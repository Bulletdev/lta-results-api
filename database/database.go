package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var client *mongo.Client

// Connect estabelece conexão com o MongoDB
func Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		return err
	}

	// Verificar a conexão
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}

	log.Println("Conexão com o MongoDB estabelecida com sucesso")
	return nil
}

// Close encerra a conexão com o MongoDB
func Close() {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		log.Printf("Erro ao fechar conexão com o MongoDB: %v", err)
	}
}

// GetCollection retorna uma coleção específica
func GetCollection(name string) *mongo.Collection {
	return client.Database(os.Getenv("MONGODB_DATABASE")).Collection(name)
}
