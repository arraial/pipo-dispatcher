package audiosource

import (
	"math/rand"

	models "github.com/arraial/pipo-dispatcher/models"
)

type SourceManager struct {
	handlers []Handler
}

func randomize(slice []string) {
	n := len(slice)
	for i := range slice {
		j := rand.Intn(n)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func (s *SourceManager) Handle(request models.MusicRequest) []models.ProviderOperation {
	result := make([]models.ProviderOperation, 0)
	var queries = request.Query
	if request.Shuffle {
		randomize(queries)
	}
	for _, query := range queries {
		for _, handler := range s.handlers {
			if handler.iHandler.CanHandle(query) {
				result = append(
					result,
					handler.Handle(request, query),
				)
			}
		}
	}
	return result
}
