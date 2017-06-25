package main

import (
	"app"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/pat"
	"github.com/urfave/negroni"
	"log"
	"net/http"
	"os"
	"time"
)

// Redis connection pool
var pool *redis.Pool

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr)
		},
	}
}

// Request handlers
// Home handler serves the main page
func homeHandler(wr http.ResponseWriter, req *http.Request) {
	var userId string
	userIdCookie, err := req.Cookie("loveliouid")
	// User hasn't got cookie set
	if err != nil {
		wr.WriteHeader(http.StatusOK)
		fmt.Fprintf(wr, "You haven't got a cookie")
		_ = userIdCookie
	} else {
		userId = userIdCookie.Value
		wr.WriteHeader(http.StatusOK)
		fmt.Fprintf(wr, "You have an old cookie %s", userId)
	}
}

// This handles the case where a userid (uuid) is passed as the url
// If the uuid exists, it is set as the userid cookie and redirects to the frontpage
func userIdHandler(wr http.ResponseWriter, req *http.Request) {
	userId := req.URL.Query().Get(":userid")
	conn := pool.Get()
	exists, _ := redis.Bool(conn.Do("EXISTS", userId))
	if exists {
		newCookie := &http.Cookie{
			Name:   "loveliouid",
			Value:  userId,
			MaxAge: 0,
		}
		http.SetCookie(wr, newCookie)
	}
	http.Redirect(wr, req, "/", 302)
}

// This deletes the userid cookie and redirects to the homepage
// Mostly for testing etc
func forgetHandler(wr http.ResponseWriter, req *http.Request) {
	newCookie := &http.Cookie{
		Name:   "loveliouid",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(wr, newCookie)
	http.Redirect(wr, req, "/", 302)
}

// Middlewares
// UserIdCookieMiddleware guarantees that the user's request will have their ID attached as a cookie
// secure no?
func UserIdCookieMiddleware(wr http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	var userId string
	_, err := req.Cookie("loveliouid")
	if err != nil {
		// New user created
		conn := pool.Get()
		userId = app.NewUser(conn)
		app.NewBoard(conn, userId)
		newCookie := &http.Cookie{
			Name:   "loveliouid",
			Value:  userId,
			MaxAge: 0,
		}
		http.SetCookie(wr, newCookie)
	}
	next(wr, req)
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

	// Initialise Redis connection pool (globally scoped variable)
	pool = newPool(redisServer)

	// Define router and routings
	router := pat.New()

	router.Get("/forget", forgetHandler)
	router.Get("/{userid:[\\w-]{36}}", userIdHandler)
	router.Get("/", homeHandler)

	// Add middlewares
	n := negroni.Classic()
	n.Use(negroni.HandlerFunc(UserIdCookieMiddleware))
	n.UseHandler(router)

	log.Print("Serving on 127.0.0.1:" + port)
	log.Fatal(http.ListenAndServe(":"+port, n))
}
