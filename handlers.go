package main

import (
	"time"

	"github.com/mitchellh/mapstructure"
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

// ChannelStop is iota, iota means start at 0, and other variables are + 1
const (
	ChannelStop = iota
	UserStop
	MessageStop
)

// Message Name and Data
type Message struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

// Channel has an ID and a Name
type Channel struct {
	ID   string `json:"id" gorethink:"id,omitempty"`
	Name string `json:"name" gorethink:"name" `
}

// User has an ID and Name
type User struct {
	ID   string `gorethink:"id,omitempty"`
	Name string `gorethink:"name"`
}

// ChannelMessage has an ID, ChannelID body author and time created
type ChannelMessage struct {
	ID        string    `gorethink:"id,omitempty"`
	ChannelID string    `gorethink:"channelId"`
	Body      string    `gorethink:"body"`
	Author    string    `gorethink:"author"`
	CreatedAt time.Time `gorethink:"createdAt"`
}

func addChannel(client *Client, data interface{}) {
	var channel Channel
	err := mapstructure.Decode(data, &channel)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}
	go func() {
		err = r.Table("channel").Insert(channel).Exec(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
		}
	}()

}

func subscribeChannel(client *Client, data interface{}) {
	go func() {
		stop := client.NewStopChannel(ChannelStop)
		cursor, err := r.Table("channel").Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
			return
		}
		changeFeedHelper(cursor, "channel", client.send, stop)
	}()
}

func unsubscribeChannel(client *Client, data interface{}) {
	client.StopForKey(ChannelStop)
}

func editUser(client *Client, data interface{}) {
	var user User
	err := mapstructure.Decode(data, &user)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}
	client.userName = user.Name
	go func() {
		_, err := r.Table("user").Get(client.id).Update(user).RunWrite(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
		}

	}()
}

func subscribeUser(client *Client, data interface{}) {
	go func() {
		stop := client.NewStopChannel(UserStop)
		cursor, err := r.Table("user").Changes(r.ChangesOpts{IncludeInitial: true}).Run(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
			return
		}
		changeFeedHelper(cursor, "user", client.send, stop)
	}()
}

func unsubscribeUser(client *Client, data interface{}) {
	client.StopForKey(UserStop)
}

func addMessage(client *Client, data interface{}) {
	var channelMessage ChannelMessage
	err := mapstructure.Decode(data, &channelMessage)
	if err != nil {
		client.send <- Message{"error", err.Error()}
	}
	go func() {
		channelMessage.CreatedAt = time.Now()
		channelMessage.Author = client.userName
		err := r.Table("message").Insert(channelMessage).Exec(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
		}
	}()
}

func subscribeMessage(client *Client, data interface{}) {
	go func() {
		eventData := data.(map[string]interface{})
		val, ok := eventData["channelId"]
		if !ok {
			return
		}
		channelID, ok := val.(string)
		if !ok {
			return
		}
		stop := client.NewStopChannel(MessageStop)
		cursor, err := r.Table("message").OrderBy(r.OrderByOpts{Index: r.Desc("createdAt")}).
			Filter(r.Row.Field("channelId").Eq(channelID)).Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
		}
		changeFeedHelper(cursor, "message", client.send, stop)
	}()
}

func unsubscribeMessage(client *Client, data interface{}) {
	client.StopForKey(MessageStop)
}

func changeFeedHelper(cursor *r.Cursor, changeEventName string, send chan<- Message, stop <-chan bool) {
	change := make(chan r.ChangeResponse)
	cursor.Listen(change)
	for {
		eventName := ""
		var data interface{}
		select {
		case <-stop:
			cursor.Close()
			return
		case val := <-change:
			if val.NewValue != nil && val.OldValue == nil {
				eventName = changeEventName + "add"
				data = val.NewValue
			} else if val.NewValue == nil && val.OldValue != nil {
				eventName = changeEventName + "remove"
				data = val.OldValue
			} else if val.NewValue != nil && val.OldValue != nil {
				eventName = changeEventName + "edit"
				data = val.OldValue
			}
			send <- Message{eventName, data}
		}
	}
}
