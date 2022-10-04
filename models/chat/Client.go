package chat

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	UserID   string
	Chatroom *Chatroom
	Send     chan Chat
	Conn     *websocket.Conn
}

type Chat struct {
	Message string `json:"message"`
	UserID  string `json:"userId"`
}

func (c *Client) ReadPump() {
	defer func() {
		c.Chatroom.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		chat := Chat{}
		err := c.Conn.ReadJSON(&chat)
		if err != nil {
			fmt.Println(err)
			break
		}
		err = c.Chatroom.Hub.ChatService.InsertChat(c.Chatroom.ID, chat.UserID, chat.Message)
		if err != nil {
			fmt.Println(err)
			break
		}

		c.Chatroom.Send <- chat
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			jsonString, err := json.Marshal(message)
			if err != nil {
				return
			}
			w.Write(jsonString)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				msg, err := json.Marshal(<-c.Send)
				if err != nil {
					return
				}
				w.Write(msg)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
