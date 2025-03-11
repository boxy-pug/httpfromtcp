package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHeaders(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Req line and then invalid host missing colon
	headers = NewHeaders()
	data = []byte("GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("    User-Agent: curl/7.81.0     \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "curl/7.81.0", headers["user-agent"])
	assert.Equal(t, 34, n)
	assert.False(t, done)

	// Test: Valid done
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 2, n) // Only the CRLF consumed
	assert.True(t, done)

	// Test: Invalid spacing header (no spaces around colon)
	headers = NewHeaders()
	data = []byte("Invalid:HeaderFormat\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "HeaderFormat", headers["invalid"])
	assert.Equal(t, 22, n) // 22 is the length of the string "Invalid:HeaderFormat" + "\r\n"
	assert.False(t, done)

	// Test: Invalid char in field name
	headers = NewHeaders()
	data = []byte("H©ŸÆst: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid multiple values for
	headers = Headers{
		"user-agent": "testing123",
	}
	data = []byte("User-Agent: curl/7.81.0\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "testing123, curl/7.81.0", headers["user-agent"])
	assert.False(t, done)

	// Test: Valid 2 headers with existing headers
	headers = map[string]string{"host": "localhost:42069"}
	data = []byte("User-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "curl/7.81.0", headers["user-agent"])
	assert.Equal(t, 25, n)
	assert.False(t, done)

}

func TestValidateHeaderKey(t *testing.T) {
	str := "sdf09&243*&"
	res := validateHeaderKey(str)
	assert.True(t, res)

	str = "H©ŸÆst: localhost:42069"
	res = validateHeaderKey(str)
	assert.False(t, res)

	str = "!#$%&'*+-.^_`|~"
	res = validateHeaderKey(str)
	assert.True(t, res)
}
