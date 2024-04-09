package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func handleConnection(con net.Conn, path string) {
	defer con.Close()

	req := make([]byte, 1024)
	n, err := con.Read(req)
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}
	parsedResponse := string(req[:n])
	var response string
	if strings.HasPrefix(parsedResponse, "GET /echo/") {
		param := strings.Split(parsedResponse, " ")
		url := strings.TrimPrefix(param[1], "/echo/")
		response = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + strconv.Itoa(len(url)) + "\r\n\r\n" + url
		con.Write([]byte(response))
	} else if strings.HasPrefix(parsedResponse, "GET /user-agent") {
		param := strings.Split(parsedResponse, "\r\n")
		var url string
		for i, v := range param {
			if strings.HasPrefix(v, "User-Agent: ") {
				url = strings.TrimPrefix(param[i], "User-Agent: ")
				break
			}
		}
		response = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + strconv.Itoa(len(url)) + "\r\n\r\n" + url
		con.Write([]byte(response))
	} else if strings.Contains(parsedResponse, "GET / ") {
		response = "HTTP/1.1 200 OK\r\n\r\n"
		con.Write([]byte(response))
	} else if strings.HasPrefix(parsedResponse, "GET /files/") {
		param := strings.Split(parsedResponse, " ")
		url := strings.TrimPrefix(param[1], "/files/")
		filePath := path + `/` + url
		fi, err := os.ReadFile(filePath)
		if err != nil {
			response = "HTTP/1.1 404 Not Found\r\n\r\n"
			con.Write([]byte(response))
		} else {
			response = "HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: " + strconv.Itoa(len(fi)) + "\r\n\r\n"
			con.Write(append([]byte(response), fi...))
		}
	} else if strings.HasPrefix(parsedResponse, "POST /files/") {
		param := strings.Split(parsedResponse, " ")
		fmt.Println("testint", param)
		url := strings.TrimPrefix(param[1], "/files/")
		filePath := path + `/` + url
		fi, err := os.ReadFile(filePath)
		if err != nil {
			response = "HTTP/1.1 404 Not Found\r\n\r\n"
			con.Write([]byte(response))
		} else {
			response = "HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: " + strconv.Itoa(len(fi)) + "\r\n\r\n"
			con.Write(append([]byte(response), fi...))
		}
	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
		con.Write([]byte(response))
	}
}

func main() {
	path := flag.String("directory", "tfk", "path to file")
	flag.Parse()
	fmt.Println(" path :", *path)

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221:", err)
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Server started, listening on port 4221")

	for {
		con, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(con, *path)
	}
}
