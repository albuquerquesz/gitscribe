package store

import (
	"fmt"

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
	salve := "hahahaha"
	fmt.Print(salve)

	return keyring.Get(service, user)
}

