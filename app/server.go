package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
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

	if s.HasPrefix(parsedResponse, "GET /echo/") {
		param := s.Split(parsedResponse, " ")
		url := s.TrimPrefix(param[1], "/echo/")
		resp := []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-length: " + strconv.Itoa(len(url)) + "\r\n\r\n" + url)
		con.Write(resp)
	} else if s.Contains(parsedResponse, "GET / ") {
		resp := []byte("HTTP/1.1 200 OK\r\n\r\n")
		con.Write(resp)
	} else {
		resp := []byte("HTTP/1.1 404 Not Found\r\n\r\n")
		con.Write(resp)
	}
	con.Close()
}
