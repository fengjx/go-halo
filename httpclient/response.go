// Copyright (c) 2015-2023 Jeevanandam M (jeeva@myjeeva.com), All rights reserved.
// resty source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package httpclient

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fengjx/go-halo/json"
)

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Response struct and methods
//_______________________________________________________________________

// Response struct holds response values of executed request.
type Response struct {
	RawResponse *http.Response

	body       []byte
	size       int64
	sendAt     time.Time
	receivedAt time.Time
}

// Body method returns HTTP response as []byte array for the executed request.
//
// Note: `Response.Body` might be nil, if `Request.SetOutput` is used.
func (r *Response) Body() []byte {
	if r.RawResponse == nil {
		return []byte{}
	}
	return r.body
}

// Status method returns the HTTP status string for the executed request.
//
//	Example: 200 OK
func (r *Response) Status() string {
	if r.RawResponse == nil {
		return ""
	}
	return r.RawResponse.Status
}

// StatusCode method returns the HTTP status code for the executed request.
//
//	Example: 200
func (r *Response) StatusCode() int {
	if r.RawResponse == nil {
		return 0
	}
	return r.RawResponse.StatusCode
}

// Proto method returns the HTTP response protocol used for the request.
func (r *Response) Proto() string {
	if r.RawResponse == nil {
		return ""
	}
	return r.RawResponse.Proto
}

// Header method returns the response headers
func (r *Response) Header() http.Header {
	if r.RawResponse == nil {
		return http.Header{}
	}
	return r.RawResponse.Header
}

// Cookies method to access all the response cookies
func (r *Response) Cookies() []*http.Cookie {
	if r.RawResponse == nil {
		return make([]*http.Cookie, 0)
	}
	return r.RawResponse.Cookies()
}

// String method returns the body of the server response as String.
func (r *Response) String() string {
	if r.body == nil {
		return ""
	}
	return strings.TrimSpace(string(r.body))
}

// Time method returns the time of HTTP response time that from request we sent and received a request.
//
// See `Response.ReceivedAt` to know when client received response and see `Response.Request.Time` to know
// when client sent a request.
func (r *Response) Time() time.Duration {
	return r.receivedAt.Sub(r.sendAt)
}

// ReceivedAt method returns when response got received from server for the request.
func (r *Response) ReceivedAt() time.Time {
	return r.receivedAt
}

// Size method returns the HTTP response size in bytes. Ya, you can relay on HTTP `Content-Length` header,
// however it won't be good for chucked transfer/compressed response. Since Resty calculates response size
// at the client end. You will get actual size of the http response.
func (r *Response) Size() int64 {
	return r.size
}

// RawBody method exposes the HTTP raw response body. Use this method in-conjunction with `SetDoNotParseResponse`
// option otherwise you get an error as `read err: http: read on closed response body`.
//
// Do not forget to close the body, otherwise you might get into connection leaks, no connection reuse.
// Basically you have taken over the control of response parsing from `Resty`.
func (r *Response) RawBody() io.ReadCloser {
	if r.RawResponse == nil {
		return nil
	}
	return r.RawResponse.Body
}

// IsSuccess method returns true if HTTP status `code >= 200 and <= 299` otherwise false.
func (r *Response) IsSuccess() bool {
	return r.StatusCode() > 199 && r.StatusCode() < 300
}

// IsError method returns true if HTTP status `code >= 400` otherwise false.
func (r *Response) IsError() bool {
	return r.StatusCode() > 399
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Response Unexported methods
//_______________________________________________________________________

func (r *Response) setReceivedAt() {
	r.receivedAt = time.Now()
}

func (r *Response) fmtBody(model interface{}) error {
	return json.FromBytes(r.Body(), model)
}
