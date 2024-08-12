package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"io"

	"github.com/ic-it/ezapi/examples/todo"
)

// Simple "TODO app client" example

func main() {
	host := "http://localhost:8080"

	for {
		fmt.Println("--------------------")
		fmt.Println("Enter command:")
		fmt.Println("1. Create")
		fmt.Println("2. Get")
		fmt.Println("3. Get all")
		fmt.Println("4. Update")
		fmt.Println("5. Delete")
		fmt.Println("6. Exit")

		var command int
		_, err := fmt.Scanf("%d", &command)
		if err != nil {
			log.Println("Invalid command")
			continue
		}

		switch command {
		case 1:
			create(host)
		case 2:
			getOne(host)
		case 3:
			getAll(host)
		case 4:
			updateOne(host)
		case 5:
			deleteOne(host)
		case 6:
			return
		default:
			log.Println("Invalid command")
		}
	}
}

func create(host string) {
	var title, description string
	fmt.Println("Enter title:")
	fmt.Scanf("%s", &title)
	fmt.Println("Enter description (optional):")
	fmt.Scanf("%s", &description)

	body := todo.BaseTodo{
		Title:       title,
		Description: description,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		log.Println("Failed to marshal body")
		return
	}

	resp, err := post(host+"/todo", bodyBytes)
	if err != nil {
		log.Println("Failed to create todo")
		return
	}

	var todoID todo.TodoIDOnly
	err = json.Unmarshal(resp, &todoID)
	if err != nil {
		log.Println("Failed to unmarshal response:", string(resp))
		return
	}

	fmt.Printf("Created todo with ID: %s\n", todoID.ID)
}

func getOne(host string) {
	var id string
	fmt.Println("Enter ID:")
	fmt.Scanf("%s", &id)

	resp, err := get(host + "/todo/" + id + "/get")
	if err != nil {
		log.Println("Failed to get todo")
		return
	}

	var todoObj todo.Todo
	err = json.Unmarshal(resp, &todoObj)
	if err != nil {
		log.Println("Failed to unmarshal response:", string(resp))
		return
	}

	fmt.Printf("Got todo: %+v\n", todoObj)
}

func getAll(host string) {
	var title, description string
	fmt.Println("Enter title (optional):")
	fmt.Scanf("%s", &title)
	fmt.Println("Enter description (optional):")
	fmt.Scanf("%s", &description)

	resp, err := get(host + "/todos?title=" + title + "&description=" + description)
	if err != nil {
		log.Println("Failed to get todos")
		return
	}

	type todosResp struct {
		Todos []todo.Todo `json:"todos"`
	}
	var todos todosResp
	err = json.Unmarshal(resp, &todos)
	if err != nil {
		log.Println("Failed to unmarshal response:", string(resp))
		return
	}

	fmt.Printf("Got todos: %+v\n", todos)
}

func updateOne(host string) {
	var id, newTitle, newDescription string
	fmt.Println("Enter ID:")
	fmt.Scanf("%s", &id)
	fmt.Println("Enter new title (optional):")
	fmt.Scanf("%s", &newTitle)
	fmt.Println("Enter new description (optional):")
	fmt.Scanf("%s", &newDescription)

	type updateBody struct {
		NewTitle       string `json:"newTitle,omitempty"`
		NewDescription string `json:"newDescription,omitempty"`
	}

	body := updateBody{
		NewTitle:       newTitle,
		NewDescription: newDescription,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		log.Println("Failed to marshal body")
		return
	}

	resp, err := put(host+"/todo/"+id+"/update", bodyBytes)
	if err != nil {
		log.Println("Failed to update todo")
		return
	}

	var todoID todo.TodoIDOnly
	err = json.Unmarshal(resp, &todoID)
	if err != nil {
		log.Println("Failed to unmarshal response:", string(resp))
		return
	}

	fmt.Printf("Updated todo with ID: %s\n", todoID.ID)
}

func deleteOne(host string) {
	var id string
	fmt.Println("* Enter ID:")
	fmt.Scanf("%s", &id)

	resp, err := delete(host + "/todo/" + id + "/delete")
	if err != nil {
		log.Println("Failed to delete todo")
		return
	}

	var todoID todo.TodoIDOnly
	err = json.Unmarshal(resp, &todoID)
	if err != nil {
		log.Println("Failed to unmarshal response:", string(resp))
		return
	}

	fmt.Printf("Deleted todo with ID: %s\n", todoID.ID)
}

// Helper functions
var client = &http.Client{}

func post(url string, body []byte) ([]byte, error) {
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(io.Reader(resp.Body))
}

func get(url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(io.Reader(resp.Body))
}

func put(url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(io.Reader(resp.Body))
}

func delete(url string) ([]byte, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(io.Reader(resp.Body))
}
