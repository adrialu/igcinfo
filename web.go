package main

import (
	"os"
	"fmt"
	"log"
	"regexp"
	"encoding/json"
	"time"
	"net/http"

	"github.com/p3lim/iso8601" // I wrote and published this since I couldn't find anything like it
)

const (
	VERSION = "v1"
	DESC    = "Service for IGC tracks."
)

type Api struct {
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

var pattern = regexp.MustCompile("^/api(/igc(/([0-9]+)(/([a-zA-Z_]+))?)?)?$")

var startTime time.Time

// Responds with the current status of the API
func getAPI(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Api{
		Uptime: iso8601.Format(time.Since(startTime)),
		Info: DESC,
		Version: VERSION,
	})
}

// Reponds with the recorded IDs
func getIGC(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dbListTracks())
}

// Reponds with the track data for the recorded ID, if any
func getID(w http.ResponseWriter, id string) {
	if data, err := dbGetTrack(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

// Responds with the field of the given name for the ID, if both parameters exist
func getField(w http.ResponseWriter, id string, field string){
	if data, err := dbGetTrack(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		if value, err := data.GetField(field); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintln(w, value)
			// w.Write([]byte(value))
		}
	}
}

// Records the track by URL and returns its stored ID, if valid
func postIGC(w http.ResponseWriter, r *http.Request){
	var data IGCReq
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid body.", http.StatusBadRequest)
	} else {
		if id, err := dbCreateTrack(data.URL); err != nil {
			http.Error(w, "Url did not contain track data.", http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(IGCRes{
				Id: id,
			})
		}
	}
}

func handleFunc(w http.ResponseWriter, r *http.Request) {
	match := pattern.FindStringSubmatch(r.URL.Path)
	if match != nil {
		if match[5] != "" {
			if r.Method == http.MethodGet {
				getField(w, match[3], match[5])
				return
			}
		} else if match[3] != "" {
			if r.Method == http.MethodGet {
				getID(w, match[3])
				return
			}
		} else if match[0] == "/api/igc" {
			if r.Method == http.MethodGet {
				getIGC(w)
				return
			} else if r.Method == http.MethodPost {
				postIGC(w, r)
				return
			}
		} else if match[0] == "/api" {
			if r.Method == http.MethodGet {
				getAPI(w)
				return
			}
		}
	}

	// default response
	http.NotFound(w, r)
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

	// start webserver
	http.HandleFunc("/", handleFunc)
	log.Fatal(http.ListenAndServe(":" + port, nil))
}
