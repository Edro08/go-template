package utils

import "uuid"

// NewUUID generates a new UUID v7.
func NewUUID() string {
	return uuid.NewV7().String()
}
