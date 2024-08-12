package goreapi

import "net/http"

type BaseContext interface {
	// Get the http.Request object
	GetR() *http.Request
	// Get the http.ResponseWriter object
	GetW() http.ResponseWriter
}

type GoreContext[T any] interface {
	BaseContext

	// Get the GoreRequest object
	GetReq() T
}

type goreContext[T any] struct {
	r *http.Request
	w http.ResponseWriter

	req T
}

func (c goreContext[T]) GetR() *http.Request {
	return c.r
}

func (c goreContext[T]) GetW() http.ResponseWriter {
	return c.w
}

func (c goreContext[T]) GetReq() T {
	return c.req
}
