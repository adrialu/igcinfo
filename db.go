package main

import (
	"fmt"
	"time"
	"errors"
	"strconv"
	"reflect"

	"github.com/marni/goigc" // required for the assignment
)


// Track data model
type Track struct {
	Pilot       string    `json:"pilot"`
	Glider      string    `json:"glider"`
	GliderID    string    `json:"glider_id"`
	TrackLength float64   `json:"track_length"`
	H_date      time.Time `json:"H_date"`
}

// There's probably a better way of doing in-memory databases, but this works
var database map[int]Track

func dbInit() {
	database = make(map[int]Track)
}

// Creates a track by the given ICG file from a remote location
func dbCreateTrack(url string) (int, error) {
	if track, err := igc.ParseLocation(url); err != nil {
		return 0, err
	} else {
		// calculate the total distance of the track data
		distance := 0.0
		for i := 0; i < len(track.Points) - 1; i++ {
			distance += track.Points[i].Distance(track.Points[i + 1])
		}

		// get next available ID
		id := len(database) + 1

		// store in database using our data model
		database[id] = Track{
			Pilot: track.Pilot,
			Glider: track.GliderType,
			GliderID: track.GliderID,
			TrackLength: distance,
			H_date: track.Date,
		}

		return id, nil
	}
}

// Returns an array of all track IDs in the database
func dbListTracks() []int {
	// iterate through the database and store the IDs in a new slice
	ids := make([]int, 0, len(database))
	for id := range database {
		ids = append(ids, id)
	}

	return ids
}

// Returns a track by the given ID, if any
func dbGetTrack(id_s string) (Track, error) {
	var t Track

	// convert the ID to an int
	id, err := strconv.Atoi(id_s)
	if err != nil {
		return t, errors.New("Invalid ID.")
	}

	// return the data if it exists
	if data, ok := database[id]; ok {
		return data, nil
	} else {
		return t, errors.New("Track with ID '" + id_s + "' doesn't exist.")
	}
}

// Method that returns the track field value by the given json tag
func (t *Track) GetField(tag string) (string, error) {
	// iterate through the struct fields and return the value of the field that matches
	// the given json tag
	val := reflect.ValueOf(t).Elem()
	for i := 0; i < val.NumField(); i++ {
		if val.Type().Field(i).Tag.Get("json") == tag {
			return fmt.Sprint(val.Field(i)), nil
		}
	}

	return "", errors.New("Track with has no field '" + tag + "'.")
}
