package request

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	r, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	r, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Invalid number of parts in request line
	_, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Lowercase in method name
	_, err = RequestFromReader(strings.NewReader("gET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Wrong HTTP version
	_, err = RequestFromReader(strings.NewReader("GET / HTTP/1.2\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	if n > cr.numBytesPerRead {
		n = cr.numBytesPerRead
		cr.pos -= n - cr.numBytesPerRead
	}
	return n, nil
}

func TestRequestLineChunks(t *testing.T) {
	tests := []struct {
		name            string
		request         string
		numBytesPerRead int
		wantMethod      string
		wantTarget      string
		wantVersion     string
		wantErr         bool
	}{
		{
			name:            "Good GET Request",
			request:         "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n",
			numBytesPerRead: 100,
			wantMethod:      "GET",
			wantTarget:      "/",
			wantVersion:     "1.1",
			wantErr:         false,
		},
		{
			name:            "Good GET req one byte read",
			request:         "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n",
			numBytesPerRead: 1,
			wantMethod:      "GET",
			wantTarget:      "/",
			wantVersion:     "1.1",
			wantErr:         false,
		},
		{
			name:            "Chunked Request",
			request:         "GE\r\nHost: localhost\r\n\r\n",
			numBytesPerRead: 2,
			wantMethod:      "",
			wantTarget:      "",
			wantVersion:     "",
			wantErr:         true,
		},
		{
			name:            "Very slow chunked valid request",
			request:         "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n",
			numBytesPerRead: 2, // Only read 2 bytes at a time
			wantMethod:      "GET",
			wantTarget:      "/",
			wantVersion:     "1.1",
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &chunkReader{data: tt.request, numBytesPerRead: tt.numBytesPerRead}
			r, err := RequestFromReader(reader)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, r)
				assert.Equal(t, tt.wantMethod, r.RequestLine.Method)
				assert.Equal(t, tt.wantTarget, r.RequestLine.RequestTarget)
				assert.Equal(t, tt.wantVersion, r.RequestLine.HttpVersion)
			}

		})

	}
}

func TestRequestFromReader(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])

	// Test: Nonsense not valid
	reader = &chunkReader{
		data:            "aølskjq48nyqcnhfdhnaewjkhd::MNijKJ",
		numBytesPerRead: 30,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

}

// Additional test cases

func TestEmptyHeaders(t *testing.T) {
	// Test: Headers with empty values
	reader := strings.NewReader("GET / HTTP/1.1\r\nHost:\r\nUser-Agent:\r\n\r\n")
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", r.Headers["host"])
	assert.Equal(t, "", r.Headers["user-agent"])
}

func TestDuplicateHeaders(t *testing.T) {
	// Test: Duplicate Headers
	reader := strings.NewReader("GET / HTTP/1.1\r\nHost: localhost\r\nHost: example.com\r\n\r\n")
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	// Assuming the parser keeps the last value for duplicate headers
	assert.Equal(t, "localhost, example.com", r.Headers["host"])
}

func TestCaseInsensitiveHeaders(t *testing.T) {
	// Test: Headers with mixed cases
	reader := strings.NewReader("GET / HTTP/1.1\r\nhOsT: localhost\r\nUser-Agent: curl\r\nACCEPT: */*\r\n\r\n")
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost", r.Headers["host"])
	assert.Equal(t, "curl", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])
}

func TestMissingEndOfHeaders(t *testing.T) {
	// Test: Missing end of headers (\r\n\r\n)
	reader := strings.NewReader("GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: curl\r\n")
	// The reader does not have the final \r\n to indicate end of headers
	_, err := RequestFromReader(reader)
	require.Error(t, err)
}

func TestMalformedHeader(t *testing.T) {
	// Test: Missing end of headers (\r\n\r\n)
	reader := strings.NewReader("GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n")
	// The reader does not have the final \r\n to indicate end of headers
	_, err := RequestFromReader(reader)
	require.Error(t, err)
}

func TestBodyParser(t *testing.T) {
	// Test: Standard Body
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))

	// Test: Body shorter than reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	// "Empty Body, 0 reported content length" (valid)
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n" +
			"",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body))

	// "Empty Body, no reported content length" (valid)
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n" +
			"",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body))

	// "No Content-Length but Body Exists" (shouldn't error, we're assuming Content-Length will be present if a body exists)
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n" +
			"Hello testing\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body))

}
