package models

type MusicRequest struct {
	ServerData
	Shuffle bool     `json:"shuffle"`
	Query   []string `json:"query"`
}
