package models

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"io"
	"net"
)

var (
	DELIMITER = []byte(`||`)
)

type Client struct {
	Conn        net.Conn       //The TCP connect
	Outbound    chan<- Command //This channel receive the commands
	Register    chan<- *Client //This channel receive the client that want to join a channel
	Deregister  chan<- *Client //This channel receive the client that want to leave a channel
	macaddr     []string       //Mac address for identify the client
	username    string         // The name of the client
	isReceiving bool
}

func NewClient(conn net.Conn, o chan<- Command, r chan<- *Client, d chan<- *Client) *Client {
	return &Client{
		Conn:       conn,
		Outbound:   o,
		Register:   r,
		Deregister: d,
	}
}

func (c *Client) Read() error {
	for {
		msg := make([]byte, 3)
		_, err := c.Conn.Read(msg)

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
		if err := c.suscribeToChannel(); err != nil {
			c.err(err)
		}
	case "UNS":
		// if err := c.unsuscribeFromChannel(); err != nil {
		// 	c.err(err)
		// }
	case "SND":
		if err := c.sendFile(); err != nil {
			c.err(err)
		}
	case "LCH":
		c.listChannels()
	default:
		c.err(fmt.Errorf("unknown command %s", cmd))
	}
}

func (c *Client) registerClient() error {
	args := make([]byte, 11)

	_, err := c.Conn.Read(args)

	if err != nil {
		return fmt.Errorf("error en recibir datos")
	}

	clientName := bytes.TrimSpace(args)

	if clientName[0] != '@' {
		return fmt.Errorf("username must begin with @")
	}
	if len(clientName[1:]) == 0 {
		return fmt.Errorf("username cannot be blank")
	}
	c.username = strings.Trim(string(clientName), ":")

	c.Register <- c

	return nil
}

func (c *Client) err(e error) {
	c.Conn.Write([]byte("ERR " + e.Error() + "\n"))
}

func (c *Client) suscribeToChannel() error {
	isRegistered := c.isNamed()

	if !isRegistered {
		return fmt.Errorf("Client is no registered, use REGISTER command for sign up")
	}

	args := make([]byte, 11)

	_, err := c.Conn.Read(args)

	if err != nil {
		return fmt.Errorf("ERR fail reading the data")
	}

	channelID := bytes.TrimSpace(args)

	if channelID[0] != '#' {
		return fmt.Errorf("ERR Channel ID must begin with #")
	}

	channelName := strings.Trim(string(channelID), ":")

	c.Outbound <- Command{
		channel: channelName,
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

func (c *Client) sendFile() error {
	isRegistered := c.isNamed()

	if !isRegistered {
		return fmt.Errorf("Client is no registered, use REGISTER command for sign up")
	}

	args := make([]byte, 11)

	if _, err := c.Conn.Read(args); err != nil {
		return fmt.Errorf("ERR fail reading the data")
	}

	args = bytes.TrimSpace(args)

	if args[0] != '#' {
		return fmt.Errorf("recipient must be a channel ('#name')")
	}

	recipient := strings.Trim(string(args), ":")

	if len(recipient[1:]) == 0 {
		return fmt.Errorf("channel must have a name")
	}

	metad := make(map[string][]byte)
	bufferFileSize := make([]byte, 10)

	if _, err := c.Conn.Read(bufferFileSize); err != nil {
		return fmt.Errorf("ERR fail reading the data")
	}

	bufferFileName := make([]byte, 64)

	if _, err := c.Conn.Read(bufferFileName); err != nil {
		return fmt.Errorf("ERR fail reading the data")
	}

	datachn := make(chan []byte, 1024)

	metad["fileSize"] = bufferFileSize
	metad["fileName"] = bufferFileName

	c.Outbound <- Command{
		channel:  string(recipient),
		sender:   c.username,
		metadata: metad,
		body:     datachn,
		id:       SEND,
	}

	const BUFFERSIZE = 1024
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)
	var receivedBytes int64

	for {
		if (fileSize - receivedBytes) < BUFFERSIZE {
			data := make([]byte, (fileSize - receivedBytes))
			c.Conn.Read(data)
			datachn <- data
			c.Conn.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
			break
		}
		// io.CopyN(newFile, connection, BUFFERSIZE)
		data := make([]byte, BUFFERSIZE)
		c.Conn.Read(data)
		datachn <- data
		receivedBytes += BUFFERSIZE
	}

	fmt.Println("llegue al final de la func", string(bufferFileName))

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
