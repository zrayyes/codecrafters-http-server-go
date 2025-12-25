// RFC 9112 - HTTP/1.1 - https://datatracker.ietf.org/doc/html/rfc9112

package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
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

	if n, found := req.Headers.Get("Content-Length"); found && n != "0" {
		num, err := strconv.Atoi(n)
		if err != nil {
			return nil, err
		}
		buf := make([]byte, num)

		_, err = io.ReadFull(reader, buf)
		if err != nil {
			return nil, err
		}

		req.Body = string(buf)
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

// NewResponse creates a new Response with sensible defaults (HTTP/1.1 200 OK).
func NewResponse() *Response {
	return &Response{
		StatusLine: StatusLine{
			HTTPVersion:  "HTTP/1.1",
			StatusCode:   200,
			ReasonPhrase: "OK",
		},
		Headers: NewHeaders(),
	}
}

type HandlerFunc func(req *Request) *Response

type Route struct {
	Pattern  string
	IsPrefix bool
	Handler  HandlerFunc
}

type Router struct {
	routes []Route
}

func NewRouter() Router {
	return Router{}
}

func (r *Router) HandleExact(path string, handler HandlerFunc) {
	r.routes = append(r.routes, Route{path, false, handler})
}

func (r *Router) HandlePrefix(prefix string, handler HandlerFunc) {
	r.routes = append(r.routes, Route{prefix, true, handler})
}

func (r *Router) Route(req *Request) *Response {
	for _, route := range r.routes {
		if route.IsPrefix {
			if strings.HasPrefix(req.RequestURI, route.Pattern) {
				return route.Handler(req)
			}
		} else {
			if req.RequestURI == route.Pattern {
				return route.Handler(req)
			}
		}
	}
	res := NewResponse()
	res.StatusCode = 404
	res.ReasonPhrase = "Not Found"
	return res
}
