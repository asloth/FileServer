package models

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type Channel struct {
	name    string           //Name of the channel
	clients map[*Client]bool //All the clients that are suscribe to the channel
}

func (c *Channel) broadcastFile(senderName string, sender net.Conn) error {
	const BUFFERSIZE = 1024
	fmt.Println("flag8")
	scanner := bufio.NewScanner(sender)
	name := senderName
	fmt.Println("flag8.2")

	bufferFileSize := make([]byte, 10)
	scanner.Buffer(bufferFileSize, BUFFERSIZE)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)
	c.broadcast(name, bufferFileSize)

	fmt.Println("flag9", fileSize)

	bufferFileName := make([]byte, 64)
	scanner.Buffer(bufferFileName, BUFFERSIZE)
	c.broadcast(name, bufferFileName)

	fmt.Println("flag9" + string(bufferFileName))

	// var receivedBytes int64

	// for {
	// 	for c1 := range c.clients {
	// 		if c1.isReceiving {
	// 			reader := ScannerToReader(scanner)
	// 			if (fileSize - receivedBytes) < BUFFERSIZE {
	// 				io.CopyN(c1.Conn, reader, (fileSize - receivedBytes))
	// 				break
	// 			}
	// 			io.CopyN(c1.Conn, reader, BUFFERSIZE)
	// 			receivedBytes += BUFFERSIZE
	// 		}
	// 	}
	// }
	return nil
}

func (c *Channel) broadcast(s string, body []byte) {
	for cl := range c.clients {
		connection := cl.Conn
		if cl.username != s {
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
			connection.Write([]byte("REC"))
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
