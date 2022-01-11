package auth

import (
	"encoding/base64"
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
	// TODO add secrets management and roll this secret before making the repository public
	accessKey := &DoorDashAccessKey{
		developerId:   "14e84291-d900-4c20-8528-ed6ca8de660f",
		keyId:         "9523c67d-e0c5-41c9-9702-9a67111338c4",
		signingSecret: "xbEg06vXu2zSQQEKRCufcRPKkv7wJOTvihSgaj9G_cc",
	}

	t := jwt.New(jwt.SigningMethodHS256)

	t.Header["dd-ver"] = "DD-JWT-V1"

	t.Claims = &DoorDashClaims{
		&jwt.StandardClaims{

			Issuer:    accessKey.developerId,
			Audience:  "doordash",
			ExpiresAt: time.Now().Add(time.Minute * 30).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		accessKey.keyId,
	}

	decodedString, err := base64.URLEncoding.DecodeString(accessKey.signingSecret)

	ss, err := t.SignedString(decodedString)

	if err != nil {
		return "", err
	} else {
		return ss, nil
	}
}
