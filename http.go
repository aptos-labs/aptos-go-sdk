package aptos

import (
	"fmt"
	"io"
	"net/http"
)

type HttpError struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Header     http.Header
	Body       []byte
}

func NewHttpError(response *http.Response) *HttpError {
	body, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	return &HttpError{
		Status:     response.Status,
		StatusCode: response.StatusCode,
		Header:     response.Header,
		Body:       body,
	}
}

// implement error interface
func (he *HttpError) Error() string {
	return fmt.Sprintf("HttpError %#v (%d bytes %s)", he.Status, len(he.Body), he.Header.Get("Content-Type"))
}
