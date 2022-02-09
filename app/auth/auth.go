package auth

import (
	"encoding/base64"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// All of the DoorDash Access Key fields needed to generate a JWT.

type DoorDashAccessKey struct {
	developerId   string
	keyId         string
	signingSecret string
}

type DoorDashClaims struct {
	*jwt.StandardClaims
	KeyId string `json:"kid"`
}

func GetJWT() (string, error) {
	accessKey := &DoorDashAccessKey{
		developerId:   os.Getenv("DOORDASH_DEVELOPER_ID"),
		keyId:         os.Getenv("DOORDASH_KEY_ID"),
		signingSecret: os.Getenv("DOORDASH_SIGNING_SECRET"),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, &DoorDashClaims{
		&jwt.StandardClaims{

			Issuer:    accessKey.developerId,
			Audience:  "doordash",
			ExpiresAt: time.Now().Add(time.Minute * 30).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		accessKey.keyId,
	})
	t.Header["dd-ver"] = "DD-JWT-V1"

	decodedSigningSecret, err := base64.RawURLEncoding.DecodeString(accessKey.signingSecret)
	if err != nil {
		return "", err
	}

	jwt, err := t.SignedString(decodedSigningSecret)
	if err != nil {
		return "", err
	}
	return jwt, nil
}
