// model.go

package main

import (
	"database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql"
)

// Base struct to store guest info
type Guest struct {
	Name               string `json:"name,omitempty"`
	Table              int    `json:"table,omitempty"`
	AccompanyingGuests int    `json:"accompanying_guests"`
	TimeArrived        string `json:"time_arrived,omitempty"`
	Arrived            int    `json:"arrived,omitempty"`
}

// Struct used multiple guest body responses
type GuestList struct {
	Guests []Guest `json:"guests"`
}

// Struct used for /seats_empty endpoint body
type SeatsEmpty struct {
	Seats int `json:"seats_empty"`
}

//Adds a new table to the venue table
func addTable(db *sql.DB, seats int) error {
	_, err := db.Exec("INSERT INTO venue (seats) values (?)", seats)

	return err
}

// Handles the addition of new guests to the guestlist
func (g *Guest) addGuest(db *sql.DB) error {

	// Checking number of free seats instead of relying on DBs strict mode with UNSIGNED
	freeSeats, err := getFreeSeats(db, g.Table, false)
	freeSeats = freeSeats - g.AccompanyingGuests - 1 // main guest is not accounted by AccompanyingGuests

	if err != nil {
		return err
	}

	// if there aren't enough sits
	if freeSeats < 0 {
		return errors.New("unable to add guest")
	}

	// Adds guest to guestlist table
	_, err = db.Exec("INSERT INTO guestlist (guest_name, table_number, accompanying_guests, arrived) values (?, ?, ?, ?)", g.Name, g.Table, g.AccompanyingGuests, false)

	return err
}

// Updates DB entry with time_arrived and sets arrived flag to "true"
func (g *Guest) updateGuest(db *sql.DB) error {

	// Get previous ammount of accompanying guests
	var previousAccompanyingGuests int
	err := db.QueryRow("SELECT table_number, accompanying_guests FROM guestlist WHERE guest_name = ?", g.Name).Scan(&g.Table, &previousAccompanyingGuests)

	if err != nil {
		return err
	}

	// if there are no changes in accompanying guests doesn't check sits
	// else checks sits
	if previousAccompanyingGuests == g.AccompanyingGuests {

		// updates guest on DB
		_, err = db.Exec("UPDATE guestlist SET time_arrived=NOW(), arrived=? WHERE guest_name=?", true, g.Name)

		return err

	} else {
		// Checking number of free seats
		var freeSeats int
		freeSeats, err = getFreeSeats(db, g.Table, false)

		freeSeats = freeSeats + previousAccompanyingGuests - g.AccompanyingGuests // new free seats count

		if err != nil {
			return err
		}

		// if there aren't enough sits
		if freeSeats < 0 {
			return errors.New("unable to add guest")
		}

		// updates guest on DB
		_, err = db.Exec("UPDATE guestlist SET accompanying_guests=?, time_arrived=NOW(), arrived=? WHERE guest_name=?", g.AccompanyingGuests, true, g.Name)
		return err
	}

}

// Queries databse and returns a GuestList struct with all guests on the guestlist table
func getGuestList(db *sql.DB) (GuestList, error) {
	gl := GuestList{}
	gl.Guests = []Guest{}

	// Get all guests from guestlist
	rows, err := db.Query("SELECT guest_name, table_number, accompanying_guests FROM guestlist")

	if err != nil {
		return gl, err
	}

	defer rows.Close()

	// Foreach guest
	for rows.Next() {
		var g Guest

		if err := rows.Scan(&g.Name, &g.Table, &g.AccompanyingGuests); err != nil {
			return gl, err
		}

		gl.Guests = append(gl.Guests, g)
	}

	return gl, nil // append guest to GuestList
}

// Queries databse and returns a GuestList struct with arrived guests
func getArrivedGuests(db *sql.DB) (GuestList, error) {
	gl := GuestList{}
	gl.Guests = []Guest{}

	// Get all guests with arrived=true from guestlist
	rows, err := db.Query("SELECT guest_name, table_number, time_arrived FROM guestlist WHERE arrived=1")

	if err != nil {
		return gl, err
	}

	defer rows.Close()

	// Foreach guest
	for rows.Next() {
		var g Guest

		if err := rows.Scan(&g.Name, &g.Table, &g.TimeArrived); err != nil {
			return gl, err
		}

		gl.Guests = append(gl.Guests, g) // append guest to GuestList
	}

	return gl, nil
}

// Deletes guest entry from DB
func deleteGuest(db *sql.DB, name string) error {

	_, err := db.Exec("DELETE FROM guestlist WHERE guest_name = ?", name)

	return err
}

/* Queries database for the number of free seats
	If all = false, returns amount of free seats on table
 	If all = true, returns all available seats
*/
func getFreeSeats(db *sql.DB, table int, all bool) (int, error) {

	var freeSeats int
	var usedSeats int
	var err error

	if all { // Get free seats

		// query DB for available sits
		err = db.QueryRow("SELECT SUM(seats) FROM venue").Scan(&freeSeats)
		if err != nil {
			return 0, err
		}

		// query DB for used sits
		err = db.QueryRow("SELECT SUM(accompanying_guests + 1) FROM guestlist").Scan(&usedSeats)

	} else { // Get free seats from specified table number

		// query DB for available on specified table
		err = db.QueryRow("SELECT seats FROM venue WHERE table_number=?", table).Scan(&freeSeats)
		if err != nil {
			return 0, err
		}

		// query DB for used sits on specified table
		err = db.QueryRow("SELECT SUM(accompanying_guests + 1) FROM guestlist WHERE table_number=?", table).Scan(&usedSeats)
	}

	//TODO properly handle error when there are no guests on guestlist for specified table
	//BUG workaround setting err nil
	if err != nil {
		return freeSeats, nil
	}

	return freeSeats - usedSeats, err
}

// Get guest (name) from guestlist
func getGuest(db *sql.DB, name string) (Guest, error) {
	var g Guest
	g.Name = name

	err := db.QueryRow("SELECT table_number, accompanying_guests, arrived FROM guestlist WHERE guest_name=?", g.Name).Scan(&g.Table, &g.AccompanyingGuests, &g.Arrived)

	return g, err
}
