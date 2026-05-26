package service

import (
	"context"
	"fmt"

	"github.com/keykibatyr/triad-chat/internal/ai/models"
	"google.golang.org/genai"
)

type AiServiceInterface interface {
	Ask(ctx context.Context, aiRequest models.AiRequest) (string, error) 
}

func NewAiService(client *genai.Client, modelName string) AiServiceInterface {
	return &AiService{
		Client: client,
		Model: modelName,
	}
}

type AiService struct {
	Client *genai.Client
	Model  string
	Config *genai.GenerateContentConfig
}



func (s *AiService) Ask(ctx context.Context, aiRequest models.AiRequest) (string, error) {
	config := &genai.GenerateContentConfig{}
	s.Config = config
	s.Config.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{{
				Text: aiRequest.SystemText,
			}},
		}

	contents := make([]*genai.Content, 0, len(aiRequest.Messages))

	for _, msg := range aiRequest.Messages {
		var content genai.Content
		switch msg.Role {
		case "ai": 
		content.Role = genai.RoleModel
		content.Parts = []*genai.Part{{Text: msg.Content}}
		contents = append(contents, &content)
		case "user":
		content.Role = genai.RoleUser
		content.Parts = []*genai.Part{{Text: msg.Content}}
		contents = append(contents, &content)
		}
	}

    fmt.Println("bug ai service contents")
	fmt.Println(contents)
	fmt.Println("bug ai service contents")

	result, err := s.Client.Models.GenerateContent(
		ctx,
		s.Model,
		contents,
		config,
	)

	fmt.Println("bug ai service")
	fmt.Println(result)
	fmt.Println("bug ai service")

	if err != nil {
		return "", fmt.Errorf("could not send request to AI: %w", err)
	}

	return result.Text(), nil
}
