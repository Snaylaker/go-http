package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type HttpResponse struct {
	httpVersion   string
	httpStatus    int
	contentType   string
	contentLength *int
	body          *string
}

func (r HttpResponse) byte() []byte {
	var statusText string
	if r.httpStatus >= 400 && r.httpStatus < 600 {
		statusText = "Error"
	} else {
		statusText = "OK"
	}
	response := fmt.Sprintf("%s %d %s\r\n", r.httpVersion, r.httpStatus, statusText)
	response += fmt.Sprintf("Content-Type: %s\r\n", r.contentType)
	response += fmt.Sprintf("Content-Length: %d\r\n\r\n", r.contentLength)
	response += *r.body
	return []byte(response)
}

type HttpRequest struct {
	httpVerb string
	path     string
}

func toHttpRequest(request string) HttpRequest {
	param := strings.Split(request, " ")
	return HttpRequest{httpVerb: param[0], path: param[1]}
}

func handleConnection(con net.Conn, path string) {
	defer con.Close()

	req := make([]byte, 1024)
	n, err := con.Read(req)
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}
	parsedReq := string(req[:n])

	httpRequest := toHttpRequest(parsedReq)
	if httpRequest.httpVerb == "GET" && strings.HasPrefix(httpRequest.path, "/echo/") {
		param := strings.TrimPrefix(httpRequest.path, "/echo/")
		length := len(param)
		response := HttpResponse{
			httpVersion:   "HTTP/1.1",
			httpStatus:    200,
			contentType:   "text/plain",
			contentLength: &length,
			body:          &param,
		}
		con.Write(response.byte())
	} else if strings.HasPrefix(parsedReq, "GET /user-agent") {
		param := strings.Split(parsedReq, "\r\n")
		var url string
		for i, v := range param {
			if strings.HasPrefix(v, "User-Agent: ") {
				url = strings.TrimPrefix(param[i], "User-Agent: ")
				break
			}
		}
		length := len(url)
		response := HttpResponse{
			httpVersion:   "HTTP/1.1",
			httpStatus:    200,
			contentType:   "text/plain",
			contentLength: &length,
			body:          &url,
		}
		con.Write(response.byte())
	} else if strings.Contains(parsedReq, "GET / ") {
		response := HttpResponse{
			httpVersion: "HTTP/1.1",
			httpStatus:  200,
			contentType: "text/plain",
		}
		con.Write(response.byte())
	} else if strings.HasPrefix(parsedReq, "GET /files/") {
		param := strings.Split(parsedReq, " ")
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
	} else if strings.HasPrefix(parsedReq, "POST /files/") {
		param := strings.Split(parsedReq, "\r\n")
		tmp := strings.Split(param[0], " ")
		url := strings.TrimPrefix(tmp[1], "/files/")
		filePath := path + `/` + url
		response = "HTTP/1.1 201 OK\r\n\r\n"
		d1 := []byte(param[6])
		err := os.WriteFile(filePath, d1, 0644)
		if err != nil {
			response = "HTTP/1.1 404 Not Found\r\n\r\n"
		}
		con.Write([]byte(response))
	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
		con.Write([]byte(response))
	}
}

func main() {
	path := flag.String("directory", "tfk", "path to file")
	flag.Parse()

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
