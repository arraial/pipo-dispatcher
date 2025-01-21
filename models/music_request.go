package models

type MusicRequest struct {
	UUID      myUUID   `json:"uuid"`
	Server_id string   `json:"server_id"`
	Shuffle   bool     `json:"shuffle"`
	Query     []string `json:"query"`
}
