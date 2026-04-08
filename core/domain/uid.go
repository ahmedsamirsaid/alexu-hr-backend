package domain

import (
	"strings"

	"github.com/google/uuid"
)

func GenerateUID(prefix string) string {
	id := uuid.New().String()
	id = strings.ReplaceAll(id, "-", "")
	return prefix + "_" + id
}
