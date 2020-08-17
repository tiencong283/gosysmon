package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const (
	interval = 3 // seconds
)

type Payload struct {
	Timestamp int64 // milliseconds
	EventRate int
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	// The websocket connection.
	conn *websocket.Conn
}

type EventRateHooker struct {
	MessageCh chan *Message
	Clients   []*Client
}

func NewEventRateHooker() *EventRateHooker {
	return &EventRateHooker{
		MessageCh: make(chan *Message, MsgChBufSize),
	}
}

func (hooker *EventRateHooker) Start() {
	numOfEvents := 0
	ticker := time.NewTicker(interval * time.Second)
	for {
		select {
		case currentTime := <-ticker.C:
			payload := &Payload{
				EventRate: numOfEvents / interval,
				Timestamp: currentTime.Unix() * 1000,
			}
			for i := 0; i < len(hooker.Clients); {
				if err := hooker.Clients[i].conn.WriteJSON(payload); err != nil {
					hooker.Clients = append(hooker.Clients[:i], hooker.Clients[i+1:]...)
				} else {
					i++
				}
			}
			numOfEvents = 0
		case <-hooker.MessageCh:
			numOfEvents++
		}
	}
}

func (hooker *EventRateHooker) ServeEventRateWs(c *gin.Context) {
	// ignore origin policy
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	hooker.Clients = append(hooker.Clients, &Client{
		conn: conn,
	})
}
