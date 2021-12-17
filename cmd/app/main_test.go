// main_test.go

// running tests: <CGO_ENABLED=0> go test -v cmd/app/app.go cmd/app/model.go cmd/app/main.go cmd/app/main_test.go
package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"testing"
)

// Used to create the "venue" table
const VenueCreationQuery = `CREATE TABLE IF NOT EXISTS venue
(
	table_number INT NOT NULL auto_increment,
	seats INT UNSIGNED NOT NULL DEFAULT 4,
	
	PRIMARY KEY (table_number)
);`

// Used to create the "guestlist" table
const GuestListCreationQuery = `CREATE TABLE IF NOT EXISTS guestlist (
	id INT NOT NULL auto_increment,
	guest_name VARCHAR (64) CHARACTER SET utf8 UNIQUE,
	table_number INT NOT NULL,
	accompanying_guests INT UNSIGNED NOT NULL, 
	time_arrived TIMESTAMP,
	arrived BOOLEAN DEFAULT FALSE,
	
	PRIMARY KEY (id),
	FOREIGN KEY (table_number) REFERENCES venue(table_number)
  );`

var a App

func TestMain(m *testing.M) {

	//test database info
	username := "user"
	password := "password"
	database := "maindatabase"
	host := "mysql"
	port := "3306"

	// init DB
	a.Init(username, password, host, port, database)

	//making sure tables exist
	createTables()

	code := m.Run()

	resetDB()
	os.Exit(code)

}

//Creates the necessary tables if they don't exist
func createTables() {
	if _, err := a.DB.Exec(VenueCreationQuery); err != nil {
		log.Fatal(err)
	}
	if _, err := a.DB.Exec(GuestListCreationQuery); err != nil {
		log.Fatal(err)
	}
}

//Resets database's tables
func resetDB() {
	a.DB.Exec("DELETE FROM guestlist")
	a.DB.Exec("ALTER TABLE guestlist AUTO_INCREMENT = 1")
	a.DB.Exec("DELETE FROM venue")
	a.DB.Exec("ALTER TABLE venue AUTO_INCREMENT = 1")

}

func initializeDB() {
	resetDB()
	addTable(a.DB, 12)
	addTable(a.DB, 12)
	addTable(a.DB, 12)
}

// Adds guests to DB, if arrived = true it alternates between "arrived" guests and regular additions to guestlist
func addGuests(count int, arrived bool) {

	if arrived { // sets arrived flag = true every other guest
		for i := 1; i <= count; i++ {
			//if it's an arrived guest, guestlist table arrived field = 1; mysql automatically sets time_arrived=NOW()
			a.DB.Exec("INSERT INTO guestlist(guest_name, table_number, accompanying_guests, arrived) VALUES(?, ?, ?, ?)", "TestGuest"+strconv.Itoa(i), i%3+1, (i * 4 % 12), i%2)
		}
	} else { // sets arrived flag = false
		for i := 1; i <= count; i++ {
			a.DB.Exec("INSERT INTO guestlist(guest_name, table_number, accompanying_guests, arrived) VALUES(?, ?, ?, ?)", "TestGuest"+strconv.Itoa(i), i%3+1, (i * 4 % 12), 0)
		}
	}
}

// Executes a given query
func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

// Checks the actual response against the expected response
func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

