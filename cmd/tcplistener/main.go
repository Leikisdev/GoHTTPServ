package main

import (
	"fmt"
	"log"
	"net"

	"github.com/leikisdev/GoHTTPServ/internal/requests"
)

const port = ":42069"

func main() {
	tcpListener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Unable to set up TCP listener")
	}
	defer tcpListener.Close()

	for {
		fmt.Println("awaiting connection")
		conn, err := tcpListener.Accept()
		if err != nil {
			log.Fatalf("failed to establish connection")
		}

		fmt.Println("Accepted connection from", conn.RemoteAddr())

		request, err := requests.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("failed to parse request with error: %v", err)
		}

		fmt.Println("Request line:")
		fmt.Println("- Method:", request.RequestLine.Method)
		fmt.Println("- Version:", request.RequestLine.HttpVersion)
		fmt.Println("- Path:", request.RequestLine.RequestTarget)
		fmt.Println("Headers:")
		for k, v := range request.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}

		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}

	// reader := &requests.ChunkReader{
	// 	Data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
	// 	NumBytesPerRead: 3,
	// }
	// r, err := requests.RequestFromReader(reader)
	// if err != nil {
	// 	fmt.Printf("err %s", err.Error())
	// 	return
	// }
	// fmt.Println("req", r)

}
