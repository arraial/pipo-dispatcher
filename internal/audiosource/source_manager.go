package audiosource

import (
	"math/rand"
	"sync"

	"github.com/arraial/pipo-dispatcher/internal/common"
	models "github.com/arraial/pipo-dispatcher/models"
)

type SourceManager struct {
	Handlers []*Handler
}

func NewSourceManager(handlers []*Handler) *SourceManager {
	return &SourceManager{Handlers: handlers}
}

func randomize(slice []string) {
	n := len(slice)
	var j int
	for i := range slice {
		j = rand.Intn(n)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// TODO return error
// go routines will be leveraged only if randomization is enabled
func (s *SourceManager) Handle(request *models.MusicRequest, messages chan<- *models.ProviderOperation) (operations []*models.ProviderOperation, err error) {
	log := common.GetLogger()
	var wg sync.WaitGroup
	queries := request.Query
	wg.Add(len(queries))
	operations = make([]*models.ProviderOperation, len(queries))
	if request.Shuffle {
		randomize(queries)
		log.Info("Shuffled requested query")
	}
	for i, query := range queries {
		log.Infow("Launched operation model creation for", "query", query)
		go func(m chan<- *models.ProviderOperation, indx int, q string) {
			defer wg.Done()
			for _, handler := range s.Handlers {
				if handler.iHandler.CanHandle(q) {
					log.Infow("Query will be handled", "query", q, "handler", handler.iHandler.Provider)
					operation := handler.Handle(request, q)
					operations[indx] = operation
					if request.Shuffle {
						m <- operation
					}
					return
				}
			}
		}(messages, i, query)
	}
	wg.Wait()
	if !request.Shuffle {
		for _, op := range operations {
			log.Infow("Sending operation to publisher", "operation", op)
			if op != nil {
				messages <- op
			}
		}
	}
	return
}
