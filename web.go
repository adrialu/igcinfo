package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/p3lim/iso8601" // I wrote and published this since I couldn't find anything like it
)

const (
	VERSION = "v1"
	DESC    = "Service for IGC tracks."
)

// data model for api status
type Status struct {
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
	Version string `json:"version"`
}

// used to store the time the webservers started
var startTime time.Time

// Responds with the current status of the API
func getStatus(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Status{
		Uptime:  iso8601.Format(time.Since(startTime)),
		Info:    DESC,
		Version: VERSION,
	})
}

// Reponds with the stored track IDs
func getTracks(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, dbListTracks())
}

// Reponds with the track data for the given track ID, if any
func getTrack(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if data, err := dbGetTrack(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		render.JSON(w, r, data)
	}
}

// Responds with the field of the given name for the track ID, if both exist
func getTrackField(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if data, err := dbGetTrack(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		field := chi.URLParam(r, "field")
		if value, err := data.GetField(field); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			// text/plain is the default content type, we can just write the value directly
			w.Write([]byte(value))
		}
	}
}

// Stores the track from url and responds with the stored track ID, if data is valid
func createTrack(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid body.", http.StatusBadRequest)
	} else {
		if url, ok := data["url"]; !ok {
			http.Error(w, "Invalid body.", http.StatusBadRequest)
		} else {
			if id, err := dbCreateTrack(url); err != nil {
				http.Error(w, "Url did not contain track data.", http.StatusBadRequest)
			} else {
				response := make(map[string]int)
				response["id"] = id
				render.JSON(w, r, response)
			}
		}
	}
}

func main() {
	// set init time
	startTime = time.Now()

	// init "database"
	dbInit()

	// get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("No port specified, using 8080")
		port = "8080"
	}

	// create webserver routes
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		r.Get("/", getStatus)
		r.Route("/igc", func(r chi.Router) {
			r.Get("/", getTracks)
			r.Post("/", createTrack)
			r.Route("/{id:[0-9]+}", func(r chi.Router) {
				r.Get("/", getTrack)
				r.Get("/{field:[A-Za-z_]+}", getTrackField)
			})
		})
	})

	// start webserver
	log.Fatal(http.ListenAndServe(":"+port, router))
}
