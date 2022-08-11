package models

// A enum is defined with all the commands the server accept
type ID int

const (
	REGISTER   ID = iota + 1 //Command for registering a client in the server
	SUSCRIBE                 // Command for suscribe to a channel
	UNSUSCRIBE               // Command for unsuscribe to a channel
	SEND                     // Command for send a file to a channel
	LCHANNELS                //Command for listing all the channels
)

// A class command is defined and its params as class attributes
type Command struct {
	id       ID                // The command
	sender   string            // The id of the sender that is sending the file
	channel  string            // The name of the channel to send the file to
	metadata map[string][]byte //Data about the body
	body     chan []byte       // The file in bytes
}
