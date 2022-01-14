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
		keyId:         "6cdda623-edcc-4c27-8a5e-58750f24abde",
		signingSecret: "AeSSabdqLknKnZND_lnGBPbS9VzzamFTOw9mvLFheBQ",
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
