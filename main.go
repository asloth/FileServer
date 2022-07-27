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
		//Accept connection
		conn, err := ln.Accept()

		//Verify errors
		if err != nil {
			log.Printf("%v", err)
		}

		//Obtaining the mac address of the client for identify it
		macadd, err := getMacAddr()

		if err != nil {
			log.Printf("%v", err)
		}

		//Creating a new client
		c := models.NewClient(
			conn,                //Set the conection
			hub.Commands,        //Channel of the hub for gettin the client's commands
			hub.Registrations,   //Channel of the hub fot getting the registrations
			hub.Deregistrations, //Channel of the hub fot getting the registrations
			macadd,              //setting the mac addres
		)
		//Client starts reading
		go c.Read()
	}

}

func getMacAddr() ([]string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var as []string
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			as = append(as, a)
		}
	}
	return as, nil
}
