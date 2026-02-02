package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// PKCE holds the PKCE parameters
type PKCE struct {
	Verifier        string
	Challenge       string
	ChallengeMethod string
}

// GeneratePKCE generates a new PKCE code verifier and challenge
func GeneratePKCE() (*PKCE, error) {
	// Generate 32 bytes of random data (sufficient entropy)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	// Base64URL encode without padding
	verifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Generate code challenge using S256 method
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return &PKCE{
		Verifier:        verifier,
		Challenge:       challenge,
		ChallengeMethod: "S256",
	}, nil
}

// GenerateState generates a cryptographically secure state parameter
func GenerateState() (string, error) {
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(stateBytes), nil
}

// ClearVerifier clears the verifier from memory (best effort)
func (p *PKCE) ClearVerifier() {
	// Note: In Go, we can't truly zero memory, but we can overwrite
	// This is more of a symbolic gesture in Go's garbage collected environment
	for i := range p.Verifier {
		p.Verifier = p.Verifier[:i] + "0" + p.Verifier[i+1:]
	}
	p.Verifier = ""
}
