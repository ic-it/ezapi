package ezapi

import "net/http"

type BaseContext interface {
	// Get the http.Request object
	GetR() *http.Request
	// Get the http.ResponseWriter object
	GetW() http.ResponseWriter
}

type Context[T any] interface {
	BaseContext

	// Get the request object
	GetReq() T
}

type ezapiContext[T any] struct {
	r *http.Request
	w http.ResponseWriter

	req T
}

func (c ezapiContext[T]) GetR() *http.Request {
	return c.r
}

func (c ezapiContext[T]) GetW() http.ResponseWriter {
	return c.w
}

func (c ezapiContext[T]) GetReq() T {
	return c.req
}
