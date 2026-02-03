package auth

import (
	"fmt"
)

type Provider interface {
	Name() string
}

var (
	ErrTimeout          = fmt.Errorf("authentication timeout")
	ErrPortInUse        = fmt.Errorf("port already in use")
	ErrAPIKeyGeneration = fmt.Errorf("API key generation failed")
)
