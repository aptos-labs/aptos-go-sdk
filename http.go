package aptos

import (
	"fmt"
	"net/http"

	"github.com/valyala/fasthttp"
)

// HttpErrSummaryLength is the maximum length of the body to include in the error message
const HttpErrSummaryLength = 1000

// HttpError is an error type that represents an error from a http request
type HttpError struct {
	Status     string      // HTTP status e.g. "200 OK"
	StatusCode int         // HTTP status code e.g. 200
	Header     http.Header // HTTP headers
	Method     string      // HTTP method e.g. "GET"
	RequestUrl string      // URL of the request
	Body       []byte      // Body of the response
}

// NewHttpError creates a new HttpError from a http.Response
func NewHttpError(response *fasthttp.Response, request *fasthttp.Request) *HttpError {
	body := response.Body()
	var headers http.Header = make(map[string][]string)
	response.Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = []string{string(value)}
	})

	return &HttpError{
		Status:     string(response.Header.StatusMessage()),
		StatusCode: response.StatusCode(),
		Header:     headers,
		Body:       body,
		Method:     string(request.Header.Method()),
		RequestUrl: string(request.URI().FullURI()),
	}
}

// Error returns a string representation of the HttpError
//
// Implements:
//   - [Error]
func (he *HttpError) Error() string {
	if len(he.Body) < HttpErrSummaryLength {
		return fmt.Sprintf("HttpError %s %#v -> %#v %#v",
			he.Method, he.RequestUrl, he.Status,
			string(he.Body),
		)
	} else {
		// Trim if the error is too long
		return fmt.Sprintf("HttpError %s %#v -> %#v %s %#v...[+%d]",
			he.Method, he.RequestUrl, he.Status,
			he.Header.Get("Content-Type"),
			string(he.Body)[:HttpErrSummaryLength-10], len(he.Body)-(HttpErrSummaryLength-10),
		)
	}
}
