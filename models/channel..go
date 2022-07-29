package models

type Channel struct {
	name    string           //Name of the channel
	clients map[*Client]bool //All the clients that are suscribe to the channel
}

func (c *Channel) broadcast(s string, m []byte) {
	// msg := append([]byte(s), ": "...)
	// msg = append(msg, m...)
	// msg = append(msg, '\n')

	for cl := range c.clients {
		connection := cl.Conn

		if cl.username != s {
			connection.Write(m)
		}
	}

}
