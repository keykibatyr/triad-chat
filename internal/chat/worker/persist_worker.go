package worker

import (
	"context"
	"log"

	"github.com/keykibatyr/triad-chat/internal/chat/models"
	"github.com/keykibatyr/triad-chat/internal/chat/service"
)

const WorkerCount = 5

type PersistWorker struct {
	MessageService service.MessageServiceInterface
	Jobs           chan *models.Message
}

func NewPersistWorker(
	messageService service.MessageServiceInterface,
) *PersistWorker {
	return &PersistWorker{
		MessageService: messageService,
		Jobs:           make(chan *models.Message, 100),
	}
}

func (w *PersistWorker) SaveToDB(ctx context.Context) {
	for i := 0; i < WorkerCount; i++{
		go func() {
			for {
				select{
				case job, ok := <- w.Jobs:
					if !ok || job == nil{
						log.Printf("channel Jobs closed or there is no Job: %v and %v", ok, job)
						return 
					}
					err := w.MessageService.SendMessage(ctx, job)
					if err != nil {
						log.Println("fail creating MEssage in DB")
					}
				case <- ctx.Done():
					log.Printf("context is done")
					return 
				}
			}
		}()
	}
}

func (w *PersistWorker) Enqueue(msg *models.Message){
	w.Jobs <- msg
	log.Println("ENQUEUE SUCCESS")
}