package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/ic-it/ezapi"
	"github.com/ic-it/ezapi/examples/todo"
)

// Simple "TODO app" example

var todos map[uuid.UUID]todo.Todo = make(map[uuid.UUID]todo.Todo)

func main() {
	mux := http.NewServeMux()

	// Create
	mux.HandleFunc(
		"/todo",
		ezapi.H(
			func(ctx ezapi.Context[CreateTodoReq]) (todo.TodoIDOnly, ezapi.RespError) {
				log.Println("create-todo", ctx.GetReq().JSONBody)
				req := ctx.GetReq()
				newTodo := todo.Todo{
					TodoIDOnly: todo.TodoIDOnly{ID: uuid.New()},
					BaseTodo:   *req.JSONBody,
				}
				todos[newTodo.ID] = newTodo
				log.Println("newTodo", newTodo)
				return todo.TodoIDOnly{ID: newTodo.ID}, nil
			},
		))

	// Get
	mux.HandleFunc(
		"/todo/{id}/get",
		ezapi.H(
			func(ctx ezapi.Context[GetTodoReq]) (*todo.Todo, ezapi.RespError) {
				log.Println("get-todo", ctx.GetReq().PathParams)
				req := ctx.GetReq()
				todo, ok := todos[req.PathParams.ID]
				if !ok {
					return nil, TodoNotFoundError{ID: req.PathParams.ID}
				}
				log.Println("todo", todo)
				return &todo, nil
			},
		))

	// Get all
	mux.HandleFunc(
		"/todos",
		ezapi.H(
			func(ctx ezapi.Context[GetAllTodosReq]) (GetAllTodosRep, ezapi.RespError) {
				log.Println("get-all-todos", ctx.GetReq().QueryParams)
				req := ctx.GetReq()
				var filteredTodos []todo.Todo
				for _, todo := range todos {
					if req.QueryParams.Title != "" && todo.Title != req.QueryParams.Title {
						continue
					}
					if req.QueryParams.Description != "" && todo.Description != req.QueryParams.Description {
						continue
					}
					filteredTodos = append(filteredTodos, todo)
				}
				log.Println("filteredTodos", filteredTodos)
				return GetAllTodosRep{Todos: filteredTodos}, nil
			},
		))

	// Update
	mux.HandleFunc(
		"/todo/{id}/update",
		ezapi.H(
			func(ctx ezapi.Context[UpdateTodoReq]) (*todo.TodoIDOnly, ezapi.RespError) {
				log.Println("update-todo", ctx.GetReq().PathParams, ctx.GetReq().JSONBody)
				req := ctx.GetReq()
				updTodo, ok := todos[req.PathParams.ID]
				if !ok {
					return nil, TodoNotFoundError{ID: req.PathParams.ID}
				}
				if req.JSONBody.NewTitle != "" {
					updTodo.Title = req.JSONBody.NewTitle
				}
				if req.JSONBody.NewDescription != "" {
					updTodo.Description = req.JSONBody.NewDescription
				}
				todos[updTodo.ID] = updTodo
				log.Println("updTodo", updTodo)
				return &todo.TodoIDOnly{ID: updTodo.ID}, nil
			},
		))

	// Delete
	mux.HandleFunc(
		"/todo/{id}/delete",
		ezapi.H(
			func(ctx ezapi.Context[DeleteTodoReq]) (*todo.TodoIDOnly, ezapi.RespError) {
				log.Println("delete-todo", ctx.GetReq().PathParams)
				req := ctx.GetReq()
				dTodo, ok := todos[req.PathParams.ID]
				if !ok {
					return nil, TodoNotFoundError{ID: req.PathParams.ID}
				}
				delete(todos, dTodo.ID)
				log.Println("dTodo", dTodo)
				return &todo.TodoIDOnly{ID: dTodo.ID}, nil
			},
		))

	log.Println("Listening on :8080")
	http.ListenAndServe(":8080", mux)
}
