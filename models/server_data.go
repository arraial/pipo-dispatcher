package models

import uuid "github.com/gofrs/uuid/v5"

type ServerData struct {
	UUID      uuid.UUID `json:"uuid"`
	Server_id string    `json:"server_id"`
}
