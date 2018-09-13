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
func getAPI(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(Api{
		Uptime: iso8601.Format(time.Since(startTime)),
		Info: DESC,
		Version: VERSION,
	})
}

// Reponds with the recorded IDs
func getIGC(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(dbListTracks())
}

// Reponds with the track data for the recorded ID, if any
func getID(res http.ResponseWriter, id string) {
	if data, err := dbGetTrack(id); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	} else {
		res.Header().Set("Content-Type", "application/json")
		json.NewEncoder(res).Encode(data)
	}
}

// Responds with the field of the given name for the ID, if both parameters exist
func getField(res http.ResponseWriter, id string, field string){
	if data, err := dbGetTrack(id); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	} else {
		if value, err := data.GetField(field); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		} else {
			res.Header().Set("Content-Type", "text/plain")
			fmt.Fprintln(res, value)
			// w.Write([]byte(value))
		}
	}
}

// Records the track by URL and returns its stored ID, if valid
func postIGC(res http.ResponseWriter, req *http.Request){
	var data IGCReq
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		http.Error(res, "Invalid body.", http.StatusBadRequest)
	} else {
		if id, err := dbCreateTrack(data.URL); err != nil {
			http.Error(res, "Url did not contain track data.", http.StatusBadRequest)
		} else {
			res.Header().Set("Content-Type", "application/json")
			json.NewEncoder(res).Encode(IGCRes{
				Id: id,
			})
		}
	}
}

func handleFunc(res http.ResponseWriter, req *http.Request) {
	match := pattern.FindStringSubmatch(req.URL.Path)
	if match != nil {
		if match[5] != "" {
			if req.Method == http.MethodGet {
				getField(res, match[3], match[5])
				return
			}
		} else if match[3] != "" {
			if req.Method == http.MethodGet {
				getID(res, match[3])
				return
			}
		} else if match[0] == "/api/igc" {
			if req.Method == http.MethodGet {
				getIGC(res)
				return
			} else if req.Method == http.MethodPost {
				postIGC(res, req)
				return
			}
		} else if match[0] == "/api" {
			if req.Method == http.MethodGet {
				getAPI(res)
				return
			}
		}
	}

	// default response
	http.NotFound(res, req)
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
