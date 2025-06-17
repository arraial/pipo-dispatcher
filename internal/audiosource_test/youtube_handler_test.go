package audiosource_test

import (
	"testing"

	audiosource "github.com/arraial/pipo-dispatcher/internal/audiosource"
	models "github.com/arraial/pipo-dispatcher/models"
	uuid "github.com/gofrs/uuid/v5"
	"github.com/google/go-cmp/cmp"
)

func TestYoutubeHandler(t *testing.T) {
	t.Parallel()
	test_uuid, _ := uuid.FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	const server = "test"
	const shuffle = true

	tests := []struct {
		name               string
		query              string
		expected_operation string
	}{
		{"URL HTTP", "http://www.youtube.com/watch?v=1V_xRb0x9aw", "url"},
		{"URL HTTPS", "https://www.youtube.com/watch?v=1V_xRb0x9aw", "url"},
		{"Playlist source", "https://www.youtube.com/playlist?list=PL4lCao7KL_QFVb7Iudeipvc2BCavECqzc", "playlist"},
		{"Playlist indexed", "https://www.youtube.com/watch?v=BaW_jenozKc&list=PL4lCao7KL_QFVb7Iudeipvc2BCavECqzc&index=1", "playlist"},
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
			handler := audiosource.NewYoutubeHandler()
			got := handler.Handle(&model, tt.query)
			expected := &models.ProviderOperation{
				ServerData: models.ServerData{
					UUID:      test_uuid,
					Server_id: server,
				},
				Provider:  models.Youtube,
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
