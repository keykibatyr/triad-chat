package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/keykibatyr/triad-chat/internal/chat/models"
	"github.com/keykibatyr/triad-chat/internal/chat/repository"
)

type MessageServiceInterface interface {
	SendMessage(ctx context.Context, msg *models.Message) error
}

type MessageService struct {
	ConvoRepo   repository.ConvoRepository
	MessageRepo repository.MessageRepository
}

func NewMessageService(
	convoRepo repository.ConvoRepository,
	messageRepo repository.MessageRepository,
) MessageServiceInterface {
	return &MessageService{
		ConvoRepo:   convoRepo,
		MessageRepo: messageRepo,
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
