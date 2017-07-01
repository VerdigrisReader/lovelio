package main

import (
	"app"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/pat"
	"github.com/gorilla/websocket"
	"github.com/urfave/negroni"
	"html/template"
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

// HTTP Request handlers
// Home handler serves the main page
func homeHandler(wr http.ResponseWriter, req *http.Request) {
	// Get userid
	var userId string
	userIdCookie, err := req.Cookie("loveliouid")
	userId = userIdCookie.Value

	// Get boards
	conn := pool.Get()
	defer conn.Close()
	boardNames := app.GetUserBoards(conn, userId)

	type Renderable struct {
		UserId     string
		BoardNames []app.BoardName
	}
	var render Renderable = Renderable{
		userId,
		boardNames,
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(wr, render); err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
	}

}

// This handles the case where a userid (uuid) is passed as the url
// If the uuid exists, it is set as the userid cookie and redirects to the frontpage
func userIdHandler(wr http.ResponseWriter, req *http.Request) {
	userId := req.URL.Query().Get(":userid")
	conn := pool.Get()
	defer conn.Close()
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

func healthHandler(wr http.ResponseWriter, req *http.Request) {
	wr.WriteHeader(200)
	wr.Write([]byte("ok"))
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

// Websocket stuff
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Message struct {
	MessageType string      `json:"type"`
	Body        interface{} `json:"body"`
}

func websocketHandler(wr http.ResponseWriter, req *http.Request) {
	sock, _ := upgrader.Upgrade(wr, req, nil)
	userIdCookie, _ := req.Cookie("loveliouid")
	userId := userIdCookie.Value

	go func() {
		for {
			var msg Message
			err := sock.ReadJSON(&msg)
			if err != nil {
				return
			}
			conn := pool.Get()
			defer conn.Close()
			switch msg.MessageType {
			case "newBoard":
				fmt.Print("New board")
				boardInfo := app.NewBoard(conn, userId)
				var reply = Message{"NewBoard", boardInfo}
				sock.WriteJSON(reply)
			case "getBoardItems":
				body, _ := msg.Body.(map[string]interface{})
				boardId, _ := body["boardId"].(string)
				boardItems := app.GetBoardItems(conn, boardId)
				var reply = Message{"getBoardItems", boardItems}
				sock.WriteJSON(reply)
			case "mutateItem":
				body, _ := msg.Body.(map[string]interface{})
				boardId, _ := body["boardId"].(string)
				itemName, _ := body["itemName"].(string)
				delta, _ := body["delta"].(string)
				if delta == "incr" {
					app.IncrementBoardItem(conn, boardId, itemName)
				} else {
					app.DecrementBoardItem(conn, boardId, itemName)
				}
			}

		}
	}()
}

// Middlewares
// UserIdCookieMiddleware guarantees that the user's request will have their ID attached as a cookie
// secure no?
func UserIdCookieMiddleware(wr http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	// Map containing exceptions that don't require stupid cookie auth
	// like a healthcheck endpoint
	var exceptions = map[string]bool{
		"/health": true,
	}
	_, ok := exceptions[req.URL.Path]
	if ok {
		next(wr, req)
		return
	}
	var userId string
	_, err := req.Cookie("loveliouid")
	if err != nil {
		// New user created
		conn := pool.Get()
		defer conn.Close()
		userId = app.NewUser(conn)
		app.NewBoard(conn, userId)
		newCookie := &http.Cookie{
			Name:   "loveliouid",
			Value:  userId,
			MaxAge: 0,
		}
		http.SetCookie(wr, newCookie)
		req.AddCookie(newCookie)
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

	router.Get("/health", healthHandler)
	router.Get("/forget", forgetHandler)
	router.Get("/ws", websocketHandler)
	router.Get("/{userid:[\\w-]{36}}", userIdHandler)
	router.Get("/", homeHandler)

	// Add middlewares
	n := negroni.Classic()
	n.Use(negroni.HandlerFunc(UserIdCookieMiddleware))
	n.UseHandler(router)

	log.Print("Serving on 127.0.0.1:" + port)
	log.Fatal(http.ListenAndServe(":"+port, n))
}
