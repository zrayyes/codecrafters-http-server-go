package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

var FILE_DIRECTORY = "/tmp/"

func homeHandler(req *Request) *Response {
	return NewResponse()
}

func echoHandler(req *Request) *Response {
	res := NewResponse()
	value := strings.TrimPrefix(req.RequestURI, "/echo/")
	res.Headers.Set("Content-Type", "text/plain")
	res.Headers.Set("Content-Length", strconv.Itoa(utf8.RuneCountInString(value)))
	res.Body = value
	return res
}

func userAgentHandler(req *Request) *Response {
	res := NewResponse()
	if ua, found := req.Headers.Get("User-Agent"); found {
		res.Headers.Set("Content-Type", "text/plain")
		res.Headers.Set("Content-Length", strconv.Itoa(utf8.RuneCountInString(ua)))
		res.Body = ua
	}
	return res
}

func fileReturnHandler(req *Request) *Response {
	res := NewResponse()

	filePath := strings.TrimPrefix(req.RequestURI, "/files/")
	filePath = filepath.Join(FILE_DIRECTORY, filePath)
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
		return res
	}

	res.Headers.Set("Content-Type", "application/octet-stream")
	res.Headers.Set("Content-Length", strconv.Itoa(utf8.RuneCountInString(string(dat))))
	res.Body = string(dat)

	return res
}

func fileCreateHandler(req *Request) *Response {
	res := NewResponse()

	filePath := strings.TrimPrefix(req.RequestURI, "/files/")
	filePath = filepath.Join(FILE_DIRECTORY, filePath)

	err := os.WriteFile(filePath, []byte(req.Body), 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		res.StatusCode = 500
		res.ReasonPhrase = "Internal Server Error"
		return res
	}

	res.StatusCode = 201
	res.ReasonPhrase = "Created"
	return res
}

func fileHandler(req *Request) *Response {
	if req.Method == "POST" {
		return fileCreateHandler(req)
	}
	return fileReturnHandler(req)
}

func handleConnection(conn net.Conn, router *Router) {
	defer conn.Close()

	req, err := ParseRequest(bufio.NewReader(conn))
	if err != nil {
		if err == io.EOF {
			return
		}
		fmt.Println("Error reading from connection: ", err.Error())
		return
	}

	res := router.Route(req)

	_, err = conn.Write([]byte(res.String()))
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		return
	}
}

func main() {
	directory := flag.String("directory", "/tmp/", "Specifies the directory where the files are stored, as an absolute path.")

	flag.Parse()

	FILE_DIRECTORY = *directory

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	router := &Router{}

	router.HandleExact("/", homeHandler)
	router.HandleExact("/user-agent", userAgentHandler)
	router.HandlePrefix("/echo/", echoHandler)
	router.HandlePrefix("/files/", fileHandler)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn, router)
	}
}
