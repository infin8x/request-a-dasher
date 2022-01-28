package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"time"

	auth "github.com/infin8x/deliverate/backend/auth"

	"github.com/gorilla/mux"
)

var DoorDashV2APIPrefix string = "https://openapi.doordash.com/drive/v2/"

type Delivery struct {
	ExternalDeliveryId  string `json:"external_delivery_id"`
	PickupAddress       string `json:"pickup_address"`
	PickupBusinessName  string `json:"pickup_business_name"`
	PickupPhoneNumber   string `json:"pickup_phone_number"`
	PickupInstructions  string `json:"pickup_instructions"`
	PickupReferenceTag  string `json:"pickup_reference_tag"`
	DropoffAddress      string `json:"dropoff_address"`
	DropoffBusinessName string `json:"dropoff_business_name"`
	DropoffPhoneNumber  string `json:"dropoff_phone_number"`
	DropoffInstructions string `json:"dropoff_instructions"`
	OrderValue          int    `json:"order_value"`
	Currency            string `json:"currency"`
	Tip                 int    `json:"tip"`
}

func main() {
	fmt.Printf("Deliverate web server starting on port 8080\n")

	// Initialize the Gorilla mux
	r := mux.NewRouter()

	// Register API handlers
	// TODO delete this nonsense, eventually
	r.HandleFunc("/whoami", whoamiHandler).Methods("GET")
	r.HandleFunc("/doordash/deliveries/{id}", GETDeliveryHandler).Methods("GET")
	r.HandleFunc("/doordash/deliveries", POSTDeliveryHandler).Methods("POST")

	// Register website handlers
	r.HandleFunc("/", indexHandler).Methods("GET")
	r.HandleFunc("/", indexPOSTHandler).Methods("POST")

	// Register static file handler
	r.PathPrefix("/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./static/"))))

	// Start server
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}

}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmplPath := path.Join("templates", "index.html")
	tmpl, err := template.ParseFiles(tmplPath)

	if err != nil {
		fmt.Printf("Unable to parse template: %v\n", err.Error())
		http.Error(w, "oh snap", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		fmt.Printf("Unable to execute template: %v\n", err.Error())
		http.Error(w, "oh snap", http.StatusInternalServerError)
	}
}

func indexPOSTHandler(w http.ResponseWriter, r *http.Request) {
	// Get the form details
	if err := r.ParseForm(); err != nil {
		fmt.Printf("Unable to parse form: %v\n", err.Error())
		http.Error(w, "oh snap", http.StatusInternalServerError)
		return
	}

	// Get a token
	token, err := auth.GetJWT()
	if err != nil {
		fmt.Printf("Unable to get a JWT: %v\n", err.Error())
		http.Error(w, "Couldn't authenticate with DoorDash", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Bearer token\n============\n%v\n", token)

	// Prepare the request
	body := Delivery{
		ExternalDeliveryId:  fmt.Sprint(time.Now().Unix()),
		PickupAddress:       r.FormValue("whereFrom"),
		PickupBusinessName:  "",
		PickupPhoneNumber:   r.FormValue("pickupPhone"),
		PickupInstructions:  r.FormValue("pickupInstructions"),
		PickupReferenceTag:  "Request a Dasher",
		DropoffAddress:      r.FormValue("whereTo"),
		DropoffBusinessName: "",
		DropoffPhoneNumber:  r.FormValue("dropoffPhone"),
		DropoffInstructions: r.FormValue("dropoffInstructions"),
		OrderValue:          0,
		Currency:            "usd",
		Tip:                 0,
	}
	bodyJson, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("Unable to create the request body: %v\n", err.Error())
		http.Error(w, "oh snap", http.StatusInternalServerError)
		return
	}

	// Create a client
	client := &http.Client{}
	req, err := http.NewRequest("POST", DoorDashV2APIPrefix+"deliveries", bytes.NewBuffer(bodyJson))
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
