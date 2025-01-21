package models

type Provider int

const (
	Youtube Provider = iota
	Spotify
)

func (p Provider) String() string {
	return []string{"youtube", "spotify"}[p]
}

type ProviderOperation struct {
	UUID      myUUID   `json:"uuid"`
	Provider  Provider `json:"provider"`
	Server_id string   `json:"server_id"`
	Operation string   `json:"operation"`
	Shuffle   bool     `json:"shuffle"`
	Query     string   `json:"query"`
}
