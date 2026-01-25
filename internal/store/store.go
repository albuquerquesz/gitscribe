package store

import (
	"github.com/zalando/go-keyring"
)

var (
	service = "gitscribe"
	user    = "anon"
)

func Save(apiKey string) error {
	return keyring.Set(service, user, apiKey)
}

func Get() (string, error) {
	return keyring.Get(service, user)
}

