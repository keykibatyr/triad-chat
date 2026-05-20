package service

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/keykibatyr/triad-chat/internal/chat/models"
	"github.com/keykibatyr/triad-chat/internal/chat/repository"
)

type ConversationServiceInterface interface {
	SaveConvoToDB(ctx context.Context, name, convoType string) (*models.Convo, error)
	GetConvosFromDB(ctx context.Context, name string) ([]models.Convo, error)
	GetConvoByID(ctx context.Context, id int64) (*models.Convo, error)
	AddParticipant(ctx context.Context, convoID, userID int64) error
	IsParticipant(ctx context.Context, convoID, userID int64) (bool,error)
}

type ConversationService struct {
	ConvoRepo repository.ConvoRepository
}

func NewConvoService(convoRepo repository.ConvoRepository) ConversationServiceInterface {
	return &ConversationService{
		ConvoRepo: convoRepo,
	}
}

func (s *ConversationService) SaveConvoToDB(ctx context.Context, name, convoType string) (*models.Convo, error) {
	convoTypes := []string{"triad", "duo"}

	if !slices.Contains(convoTypes, convoType) {
		convoType = "triad"
	}

	convo := models.Convo{
		Type: convoType,
		Name: name,
	}

	return s.ConvoRepo.CreateConvo(ctx, &convo)

}

func (s *ConversationService) GetConvosFromDB(ctx context.Context, name string) ([]models.Convo, error) {
	name = strings.ToLower(name)

	convoList, err := s.ConvoRepo.GetConvoListByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("fail retrieving convo list: %w", err)
	}

	return convoList, nil
}

func (s *ConversationService) GetConvoByID(ctx context.Context, id int64) (*models.Convo, error) {
	return s.ConvoRepo.GetConvoByID(ctx, id)
}

func(s *ConversationService) AddParticipant(ctx context.Context, convoID, userID int64) error {
	return s.ConvoRepo.AddParticipant(ctx, convoID, userID)
}

func (s *ConversationService) 	IsParticipant(ctx context.Context, convoID, userID int64) (bool,error) {
	return s.ConvoRepo.IsParticipant(ctx, convoID, userID)
}