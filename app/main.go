package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

func handleFileReturn(req *Request, res *Response) {
	filePath := strings.TrimPrefix(req.RequestURI, "/files/")
	filePath = filepath.Join("/tmp/", filePath)
	dat, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("File '%s' not found, need to create it\n", filePath)
			res.StatusCode = 404
			res.ReasonPhrase = "Not Found"
		} else {
			fmt.Printf("Error opening file: %v\n", err)
			res.StatusCode = 500
			res.ReasonPhrase = "Internal Server Error"
		}
		return
	}

	res.Headers.Set("Content-Type", "application/octet-stream")
	res.Headers.Set("Content-Length", strconv.Itoa(utf8.RuneCountInString(string(dat))))
	res.Body = string(dat)
}

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

	res := &Response{
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
			res.Headers.Set("Content-Type", "text/plain")
			res.Headers.Set("Content-Length", strconv.Itoa(utf8.RuneCountInString(ua)))
			res.Body = ua
		}

	case strings.HasPrefix(req.RequestURI, "/echo/"):
		value := strings.TrimPrefix(req.RequestURI, "/echo/")
		res.Headers.Set("Content-Type", "text/plain")
		res.Headers.Set("Content-Length", strconv.Itoa(utf8.RuneCountInString(value)))
		res.Body = value

	case strings.HasPrefix(req.RequestURI, "/files/"):
		handleFileReturn(req, res)

	default:
		res.StatusCode = 404
		res.ReasonPhrase = "Not Found"
	}

	_, err = conn.Write([]byte(res.String()))
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