// Tests for an empty guest list GET /guest_list
func TestEmptyGuestList(t *testing.T) {
	resetDB()

	req, _ := http.NewRequest("GET", "/guest_list", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	// expected API response
	expectedResponse := `{"guests":[]}`
	// checking response body
	if response.Body.String() != expectedResponse {
		t.Errorf("Expected response: `%s`\nGot: '%s'", expectedResponse, response.Body.String())
	}
}

// Tests handlerAddTable() POST /venue
func TestHandlerAddTable(t *testing.T) {
	resetDB()

	// preparing request
	var jsonStr = []byte(`{"seats": 4}`)
	req, _ := http.NewRequest("POST", "/venue", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	// Query table for contents
	var table_number, seats int
	err := a.DB.QueryRow("SELECT * FROM venue").Scan(&table_number, &seats)

	if err != nil {
		t.Errorf("database issue %s", err)
	}
	if table_number != 1 {
		t.Errorf("Expected table_number to be 1. Got '%d'", table_number)
	}
	if seats != 4 {
		t.Errorf("Expected seats to be 4. Got '%d'", seats)
	}
}

// Tests handlerAddGuest() POST /guest_list/name
func TestHandlerAddGuest(t *testing.T) {
	initializeDB()

	// preparing request
	var jsonStr = []byte(`{"table": 1, "accompanying_guests": 2}`)
	req, _ := http.NewRequest("POST", "/guest_list/TestGuest1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	// expected API response
	expectedResponse := `{"name":"TestGuest1"}`

	// checking response body
	if response.Body.String() != expectedResponse {
		t.Errorf("Expected response: `%s`\nGot: '%s'", expectedResponse, response.Body.String())
	}
	// Tests handlerAddGuest for failure to add a guest on edge case
	// (12 free seats, 12 accompanying guests)

	jsonStr = []byte(`{"table": 1, "accompanying_guests": 12}`)
	req, _ = http.NewRequest("POST", "/guest_list/TestGuest2", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusConflict, response.Code)
}

//Tests handlerGuestArrives() PUT /guests/name
func TestHandlerGuestArrives(t *testing.T) {
	initializeDB()

	addGuests(3, false) //adding 3 guests

	// preparing request
	var jsonStr = []byte(`{"accompanying_guests": 6}`)
	req, _ := http.NewRequest("PUT", "/guests/TestGuest1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	// expected API response
	expectedResponse := `{"name":"TestGuest1"}`

	// checking response body
	if response.Body.String() != expectedResponse {
		t.Errorf("Expected response: `%s`\nGot: '%s'", expectedResponse, response.Body.String())
	}

	// Tests for failure to update guest on edge case
	// (12 free seats, 12 accompanying guests)

	// preparing request
	jsonStr = []byte(`{"accompanying_guests": 12}`)
	req, _ = http.NewRequest("PUT", "/guests/TestGuest2", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusConflict, response.Code)

	// Tests updating guest status when arriving with
	// the same amount of accompanying guests

	// preparing request
	jsonStr = []byte(`{"accompanying_guests": 0}`)
	req, _ = http.NewRequest("PUT", "/guests/TestGuest3", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	// expected API response
	expectedResponse = `{"name":"TestGuest3"}`

	// checking response body
	if response.Body.String() != expectedResponse {
		t.Errorf("Expected response: `%s`\nGot: '%s'", expectedResponse, response.Body.String())
	}
}

//Tests handlerGuestList() GET /guest_list
func TestHandlerGuestList(t *testing.T) {
	initializeDB()

	addGuests(3, false) //adding 3 guests

	// preparing request
	var jsonStr = []byte(`{}`)
	req, _ := http.NewRequest("GET", "/guest_list", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	// expected API response
	expectedResponse := `{"guests":[{"name":"TestGuest1","table":2,"accompanying_guests":4},{"name":"TestGuest2","table":3,"accompanying_guests":8},{"name":"TestGuest3","table":1,"accompanying_guests":0}]}`

	// checking response body
	if response.Body.String() != expectedResponse {
		t.Errorf("Expected response: `%s`\nGot: '%s'", expectedResponse, response.Body.String())
	}

}

//Tests handlerGetGuest() GET /guests/name (NEW Endpoint)
func TestHandlerGetGuest(t *testing.T) {
	initializeDB()

	addGuests(1, false) //adding 1 guest

	req, _ := http.NewRequest("GET", "/guests/TestGuest1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/guests/TestGuest2", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

//Tests handlerGuestLeaves() DELETE /guests/name
func TestHandlerGuestLeaves(t *testing.T) {
	initializeDB()

	addGuests(1, false) //adding 1 guest

	req, _ := http.NewRequest("GET", "/guests/TestGuest1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	g := Guest{}
	json.Unmarshal(response.Body.Bytes(), &g)

	if g.Name != "TestGuest1" {
		t.Errorf("Expected name to be 'TestGuest1'. Got '%v'", g.Name)
	}

	req, _ = http.NewRequest("DELETE", "/guests/TestGuest1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/guests/TestGuest1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

// testing hanbdlerArrivedGuests() GET /guests
func TestHandlerArrivedGuests(t *testing.T) {
	initializeDB()

	addGuests(3, false) //adding 3 guests

	var jsonStr = []byte(`{}`)
	req, _ := http.NewRequest("GET", "/guests", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	gl := GuestList{}
	json.Unmarshal(response.Body.Bytes(), &gl)

	// expected API response
	expectedResponse := `{"guests":[]}`

	// checking response body
	if response.Body.String() != expectedResponse {
		t.Errorf("Expected response: `%s`\nGot: '%s'", expectedResponse, response.Body.String())
	}

	initializeDB()

	addGuests(3, true) //adding 3 guests (2)

	req, _ = http.NewRequest("GET", "/guests", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	gl = GuestList{}
	json.Unmarshal(response.Body.Bytes(), &gl)

	// expected API response regular expression
	expectedResponse = `\{"guests":\[\{"name":"TestGuest1","table":2,"accompanying_guests":0,"time_arrived":"[0-9][0-9][0-9][0-9]\-[0-9][0-9]\-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]:[0-9][0-9]Z"\},\{"name":"TestGuest3","table":1,"accompanying_guests":0,"time_arrived":"[0-9][0-9][0-9][0-9]\-[0-9][0-9]\-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]:[0-9][0-9]Z"\}\]\}`

	expectedResponseRegex, _ := regexp.Compile(expectedResponse)
	// checking response body
	if !expectedResponseRegex.MatchString(response.Body.String()) {
		t.Errorf("Expected response: `%s`\nGot: '%s'", expectedResponse, response.Body.String())
	}

}

//Tests handlerSeatsEmpty() GET /seats_empty
func TestHandlerSeatsEmpty(t *testing.T) {
	initializeDB()

	//testing for all empty seats (36)
	var jsonStr = []byte(`{}`)
	req, _ := http.NewRequest("GET", "/seats_empty", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// expected API response
	expectedResponse := `{"seats_empty":36}`
	// checking response body
	if response.Body.String() != expectedResponse {
		t.Errorf("Expected response: `%s`\nGot: '%s'", expectedResponse, response.Body.String())
	}

	//testing for 22 empty seats

	addGuests(2, false) //adding 2 guests with 4 and 8 guests; total=14

	jsonStr = []byte(`{}`)
	req, _ = http.NewRequest("GET", "/seats_empty", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	// expected API response
	expectedResponse = `{"seats_empty":22}`
	// checking response body
	if response.Body.String() != expectedResponse {
		t.Errorf("Expected response: `%s`\nGot: '%s'", expectedResponse, response.Body.String())
	}
}
