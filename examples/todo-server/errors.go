package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/ic-it/ezapi"
)

// Error: todo not found
type TodoNotFoundError struct {
	ID uuid.UUID
}

func (e TodoNotFoundError) Error() string {
	return "todo not found"
}

func (e TodoNotFoundError) Render(ctx ezapi.BaseContext) error {
	ctx.GetW().WriteHeader(http.StatusNotFound)
	ctx.GetW().Write([]byte(e.Error()))
	return nil
}

// Validation error: title cannot be empty
type TodoTitleEmptyError struct{}

func (e TodoTitleEmptyError) Error() string {
	return "title cannot be empty"
}

func (e TodoTitleEmptyError) Render(ctx ezapi.BaseContext) error {
	log.Println("rendering: todo title empty")
	ctx.GetW().WriteHeader(http.StatusBadRequest)
	ctx.GetW().Write([]byte(e.Error()))
	return nil
}

// Validation error: description or title should be provided
type TodoTitleOrDescriptionEmptyError struct{}

func (e TodoTitleOrDescriptionEmptyError) Error() string {
	return "title or description should be provided"
}

func (e TodoTitleOrDescriptionEmptyError) Render(ctx ezapi.BaseContext) error {
	log.Println("rendering: todo title or description empty")
	ctx.GetW().WriteHeader(http.StatusBadRequest)
	ctx.GetW().Write([]byte(e.Error()))
	return nil
}
