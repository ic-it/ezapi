# GoreAPI

> [!WARNING]
> THIS IS JUST A POC. DO NOT USE IT IN PRODUCTION.

GoreAPI is a framework for building `http.HandlerFunc`.  

GoreAPI allows devtime to run the server, but since all the magic only works 
because of reflect, the final build requires building the final version of each handler. 

## Usage

**main.go**
```go
package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/create-user/{id}", CreateUserHandler)
	// /create-user/1
	// {"name":"John","age":30,"email":"test123@mail.com","password":"password"}
	mux.HandleFunc("/create-user-2", CreateUserHandler2)
	// /create-user-2?id=1&user={"name":"John","age":30,"email":"test123@mail.com","password":"password"}
	mux.HandleFunc("/get-user/{id}", GetUserHandler)
	// /get-user/1
	mux.HandleFunc("/get-user-2", GetUserHandler2)
	// /get-user-2?id=1
	log.Println("Listening on :8080")

	http.ListenAndServe(":8080", mux)
}
```

**handler.go**
```go
//go:build gore
// +build gore

package main

import (
	"fmt"

	"github.com/ic-it/goreapi"
)

type User struct {
	Id       int    `json:"id,omitempty"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var (
	CreateUserHandler = goreapi.H(
		goreapi.Param("id", &goreapi.Int{}),
		goreapi.Body("user", &User{}),
		goreapi.Response(&goreapi.Int{}),
		func(id int, user *User) int {
			fmt.Printf("Creating user: %s\n", user.Name)
			return id
		},
	)

	CreateUserHandler2 = goreapi.H(
		goreapi.Query("id", &goreapi.Int{}),
		goreapi.Query("user", &User{}),
		goreapi.Response(&goreapi.Int{}),
		func(id int, user *User) int {
			fmt.Printf("Creating user: %s\n", user.Name)
			return id
		},
	)

	GetUserHandler = goreapi.H(
		goreapi.Param("id", &goreapi.Int{}),
		goreapi.Response(&User{}),
		func(id int) *User {
			return &User{
				Id:       id,
				Name:     "John",
				Age:      30,
				Email:    "test@mail.com",
				Password: "password",
			}
		},
	)

	GetUserHandler2 = goreapi.H(
		goreapi.Query("id", &goreapi.Int{}),
		goreapi.Response(&User{}),
		func(id int) *User {
			return &User{
				Id:       id,
				Name:     "John",
				Age:      30,
				Email:    "asdasd@asdasd.com",
				Password: "password",
			}
		},
	)
)
```

## Run code Generation

```bash
go run ./cmd/gen/...
```

**Result of the code generation Example**

[Here](./cmd/web/handlers-generated.go)