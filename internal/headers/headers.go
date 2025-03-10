package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	// Look for a CRLF, if it doesn't find one, assume you haven't been given enough data yet.
	// Consume no data, return false for done, and nil for err.
	// If you do find a CRLF, but it's at the start of the data, you've found the end of the headers,
	// so return the proper values immediately.
	// Look for \r\n in the input bytes
	idxHeaderBytes := bytes.Index(data, []byte("\r\n"))
	if idxHeaderBytes == -1 {
		return 0, false, nil
	} else if idxHeaderBytes == 0 {
		return 2, true, nil
	}

	// Include the \r\n in our byte count
	bytesConsumed := idxHeaderBytes + 2

	headerString := strings.TrimSpace(string(data[:idxHeaderBytes]))

	colonIndex := strings.Index(headerString, ":")
	if colonIndex == -1 {
		return 0, false, errors.New("invalid format, no colon")
	}
	headerKey := headerString[:colonIndex]
	headerValue := strings.TrimSpace(headerString[colonIndex+1:])

	if len(headerKey) != len(strings.TrimRight(headerKey, " ")) {
		return 0, false, errors.New("space between colon and field name")
	}

	if !validateHeaderKey(headerKey) {
		return 0, false, errors.New("field name contains illegal char")
	}

	headerKeyLower := strings.ToLower(headerKey)

	// fmt.Printf("Testprint: headerKey: %v\nheaderValue: %v\n", headerKey, headerValue)

	if _, exists := h[headerKeyLower]; exists {
		h[headerKeyLower] += fmt.Sprintf(", %s", headerValue)
	} else {
		h[headerKeyLower] = headerValue
	}

	return bytesConsumed, false, nil
}

func validateHeaderKey(key string) bool {
	for _, ch := range key {
		if !isValidHeaderChar(ch) {
			return false
		}
	}
	return true
}

func isValidHeaderChar(r rune) bool {
	// fmt.Printf("curchar as rune: %v and as str: %s\n", r, string(r))
	return unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("!#$%&'*+-.^ _`|~", r) && r < 128
}
