package service

import (
	"context"
	"fmt"
	"strconv"

	aimodels "github.com/keykibatyr/triad-chat/internal/ai/models"
	"github.com/keykibatyr/triad-chat/internal/ai/service"
	"github.com/keykibatyr/triad-chat/internal/chat/models"
	"github.com/keykibatyr/triad-chat/internal/chat/repository"
)

type ConversationSummaryServiceInterface interface {
	Summerize(ctx context.Context, convoID int64, messageLimit int) (*models.ConvoSummary, error)
}

func NewConvoSummaryService(
	messageRepo repository.MessageRepository, 
	convoSummaryRepo repository.ConvoSummaryRepository, 
	ai service.AiServiceInterface) ConversationSummaryServiceInterface {
	return &ConversationSummaryService{
		MessageRepo:     messageRepo,
		ConoSummaryRepo: convoSummaryRepo,
		AI: ai,
	}
}

type ConversationSummaryService struct {
	MessageRepo     repository.MessageRepository
	ConoSummaryRepo repository.ConvoSummaryRepository
	AI  service.AiServiceInterface
}

func (s *ConversationSummaryService) Summerize(ctx context.Context, convoID int64, messageLimit int) (*models.ConvoSummary, error) {
	convoIDstring := strconv.FormatInt(convoID, 10)
	
	
	messages, err := s.MessageRepo.GetXMeesages(ctx, convoID, messageLimit)
	if err != nil {
		return nil, fmt.Errorf("could not get X messages")
	}

	lastMessageId := int(messages[0].ID)

	convoSummary, err := s.ConoSummaryRepo.GetConvoSummaryByConvoID(ctx, convoID)
	if err != nil {
		return nil, err
	}
	//set some default summary about the chat 
	if convoSummary == nil {
		convoSummary = &models.ConvoSummary{
		ConvoID:       convoIDstring,
		SummaryText:   "",
		LastMessageID: strconv.Itoa(lastMessageId),
		}

		convoSummary, err = s.ConoSummaryRepo.CreateConvoSummary(ctx, convoSummary)
		if err != nil {
			return nil, fmt.Errorf("could not create summary: %w", err)
		}
	}


	for i, j := 0, len(messages) - 1; i < j; i, j = i+1, j+1{
		messages[i], messages[j] = messages[j], messages[i]
		} 
		
		
	var aiMessages []aimodels.AiMessage

	for _, message := range messages {
		var aiMessage aimodels.AiMessage
		aiMessage.Role = message.SenderType
		aiMessage.Content = message.Content

		aiMessages = append(aiMessages, aiMessage)
	}

	persona := "You are an AI assistant in the room " + convoIDstring + ", answer the questions. Here is the summary for the chat so far: "
	specPrompt := ". Update the summary using the new messages"

	aiRequest := aimodels.AiRequest{
		Messages: aiMessages,
		SystemText: persona + convoSummary.SummaryText + specPrompt,
	}

	
	aiText, err := s.AI.Ask(ctx, aiRequest)
	if err != nil {
		return nil, fmt.Errorf("could not generate AI text: %w", err)
	}


	convoSummary = &models.ConvoSummary{
		ConvoID: convoIDstring,
		SummaryText: aiText,
		LastMessageID: strconv.Itoa(lastMessageId),
	}
	
	lastMessageID64, err := strconv.ParseInt(convoSummary.LastMessageID, 10, 64)
    
    if err != nil {
        fmt.Println("Error:", err)
        return nil, fmt.Errorf("could not convert lastMessageID to int64: %w", err)
    }
    
	err = s.ConoSummaryRepo.UpdateConvoSummary(ctx, convoID, lastMessageID64, convoSummary.SummaryText)
	if err != nil {
		return nil, fmt.Errorf("could not update the summary: %w", err)
	}

	return convoSummary, nil
}
