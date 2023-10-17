package utils

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"

	_ "time/tzdata"
)

type TokenMetadata struct {
	UserID      uuid.UUID
	Credentials map[string]bool
	Expires     int64
}

func CreateToken(ttl time.Duration, id uuid.UUID, privateKey string, credentials []string) (string, error) {
	decodedPrivateKey, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", fmt.Errorf("could not decode key: %w", err)
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(decodedPrivateKey)

	if err != nil {
		return "", fmt.Errorf("create: parse key: %w", err)
	}

	loc, _ := time.LoadLocation("Asia/Singapore")
	now := time.Now().In(loc)

	claims := make(jwt.MapClaims)
	claims["id"] = id
	claims["exp"] = now.Add(ttl).Unix()
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()

	claims["user:add"] = false
	claims["user:delete"] = false
	claims["user:get"] = false

	claims["andalalin:get"] = false
	claims["andalalin:persyaratan"] = false
	claims["andalalin:status"] = false
	claims["andalalin:update"] = false
	claims["andalalin:tambahpetugas"] = false
	claims["andalalin:petugas"] = false
	claims["andalalin:survey"] = false
	claims["andalalin:ticket1"] = false
	claims["andalalin:ticket2"] = false
	claims["andalalin:persetujuan"] = false
	claims["andalalin:bap"] = false
	claims["andalalin:sk"] = false
	claims["andalalin:pengajuan"] = false
	claims["andalalin:kelola"] = false
	claims["andalalin:keputusan"] = false
	claims["andalalin:kepuasan"] = false

	claims["product:add"] = false
	claims["product:delete"] = false
	claims["product:update"] = false

	for _, credential := range credentials {
		claims[credential] = true
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)

	if err != nil {
		return "", fmt.Errorf("create: sign token: %w", err)
	}

	return token, nil
}

func ValidateToken(token string, publicKey string) (*TokenMetadata, error) {
	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, fmt.Errorf("could not decode: %w", err)
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(decodedPublicKey)

	if err != nil {
		return nil, fmt.Errorf("validate: parse key: %w", err)
	}

	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected method: %s", t.Header["alg"])
		}
		return key, nil
	})

	if err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if ok && parsedToken.Valid {
		userID, err := uuid.Parse(claims["id"].(string))
		if err != nil {
			return nil, err
		}

		// Expires time.
		expires := int64(claims["exp"].(float64))

		// User credentials.
		credentials := map[string]bool{
			"user:add":    claims["user:add"].(bool),
			"user:delete": claims["user:delete"].(bool),
			"user:get":    claims["user:get"].(bool),

			"andalalin:get":           claims["andalalin:get"].(bool),
			"andalalin:persyaratan":   claims["andalalin:persyaratan"].(bool),
			"andalalin:status":        claims["andalalin:status"].(bool),
			"andalalin:update":        claims["andalalin:update"].(bool),
			"andalalin:tambahpetugas": claims["andalalin:tambahpetugas"].(bool),
			"andalalin:petugas":       claims["andalalin:petugas"].(bool),
			"andalalin:survey":        claims["andalalin:survey"].(bool),
			"andalalin:ticket1":       claims["andalalin:ticket1"].(bool),
			"andalalin:ticket2":       claims["andalalin:ticket2"].(bool),
			"andalalin:persetujuan":   claims["andalalin:persetujuan"].(bool),
			"andalalin:bap":           claims["andalalin:bap"].(bool),
			"andalalin:sk":            claims["andalalin:sk"].(bool),
			"andalalin:pengajuan":     claims["andalalin:pengajuan"].(bool),
			"andalalin:kelola":        claims["andalalin:kelola"].(bool),
			"andalalin:keputusan":     claims["andalalin:keputusan"].(bool),
			"andalalin:kepuasan":      claims["andalalin:kepuasan"].(bool),

			"product:add":    claims["product:add"].(bool),
			"product:delete": claims["product:delete"].(bool),
			"product:update": claims["product:update"].(bool),
		}

		return &TokenMetadata{
			UserID:      userID,
			Credentials: credentials,
			Expires:     expires,
		}, nil
	}
	return nil, err
}

func GetIdByToken(token string, publicKey string) *TokenMetadata {
	decodedPublicKey, _ := base64.StdEncoding.DecodeString(publicKey)

	key, _ := jwt.ParseRSAPublicKeyFromPEM(decodedPublicKey)

	parsedToken, _ := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected method: %s", t.Header["alg"])
		}
		return key, nil
	})

	claims, _ := parsedToken.Claims.(jwt.MapClaims)
	userID, _ := uuid.Parse(claims["id"].(string))

	return &TokenMetadata{
		UserID: userID,
	}
}
