package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	middleware "github.com/paulcager/go-http-middleware"
	"github.com/paulcager/osgridref"
	flag "github.com/spf13/pflag"
)

const (
	apiVersion = "v5"
)

var (
	staticCache time.Duration
	listenPort  string
)

func main() {
	flag.StringVar(&listenPort, "port", ":9090", "Port to listen on")
	flag.DurationVar(&staticCache, "static-cache-max-age", 1*time.Hour, "If not zero, the max-age property to set in Cache-Control for responses")
	flag.Parse()

	server := makeHTTPServer(listenPort)
	log.Fatal(server.ListenAndServe())
}

type Reply struct {
	OSGridRef  string  `json:"osGridRef"`
	Easting    int     `json:"easting"`
	Northing   int     `json:"northing"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	UsageCount int     `json:"usageCount"`
}

func makeHTTPServer(listenPort string) *http.Server {
	http.Handle("/"+apiVersion+"/gridref/", middleware.MakeCachingHandler(1*time.Hour, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gridRefStr := r.URL.Path[len("/"+apiVersion+"/gridref/"):]
			gridRef, err := osgridref.ParseOsGridRef(gridRefStr)
			if err != nil {
				handleError(w, r, gridRefStr, err)
				return
			}

			lat, lon := gridRef.ToLatLon()
			handle(w, r, gridRef, lat, lon)
		})))

	http.Handle("/"+apiVersion+"/latlon/", middleware.MakeCachingHandler(1*time.Hour, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			latLonStr := r.URL.Path[len("/"+apiVersion+"/latlon/"):]
			latLon, err := osgridref.ParseLatLon(latLonStr, 0, osgridref.WGS84)
			if err != nil {
				handleError(w, r, latLonStr, err)
				return
			}

			gridRef := latLon.ToOsGridRef()
			handle(w, r, gridRef, latLon.Lat, latLon.Lon)
		})))

	if !strings.Contains(listenPort, ":") {
		listenPort = ":" + listenPort
	}

	log.Println("Starting HTTP server on " + listenPort)
	s := &http.Server{
		ReadHeaderTimeout: 20 * time.Second,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       10 * time.Minute,
		Handler:           middleware.MakeLoggingHandler(http.DefaultServeMux),
		Addr:              listenPort,
	}

	return s
}

func md5hash(text string) string {
	hash := md5.New()
	hash.Write([]byte(text))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func handleError(w http.ResponseWriter, _ *http.Request, str string, _ error) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "Invalid request: %q\n", str)
}

func handle(w http.ResponseWriter, r *http.Request, ref osgridref.OsGridRef, lat float64, lon float64) {
	authorization := r.Header.Get("Authorization")

	// extract bearer token from authorization header
	if authorization != "" {
		authorization = strings.TrimPrefix(strings.ToLower(authorization), "bearer ")
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	usageCount := 0

	// if file exists, increment usage count
	if authorization != "" {
		authorization = md5hash(authorization)
		filePath := "keys/" + authorization
		if _, err := os.Stat(filePath); err == nil {
			data, err := os.ReadFile(filePath)
			if err == nil {
				usageCount, err = strconv.Atoi(strings.TrimSpace(string(data)))
				if err == nil {
					usageCount++
					err = os.WriteFile(filePath, []byte(strconv.Itoa(usageCount)), 0644)
					if err != nil {
						log.Printf("Failed to write usage count: %s", err)
					}
				} else {
					log.Printf("Failed to parse usage count: %s", err)
				}
			} else {
				log.Printf("Failed to read file: %s", err)
			}
		} else {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	reply := Reply{
		OSGridRef:  ref.StringNCompact(8),
		Easting:    ref.Easting,
		Northing:   ref.Northing,
		Lat:        lat,
		Lon:        lon,
		UsageCount: usageCount,
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err := enc.Encode(reply)
	if err != nil {
		log.Printf("Failed to write response: %s", err)
		w.WriteHeader(http.StatusBadGateway)
	}
}
