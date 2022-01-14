package main

import (
	"bytes"
	"encoding/json"
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
	r.HandleFunc("/doordash/deliveries/{id}", GETDeliveryHandler).Methods("GET")
	r.HandleFunc("/doordash/deliveries", POSTDeliveryHandler).Methods("POST")

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

func GETDeliveryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Printf("Requesting details of delivery with ID: %v\n", vars["id"])

	// Get a token
	token, err := auth.GetJWT()
	if err != nil {
		fmt.Printf("Unable to get a JWT: %v\n", err.Error())
		http.Error(w, "Couldn't authenticate with DoorDash", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Bearer token\n============\n%v\n", token)

	// Create a client and prepare the request
	client := &http.Client{}
	req, err := http.NewRequest("GET", DoorDashV2APIPrefix+"deliveries/"+vars["id"], nil)
	if err != nil {
		fmt.Printf("Unable to create an http client: %v\n", err.Error())
		http.Error(w, "Could connect to DoorDash", http.StatusInternalServerError)
		return
	}

	// Add the authorization header and do the request
	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		// TODO better/more specific error code handling
		fmt.Printf("Unable to request details of the delivery: %v\n", err.Error())
		// TODO print a machine-readable error to http.Error
		http.Error(w, "oh snap", http.StatusInternalServerError)
		return
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Unable to parse details of the delivery: %v\n", err.Error())
		// TODO print a machine-readable error to http.Error
		http.Error(w, "oh snap", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(responseData))
}

func POSTDeliveryHandler(w http.ResponseWriter, r *http.Request) {
	// Get request body
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Unable to parse the request body: %v\n", err.Error())
		http.Error(w, "Couldn't understand your request", http.StatusBadRequest)
		return
	}

	// Parse request JSON
	var bodyJson map[string]interface{}
	json.Unmarshal([]byte(bodyBytes), &bodyJson)
	fmt.Printf("Creating a new delivery with external_delivery_id: %v\n", bodyJson["external_delivery_id"])

	// Get a token
	token, err := auth.GetJWT()
	if err != nil {
		fmt.Printf("Unable to get a JWT: %v\n", err.Error())
		http.Error(w, "Couldn't authenticate with DoorDash", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Bearer token\n============\n%v\n", token)

	// Create a client and prepare the request
	client := &http.Client{}
	req, err := http.NewRequest("POST", DoorDashV2APIPrefix+"deliveries", bytes.NewBuffer(bodyBytes))
	if err != nil {
		fmt.Printf("Unable to create an http client: %v\n", err.Error())
		http.Error(w, "Could connect to DoorDash", http.StatusInternalServerError)
		return
	}

	// Add the authorization header and do the request
	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		// TODO better/more specific error code handling
		fmt.Printf("Unable to request creation of the delivery: %v\n", err.Error())
		// TODO print a machine-readable error to http.Error
		http.Error(w, "oh snap", http.StatusInternalServerError)
		return
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Unable to parse ther response details of the newly-created delivery: %v\n", err.Error())
		// TODO print a machine-readable error to http.Error
		http.Error(w, "oh snap", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(responseData))
}
