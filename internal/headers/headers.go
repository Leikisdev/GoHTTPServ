package headers

import (
	"bytes"
	"fmt"
	"strings"
)

var LineSeparator = []byte("\r\n")
var HeaderSeparator = []byte(":")
var ValidSpecialChars = "!#$%&'*+-.^_`|~"

var ErrorHeaderMalformed = fmt.Errorf("malformed header")

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, LineSeparator)
	switch idx {
	case -1:
		return
	case 0:
		done = true
		n = len(LineSeparator)
		return
	}

	fieldLine := bytes.Trim(data[:idx], " ")
	key, val, found := bytes.Cut(fieldLine, HeaderSeparator)
	keyString := strings.ToLower(string(key))
	if !found || keyString[len(keyString)-1] == ' ' || !validateHeaderName(keyString) {
		return 0, false, ErrorHeaderMalformed
	}

	headerVal := strings.Trim(string(val), " ")

	if _, found := h[keyString]; found {
		headerVal = h[keyString] + "," + headerVal
	}

	h[keyString] = headerVal
	n = idx + len(LineSeparator)
	return
}

func validateHeaderName(s string) bool {
	for _, char := range s {
		if char >= 'a' && char <= 'z' {
			continue
		} else if char >= '0' && char <= '9' {
			continue
		} else if strings.ContainsRune(ValidSpecialChars, char) {
			continue
		}
		return false
	}

	return true
}
