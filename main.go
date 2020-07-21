package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// With apologies to SWAPI

var personId int
var vehicleId int
var people []Person
var vehicles []Vehicle

type Person struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Species   string `json:"species"`
	Homeworld string `json:"homeworld"`
	YOB       string `json:"YOB"`
}

func newPerson(name string, species string, homeworld string, yearOfBirth string) Person {
	personId++
	return Person{
		Id:        personId,
		Name:      name,
		Species:   species,
		Homeworld: homeworld,
		YOB:       yearOfBirth}
}

type Vehicle struct {
	Id                 int    `json:"id"`
	Name               string `json:"name"`
	Cost_in_credits    int    `json:"cost_in_credits"`
	Passenger_capacity int    `json:"crew"`
	Vehicle_class      string `json:"vehicle_class"`
}

func newVehicle(name string, costInCredits int, passengerCapacity int, vehicleClass string) Vehicle {
	vehicleId++
	return Vehicle{
		Id:                 vehicleId,
		Name:               name,
		Cost_in_credits:    costInCredits,
		Passenger_capacity: passengerCapacity,
		Vehicle_class:      vehicleClass}
}

func populatePeople() {
	obiWan := newPerson(
		"Obi-Wan Kenobi",
		"Homo sapiens sapiens",
		"Stewjon",
		"57BBY") //Before the Battle of Yavin
	leia := newPerson(
		"Leia Organa",
		"Homo sapiens sapiens",
		"Alderaan",
		"19BBY")
	people = append(people, obiWan, leia)
}

func populateVehicles() {
	milleniumFalcon := newVehicle(
		"Millenium Falcon",
		100000,
		10,
		"flying")
	vehicles = append(vehicles, milleniumFalcon)
}

func getPeople(w http.ResponseWriter, r *http.Request) {
	log.Println("GETTING PEOPLE")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(people)
}

func getPerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	log.Println("GETTING PERSON", id)
	intId, strConvErr := strconv.Atoi(id)
	if strConvErr != nil {
		http.Error(w, "ID parameter appears to be malformed. Please send a valid integer.", http.StatusBadRequest)
		return //to prevent handler from continuing to run. Could probably also utilize `defer resp.Body.Close()`
	}
	p, findPersonErr := findPerson(intId)
	if findPersonErr != nil {
		http.Error(w, findPersonErr.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func createPerson(w http.ResponseWriter, r *http.Request) {
	var p Person
	log.Println("CREATING PERSON")
	// Decode the request body into the struct.
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	personId++
	p.Id = personId
	people = append(people, p)
	json.NewEncoder(w).Encode(p)
}

func deletePerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	intId, strConvErr := strconv.Atoi(params["id"])
	if strConvErr != nil {
		http.Error(w, "ID parameter appears to be malformed. Please send a valid integer.", http.StatusBadRequest)
		return
	}
	log.Println("DELETING PERSON", intId)
	var toDelete Person
	for ind, val := range people {
		if val.Id == intId {
			toDelete = val
			people = append(people[:ind], people[ind+1:]...)
			break
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toDelete)
}

func updatePerson(w http.ResponseWriter, r *http.Request) {
	//Should also be able to update from a partial object
	params := mux.Vars(r)
	intId, strConvErr := strconv.Atoi(params["id"])
	if strConvErr != nil {
		http.Error(w, "ID parameter appears to be malformed. Please send a valid integer.", http.StatusBadRequest)
		return
	}
	log.Println("UPDATING PERSON", intId)
	var p Person
	_ = json.NewDecoder(r.Body).Decode(&p)
	for ind, val := range people {
		if val.Id == intId {
			people = append(people[:ind], people[ind+1:]...) //Remove previous
			p.Id = val.Id
			people = append(people, p)
			json.NewEncoder(w).Encode(&p)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func findPerson(id int) (Person, error) {
	for _, val := range people {
		if val.Id == id {
			return val, nil
		}
	}
	// Better return value?
	return Person{}, errors.New("No person found with that ID.")
}

func getVehicles(w http.ResponseWriter, r *http.Request) {
	log.Println("GETTING VEHICLES")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vehicles)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("GREETINGS")
}

func main() {
	populatePeople()
	populateVehicles()
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/people", getPeople).Methods("GET")
	r.HandleFunc("/people/{id}", getPerson).Methods("GET")
	r.HandleFunc("/people", createPerson).Methods("POST")
	r.HandleFunc("/people/{id}", updatePerson).Methods("PUT")
	r.HandleFunc("/people/{id}", deletePerson).Methods("DELETE")
	r.HandleFunc("/vehicles", getVehicles).Methods("GET")
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":4500", nil))
}
