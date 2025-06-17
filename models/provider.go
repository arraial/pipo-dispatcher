package models

type Provider string

const (
	Youtube Provider = "youtube"
	Spotify Provider = "spotify"
)

type ProviderOperation struct {
	ServerData
	Provider  Provider `json:"provider"`
	Operation string   `json:"operation"`
	Shuffle   bool     `json:"shuffle"`
	Query     string   `json:"query"`
}
