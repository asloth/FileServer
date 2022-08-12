package models

import (
	"bufio"
	"fmt"
	"io"
)

type Channel struct {
	name    string           //Name of the channel
	clients map[*Client]bool //All the clients that are suscribe to the channel
}

func (c *Channel) broadcast(body []byte) {
	for cl := range c.clients {
		connection := cl.Conn
		if cl.isReceiving {
			fmt.Println("broadcastmethod")
			connection.Write(body)
		}
	}
}

func (c *Channel) setReceivingMode(s string) {
	for cl := range c.clients {
		connection := cl.Conn
		if cl.username != s {
			cl.isReceiving = true
			connection.Write([]byte("RC"))
		}
	}
}

func ScannerToReader(scanner *bufio.Scanner) io.Reader {
	reader, writer := io.Pipe()

	go func() {
		defer writer.Close()
		for scanner.Scan() {
			writer.Write(scanner.Bytes())
		}
	}()

	return reader
}
