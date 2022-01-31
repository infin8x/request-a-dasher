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
	"strconv"
	"strings"
	"time"

	auth "github.com/infin8x/request-a-dasher/app/auth"

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

	DeliveryStatus       string    `json:"delivery_status"`
	TrackingUrl          string    `json:"tracking_url"`
	Fee                  int       `json:"fee"`
	PickupTimeEstimated  time.Time `json:"pickup_time_estimated"`
	PickupTimeActual     time.Time `json:"pickup_time_actual"`
	DropoffTimeEstimated time.Time `json:"dropoff_time_estimated"`
	DropoffTimeActual    time.Time `json:"dropoff_time_actual"`
	ContactlessDropoff   bool      `json:"contactless_dropoff"`

	DeprefixedDeliveryId             string
	DeliveryStatusFriendly           string
	DeliveryStatusPercentage         int
	DeliveryStatusProgressBarClasses string
	PickupTime                       time.Time
	DropoffTime                      time.Time
}

func main() {
	fmt.Printf("RAD web server starting on port 8080\n")

	// Initialize the Gorilla mux
	r := mux.NewRouter()

	// Register website handlers
	r.HandleFunc("/", indexGETHandler).Methods("GET")
	r.HandleFunc("/", indexPOSTHandler).Methods("POST")
	r.HandleFunc("/deliveries", deliveriesGETHandler).Methods("GET")
	r.HandleFunc("/deliveries", deliveriesPOSTHandler).Methods("POST")
	r.HandleFunc("/deliveries/{id}", deliveriesGETHandler).Methods("GET")

	// Register API handlers - can probably delete these
	r.HandleFunc("/doordash/deliveries/{id}", dddeliveriesGETHandler).Methods("GET")
	r.HandleFunc("/doordash/deliveries", dddeliveriesPOSTHandler).Methods("POST")

	// Lastly, register static file handlers
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

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
	fmt.Print("Index page form submission\n==========================\n")

	// Get the form details
	if err := r.ParseForm(); err != nil {
		fmt.Printf("Unable to parse form: %v\n", err.Error())
		http.Error(w, "oh snap", http.StatusInternalServerError)
		return
	}

	// Create the request body
	orderValue, err := strconv.ParseFloat(r.FormValue("orderValue"), 64)
	if err != nil {
		fmt.Printf("Unable to parse order value: %v\n", err.Error())
		http.Error(w, "oh snap", http.StatusInternalServerError)
		return
	}

	tip, err := strconv.ParseFloat(r.FormValue("tip"), 64)
	if err != nil {
		fmt.Printf("Unable to parse tip: %v\n", err.Error())
		http.Error(w, "oh snap", http.StatusInternalServerError)
		return
	}

	body := DeliveryRequest{
		ExternalDeliveryId:  prefixDeliveryId(fmt.Sprint(time.Now().Unix())),
		PickupAddress:       r.FormValue("whereFrom"),
		PickupBusinessName:  "DoorDash",
		PickupPhoneNumber:   "+1" + strings.Map(mapFilterPhoneNumber, r.FormValue("pickupPhone")),
		PickupInstructions:  r.FormValue("pickupInstructions"),
		PickupReferenceTag:  "Request a Dasher",
		DropoffAddress:      r.FormValue("whereTo"),
		DropoffBusinessName: r.FormValue("dropoffBusinessName"),
		DropoffPhoneNumber:  "+1" + strings.Map(mapFilterPhoneNumber, r.FormValue("pickupPhone")),
		DropoffInstructions: r.FormValue("dropoffInstructions"),
		OrderValue:          int(orderValue * 100), // DoorDash API expects all money in cents
		Currency:            "usd",
		Tip:                 int(tip * 100), // DoorDash API expects all money in cents
	}

	// TODO move cents-handling logic to the SDK layer eventually

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

	fmt.Print(response)

	http.Redirect(w, r, fmt.Sprintf("/deliveries/%s", body.ExternalDeliveryId), http.StatusFound)
}

func deliveriesGETHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Print("Status page\n===========\n")

	tmplPath := path.Join("templates", "deliveries.html")
	tmpl, err := template.ParseFiles(tmplPath)

	if err != nil {
		fmt.Printf("Unable to parse template: %v\n", err.Error())
		http.Error(w, "oh snap", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	if vars["id"] == "" {
		fmt.Print("Requesting status page with no delivery ID\n")

		if err := tmpl.Execute(w, DeliveryResponse{}); err != nil {
			fmt.Printf("Unable to execute template: %v\n", err.Error())
			http.Error(w, "oh snap", http.StatusInternalServerError)
		}
		fmt.Print("\n")
		return
	}

	fmt.Printf("Requesting details of delivery with ID: %v\n", vars["id"])
	response, err := getDelivery(vars["id"])
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Couldn't get your delivery because %v! We've logged an error and will take a look.", err.Error()),
			http.StatusInternalServerError)
		return
	}

	response.DeprefixedDeliveryId = deprefixDeliveryId(response.ExternalDeliveryId)
	response = addFriendlyResponseInfo(response)
	if err := tmpl.Execute(w, response); err != nil {
		fmt.Printf("Unable to execute template: %v\n", err.Error())
		http.Error(w, "oh snap", http.StatusInternalServerError)
	}
	fmt.Print("\n")

}

func deliveriesPOSTHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Print("Status page form submission\n===========================\n")
	// Get the form details
	if err := r.ParseForm(); err != nil {
		fmt.Printf("Unable to parse form: %v\n", err.Error())
		http.Error(w, "oh snap", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/deliveries/%s", prefixDeliveryId(r.FormValue("externalDeliveryId"))), http.StatusFound)
}

func dddeliveriesGETHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := prefixDeliveryId(vars["id"])
	fmt.Printf("Requesting details of delivery with ID: %v\n", id)

	response, err := getDelivery(id)
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

func dddeliveriesPOSTHandler(w http.ResponseWriter, r *http.Request) {
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

func mapFilterPhoneNumber(r rune) rune {
	if r >= '0' && r <= '9' {
		return r
	}

	return -1
}

func prefixDeliveryId(uniqueIdentifier string) string {
	return "RAD-" + uniqueIdentifier
}

func deprefixDeliveryId(deliveryId string) string {
	return strings.TrimLeft(deliveryId, "RAD-")
}

func addFriendlyResponseInfo(delivery DeliveryResponse) DeliveryResponse {
	delivery.PickupPhoneNumber = strings.Trim(delivery.PickupPhoneNumber, "+1")
	delivery.DropoffPhoneNumber = strings.Trim(delivery.DropoffPhoneNumber, "+1")

	switch delivery.DeliveryStatus {
	case "created":
		delivery.DeliveryStatusFriendly = "Your delivery has been created and is awaiting assignment to a Dasher."
		delivery.DeliveryStatusPercentage = 12
		delivery.DeliveryStatusProgressBarClasses = "progress-bar-striped progress-bar-animated"
		delivery.PickupTime = delivery.PickupTimeEstimated
		delivery.DropoffTime = delivery.DropoffTimeEstimated
	case "confirmed":
		delivery.DeliveryStatusFriendly = "Your delivery has been assigned to, and confirmed by, a Dasher."
		delivery.DeliveryStatusPercentage = 24
		delivery.DeliveryStatusProgressBarClasses = "progress-bar-striped progress-bar-animated"
		delivery.PickupTime = delivery.PickupTimeEstimated
		delivery.DropoffTime = delivery.DropoffTimeEstimated
	case "enroute_to_pickup":
		delivery.DeliveryStatusFriendly = "The Dasher is en route to your pick up location."
		delivery.DeliveryStatusPercentage = 36
		delivery.DeliveryStatusProgressBarClasses = "progress-bar-striped progress-bar-animated"
		delivery.PickupTime = delivery.PickupTimeEstimated
		delivery.DropoffTime = delivery.DropoffTimeEstimated
	case "arrived_at_pickup":
		delivery.DeliveryStatusFriendly = "The Dasher has arrived at your pick up location."
		delivery.DeliveryStatusPercentage = 48
		delivery.DeliveryStatusProgressBarClasses = "progress-bar-striped progress-bar-animated"
		delivery.PickupTime = delivery.PickupTimeEstimated
		delivery.DropoffTime = delivery.DropoffTimeEstimated
	case "picked_up":
		delivery.DeliveryStatusFriendly = "The Dasher has picked up your items and is heading to the drop off."
		delivery.DeliveryStatusPercentage = 60
		delivery.DeliveryStatusProgressBarClasses = "progress-bar-striped progress-bar-animated"
		delivery.DropoffTime = delivery.DropoffTimeEstimated
		delivery.PickupTime = delivery.PickupTimeActual
	case "enroute_to_dropoff":
		delivery.DeliveryStatusFriendly = "The Dasher is heading to the drop off location."
		delivery.DeliveryStatusPercentage = 72
		delivery.DeliveryStatusProgressBarClasses = "progress-bar-striped progress-bar-animated"
		delivery.DropoffTime = delivery.DropoffTimeEstimated
		delivery.PickupTime = delivery.PickupTimeActual
	case "arrived_at_dropoff":
		delivery.DeliveryStatusFriendly = "The Dasher has arrived at the drop off location."
		delivery.DeliveryStatusPercentage = 84
		delivery.DeliveryStatusProgressBarClasses = "progress-bar-striped progress-bar-animated"
		delivery.DropoffTime = delivery.DropoffTimeEstimated
		delivery.PickupTime = delivery.PickupTimeActual
	case "delivered":
		delivery.DeliveryStatusFriendly = "Your delivery is complete."
		delivery.DeliveryStatusPercentage = 100
		delivery.DeliveryStatusProgressBarClasses = "bg-success"
		delivery.DropoffTime = delivery.DropoffTimeActual
		delivery.PickupTime = delivery.PickupTimeActual
	case "cancelled":
		delivery.DeliveryStatusFriendly = "Your delivery was cancelled."
		delivery.DeliveryStatusPercentage = 100
		delivery.DeliveryStatusProgressBarClasses = "bg-danger"
	}
	return delivery
}