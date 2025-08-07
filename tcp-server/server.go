package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	fmt.Println("Lite HTTP server on :8080")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)

	// 1) Request line: METHOD SP PATH SP VERSION CRLF
	reqLine, err := readLine(r)
	if err != nil {
		return
	}

	parts := strings.Split(reqLine, " ")
	if len(parts) != 3 {
		writeBadRequest(conn, "malformed request line")
		return
	}

	method, path, version := parts[0], parts[1], parts[2]
	if version != "HTTP/1.1" && version != "HTTP/1.0" {
		writeBadRequest(conn, "unsupported HTTP version")
		return
	}

	// 2) Headers: key: value CRLF ... then a blank line CRLF
	headers := make(map[string]string)
	for {
		line, err := readLine(r)
		if err != nil {
			return
		}
		if line == "" {
			break
		}

		colon := strings.Index(line, ":")
		if colon < 0 {
			writeBadRequest(conn, "malformed header")
			return
		}

		key := strings.TrimSpace(line[:colon])
		val := strings.TrimSpace(line[colon+1:])
		headers[strings.ToLower(key)] = val
	}

	// 3) Optional body if Content-Length present
	var body []byte
	if cl, ok := headers["content-length"]; ok {
		n, err := strconv.Atoi(cl)
		if err != nil || n < 0 {
			writeBadRequest(conn, "bad Content-Length")
			return
		}

		body = make([]byte, n)
		if _, err := io.ReadFull(r, body); err != nil {
			return
		}
	}

	// 4) Route
	switch {
	case method == "GET" && path == "/":
		writeResponse(conn, 200, "OK", "text/plain", []byte("Hello from a tiny HTTP server!\n"))
	case method == "POST" && path == "/echo":
		writeResponse(conn, 200, "OK", "text/plain", body)
	default:
		writeResponse(conn, 404, "Not Found", "text/plain", []byte("route not found\n"))
	}
}

func readLine(r *bufio.Reader) (string, error) {
	// Reads up to CRLF, return line without CRLF
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	// HTTP lines end with \r\n; trim both
	return strings.TrimRight(line, "\r\n"), nil
}

func writeBadRequest(w io.Writer, msg string) {
	writeResponse(w, 400, "Bad Request", "text/plain", []byte(msg+"\n"))
}

func writeResponse(w io.Writer, code int, status, ctype string, body []byte) {
	fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", code, status)
	fmt.Fprintf(w, "Content-Type: %s\r\n", ctype)
	fmt.Fprintf(w, "Content-Length: %d\r\n", len(body))
	// We’ll explicitly close after this response
	fmt.Fprintf(w, "Connection: close\r\n")
	fmt.Fprintf(w, "\r\n")
	w.Write(body)
}
