package models

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

const BUFFERSIZE = 1024

type Channel struct {
	name    string           //Name of the channel
	clients map[*Client]bool //All the clients that are suscribe to the channel
}

func (c *Channel) broadcast(s string, m []byte) {
	// msg := append([]byte(s), ": "...)
	// msg = append(msg, m...)
	// msg = append(msg, '\n')
	filePath := string(m)

	for cl := range c.clients {
		connection := cl.Conn

		if cl.username != s {
			file, err := os.Open(filePath)
			if err != nil {
				fmt.Println(err)
				return
			}
			fileInfo, err := file.Stat()
			if err != nil {
				fmt.Println(err)
				return
			}
			fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
			fileName := fillString(fileInfo.Name(), 64)
			fmt.Println("Sending filename and filesize!")
			connection.Write([]byte(fileSize))
			connection.Write([]byte(fileName))
			sendBuffer := make([]byte, BUFFERSIZE)
			fmt.Println("Start sending file!")
			for {
				_, err = file.Read(sendBuffer)
				if err == io.EOF {
					break
				}
				connection.Write(sendBuffer)
			}
			fmt.Println("File has been sent")
		}
	}

}

func fillString(retunString string, toLength int) string {
	for {
		lengtString := len(retunString)
		if lengtString < toLength {
			retunString = retunString + ":"
			continue
		}
		break
	}
	return retunString
}
