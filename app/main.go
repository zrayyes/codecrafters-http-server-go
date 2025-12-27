package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

var FILE_DIRECTORY = "/tmp/"

func homeHandler(req *Request, res *Response) {
}

func echoHandler(req *Request, res *Response) {
	value := strings.TrimPrefix(req.RequestURI, "/echo/")
	res.Headers.Set("Content-Type", "text/plain")
	res.Headers.Set("Content-Length", strconv.Itoa(utf8.RuneCountInString(value)))
	res.Body = value
}

func userAgentHandler(req *Request, res *Response) {
	if ua, found := req.Headers.Get("User-Agent"); found {
		res.Headers.Set("Content-Type", "text/plain")
		res.Headers.Set("Content-Length", strconv.Itoa(utf8.RuneCountInString(ua)))
		res.Body = ua
	}
}

func fileReturnHandler(req *Request, res *Response) {
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
		return
	}

	res.Headers.Set("Content-Type", "application/octet-stream")
	res.Headers.Set("Content-Length", strconv.Itoa(utf8.RuneCountInString(string(dat))))
	res.Body = string(dat)
}

func fileCreateHandler(req *Request, res *Response) {
	filePath := strings.TrimPrefix(req.RequestURI, "/files/")
	filePath = filepath.Join(FILE_DIRECTORY, filePath)

	err := os.WriteFile(filePath, []byte(req.Body), 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		res.StatusCode = 500
		res.ReasonPhrase = "Internal Server Error"
		return
	}

	res.StatusCode = 201
	res.ReasonPhrase = "Created"
}

func fileHandler(req *Request, res *Response) {
	if req.Method == "POST" {
		fileCreateHandler(req, res)
		return
	}
	fileReturnHandler(req, res)
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

	err = router.start(l)
	if err != nil {
		os.Exit(1)
	}
}
