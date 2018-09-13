package main

import (
	"os"
	"fmt"
	"log"
	"encoding/json"
	"time"
	"net/http"

	"github.com/p3lim/iso8601" // I wrote and published this since I couldn't find anything like it
	"github.com/go-chi/render"
	"github.com/go-chi/chi"
)

const (
	VERSION = "v1"
	DESC    = "Service for IGC tracks."
)

type Status struct {
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
	Version string `json:"version"`
}

type IGCReq struct {
	URL string `json:"url"`
}

type IGCRes struct {
	Id int `json:"id"`
}

var startTime time.Time

// Responds with the current status of the API
func getStatus(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Status{
		Uptime: iso8601.Format(time.Since(startTime)),
		Info: DESC,
		Version: VERSION,
	})
}

// Reponds with the recorded IDs
func getTracks(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, dbListTracks())
}

// Reponds with the track data for the recorded ID, if any
func getTrack(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if data, err := dbGetTrack(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		render.JSON(w, r, data)
	}
}

// Responds with the field of the given name for the ID, if both parameters exist
func getTrackField(w http.ResponseWriter, r *http.Request){
	id := chi.URLParam(r, "id")
	if data, err := dbGetTrack(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		field := chi.URLParam(r, "field")
		if value, err := data.GetField(field); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			w.Write([]byte(value))
		}
	}
}

// Records the track by URL and returns its stored ID, if valid
func createTrack(w http.ResponseWriter, r *http.Request){
	var data IGCReq
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid body.", http.StatusBadRequest)
	} else {
		if id, err := dbCreateTrack(data.URL); err != nil {
			http.Error(w, "Url did not contain track data.", http.StatusBadRequest)
		} else {
			render.JSON(w, r, IGCRes{Id: id})
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
		fmt.Println("No port specified, using 8080")
		port = "8080"
	}

	// webserver
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
	log.Fatal(http.ListenAndServe(":" + port, router))
}
