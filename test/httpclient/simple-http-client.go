package httpclient

import (
	"io"
	"net/http"
)

type simpleHTTPClient struct{}

func NewSimpleHTTPClient() simpleHTTPClient {
	return simpleHTTPClient{}
}

func (c simpleHTTPClient) Get(url string) (status int, body string, e error) {
	var (
		bodyBytes []byte
		response  *http.Response
	)

	response, e = http.Get(url)
	if e != nil {
		return
	}

	status = response.StatusCode

	bodyBytes, e = io.ReadAll(response.Body)

	response.Body.Close()

	body = string(bodyBytes)

	return
}
