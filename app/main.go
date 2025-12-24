package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	req, err := ParseRequest(bufio.NewReader(conn))
	if err != nil {
		if err == io.EOF {
			return
		}
		fmt.Println("Error reading from connection: ", err.Error())
		return
	}

	r := Response{
		StatusLine: StatusLine{
			HTTPVersion:  "HTTP/1.1",
			StatusCode:   200,
			ReasonPhrase: "OK",
		},
		Headers: NewHeaders(),
	}

	switch {
	case req.RequestURI == "/":
		// Home - just return 200 OK

	case req.RequestURI == "/user-agent":
		if ua, found := req.Headers.Get("User-Agent"); found {
			r.Headers.Set("Content-Type", "text/plain")
			r.Headers.Set("Content-Length", strconv.Itoa(utf8.RuneCountInString(ua)))
			r.Body = ua
		}

	case strings.HasPrefix(req.RequestURI, "/echo/"):
		value := strings.TrimPrefix(req.RequestURI, "/echo/")
		r.Headers.Set("Content-Type", "text/plain")
		r.Headers.Set("Content-Length", strconv.Itoa(utf8.RuneCountInString(value)))
		r.Body = value

	default:
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
