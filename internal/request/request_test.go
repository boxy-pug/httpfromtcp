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
