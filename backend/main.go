package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	auth "github.com/infin8x/deliverate/backend/auth"

	"github.com/gorilla/mux"
)

var DoorDashV2APIPrefix string = "https://openapi.doordash.com/drive/v2/"

func main() {
	fmt.Printf("Deliverate web server starting on port 8080\n")

	// Initialize the Gorilla mux
	r := mux.NewRouter()

	// Register all handlers
	// TODO delete this nonsense, eventually
	r.HandleFunc("/whoami", whoamiHandler)
	r.HandleFunc("/doordash/deliveries/{id}", DeliveryHandler)

	// Start server
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}

}

func whoamiHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/whoami" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method is not supported", http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprint(w, "I am Deliverate! Like Deliberate but with Delivery :troll_face:")
}

func DeliveryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Handling delivery with ID: %v\n", vars["id"])
	// Get a token
	token, err := auth.GetJWT()

	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(token)

	client := &http.Client{}
	req, err := http.NewRequest("GET", DoorDashV2APIPrefix+"deliveries/"+vars["id"], nil)

	if err != nil {
		fmt.Print(err.Error())
	}

	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)

	if err != nil {
		fmt.Print(err.Error())
	}

	responseData, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Print(err.Error())
	}

	fmt.Print(string(responseData))
}
