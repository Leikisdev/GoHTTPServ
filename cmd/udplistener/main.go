package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

type HTTPServer struct {
	port   int
	domain string
}

type Request struct {
	method  string
	path    string
	version string
	headers []Header
	body    File
}

func (req Request) parse(payload []byte) Request {
	return Request{}
}

type Response struct {
	version    string
	statusCode int
	headers    []Header
	body       File
}

type Header struct {
}

type File struct {
}

const port = ":42069"

func getLinesChannel(f io.ReadCloser) <-chan string {
	channel := make(chan string)

	go func() {
		defer f.Close()
		defer close(channel)

		buffer := make([]byte, 8)
		currLine := ""
		for {
			n, err := f.Read(buffer)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s", err.Error())
				break
			}
			currString := string(buffer[:n])
			parts := strings.Split(currString, "\n")
			for i := 0; i < len(parts)-1; i++ {
				channel <- currLine + parts[i]
				currLine = ""
			}
			currLine += parts[len(parts)-1]
		}

		if currLine != "" {
			channel <- currLine
		}
	}()

	return channel
}

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("Unable to resolve UDP address")
	}

	udpSender, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal("Unable to set up UDP sender")
	}
	defer udpSender.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Unable to read message")
			continue
		}
		fmt.Println("sending:", input)
		if _, err := udpSender.Write([]byte(input)); err != nil {
			log.Printf("error: failed to write to UDP with %s", err.Error())
		}
	}
}
