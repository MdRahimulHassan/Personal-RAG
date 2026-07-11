package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/stackloklabs/gorag/pkg/backend"
	"github.com/stackloklabs/gorag/pkg/db"
)

const (
	ollamaHost  = "http://localhost:11434"
	embedModel  = "mxbai-embed-large"
	genModel    = "llama3.2"
	databaseURL = "postgres://user:password@127.0.0.1:5433/dbname?sslmode=disable")

var (
	embeddingBackend  *backend.OllamaBackend
	generationBackend *backend.OllamaBackend
	vectorDB          *db.PGVector
)

func initBackends() error {
	embeddingBackend = backend.NewOllamaBackend(ollamaHost, embedModel, 30*time.Second)
	generationBackend = backend.NewOllamaBackend(ollamaHost, genModel, 60*time.Second)

	var err error
	vectorDB, err = db.NewPGVector(databaseURL)
	if err != nil {
		return fmt.Errorf("connecting to postgres: %w", err)
	}
	return nil
}

// ingestFile extracts text from a PDF, chunks it, embeds each chunk, and stores it.
func ingestFile(ctx context.Context, path string) (int, error) {
	text, err := extractText(path)
	if err != nil {
		return 0, fmt.Errorf("extracting pdf text: %w", err)
	}

	chunks := chunkText(text, 800, 150)
	//headers := map[string]string{"Content-Type": "application/json"}

	for i, chunk := range chunks {
		embedding, err := embeddingBackend.Embed(ctx, chunk)
		if err != nil {
			return i, fmt.Errorf("embedding chunk %d: %w", i, err)
		}
		if err := vectorDB.InsertDocument(ctx, chunk, embedding); err != nil {
			return i, fmt.Errorf("inserting chunk %d: %w", i, err)
		}
	}
	return len(chunks), nil
}

// answerQuestion embeds the query, retrieves relevant chunks, and generates a grounded answer.
func answerQuestion(ctx context.Context, question string) (string, error) {
	//headers := map[string]string{"Content-Type": "application/json"}

	queryEmbedding, err := embeddingBackend.Embed(ctx, question)
	if err != nil {
		return "", fmt.Errorf("embedding query: %w", err)
	}

	retrievedDocs, err := vectorDB.QueryRelevantDocuments(ctx, queryEmbedding, "ollama")
	if err != nil {
		return "", fmt.Errorf("retrieving documents: %w", err)
	}

	augmentedQuery := db.CombineQueryWithContext(question, retrievedDocs)

	prompt := backend.NewPrompt().
		AddMessage("system", "You are an AI assistant. Answer only using the provided context. If the context doesn't contain the answer, say so.").
		AddMessage("user", augmentedQuery).
		SetParameters(backend.Parameters{
			MaxTokens:   500,
			Temperature: 0.3,
			TopP:        0.9,
		})

	response, err := generationBackend.Generate(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("generating response: %w", err)
	}
	return response, nil
}

func main() {
	if err := initBackends(); err != nil {
		log.Fatal(err)
	}
	defer vectorDB.Close()

	startServer(":8080")
}
