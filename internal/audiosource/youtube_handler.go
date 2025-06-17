package audiosource

import (
	"strings"

	models "github.com/arraial/pipo-dispatcher/models"
)

type YoutubeHandler struct {
	Handler
}

func NewYoutubeHandler() *Handler {
	handler := &YoutubeHandler{}
	return &Handler{handler}
}

func (s *YoutubeHandler) CanHandle(query string) bool {
	return validUrl(query)
}

func (s *YoutubeHandler) Provider(query string) models.Provider {
	return models.Youtube
}

func (s *YoutubeHandler) Operation(query string) string {
	if strings.Contains(query, "list=") {
		return string(PLAYLIST)
	}
	return string(URL)
}

type YoutubeQueryHandler struct {
	Handler
}

func NewYoutubeQueryHandler() *Handler {
	handler := &YoutubeQueryHandler{}
	return &Handler{handler}
}

func (s *YoutubeQueryHandler) CanHandle(query string) bool {
	return !validUrl(query)
}

func (s *YoutubeQueryHandler) Provider(query string) models.Provider {
	return models.Youtube
}

func (s *YoutubeQueryHandler) Operation(query string) string {
	return string(QUERY)
}
