package aptos

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// HttpErrSummaryLength is the maximum length of the body to include in the error message
const HttpErrSummaryLength = 1000

// HttpError is an error type that represents an error from a http request
type HttpError struct {
	Status     string      // HTTP status e.g. "200 OK"
	StatusCode int         // HTTP status code e.g. 200
	Header     http.Header // HTTP headers
	Method     string      // HTTP method e.g. "GET"
	RequestUrl url.URL     // URL of the request
	Body       []byte      // Body of the response
}

// NewHttpError creates a new HttpError from a http.Response
func NewHttpError(response *http.Response) *HttpError {
	body, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	return &HttpError{
		Status:     response.Status,
		StatusCode: response.StatusCode,
		Header:     response.Header,
		Body:       body,
		Method:     response.Request.Method,
		RequestUrl: *response.Request.URL,
	}
}

// Error returns a string representation of the HttpError
//
// Implements:
//   - [Error]
func (he *HttpError) Error() string {
	if len(he.Body) < HttpErrSummaryLength {
		return fmt.Sprintf("HttpError %s %#v -> %#v %#v",
			he.Method, he.RequestUrl.String(), he.Status,
			string(he.Body),
		)
	}

	// Trim if the error is too long
	return fmt.Sprintf("HttpError %s %#v -> %#v %s %#v...[+%d]",
		he.Method, he.RequestUrl.String(), he.Status,
		he.Header.Get("Content-Type"),
		string(he.Body)[:HttpErrSummaryLength-10], len(he.Body)-(HttpErrSummaryLength-10),
	)
}
