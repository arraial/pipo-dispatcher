package audiosource

import (
	"strings"

	models "github.com/arraial/pipo-dispatcher/models"
)

type SpotifyHandler struct {
	Handler
}

func NewSpotifyHandler() *Handler {
	handler := &SpotifyHandler{}
	return &Handler{handler}
}

func (s *SpotifyHandler) CanHandle(query string) bool {
	return strings.Contains(query, "spotify") && validUrl(query)
}

func (s *SpotifyHandler) Provider(query string) models.Provider {
	return models.Spotify
}

func (s *SpotifyHandler) Operation(query string) string {
	return string(URL)
}
