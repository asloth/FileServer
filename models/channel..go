package models

type Channel struct {
	name    string           //Name of the channel
	clients map[*Client]bool //All the clients that are suscribe to the channel
}

func (c *Channel) broadcast(s string, m []byte) {
	msg := append([]byte(s), ": "...)
	msg = append(msg, m...)
	msg = append(msg, '\n')

	for cl := range c.clients {

		if cl.username != s {
			cl.Conn.Write(msg)
		}
	}

	/* buf := make([]byte, 1024)
	// for {
	//     n, err := c.Conn.Read(buf)
	//     if err != nil {
	//         if err != io.EOF {
	//             log.Println(err)
	//         }

	//         return
	//     }
	//     log.Printf("received: %q", buf[:n])
	//     log.Printf("bytes: %d", n)

	 }*/
}
