package aptos

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpError_Error_ShortBody(t *testing.T) {
	t.Parallel()
	he := &HttpError{
		Status:     "400 Bad Request",
		StatusCode: 400,
		Method:     "GET",
		Body:       []byte("short error body"),
	}
	errStr := he.Error()
	assert.Contains(t, errStr, "HttpError")
	assert.Contains(t, errStr, "GET")
	assert.Contains(t, errStr, "400 Bad Request")
	assert.Contains(t, errStr, "short error body")
}

func TestHttpError_Error_LongBody(t *testing.T) {
	t.Parallel()
	longBody := strings.Repeat("x", HttpErrSummaryLength+500)
	he := &HttpError{
		Status:     "500 Internal Server Error",
		StatusCode: 500,
		Method:     "POST",
		Body:       []byte(longBody),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
	errStr := he.Error()
	assert.Contains(t, errStr, "HttpError")
	assert.Contains(t, errStr, "POST")
	assert.Contains(t, errStr, "500 Internal Server Error")
	assert.Contains(t, errStr, "...[+")
}

func TestNewHttpError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	resp, err := http.Get(server.URL + "/test") //nolint:gosec
	require.NoError(t, err)

	httpErr := NewHttpError(resp)
	assert.Equal(t, 404, httpErr.StatusCode)
	assert.Equal(t, "404 Not Found", httpErr.Status)
	assert.Equal(t, "GET", httpErr.Method)
	assert.Contains(t, string(httpErr.Body), "not found")
}
