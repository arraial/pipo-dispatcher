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

func defaultHandlers() []*Handler {
	return []*Handler{
		NewSpotifyHandler(),
		NewYoutubeHandler(),
		NewYoutubeQueryHandler(),
	}
}

func NewDefaultSourceManager() *SourceManager {
	return &SourceManager{Handlers: defaultHandlers()}
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
func (s *SourceManager) Handle(request *models.MusicRequest) (operations []*models.ProviderOperation, err error) {
	log := common.GetLogger()
	var wg sync.WaitGroup
	queries := request.Query
	wg.Add(len(queries))
	operations = make([]*models.ProviderOperation, len(queries))
	if request.Shuffle {
		randomize(queries)
		log.Info("Shuffled requested query")
	}
	for i, q := range queries {
		log.Infow("Launched operation model creation", "query", q)
		go func(indx int, query string) {
			defer wg.Done()
			for _, handler := range s.Handlers {
				if handler.iHandler.CanHandle(query) {
					log.Infow("Query will be handled", "query", query, "handler", handler.iHandler.Provider(q))
					operation := handler.Handle(request, query)
					operations[indx] = operation
					return
				}
			}
		}(i, q)
	}
	wg.Wait()
	return
}
