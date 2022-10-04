package chat

import (
	"carrot-market-clone-api/services"
	"fmt"
)

type ChatHub struct {
	ChatService services.ChatService
	Chatrooms   map[int]*Chatroom
	Register    chan *Client
	Unregister  chan *Client
}

func NewChatHub(chatService services.ChatService) ChatHub {
	return ChatHub{
		ChatService: chatService,
		Chatrooms:   make(map[int]*Chatroom),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
	}
}

func (h *ChatHub) Run() {
	for {
		select {
		case client := <-h.Register:
			userId := client.UserID
			chatroomId := client.Chatroom.ID
			// 올바른 채팅방 사용자인지 체크
			if h.ChatService.CheckCorrectUser(userId, chatroomId) {
				go client.ReadPump()
				go client.WritePump()
				if _, ok := h.Chatrooms[chatroomId]; ok {
					// 채팅방이 존재하는 경우
					client.Chatroom = h.Chatrooms[chatroomId]
					client.Chatroom.Clients[client] = true
				} else {
					// 채팅방이 존재하지 않는 경우
					h.Chatrooms[chatroomId] = client.Chatroom
					client.Chatroom.Clients[client] = true
					go client.Chatroom.open()
				}
			} else {
				fmt.Println("올바르지 않은 사용자")
			}
		case client := <-h.Unregister:
			chatroomId := client.Chatroom.ID
			if len(client.Chatroom.Clients) <= 1 {
				delete(h.Chatrooms, chatroomId)
				close(client.Send)
			} else {
				delete(client.Chatroom.Clients, client)
				close(client.Send)
			}
		}
	}
}
