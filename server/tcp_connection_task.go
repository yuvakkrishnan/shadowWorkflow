package server

import (
	"fmt"
	"net"
)

func HandleConnection(conn net.Conn) {
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err)
			return
		}

		fmt.Println("Received data:", string(buffer[:n]))

		// Echo the data back to the client
		_, err = conn.Write(buffer[:n])
		if err != nil {
			fmt.Println("Error writing:", err)
			return
		}
	}
}
