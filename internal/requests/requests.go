package requests

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/leikisdev/GoHTTPServ/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       parserState
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

var ErrorRequestState = fmt.Errorf("error request state")
var ErrorInvalidReqLine = fmt.Errorf("invalid request line")
var ErrorInvalidHttpVersion = fmt.Errorf("unsupported HTTP version")
var Separator = []byte("\r\n")

type parserState int

const (
	StateInit          parserState = 1
	StateDone          parserState = 2
	StateError         parserState = 0
	StateParsingHeader parserState = 3
)

func (r *Request) parse(data []byte) (int, error) {
	totalParsed := 0
	for !r.done() {
		n, err := r.parseSingle(data[totalParsed:])
		if err != nil {
			r.state = StateError
			return 0, err
		} else if n == 0 {
			break
		}

		totalParsed += n
	}
	return totalParsed, nil
}

func (r *Request) parseSingle(data []byte) (numConsumed int, err error) {
	switch r.state {
	case StateError:
		return 0, ErrorRequestState

	case StateInit:
		reqLine, numConsumed, err := parseRequestLine(data)
		if err != nil || numConsumed == 0 {
			return 0, err
		}

		r.RequestLine = *reqLine
		r.state = StateParsingHeader
		return numConsumed, nil

	case StateParsingHeader:
		numConsumed, doneHeaders, err := r.Headers.Parse(data)
		if err != nil || numConsumed == 0 {
			return 0, err
		} else if doneHeaders {
			r.state = StateDone
		}

		return numConsumed, nil

	default:
		return 0, ErrorRequestState
	}
}

func (r *Request) done() bool {
	return r.state == StateDone
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{state: StateInit, Headers: headers.NewHeaders()}
	buffer := make([]byte, 8)
	nRead := 0
	for !req.done() {
		n, err := reader.Read(buffer[nRead:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				req.state = StateDone
				break
			}
			return nil, err
		}
		nRead += n

		nParsed, err := req.parse(buffer[:nRead])
		if err != nil {
			return nil, err
		} else if n == 0 {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}
		copy(buffer, buffer[nParsed:nRead])
		nRead -= nParsed
	}

	return req, nil
}

func parseRequestLine(reqLine []byte) (*RequestLine, int, error) {
	idxNewLine := bytes.Index(reqLine, Separator)
	if idxNewLine == -1 {
		return nil, 0, nil
	}

	segments := strings.Split(string(reqLine[:idxNewLine]), " ")
	if len(segments) != 3 {
		return nil, 0, ErrorInvalidReqLine
	}

	method := segments[0]
	for _, l := range method {
		if l < 'A' || l > 'Z' {
			return nil, 0, ErrorInvalidReqLine
		}
	}

	version, hasPrefix := strings.CutPrefix(segments[2], "HTTP/")
	if !hasPrefix || version != "1.1" {
		return nil, 0, ErrorInvalidHttpVersion
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: segments[1],
		HttpVersion:   version,
	}, idxNewLine + len(Separator), nil
}

type ChunkReader struct {
	Data            string
	NumBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *ChunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.Data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.NumBytesPerRead
	if endIndex > len(cr.Data) {
		endIndex = len(cr.Data)
	}
	n = copy(p, cr.Data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}
