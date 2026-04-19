package service

import (
	"context"

	"github.com/keykibatyr/triad-chat/internal/chat/models"
	"github.com/keykibatyr/triad-chat/internal/chat/repository"
)

type ConversationServiceInterface interface {
	SaveConvoToDB(ctx context.Context, id, name, convoType string) (*models.Convo, error)
}

type ConversationService struct {
	ConvoRepo repository.ConvoRepository
}

func NewConvoService(convoRepo repository.ConvoRepository) ConversationServiceInterface {
	return &ConversationService{
		ConvoRepo: convoRepo,
	}
}

func (s *ConversationService) SaveConvoToDB(ctx context.Context, id, name, convoType string) (*models.Convo, error) {
	// convoID, err := strconv.ParseInt(id, 10, 64)
	// if err != nil {
	// 	return fmt.Errorf("fail at converting id to int64: %w", err)
	// }

	// convoTypes := []string{"triad", "duo"}

	// if !slices.Contains(convoTypes, convoType) {
	// 	convoType = "triad"
	// }

	// query := `INSERT INT`

	return nil, nil
}
