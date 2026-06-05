package service

import "github.com/google/uuid"

func parseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}
