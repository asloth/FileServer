package models

import (
	"bytes"
	"fmt"

	"io"
	"net"
)

var (
	DELIMITER = []byte(`||`)
)

type Client struct {
	Conn          net.Conn       //The TCP connect
	Outbound      chan<- Command //This channel receive the commands
	Register      chan<- *Client //This channel receive the client that want to join a channel
	Deregister    chan<- *Client //This channel receive the client that want to leave a channel
	macaddr       []string       //Mac address for identify the client
	username      string         // The name of the client
	isTransfering bool
	isReceiving   bool
}

func NewClient(conn net.Conn, o chan<- Command, r chan<- *Client, d chan<- *Client) *Client {
	return &Client{
		Conn:          conn,
		Outbound:      o,
		Register:      r,
		Deregister:    d,
		isTransfering: false,
	}
}

func (c *Client) Read() error {
	// reader := bufio.NewReader(c.Conn)
	for {
		msg := make([]byte, 3)
		_, err := c.Conn.Read(msg)
		fmt.Println(msg)
		if err == io.EOF {
			// Connection closed, deregister client
			c.Deregister <- c
			return nil
		}
		if err != nil {
			return err
		}

		c.Handle(msg)
	}

}

func (c *Client) Handle(message []byte) {
	//Taking the command from the received message
	cmd := bytes.ToUpper(bytes.TrimSpace(message))

	//Take the arguments of the command
	// args := bytes.TrimSpace(bytes.TrimPrefix(message, cmd))

	//Identifying the command
	switch string(cmd) {
	case "REG":
		if err := c.registerClient(); err != nil {
			c.err(err)
		}
	case "SUS":
		// if err := c.suscribeToChannel(); err != nil {
		// 	c.err(err)
		// }
	case "UNS":
		// if err := c.unsuscribeFromChannel(); err != nil {
		// 	c.err(err)
		// }
	case "SND":
		// if err := c.sendFile(); err != nil {
		// 	c.err(err)
		// }
	case "LCH":
		c.listChannels()
	default:
		c.err(fmt.Errorf("Unknown command %s", cmd))
	}
}

func (c *Client) registerClient() error {
	args := make([]byte, 7)
	_, err := c.Conn.Read(args)
	if err != nil {
		return fmt.Errorf("Error en recibir datos")
	}

	u := bytes.TrimSpace(args)
	if u[0] != '@' {
		return fmt.Errorf("Username must begin with @")
	}
	if len(u) == 0 {
		return fmt.Errorf("Username cannot be blank")
	}
	fmt.Println(string(args))
	fmt.Println("im u ", string(u))
	c.username = string(u)

	c.Register <- c

	return nil
}

func (c *Client) err(e error) {
	c.Conn.Write([]byte("ERR " + e.Error() + "\n"))
}

func (c *Client) suscribeToChannel(args []byte) error {
	isRegistered := c.isNamed()

	if !isRegistered {
		return fmt.Errorf("Client is no registered, use REGISTER command for sign up")
	}

	channelID := bytes.TrimSpace(args)
	if channelID[0] != '#' {
		return fmt.Errorf("ERR Channel ID must begin with #")
	}

	c.Outbound <- Command{
		channel: string(channelID),
		sender:  c.username,
		id:      SUSCRIBE,
	}
	return nil
}

func (c *Client) unsuscribeFromChannel(args []byte) error {
	channelID := bytes.TrimSpace(args)
	if channelID[0] == '#' {
		return fmt.Errorf("ERR channelID must start with '#'")
	}

	c.Outbound <- Command{
		channel: string(channelID),
		sender:  c.username,
		id:      UNSUSCRIBE,
	}
	return nil
}

func (c *Client) sendFile(args []byte) error {
	isRegistered := c.isNamed()

	if !isRegistered {
		return fmt.Errorf("Client is no registered, use REGISTER command for sign up")
	}

	args = bytes.TrimSpace(args)

	if args[0] != '#' {
		return fmt.Errorf("recipient must be a channel ('#name')")
	}

	recipient := bytes.Split(args, []byte(" "))[0]
	if len(recipient) == 0 {
		return fmt.Errorf("channel must have a name")
	}

	args = bytes.TrimSpace(bytes.TrimPrefix(args, recipient))

	c.Outbound <- Command{
		channel: string(recipient),
		sender:  c.username,
		body:    args,
		id:      SEND,
	}

	return nil
}

func (c *Client) listChannels() {
	c.Outbound <- Command{
		sender: c.username,
		id:     LCHANNELS,
	}
}

func (c *Client) isNamed() bool {
	return !(len(c.username) == 0)
}
