package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

type DoorDashAccessKey struct {
	developerId   string
	keyId         string
	signingSecret string
}

type DoorDashClaims struct {
	*jwt.StandardClaims
	keyId string
}

func main() {
	// TODO delete this nonsense, eventually
	http.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "I am Deliverate! Like Deliberate but with Delivery :troll_face:")
	})

	fmt.Printf("Deliverate web server starting on port 8080\n")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

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

	ss, err := t.SignedString([]byte(accessKey.signingSecret))

	if err != nil {
		fmt.Printf("%v", err)
	} else {
		fmt.Printf("%v", ss)
	}

}
