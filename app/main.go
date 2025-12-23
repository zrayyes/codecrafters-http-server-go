package main

import (
	"fmt"
	"net"
	"os"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	response := "HTTP/1.1 200 OK\r\n\r\n"
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		return
	}
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}

}
