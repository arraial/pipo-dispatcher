package audiosource

import (
	"net/url"

	models "github.com/arraial/pipo-dispatcher/models"
)

type Operations string

const (
	URL      Operations = "url"
	PLAYLIST Operations = "playlist"
	QUERY    Operations = "query"
)

func validUrl(input string) bool {
	u, err := url.ParseRequestURI(input)
	return err == nil && u.Scheme != "" && u.Host != ""
}

type IHandler interface {
	CanHandle(string) bool
	Provider(string) models.Provider
	Operation(string) string
}

type Handler struct {
	iHandler IHandler
}

func (s *Handler) Handle(request models.MusicRequest, query string) models.ProviderOperation {
	return models.ProviderOperation{
		ServerData: request.ServerData,
		Provider:   s.iHandler.Provider(query),
		Operation:  s.iHandler.Operation(query),
		Shuffle:    request.Shuffle,
		Query:      query,
	}
}
