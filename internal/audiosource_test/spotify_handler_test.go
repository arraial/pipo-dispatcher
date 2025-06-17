package audiosource_test

import (
	"testing"

	audiosource "github.com/arraial/pipo-dispatcher/internal/audiosource"
	models "github.com/arraial/pipo-dispatcher/models"
	uuid "github.com/gofrs/uuid/v5"
	"github.com/google/go-cmp/cmp"
)

func TestSpotifyHandler(t *testing.T) {
	t.Parallel()
	test_uuid, _ := uuid.FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	const server = "test"
	const shuffle = true

	tests := []struct {
		name               string
		query              string
		expected_operation string
	}{
		{"URL HTTP", "http://open.spotify.com/track/7gaA3wERFkFkgivjwbSvkG", "url"},
		{"URL HTTPS", "https://open.spotify.com/track/0q6LuUqGLUiCPP1cbdwFs3", "url"},
		{"Playlist source", "https://open.spotify.com/playlist/5XAzQsh9fmEqro13lLgD1I", "url"},
		{"Album source", "https://open.spotify.com/album/4wtZQMNTC1O79kDxMBsEan", "url"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := models.MusicRequest{
				ServerData: models.ServerData{
					UUID:      test_uuid,
					Server_id: server,
				},
				Shuffle: true,
				Query:   []string{tt.query},
			}
			handler := audiosource.NewSpotifyHandler()
			got := handler.Handle(&model, tt.query)
			expected := &models.ProviderOperation{
				ServerData: models.ServerData{
					UUID:      test_uuid,
					Server_id: server,
				},
				Provider:  models.Spotify,
				Operation: tt.expected_operation,
				Shuffle:   shuffle,
				Query:     tt.query,
			}
			if !cmp.Equal(got, expected) {
				t.Errorf("Distinct objects detected. Got: %v. Expected: %v", got, expected)
			}
		})
	}
}
