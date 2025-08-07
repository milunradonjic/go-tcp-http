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
	// 1) Connect via TCP
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	rb := "Hello Go!"
	// 2) Send a GET request
	fmt.Fprint(conn, "POST /echo HTTP/1.1\r\n")
	fmt.Fprint(conn, "Host: localhost\r\n")
	fmt.Fprint(conn, "Connection: close\r\n")
	fmt.Fprint(conn, "Content-Type: application/x-www-form-urlencoded\r\n")
	fmt.Fprintf(conn, "Content-Length: %d\r\n", len(rb))
	fmt.Fprint(conn, "\r\n")
	fmt.Fprint(conn, rb)

	// 3) Read status line
	r := bufio.NewReader(conn)
	statusLine, err := readLine(r)
	if err != nil {
		panic(err)
	}
	fmt.Println("Status:", statusLine)

	// 4) Read headers
	headers := make(map[string]string)
	for {
		line, err := readLine(r)
		if err != nil {
			panic(err)
		}
		if line == "" {
			break
		}

		colon := strings.Index(line, ":")
		if colon > 0 {
			k := strings.ToLower(strings.TrimSpace(line[:colon]))
			v := strings.TrimSpace(line[colon+1:])
			headers[k] = v
		}
	}

	// 5) Read body using Content-Length
	var body []byte
	if cl, ok := headers["content-length"]; ok {
		n, _ := strconv.Atoi(cl)
		body = make([]byte, n)
		if _, err := io.ReadFull(r, body); err != nil {
			panic(err)
		}
	}
	fmt.Printf("Body:\n%s", string(body))
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
