package models

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"

	"io"
	"net"
)

var (
	DELIMITER = []byte(`||`)
)

type Client struct {
	Conn       net.Conn       //The TCP connect
	Outbound   chan<- Command //This channel receive the commands
	Register   chan<- *Client //This channel receive the client that want to join a channel
	Deregister chan<- *Client //This channel receive the client that want to leave a channel
	Macadddr   []string       //Mac address for identify the client
	username   string         // The name of the client
}

func NewClient(conn net.Conn, o chan<- Command, r chan<- *Client, d chan<- *Client, mac []string) *Client {
	return &Client{
		Conn:       conn,
		Outbound:   o,
		Register:   r,
		Deregister: d,
		Macadddr:   mac,
	}
}

func (c *Client) Read() error {
	for {
		msg, err := bufio.NewReader(c.Conn).ReadBytes('\n')
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
	cmd := bytes.ToUpper(bytes.TrimSpace(bytes.Split(message, []byte(" "))[0]))

	//Take the arguments of the command
	args := bytes.TrimSpace(bytes.TrimPrefix(message, cmd))

	//Identifying the command
	switch string(cmd) {
	case "REGISTER":
		if err := c.registerClient(args); err != nil {
			c.err(err)
		}
	case "SUSCRIBE":
		if err := c.suscribeToChannel(args); err != nil {
			c.err(err)
		}
	case "UNSUSCRIBE":
		if err := c.unsuscribeFromChannel(args); err != nil {
			c.err(err)
		}
	case "SEND":
		if err := c.sendFile(args); err != nil {
			c.err(err)
		}
	case "LCHANNELS":
		c.listChannels()
	default:
		c.err(fmt.Errorf("Unknown command %s", cmd))
	}
}

func (c *Client) registerClient(args []byte) error {
	u := bytes.TrimSpace(args)
	if u[0] != '@' {
		fmt.Print(args)
		fmt.Print(u[0])
		return fmt.Errorf("Username must begin with @")
	}
	if len(u) == 0 {
		return fmt.Errorf("Username cannot be blank")
	}

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

	l := bytes.Split(args, DELIMITER)[0]

	length, err := strconv.Atoi(string(l))

	if err != nil {
		return fmt.Errorf("body length must be present")

	}
	if length == 0 {
		return fmt.Errorf("body length must be at least 1")
	}

	padding := len(l) + len(DELIMITER) // Size of the body length + the delimiter
	body := args[padding : padding+length]

	c.Outbound <- Command{
		channel: string(recipient),
		sender:  c.username,
		body:    body,
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
	if len(c.username) == 0 {
		return false
	}
	return true
}
