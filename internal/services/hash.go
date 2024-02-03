package services

import (
	"crypto/sha256"
	"encoding/hex"
)

func GetHash(login, password string) string {
	// TODO получение хеша нужно усложнить. добавить соль, может еще как-то

	hash := sha256.Sum256([]byte(login + password))
	return hex.EncodeToString(hash[:])

}
