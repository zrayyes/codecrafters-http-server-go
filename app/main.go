package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// The Request-Line begins with a method token,
// followed by the Request-URI and the protocol version, and ending with CRLF.
// The elements are separated by SP characters.
// No CR or LF is allowed except in the final CRLF sequence.
// Request-Line = Method SP Request-URI SP HTTP-Version CRLF
type RequestLine struct {
	Method      string // TODO: Use consts later
	RequestURI  string
	HTTPVersion string
}

type Request struct {
	RequestLine
}

func ParseRequest(reader *bufio.Reader) (*Request, error) {
	out, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	parts := strings.Split(out, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line")
	}

	req := &Request{
		RequestLine: RequestLine{
			Method:      parts[0],
			RequestURI:  parts[1],
			HTTPVersion: parts[2],
		},
	}

	return req, nil
}

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

	req, err := ParseRequest(bufio.NewReader(conn))
	if err != nil {
		fmt.Println("Error reading from connection: ", err.Error())
		return
	}

	r := Response{
		StatusLine: StatusLine{
			HTTPVersion:  "HTTP/1.1",
			StatusCode:   200,
			ReasonPhrase: "OK",
		},
	}

	if req.RequestURI != "/" {
		r.StatusCode = 404
		r.ReasonPhrase = "Not Found"
	}

	_, err = conn.Write([]byte(r.String()))
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
