package main

import (
	"log"

	"github.com/gorilla/websocket"
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

// FindHandler is a function that takes a string and retun a type Hander and a boolean
type FindHandler func(string) (Handler, bool)

// Client yadayada
type Client struct {
	send         chan Message
	socket       *websocket.Conn
	findHandler  FindHandler
	session      *r.Session
	stopChannels map[int]chan bool
	id           string
	userName     string
}

func (client *Client) Write() {
	for msg := range client.send {
		if err := client.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	client.socket.Close()
}

// NewStopChannel stops a channel after some time for not leaking stuff
func (client *Client) NewStopChannel(stopKey int) chan bool {
	client.StopForKey(stopKey)
	stop := make(chan bool)
	client.stopChannels[stopKey] = stop
	return stop

}

// StopForKey stop a channel from its key
func (client *Client) StopForKey(key int) {
	if ch, found := client.stopChannels[key]; found {
		ch <- true
		delete(client.stopChannels, key)
	}
}

func (client *Client) Read() {
	var message Message
	for {
		if err := client.socket.ReadJSON(&message); err != nil {
			break
		}
		if handler, found := client.findHandler(message.Name); found {
			handler(client, message.Data)
		}
	}
	client.socket.Close()
}

// Close a connection if the boolean changes
func (client *Client) Close() {
	for _, ch := range client.stopChannels {
		ch <- true
	}
	close(client.send)
}

// NewClient returns a pointer which makes a channel to send
func NewClient(socket *websocket.Conn, findHandler FindHandler, session *r.Session) *Client {
	var user User
	user.Name = "anon"
	res, err := r.Table("user").Insert(user).RunWrite(session)
	if err != nil {
		log.Println(err.Error())
	}
	var id string
	if len(res.GeneratedKeys) > 0 {
		id = res.GeneratedKeys[0]
	}
	return &Client{
		send:         make(chan Message),
		socket:       socket,
		findHandler:  findHandler,
		session:      session,
		stopChannels: make(map[int]chan bool),
		id:           id,
		userName:     user.Name,
	}
}
