package main

import (
	"context"
	"log"
	"net/http"
	"strings"

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
				req := ctx.GetReq()
				log.Println("create-todo", req.JSONBody)
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
				req := ctx.GetReq()
				log.Println("get-todo", req.PathParams)
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
				req := ctx.GetReq()
				log.Println("get-all-todos", req.QueryParams)
				var filteredTodos []todo.Todo
				for _, todo := range todos {
					if req.QueryParams.Title != "" && !strings.Contains(todo.Title, req.QueryParams.Title) {
						continue
					}
					if req.QueryParams.Description != "" && !strings.Contains(todo.Description, req.QueryParams.Description) {
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
				req := ctx.GetReq()
				log.Println("update-todo", req.PathParams, req.JSONBody)
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
				req := ctx.GetReq()
				log.Println("delete-todo", req.PathParams)
				dTodo, ok := todos[req.PathParams.ID]
				if !ok {
					return nil, TodoNotFoundError{ID: req.PathParams.ID}
				}
				delete(todos, dTodo.ID)
				log.Println("dTodo", dTodo)
				return &todo.TodoIDOnly{ID: dTodo.ID}, nil
			},
		))

	// Echo Hello with middleware
	mux.Handle("/hello/{name}", Middleware(
		ezapi.H(
			func(ctx ezapi.Context[*HelloReq]) (HelloRep, ezapi.RespError) {
				names := []string{ctx.GetReq().PathParams.Name}
				names = append(names, ctx.GetReq().QueryParams.Names...)
				names = append(names, ctx.GetReq().ContextParams.Names...)
				message := "Hello, " + strings.Join(names, ", ") + "!"
				log.Println("hello", message)
				return HelloRep{Message: message}, nil
			},
		)))

	log.Println("Listening on :8080")
	http.ListenAndServe(":8080", mux)
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("middleware, add name to the context")
		r = r.WithContext(context.WithValue(r.Context(), "names", []string{"Alice", "Bob"}))
		next.ServeHTTP(w, r)
	})
}
