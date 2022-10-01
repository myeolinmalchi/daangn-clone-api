package chat

type Chatroom struct {
	ID      int
	Send    chan []byte
	Clients map[*Client]bool
	Hub     *ChatHub
}

func (r *Chatroom) open() {
	for {
		select {
		case message := <-r.Send:
			for client := range r.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(r.Clients, client)
				}
			}
		}
	}
}
