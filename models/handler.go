package models

import (
	"strconv"
	"strings"
)

// The hub is going to coordinate channels and petitions from the clients
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
				h.sendFile(cmd.sender, cmd.channel, cmd.metadata, cmd.body)
			case LCHANNELS:
				h.listChannels(cmd.sender)
			default:
				// Freak out
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

func (h *Hub) joinChannel(userName string, channelName string) {
	if client, ok := h.Clients[userName]; ok {
		if channel, ok := h.Channels[channelName]; ok {
			// Channel exists, join
			channel.clients[client] = true
		} else {
			// Channel doesn't exists, create and join
			ch := newChannel(channelName)
			ch.clients[client] = true
			h.Channels[channelName] = ch
		}
		client.Conn.Write([]byte("OK"))
	}
}

func (h *Hub) leaveChannel(userName string, channelName string) {
	if client, ok := h.Clients[userName]; ok {
		if channel, ok := h.Channels[channelName]; ok {
			delete(channel.clients, client)
			client.Conn.Write([]byte("OK"))
		} else {
			client.Conn.Write([]byte("ERR channel invalid"))
		}
	} else {
		client.Conn.Write([]byte("ERR username invalid"))
	}
}

func (h *Hub) sendFile(name string, channel string, meta map[string][]byte, c chan []byte) {

	if sender, ok := h.Clients[name]; ok {
		if channel[0] == '#' {
			if channel, ok := h.Channels[channel]; ok {
				if _, ok := channel.clients[sender]; ok {
					channel.setReceivingMode(sender.username)
					fileSize, _ := strconv.ParseInt(strings.Trim(string(meta["fileSize"]), ":"), 10, 64)
					channel.broadcast(meta["fileSize"])
					channel.broadcast(meta["fileName"])
					var receivedBytes int64
					const BUFFERSIZE = 1024
					for {
						if (fileSize - receivedBytes) < BUFFERSIZE {
							fileData := <-c
							channel.broadcast(fileData)
							break
						}
						fileData := <-c
						channel.broadcast(fileData)
						receivedBytes += BUFFERSIZE
					}
					sender.Conn.Write([]byte("File sended!"))

				} else {
					sender.Conn.Write([]byte("ERR don't allowed\n"))

				}
			} else {
				sender.Conn.Write([]byte("ERR no such channel\n"))
			}
		} else {
			sender.Conn.Write([]byte("ERR MSG command\n"))
		}

	}
}

func (h *Hub) listChannels(u string) {
	if client, ok := h.Clients[u]; ok {
		var names []string

		if len(h.Channels) == 0 {
			client.Conn.Write([]byte("ERR no channels found"))
		}

		for c := range h.Channels {
			names = append(names, "-"+c+" ")
		}

		resp := strings.Join(names, ", ")

		client.Conn.Write([]byte(resp + "\n"))
	}
}

// for create new channels
func newChannel(c string) (chn *Channel) {

	return &Channel{
		name:    c,
		clients: make(map[*Client]bool),
	}
}
