package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)


type PKCE struct {
	Verifier        string
	Challenge       string
	ChallengeMethod string
}


func GeneratePKCE() (*PKCE, error) {
	
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	
	verifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return &PKCE{
		Verifier:        verifier,
		Challenge:       challenge,
		ChallengeMethod: "S256",
	}, nil
}


func GenerateState() (string, error) {
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(stateBytes), nil
}


func (p *PKCE) ClearVerifier() {
	
	
	for i := range p.Verifier {
		p.Verifier = p.Verifier[:i] + "0" + p.Verifier[i+1:]
	}
	p.Verifier = ""
}
