package service

import (
	"context"
	"fmt"
	"strconv"

	// aimodels "github.com/keykibatyr/triad-chat/internal/ai/models"

	aimodels "github.com/keykibatyr/triad-chat/internal/ai/models"
	"github.com/keykibatyr/triad-chat/internal/ai/service"
	"github.com/keykibatyr/triad-chat/internal/chat/models"
	"github.com/keykibatyr/triad-chat/internal/chat/repository"
)

const MessageLimit = 15

type Paginate struct {
	Messages []models.Message `json:"messages"`
	Cursor   int              `json:"cursor"`
}

type MessageServiceInterface interface {
	SendMessage(ctx context.Context, msg *models.Message) error
	ListMessagesAtLoad(ctx context.Context, convoID int64, messageLimit int) (*Paginate, error)
	ListMessagesAndCursor(ctx context.Context, convoID int64, messageLimit, cursor int) (*Paginate, error)
	MessageToAi(ctx context.Context, msg *models.Message) (*models.Message, error)
}

type MessageService struct {
	ConvoRepo        repository.ConvoRepository
	MessageRepo      repository.MessageRepository
	ConvoSummaryRepo repository.ConvoSummaryRepository
	AI               service.AiServiceInterface
}

func NewMessageService(
	convoRepo repository.ConvoRepository,
	messageRepo repository.MessageRepository,
	convoSummaryRepo repository.ConvoSummaryRepository,
	ai           service.AiServiceInterface,
) MessageServiceInterface {
	return &MessageService{
		ConvoRepo:        convoRepo,
		MessageRepo:      messageRepo,
		ConvoSummaryRepo: convoSummaryRepo,
		AI: ai,
	}
}

func (s *MessageService) SendMessage(ctx context.Context, msg *models.Message) error {
	userID, err := strconv.ParseInt(msg.UserID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid userID: %w", err)
	}
	convoID, err := strconv.ParseInt(msg.ConvoID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid userID: %w", err)
	}

	isParticipant, err := s.ConvoRepo.IsParticipant(ctx, convoID, userID)
	if err != nil {
		return fmt.Errorf("failed reading if participant: %w", err)
	}

	if isParticipant {
		_, err = s.MessageRepo.CreateMessage(ctx, msg)
		if err != nil {
			return fmt.Errorf("failed sending message to DB: %w", err)
		}

		return nil
	}

	return fmt.Errorf("user is not part of convo")
}

func (s *MessageService) ListMessagesAtLoad(ctx context.Context, convoID int64, messageLimit int) (*Paginate, error) {
	messages, cursor, err := s.MessageRepo.GetMeesages(ctx, convoID, messageLimit)
	if err != nil {
		return nil, fmt.Errorf("couldn't load messages on load: %w", err)
	}

	//ADD HMAC SIGNING
	fmt.Println(messages)
	fmt.Println("LENGTH OF THE MESSAGES: ", messages)
	if len(messages) == 0 {
		return &Paginate{
			Messages: []models.Message{},
			Cursor:   0,
		}, nil
	}

	messageLast := messages[len(messages)-1]
	cursor = int(messageLast.ID)

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	fmt.Println("CURSOR HERE1", cursor)

	return &Paginate{
		Messages: messages,
		Cursor:   cursor,
	}, nil
}

func (s *MessageService) ListMessagesAndCursor(ctx context.Context, convoID int64, messageLimit, cursor int) (*Paginate, error) {
	messages, cursor, err := s.MessageRepo.GetMessagesAndCursor(ctx, convoID, messageLimit, cursor)
	if err != nil {
		return nil, fmt.Errorf("couldn't load messages on load: %w", err)
	}

	//ADD HMAC SIGNING
	fmt.Println("LENGTH OF THE MESSAGES: ", messages)
	if len(messages) == 0 {
		return &Paginate{
			Messages: []models.Message{},
			Cursor:   0,
		}, nil
	}

	messageLast := messages[len(messages)-1]
	cursor = int(messageLast.ID)

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	fmt.Println("CURSOR HERE22", cursor)

	return &Paginate{
		Messages: messages,
		Cursor:   cursor,
	}, nil
}

func (s *MessageService) MessageToAi(ctx context.Context, msg *models.Message) (*models.Message, error) {
	convoID, err := strconv.ParseInt(msg.ConvoID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not convert the string into in64: %w", err)
	}

	convoSummary, err := s.ConvoSummaryRepo.GetConvoSummaryByConvoID(ctx, convoID)
	if err != nil {
		return nil, fmt.Errorf("could not get ConvoSummary by ConvoId: %w", err)
	}

	messages, err := s.MessageRepo.GetXMeesages(ctx, convoID, MessageLimit)
	if err != nil {
		return nil, fmt.Errorf("could not get the last X messages: %w", err)
	}


	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	var aiMessages []aimodels.AiMessage

	for _, message := range messages {
		var aiMessage aimodels.AiMessage
		aiMessage.Role = message.SenderType
		aiMessage.Content = message.Content

		aiMessages = append(aiMessages, aiMessage)
	}

	convoIDstring := strconv.FormatInt(convoID, 10)
	
	persona := "You are an AI assistant in the room " + convoIDstring + ", answer the questions. Here is the summary for the chat so far: "
	specPrompt := "generate the answer based on the last message and using thesummary and rest of the messages as the context if needed"

	aiRequest := aimodels.AiRequest{
		Messages: aiMessages,
		SystemText: persona + convoSummary.SummaryText + specPrompt,
	}


	aiText, err := s.AI.Ask(ctx, aiRequest)
	if err != nil {
		return nil, fmt.Errorf("could not generate AI text: %w", err)
	}

	message := &models.Message{
		Type: "ai_message",
		SenderType: "ai",
		Content: aiText,
		ConvoID: msg.ConvoID,
		UserID: msg.UserID,
	}

	message, err = s.MessageRepo.CreateMessage(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("could not store ai message in DB: %w", err)
	}

	return message, nil

}
