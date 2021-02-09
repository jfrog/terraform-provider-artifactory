package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"strings"
	"time"
)

func extractPayloadFromAccessToken(token string) (TokenPayload, error) {
	// Separate token parts.
	tokenParts := strings.Split(token, ".")

	// Decode the payload.
	if len(tokenParts) != 3 {
		return TokenPayload{}, errorutils.CheckError(errors.New("received invalid access-token"))
	}
	payload, err := base64.RawStdEncoding.DecodeString(tokenParts[1])
	if err != nil {
		return TokenPayload{}, errorutils.CheckError(err)
	}

	// Unmarshal json.
	var tokenPayload TokenPayload
	err = json.Unmarshal(payload, &tokenPayload)
	if err != nil {
		return TokenPayload{}, errorutils.CheckError(errors.New("Failed extracting payload from the provided access-token." + err.Error()))
	}
	return tokenPayload, nil
}

func ExtractUsernameFromAccessToken(token string) (string, error) {
	tokenPayload, err := extractPayloadFromAccessToken(token)
	if err != nil {
		return "", err
	}
	// Extract subject.
	if tokenPayload.Subject == "" {
		return "", errorutils.CheckError(errors.New("could not extract subject from the provided access-token"))
	}

	// Extract username from subject.
	usernameStartIndex := strings.LastIndex(tokenPayload.Subject, "/")
	if usernameStartIndex < 0 {
		return "", errorutils.CheckError(errors.New(fmt.Sprintf("Could not extract username from access-token's subject: %s", tokenPayload.Subject)))
	}
	username := tokenPayload.Subject[usernameStartIndex+1:]

	return username, nil
}

// Extracts the expiry from an access token, in seconds
func ExtractExpiryFromAccessToken(token string) (int, error) {
	tokenPayload, err := extractPayloadFromAccessToken(token)
	if err != nil {
		return -1, err
	}
	expiry := tokenPayload.ExpirationTime - tokenPayload.IssuedAt
	return expiry, nil
}

// Returns 0 if expired
func GetTokenMinutesLeft(token string) (int64, error) {
	payload, err := extractPayloadFromAccessToken(token)
	if err != nil {
		return -1, err
	}
	left := int64(payload.ExpirationTime) - time.Now().Unix()
	if left < 0 {
		return 0, nil
	}
	return left / 60, nil
}

type TokenPayload struct {
	Subject        string `json:"sub,omitempty"`
	Scope          string `json:"scp,omitempty"`
	Audience       string `json:"aud,omitempty"`
	Issuer         string `json:"iss,omitempty"`
	ExpirationTime int    `json:"exp,omitempty"`
	IssuedAt       int    `json:"iat,omitempty"`
	JwtId          string `json:"jti,omitempty"`
}

// Refreshable Tokens Constants.
var RefreshBeforeExpiryMinutes = int64(10)

const WaitBeforeRefreshSeconds = 15
