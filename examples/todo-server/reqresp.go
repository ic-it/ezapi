package main

import (
	"log"

	"github.com/google/uuid"
	"github.com/ic-it/ezapi"
	"github.com/ic-it/ezapi/examples/todo"
)

// Create
type CreateTodoReq struct {
	JSONBody *todo.BaseTodo `ezapi:"jsonBody"`
}

func (req CreateTodoReq) Validate(ctx ezapi.BaseContext) ezapi.RespError {
	log.Println("validating: create todo")
	if req.JSONBody.Title == "" {
		return TodoTitleEmptyError{}
	}
	return nil
}

// Get
type GetTodoReq struct {
	PathParams struct {
		ID uuid.UUID `ezapi:"id"`
	} `ezapi:"path"`
}

// Get all
type GetAllTodosReq struct {
	QueryParams struct {
		Title       string `ezapi:"title,optional"`       // search by title
		Description string `ezapi:"description,optional"` // search by description
	} `ezapi:"query"`
}

type GetAllTodosRep struct {
	Todos []todo.Todo `json:"todos"`
}

// Update
type UpdateTodoReq struct {
	PathParams struct {
		ID uuid.UUID `ezapi:"id"`
	} `ezapi:"path"`

	JSONBody struct {
		NewTitle       string `json:"newTitle,omitempty"`
		NewDescription string `json:"newDescription,omitempty"`
	} `ezapi:"jsonBody"`
}

func (req UpdateTodoReq) Validate(ctx ezapi.BaseContext) ezapi.RespError {
	log.Println("validating: update todo")
	if req.JSONBody.NewTitle == "" && req.JSONBody.NewDescription == "" {
		return TodoTitleOrDescriptionEmptyError{}
	}
	return nil
}

// Delete
type DeleteTodoReq struct {
	PathParams struct {
		ID uuid.UUID `ezapi:"id"`
	} `ezapi:"path"`
}
