package main

import (
	"fmt"
	"net"
	"os"
	s "strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	con, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	req := make([]byte, 1024)
	con.Read(req)

	parsedResponse := string(req)
	fmt.Printf(parsedResponse)

	if s.Contains(parsedResponse, "GET /echo/") {
		param := s.Split(parsedResponse, " ")
		url := s.Split(param[1], "/")
		resp := []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n" + url[len(url)-1])
		con.Write(resp)
		os.Exit(1)
	}

	if s.Contains(parsedResponse, "/ ") {
		resp := []byte("HTTP/1.1 200 OK\r\n\r\n")
		con.Write(resp)
		os.Exit(1)
	}
	resp := []byte("HTTP/1.1 404 Not Found\r\n\r\n")
	con.Write(resp)
	os.Exit(1)
}
