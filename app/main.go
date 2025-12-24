// RFC 9112 - HTTP/1.1 - https://datatracker.ietf.org/doc/html/rfc9112

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

// Headers is a case-insensitive map for HTTP headers.
// Keys are normalized to lowercase internally.
type Headers map[string]string

// Get retrieves a header value by key (case-insensitive).
func (h Headers) Get(key string) (string, bool) {
	val, ok := h[strings.ToLower(key)]
	return val, ok
}

// Set stores a header value with a key (case-insensitive).
func (h Headers) Set(key, value string) {
	h[strings.ToLower(key)] = value
}

// NewHeaders creates a new Headers map.
func NewHeaders() Headers {
	return make(Headers)
}

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
	Headers Headers
	Body    string
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
		Headers: NewHeaders(),
	}

	// Parse headers
	for {
		line, err := reader.ReadString('\n')
		// CRLF that marks the end of the headers
		if err != nil || line == "\r\n" {
			break
		}
		// Split into at most n substrings
		headerParts := strings.SplitN(strings.TrimSpace(line), ": ", 2)
		if len(headerParts) == 2 {
			req.Headers.Set(headerParts[0], headerParts[1])
		}
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
	Headers Headers
	Body    string
}

func (r Response) HeaderToString() string {
	if len(r.Headers) == 0 {
		return ""
	}

	var sb strings.Builder

	for k, v := range r.Headers {
		sb.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	return sb.String()
}

func (r Response) String() string {
	return fmt.Sprintf("%s %d %s\r\n%s\r\n%s", r.HTTPVersion, r.StatusCode, r.ReasonPhrase, r.HeaderToString(), r.Body)
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

	r := Response{
		StatusLine: StatusLine{
			HTTPVersion:  "HTTP/1.1",
			StatusCode:   200,
			ReasonPhrase: "OK",
		},
		Headers: NewHeaders(),
	}

	after, found := strings.CutPrefix(req.RequestURI, "/echo/")
	if found {
		r.Body = after
		r.Headers.Set("Content-Type", "text/plain")
		r.Headers.Set("Content-Length", strconv.Itoa(utf8.RuneCountInString(after)))
	} else {
		if req.RequestURI == "/user-agent" {
			userAgentHeaderValue, found := req.Headers.Get("User-Agent")
			if found {
				r.Headers.Set("Content-Type", "text/plain")
				r.Headers.Set("Content-Length", strconv.Itoa(utf8.RuneCountInString(userAgentHeaderValue)))
				r.Body = userAgentHeaderValue
			}
		} else if req.RequestURI != "/" {
			r.StatusCode = 404
			r.ReasonPhrase = "Not Found"
		}
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
