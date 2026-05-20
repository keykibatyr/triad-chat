package client

import (
	"context"
	"log"

	"google.golang.org/genai"
)

func NewClient(ctx context.Context, apikey string) (*genai.Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apikey,
		Backend: genai.BackendGeminiAPI,
	})

	if err != nil {
		log.Fatal(err)
	}

	return client, err
}
