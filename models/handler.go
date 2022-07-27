package models

import "strings"

//The hub is going to coordinate channels and petitions from the clients
type Hub struct {
	Channels        map[string]*Channel // The list of channels
	Clients         map[string]*Client  // The list of clients registrate in the server
	Commands        chan Command        // The list of commands
	Deregistrations chan *Client        // The channel for deregistrations from channels
	Registrations   chan *Client        // The channel for registrations in the channels
}

func NewHub() *Hub {
	return &Hub{
		Registrations:   make(chan *Client),
		Deregistrations: make(chan *Client),
		Clients:         make(map[string]*Client),
		Channels:        make(map[string]*Channel),
		Commands:        make(chan Command),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Registrations:
			h.register(client)
		case client := <-h.Deregistrations:
			h.deregister(client)
		case cmd := <-h.Commands:
			switch cmd.id {
			case SUSCRIBE:
				h.joinChannel(cmd.sender, cmd.channel)
			case UNSUSCRIBE:
				h.leaveChannel(cmd.sender, cmd.channel)
			case SEND:
				h.sendFile(cmd.sender, cmd.channel, cmd.body)
			case LCHANNELS:
				h.listChannels(cmd.sender)
			default:
				// Freak out?
			}
		}
	}
}

func (h *Hub) register(c *Client) {
	if _, exists := h.Clients[c.username]; exists {
		c.username = ""
		c.Conn.Write([]byte("ERR username taken\n"))
	} else {
		h.Clients[c.username] = c
		c.Conn.Write([]byte("OK\n"))
	}
}

func (h *Hub) deregister(c *Client) {
	if _, exists := h.Clients[c.username]; exists {
		delete(h.Clients, c.username)

		for _, channel := range h.Channels {
			delete(channel.clients, c)
		}
	}
}

func (h *Hub) joinChannel(u string, c string) {
	if client, ok := h.Clients[u]; ok {
		if channel, ok := h.Channels[c]; ok {
			// Channel exists, join
			channel.clients[client] = true
		} else {
			// Channel doesn't exists, create and join
			ch := newChannel(c)
			ch.clients[client] = true
			h.Channels[c] = ch
		}
		client.Conn.Write([]byte("OK\n"))
	}
}

func (h *Hub) leaveChannel(u string, c string) {
	if client, ok := h.Clients[u]; ok {
		if channel, ok := h.Channels[c]; ok {
			delete(channel.clients, client)
		}
	}
}

func (h *Hub) sendFile(u string, r string, body []byte) {
	if sender, ok := h.Clients[u]; ok {
		switch r[0] {
		case '#':
			if channel, ok := h.Channels[r]; ok {
				if _, ok := channel.clients[sender]; ok {
					channel.broadcast(sender.username, body)
				}
			} else {
				sender.Conn.Write([]byte("ERR no such channel"))
			}
		default:
			sender.Conn.Write([]byte("ERR MSG command"))
		}
	}
}

func (h *Hub) listChannels(u string) {
	if client, ok := h.Clients[u]; ok {
		var names []string

		if len(h.Channels) == 0 {
			client.Conn.Write([]byte("ERR no channels found\n"))
		}

		for c := range h.Channels {
			names = append(names, "-"+c+" ")
		}

		resp := strings.Join(names, ", ")

		client.Conn.Write([]byte(resp + "\n"))
	}
}

//for create new channels
func newChannel(c string) (chn *Channel) {

	return &Channel{
		name:    c,
		clients: make(map[*Client]bool),
	}
}
