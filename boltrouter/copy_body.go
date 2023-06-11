package boltrouter

import (
	"bytes"
	"io"
	"net/http"
)

// copyReq copies the request body into a destination request.
// this allows reading a request body multiple times without "closing" it
func CopyReqBody(src *http.Request, dest *http.Request) {
	var b bytes.Buffer
	b.ReadFrom(src.Body)
	src.Body = io.NopCloser(&b)
	dest.Body = io.NopCloser(bytes.NewReader(b.Bytes()))
}

// copyResp copies the response body and returns a new response with the copied body.
// this allows reading a response body multiple times without "closing" it.
func CopyRespBody(resp *http.Response) io.ReadCloser {
	var b bytes.Buffer
	b.ReadFrom(resp.Body)
	resp.Body = io.NopCloser(&b)
	return io.NopCloser(bytes.NewReader(b.Bytes()))
}
