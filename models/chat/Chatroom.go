package chat

type Chatroom struct {
	ChatroomID int
	Clients    map[*Client]bool
	Send       chan Chat
}

func (c *Chatroom) Open() {
	for {
		select {
		case message, _ := <-c.Send:
			for client := range c.Clients {
				client.Send <- message
			}
		}
	}
}
