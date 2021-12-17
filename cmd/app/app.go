package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux" // in order to proccess different types of requests
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// Initialize mysql with login credentials (user, password) and database name (dbname)
func (a *App) Init(user string, password string, host string, port string, dbname string) {

	var err error

	//initialize DB
	//"user:password@/database?parseTime=true"
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, dbname)

	a.DB, err = sql.Open("mysql", dataSource)

	if err != nil {
		log.Fatal(err)
	}

	//mux
	a.Router = mux.NewRouter()

	a.initializeRoutes()
}

// Runs the App on adress (addr)
func (a *App) Run(addr string) {
	err := http.ListenAndServe(addr, a.Router)

	if err != nil {
		log.Fatal(err)
	}
}

// Initialize routing
func (a *App) initializeRoutes() {

	a.Router.HandleFunc("/guest_list/{name}", a.handlerAddGuest).Methods("POST")  // Add a guest to the guestlist "POST /guest_list/name"
	a.Router.HandleFunc("/guest_list", a.handlerGuestList).Methods("GET")         // Get the guest list "GET /guest_list"
	a.Router.HandleFunc("/guests/{name}", a.handlerGuestArrives).Methods("PUT")   // Guest Arrives "PUT /guests/name"
	a.Router.HandleFunc("/guests/{name}", a.handlerGuestLeaves).Methods("DELETE") // Guest Leaves "DELETE /guests/name"
	a.Router.HandleFunc("/guests", a.handlerArrivedGuests).Methods("GET")         // Get arrived guests "GET /guests"
	a.Router.HandleFunc("/seats_empty", a.handlerSeatsEmpty).Methods("GET")       // Count number of empty seats "GET /seats_empty"
	a.Router.HandleFunc("/guests/{name}", a.handlerGetGuest).Methods("GET")       // Gets guest info "GET /guests/name"
	a.Router.HandleFunc("/venue", a.handlerAddTable).Methods("POST")              // Adds table to venue "POST /venue"
}

// Sends JSON responses
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {

	// Marshal payload into JSON format
	response, _ := json.Marshal(payload)

	// Set header
	w.Header().Set("Content-Type", "application/json")

	// Set HTTP status code
	w.WriteHeader(code)

	// Set body
	w.Write(response)
}

// Sends error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

/*
### Add a guest to the guestlist

If there is insufficient space at the specified table, throws an error (http.StatusConflict).

POST /guest_list/name
body:
{
    "table": int,
    "accompanying_guests": int
}
response:
{
    "name": "string"
}
*/
func (a *App) handlerAddGuest(w http.ResponseWriter, r *http.Request) {

	name := mux.Vars(r)["name"] // Get guest name
	var g Guest

	// Decoding request body into Guest struct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&g); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	g.Name = name

	// Adding guest to guest list
	if err := g.addGuest(a.DB); err != nil {
		respondWithError(w, http.StatusConflict, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"name": g.Name})
}

/*
### Get the guest list

GET /guest_list
response:
{
    "guests": [
        {
            "name": "string",
            "table": int,
            "accompanying_guests": int
        }, ...
    ]
}
*/
func (a *App) handlerGuestList(w http.ResponseWriter, r *http.Request) {

	var g GuestList
	var err error

	// Get all guests from guestlist
	if g, err = getGuestList(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, g)

}

/*
### Guest Arrives
HandlerGuestArrives () handles the PUT /guest/name requests

A guest may arrive with an entourage that is not the size indicated at the guest list.
If the table is expected to have space for the extras, allow them to come. Otherwise, this method throws an error (http.StatusConflict).


PUT /guests/name
body:
{
    "accompanying_guests": int
}
response:
{
    "name": "string"
}
*/
func (a *App) handlerGuestArrives(w http.ResponseWriter, r *http.Request) {

	name := mux.Vars(r)["name"] // Get guest name
	var g Guest

	// Decoding request body into Guest struct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&g); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	g.Name = name

	// Updating guest arrived time/arrived flag on the database
	if err := g.updateGuest(a.DB); err != nil {
		respondWithError(w, http.StatusConflict, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"name": name})
}

/*
### Guest Leaves

When a guest leaves, all their accompanying guests leave as well.

DELETE /guests/name
*/
func (a *App) handlerGuestLeaves(w http.ResponseWriter, r *http.Request) {

	name := mux.Vars(r)["name"] // Get guest name

	// Deleting guest by name
	if err := deleteGuest(a.DB, name); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

/*
### Get arrived guests

GET /guests
response:
{
    "guests": [
        {
            "name": "string",
            "accompanying_guests": int,
            "time_arrived": "string"
        }
    ]
}
*/
func (a *App) handlerArrivedGuests(w http.ResponseWriter, r *http.Request) {

	var g GuestList
	var err error

	// Get all guests from guestlist
	if g, err = getArrivedGuests(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, g)

}

/*
### Count number of empty seats

GET /seats_empty
response:
{
    "seats_empty": int
}
*/
func (a *App) handlerSeatsEmpty(w http.ResponseWriter, r *http.Request) {

	s := SeatsEmpty{}
	var err error

	// Get empty seats
	if s.Seats, err = getFreeSeats(a.DB, 0, true); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, s)

}

/*
### Get the specified guest

***NEW useful endpoint***

GET /guests/name
response:
{
    "name": "string",
    "table": int,
    "accompanying_guests": int,
	"arrived": bool
}
*/
func (a *App) handlerGetGuest(w http.ResponseWriter, r *http.Request) {

	name := mux.Vars(r)["name"] // Get guest name

	var g Guest
	var err error

	// Get all guests from guestlist
	if g, err = getGuest(a.DB, name); err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, g)
}

/*
### Add a new table

***NEW useful endpoint***

POST /venue

body:
{
	"seats": int
}
*/
func (a *App) handlerAddTable(w http.ResponseWriter, r *http.Request) {

	// Temporary struct used for decoding
	seats := struct {
		S int `json:"seats"`
	}{}

	// Decoding request into Guest struct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&seats); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// Adding new table
	if err := addTable(a.DB, seats.S); err != nil {
		respondWithError(w, http.StatusConflict, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"result": "success"})
}
