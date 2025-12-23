package main

import (
	"fmt"
	"net"
	"os"
)

// The first line of a Response message is the Status-Line,
// consisting of the protocol version followed by a numeric status code and its associated textual phrase,
// with each element separated by SP characters.
// No CR or LF is allowed except in the final CRLF sequence.
// Status-Line = HTTP-Version SP Status-Code SP Reason-Phrase CRLF
type StatusLine struct {
	HTTPVersion  string
	StatusCode   int
	ReasonPhrase string
}

type Response struct {
	StatusLine
}

func (r Response) String() string {
	return fmt.Sprintf("%s %d %s\r\n\r\n", r.HTTPVersion, r.StatusCode, r.ReasonPhrase)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	r := Response{
		StatusLine: StatusLine{
			HTTPVersion:  "HTTP/1.1",
			StatusCode:   200,
			ReasonPhrase: "OK",
		},
	}

	_, err := conn.Write([]byte(r.String()))
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
