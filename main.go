package main

import (
	"os"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"encoding/json"
	"time"
	"net/http"
	"reflect"

	"github.com/p3lim/iso8601" // I wrote and published this since I couldn't find anything like it
	"github.com/marni/goigc"
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

type IGCData struct {
	Pilot       string    `json:"pilot"`
	Glider      string    `json:"glider"`
	GliderID    string    `json:"glider_id"`
	TrackLength float64   `json:"track_length"`
	H_date      time.Time `json:"h_date"`
}

var pattern = regexp.MustCompile("^/api(/igc(/([0-9]+)(/([a-zA-Z_]+))?)?)?$")

var startTime time.Time
var database map[int]IGCData

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
	ids := make([]int, len(database))
	for id := range database {
		ids = append(ids, id)
	}

	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(ids)
}

// Reponds with the track data for the recorded ID, if any
func getID(res http.ResponseWriter, id_s string) {
	id, _ := strconv.Atoi(id_s) // error is already handled by the regex pattern
	if data, ok := database[id]; ok {
		res.Header().Set("Content-Type", "application/json")
		json.NewEncoder(res).Encode(data)
	} else {
		http.Error(res, "ID doesn't exist", 404)
	}
}

// Responds with the field of the given name for the ID, if both parameters exist
func getField(res http.ResponseWriter, id_s string, field string){
	id, _ := strconv.Atoi(id_s) // error is already handled by the regex pattern
	if data, ok := database[id]; ok {
		val := reflect.ValueOf(&data).Elem()

		for i := 0; i < val.NumField(); i++ {
			if val.Type().Field(i).Tag.Get("json") == field {
				res.Header().Set("Content-Type", "text/plain")
				fmt.Fprintln(res, val.Field(i))
				return
			}
		}

		http.Error(res, "Field doesn't exist", 404)
	} else {
		http.Error(res, "ID doesn't exist", 404)
	}
}

// Records the track by URL and returns its stored ID, if valid
func postIGC(res http.ResponseWriter, req *http.Request){
	var reqData IGCReq
	json.NewDecoder(req.Body).Decode(&reqData)

	track, err := igc.ParseLocation(reqData.URL)
	if err != nil {
		http.Error(res, "Failed to load track data", 400)
		return
	}

	distance := 0.0
	for i := 0; i < len(track.Points) - 1; i++ {
		distance += track.Points[i].Distance(track.Points[i + 1])
	}

	data := IGCData{
		Pilot: track.Pilot,
		Glider: track.GliderType,
		GliderID: track.GliderID,
		TrackLength: distance,
		H_date: track.Date,
	}

	id := len(database) + 1
	database[id] = data

	fmt.Println(database)

	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(IGCRes{
		Id: id,
	})
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
	database = make(map[int]IGCData)

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
