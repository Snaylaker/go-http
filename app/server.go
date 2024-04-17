package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type HttpResponse struct {
	httpVersion   string
	httpStatus    int
	contentType   string
	contentLength int
	body          []byte
}

type HttpRequest struct {
	httpVerb  string
	userAgent string
	path      string
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
	response += string(r.body)
	return []byte(response)
}
func toHttpRequest(request string) HttpRequest {
	lines := strings.Split(request, "\r\n")
	firstLine := strings.Split(lines[0], " ")
	var userAgent string
	for _, line := range lines {
		if strings.HasPrefix(line, "User-Agent: ") {
			userAgent = strings.TrimPrefix(line, "User-Agent: ")
			break
		}
	}
	return HttpRequest{
		httpVerb:  firstLine[0],
		path:      firstLine[1],
		userAgent: userAgent,
	}
}

func handleEchoRequest(path string) HttpResponse {
	param := strings.TrimPrefix(path, "/echo/")
	length := len(param)
	return HttpResponse{
		httpVersion:   "HTTP/1.1",
		httpStatus:    200,
		contentType:   "text/plain",
		contentLength: length,
		body:          []byte(param),
	}
}

func handleFileRequest(path string) (HttpResponse, error) {
	fileData, err := os.ReadFile(path)
	if err != nil {
		return HttpResponse{}, err
	}
	return HttpResponse{
		httpVersion:   "HTTP/1.1",
		httpStatus:    200,
		contentType:   "application/octet-stream",
		contentLength: len(fileData),
		body:          fileData,
	}, nil
}

func handleConnection(conn net.Conn, path string) {
	defer conn.Close()
	req, err := io.ReadAll(io.Reader(conn))
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}

	httpRequest := toHttpRequest(string(req))

	var response HttpResponse

	switch {
	case httpRequest.httpVerb == "GET" && strings.HasPrefix(httpRequest.path, "/echo/"):
		response = handleEchoRequest(httpRequest.path)
	case httpRequest.httpVerb == "GET" && strings.HasPrefix(httpRequest.path, "/files/"):
		response, err = handleFileRequest(path + httpRequest.path)
		if err != nil {
			response = HttpResponse{
				httpVersion: "HTTP/1.1",
				httpStatus:  404,
				contentType: "text/plain",
				body:        []byte("File not found"),
			}
		}
	case httpRequest.httpVerb == "GET" && httpRequest.path == "/user-agent":
		response = HttpResponse{
			httpVersion:   "HTTP/1.1",
			httpStatus:    200,
			contentType:   "text/plain",
			contentLength: len(httpRequest.userAgent),
			body:          []byte(httpRequest.userAgent),
		}
	default:
		response = HttpResponse{
			httpVersion:   "HTTP/1.1",
			httpStatus:    404,
			contentType:   "text/plain",
			contentLength: 0,
			body:          []byte("Not found"),
		}
	}

	conn.Write(response.byte())
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
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn, *path)
	}
}
