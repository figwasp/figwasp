package images

import (
	"net/http"
	"testing"
)

func TestHTTPStatusCodeServerImage(t *testing.T) {
	const (
		buildContextPath = "http-status-code-server"
	)

	var (
		image *httpStatusCodeServerImage

		e error
	)

	image, e = NewHTTPStatusCodeServerImage(http.StatusTeapot, buildContextPath)
	if e != nil {
		t.Error(e)
	}

	e = image.Build()
	if e != nil {
		t.Error(e)
	}

	defer image.Remove()
}
