package todo

import "github.com/google/uuid"

type BaseTodo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type TodoIDOnly struct {
	ID uuid.UUID `json:"id"`
}

type Todo struct {
	TodoIDOnly
	BaseTodo
}
