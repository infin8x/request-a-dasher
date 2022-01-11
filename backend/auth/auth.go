package auth

import (
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

	// decodedString, err := base64.URLEncoding.DecodeString(accessKey.signingSecret)
	// if err != nil {
	// 	return "", err
	// }

	ss, err := t.SignedString(accessKey.signingSecret)
	if err != nil {
		return "", err
	}
	return ss, nil
}
