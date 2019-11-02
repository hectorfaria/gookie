package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

// Handler is a function that is basically a constructor
type Handler func(*Client, interface{})

// Router has an array of string called rules
type Router struct {
	rules   map[string]Handler
	session *r.Session
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// NewRouter Basically a constructor that set initial parameters
func NewRouter(session *r.Session) *Router {
	return &Router{
		rules:   make(map[string]Handler),
		session: session,
	}
}

// Handle recieves msgname and a handler type
func (r *Router) Handle(msgName string, handler Handler) {
	r.rules[msgName] = handler
}

// FindHandler recieves a string and return a type handler and boolean
func (r *Router) FindHandler(msgName string) (Handler, bool) {
	handler, found := r.rules[msgName]
	return handler, found
}

// ServeHTTP recieves a res,rep and turn it to the client
func (r *Router) ServeHTTP(w http.ResponseWriter, h *http.Request) {
	socket, err := upgrader.Upgrade(w, h, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	client := NewClient(socket, r.FindHandler, r.session)
	defer client.Close()
	go client.Write()
	client.Read()
}
