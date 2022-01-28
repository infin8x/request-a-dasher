package main

import (
	"bytes"
	"encoding/json"
	"errors"
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

type DeliveryRequest struct {
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

type DeliveryResponse struct {
	ExternalDeliveryId   string    `json:"external_delivery_id"`
	Currency             string    `json:"currency"`
	DeliveryStatus       string    `json:"delivery_status"`
	Fee                  int       `json:"fee"`
	PickupAddress        string    `json:"pickup_address"`
	PickupBusinessName   string    `json:"pickup_business_name"`
	PickupPhoneNumber    string    `json:"pickup_phone_number"`
	PickupInstructions   string    `json:"pickup_instructions"`
	PickupReferenceTag   string    `json:"pickup_reference_tag"`
	DropoffAddress       string    `json:"dropoff_address"`
	DropoffBusinessName  string    `json:"dropoff_business_name"`
	DropoffPhoneNumber   string    `json:"dropoff_phone_number"`
	DropoffInstructions  string    `json:"dropoff_instructions"`
	OrderValue           int       `json:"order_value"`
	PickupTimeEstimated  time.Time `json:"pickup_time_estimated"`
	DropoffTimeEstimated time.Time `json:"dropoff_time_estimated"`
	TrackingUrl          string    `json:"tracking_url"`
	ContactlessDropoff   bool      `json:"contactless_dropoff"`
	Tip                  int       `json:"tip"`
}

func main() {
	fmt.Printf("Deliverate web server starting on port 8080\n")

	// Initialize the Gorilla mux
	r := mux.NewRouter()

	// Register website handlers
	r.HandleFunc("/", indexGETHandler).Methods("GET")
	r.HandleFunc("/", indexPOSTHandler).Methods("POST")
	r.HandleFunc("/status", statusGETHandler).Methods("GET")
	r.HandleFunc("/status/{id}", statusGETHandler).Methods("GET")

	// Register API handlers
	r.HandleFunc("/doordash/deliveries/{id}", deliveriesGETHandler).Methods("GET")
	r.HandleFunc("/doordash/deliveries", deliveriesPOSTHandler).Methods("POST")

	// Lastly, register static file handlers
	r.PathPrefix("/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./static/"))))

	// Start server
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}

}

func indexGETHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Print("Index page\n==========\n")
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

	fmt.Print("\n")
}

func indexPOSTHandler(w http.ResponseWriter, r *http.Request) {
	// Get the form details
	if err := r.ParseForm(); err != nil {
		fmt.Printf("Unable to parse form: %v\n", err.Error())
		http.Error(w, "oh snap", http.StatusInternalServerError)
		return
	}

	// Create the request body
	body := DeliveryRequest{
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

	// Prepare the request as JSON
	bodyJson, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("Unable to create the request body: %v\n", err.Error())
		http.Error(w, "oh snap", http.StatusInternalServerError)
		return
	}

	response, err := createDelivery(bodyJson)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Couldn't create your delivery because %v! We've logged an error and will take a look.", err.Error()),
			http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, response)
}

func statusGETHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Print("Status page\n===========\n")
	vars := mux.Vars(r)
	if vars["id"] != "" {
		fmt.Printf("Requesting details of delivery with ID: %v\n", vars["id"])

	} else {
		fmt.Print("Requesting status page with no delivery ID\n")
	}

	tmplPath := path.Join("templates", "status.html")
	tmpl, err := template.ParseFiles(tmplPath)

	if err != nil {
		fmt.Printf("Unable to parse template: %v\n", err.Error())
		http.Error(w, "oh snap", http.StatusInternalServerError)
		return
	}

	// Get the delivery
	response, err := getDelivery(vars["id"])
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Couldn't get your delivery because %v! We've logged an error and will take a look.", err.Error()),
			http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, response); err != nil {
		fmt.Printf("Unable to execute template: %v\n", err.Error())
		http.Error(w, "oh snap", http.StatusInternalServerError)
	}

	fmt.Print("\n")
}

func deliveriesGETHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Printf("Requesting details of delivery with ID: %v\n", vars["id"])

	response, err := getDelivery(vars["id"])
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Couldn't create your delivery because %v! We've logged an error and will take a look.", err.Error()),
			http.StatusInternalServerError)
		return
	}

	toreturn, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Couldn't create your delivery because %v! We've logged an error and will take a look.", err.Error()),
			http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(toreturn))
}

func deliveriesPOSTHandler(w http.ResponseWriter, r *http.Request) {
	// Get request body
	defer r.Body.Close()
	bodyJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Unable to parse the request body: %v\n", err.Error())
		http.Error(w, "Couldn't understand your request", http.StatusBadRequest)
		return
	}

	// Parse request JSON
	var bodyMap map[string]interface{}
	json.Unmarshal([]byte(bodyJson), &bodyMap)
	fmt.Printf("Creating a new delivery with external_delivery_id: %v\n", bodyMap["external_delivery_id"])

	response, err := createDelivery(bodyJson)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Couldn't create your delivery because %v! We've logged an error and will take a look.", err.Error()),
			http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, response)
}

func getDelivery(externalDeliveryId string) (DeliveryResponse, error) {
	// Get a token
	token, err := auth.GetJWT()
	if err != nil {
		fmt.Printf("Unable to get a JWT: %v\n", err.Error())
		return DeliveryResponse{}, errors.New("we couldn't authenticate with DoorDash")
	}
	fmt.Printf("Bearer token\n============\n%v\n", token)

	// Create a client and prepare the request
	client := &http.Client{}
	req, err := http.NewRequest("GET", DoorDashV2APIPrefix+"deliveries/"+externalDeliveryId, nil)
	if err != nil {
		fmt.Printf("Unable to create an http client: %v\n", err.Error())
		return DeliveryResponse{}, errors.New("we couldn't connect to DoorDash")
	}

	// Add the authorization header and do the request
	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		// TODO better/more specific error code handling
		fmt.Printf("Unable to request details of the delivery: %v\n", err.Error())
		return DeliveryResponse{}, errors.New("we couldn't connect to DoorDash")
	}

	// Parse the response
	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Unable to parse details of the delivery: %v\n", err.Error())
		return DeliveryResponse{}, errors.New("something went wrong")
	}

	// Convert the response to our struct
	delivery := DeliveryResponse{}
	if err := json.Unmarshal(responseData, &delivery); err != nil {
		fmt.Printf("Unable to parse details of the delivery: %v\n", err.Error())
		return DeliveryResponse{}, errors.New("something went wrong")
	}

	return delivery, nil
}

func createDelivery(apiRequestBody []byte) (string, error) {
	// Get a token
	token, err := auth.GetJWT()
	if err != nil {
		fmt.Printf("Unable to get a JWT: %v\n", err.Error())
		return "", errors.New("we couldn't authenticate with DoorDash")
	}
	fmt.Printf("Bearer token\n============\n%v\n", token)

	// Create a client and prepare the request
	client := &http.Client{}
	req, err := http.NewRequest("POST", DoorDashV2APIPrefix+"deliveries", bytes.NewBuffer(apiRequestBody))
	if err != nil {
		fmt.Printf("Unable to create an http client: %v\n", err.Error())
		return "", errors.New("we couldn't connect to DoorDash")
	}

	// Add the authorization header and do the request
	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		// TODO better/more specific error code handling
		fmt.Printf("Unable to request creation of the delivery: %v\n", err.Error())
		return "", errors.New("we couldn't connect to DoorDash")
	}

	// Parse the response
	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Unable to parse the response details of the newly-created delivery: %v\n", err.Error())
		return "", errors.New("we couldn't connect to DoorDash")
	}

	return string(responseData), nil
}
