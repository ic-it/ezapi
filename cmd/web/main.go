package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/ic-it/goreapi"
)

type CreateUserReqBody struct {
	Name     string `json:"name"`
	Age      int    `json:"age,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserReq struct {
	JSONBody   *CreateUserReqBody `gore:"jsonBody"`
	PathParams struct {
		ID uuid.UUID `gore:"id"`
	} `gore:"path"`
	QueryParams struct {
		ID     int    `gore:"id"`
		Status string `gore:"status,optional"`
	} `gore:"query"`
}

func (r CreateUserReq) Validate(ctx goreapi.BaseContext) goreapi.GoreError {
	log.Println("Validating", r)
	return nil
}

func (r CreateUserReq) OnUnmarshalError(ctx goreapi.BaseContext, err error) goreapi.GoreError {
	log.Println("Handling unmarshal error", err)

	ctx.GetW().WriteHeader(http.StatusBadRequest)
	ctx.GetW().Write([]byte(err.Error()))
	return nil
}

type TestRep struct {
	Test1 string `json:"test1"`
	Test2 struct {
		Test3 string `json:"test3"`
	} `json:"test2"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"/create-user/{id}",
		goreapi.H(
			func(ctx goreapi.GoreContext[CreateUserReq]) (TestRep, goreapi.GoreError) {
				req := ctx.GetReq()
				log.Println(req.PathParams.ID, "PathParams")
				log.Println(req.QueryParams.ID, "QueryParams")
				ctx.GetW().WriteHeader(http.StatusOK)
				return TestRep{
					Test1: "test1",
					Test2: struct {
						Test3 string `json:"test3"`
					}{
						Test3: "test3",
					},
				}, nil
			},
		))
	log.Println("Listening on :8080")
	http.ListenAndServe(":8080", mux)
}
