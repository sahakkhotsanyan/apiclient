package apiclient

import (
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type webSocketAPIClient struct {
	url        string
	connection *websocket.Conn
	disconnect chan bool
}

type subscribeQuery struct {
	Op string `json:"op"`
	Id string `json:"id"`
	Ch string `json:"ch"`
}

func (w webSocketAPIClient) Connection() error {
	c, _, err := websocket.DefaultDialer.Dial(w.url, nil)
	if err != nil {
		return err
	}
	w.connection = c
	return nil
}

func (w webSocketAPIClient) Disconnect() {
	w.disconnect <- true
	if err := w.connection.Close(); err != nil {
		log.Println(err)
		return
	}
}

func (w webSocketAPIClient) SubscribeToChannel(symbol string) error {
	query := subscribeQuery{
		Op: "sub",
		Id: "our_app",
		Ch: "bbo:" + strings.ReplaceAll(symbol, "_", "/"),
	}
	if err := w.connection.WriteJSON(query); err != nil {
		defer w.Disconnect()
		return err
	}
	return nil
}

func (w webSocketAPIClient) ReadMessagesFromChannel(ch chan<- BestOrderBook) {
	var msg BestOrderBook
	if err := w.connection.ReadJSON(&msg); err != nil {
		log.Println(err)
		return
	}
	ch <- msg
}

func (w webSocketAPIClient) WriteMessagesToChannel() {
	ticker := time.NewTicker(time.Second * 15)
	ping := map[string]string{
		"op": "ping",
	}
	for {
		select {
		case <-ticker.C:
			if err := w.connection.WriteJSON(ping); err != nil {
				log.Println(err)
				return
			}
		case <-w.disconnect:
			return
		}
	}
}

func NewWebSocket(url string) APIClient {
	return &webSocketAPIClient{url: url}
}
