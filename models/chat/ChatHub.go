package chat

import (
	"carrot-market-clone-api/services"
)

type ChatHub struct {
	ChatService services.ChatService
	Chatrooms   map[int]*Chatroom
	Clients     map[string]*Client
	Register    chan *Client
	Unregister  chan *Client
}

func NewChatHub(chatService services.ChatService) ChatHub {
	return ChatHub{
		ChatService: chatService,
		Chatrooms:   make(map[int]*Chatroom),
		Clients:     make(map[string]*Client),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
	}
}

func (h *ChatHub) Run() {
	for {
		select {
		case client := <-h.Register:
			userId := client.UserID
			if _, ok := h.Clients[userId]; !ok {
				h.Clients[userId] = client

				chatrooms, _, err := h.ChatService.GetChatrooms(userId, nil, nil)
				if err != nil {
					continue
				}

				for _, chatroom := range chatrooms {
					chatroomId := chatroom.ID
					if c, ok := h.Chatrooms[chatroomId]; ok {
						// 채팅방이 존재하는 경우
						client.Chatrooms[chatroom.ID] = c
					} else {
						// 채팅방이 존재하지 않는 경우
						c := &Chatroom{
							ChatroomID: chatroomId,
							Clients: map[*Client]bool{
								client: true,
							},
							Send: make(chan Chat),
						}
						client.Chatrooms[chatroom.ID] = c
						h.Chatrooms[chatroom.ID] = c

						go c.Open()
					}
				}

				go client.ReadPump()
				go client.WritePump()
			}
		case client := <-h.Unregister:
			for _, chatroom := range client.Chatrooms {
				if len(chatroom.Clients) <= 1 {
					delete(h.Clients, client.UserID)
					delete(h.Chatrooms, chatroom.ChatroomID)
					close(chatroom.Send)
					close(client.Send)
				} else {
					delete(h.Clients, client.UserID)
					delete(chatroom.Clients, client)
					close(client.Send)
				}
			}
		}
	}
}
