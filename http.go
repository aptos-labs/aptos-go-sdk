package aptos

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const HTTP_ERR_SUMMARY_LEN = 100

type HttpError struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Header     http.Header
	Method     string
	RequestUrl url.URL
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
		Method:     response.Request.Method,
		RequestUrl: *response.Request.URL,
	}
}

// implement error interface
func (he *HttpError) Error() string {
	if len(he.Body) < HTTP_ERR_SUMMARY_LEN {
		return fmt.Sprintf("HttpError %s %#v -> %#v %#v",
			he.Method, he.RequestUrl.String(), he.Status,
			string(he.Body),
		)

	} else {
		return fmt.Sprintf("HttpError %s %#v -> %#v %s %#v...[+%d]",
			he.Method, he.RequestUrl.String(), he.Status,
			he.Header.Get("Content-Type"),
			string(he.Body)[:HTTP_ERR_SUMMARY_LEN-10], len(he.Body)-(HTTP_ERR_SUMMARY_LEN-10),
		)
	}
}
