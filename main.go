package main

import (
	"log"
	"net/http"

	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

func main() {
	session, err := r.Connect(r.ConnectOpts{
		Address:  "localhost:28015",
		Database: "gookie",
	})
	if err != nil {
		log.Panic(err.Error())
	}

	router := NewRouter(session)
	router.Handle("channel add", addChannel)
	router.Handle("channel subscribe", subscribeChannel)
	router.Handle("channel unsubcribe", unsubscribeChannel)

	router.Handle("user edit", editUser)
	router.Handle("user subscribe", subscribeUser)
	router.Handle("user unsubscribe", unsubscribeUser)

	router.Handle("message add", addMessage)
	router.Handle("message subscribe", subscribeMessage)
	router.Handle("message unsubscribe", unsubscribeMessage)

	http.Handle("/", router)
	http.ListenAndServe(":4000", nil)
}
