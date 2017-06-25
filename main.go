package main

import (
	"github.com/gorilla/pat"
	"log"
	"net/http"
	"os"
)

func homeHandler(wr http.ResponseWriter, req *http.Request) {
	wr.WriteHeader(http.StatusOK)
	wr.Write([]byte("Just a boring old route"))
}

func main() {
	// Get variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	redisServer := os.Getenv("REDIS")
	if redisServer == "" {
		redisServer = "127.0.0.1:6379"
	}

	// Define router and routings
	router := pat.New()

	router.Get("/", homeHandler)
	http.Handle("/", router)

	log.Print("Serving on 127.0.0.1:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
