package main

import (
	"fmt"
	"log"
	"net"

	"github.com/asloth/fileserver/models"
)

func main() {
	ln, err := net.Listen("tcp", ":8081")

	if err != nil {
		log.Printf("%v", err)
	}

	hub := models.NewHub()
	go hub.Run()
	fmt.Print("Server listening in te port 8081 \n")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("%v", err)
		}

		c := models.NewClient(
			conn,
			hub.Commands,
			hub.Registrations,
			hub.Deregistrations,
		)

		go c.Read()
	}

}
