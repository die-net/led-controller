package ws

// Router maintains the set of active clients and sends and receives
// messages to/from the clients.
type Router struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	Incoming chan []byte

	// Outgoing messages to the clients.
	Outgoing chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewRouter() *Router {
	return &Router{
		Incoming:   make(chan []byte),
		Outgoing:   make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (r *Router) Worker() {
	for {
		select {
		case client := <-r.register:
			r.clients[client] = true
		case client := <-r.unregister:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
			}
		case message := <-r.Outgoing:
			for client := range r.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(r.clients, client)
				}
			}
		}
	}
}
